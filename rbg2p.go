package rbg2p

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Trans is a container for phonemes in a transcriptions
type Trans struct {
	Phonemes []string
}

// Context in which the rule applies (left hand/right hand context specified by a regular expression)
type Context struct {
	input  string
	regexp *regexp.Regexp
}

// Matches checks if the input string matches the context rule
func (c Context) Matches(s string) bool {
	if c.IsDefined() {
		return c.regexp.MatchString(s)
	}
	return true
}

// IsDefined return true if the contained regexp is defined
func (c Context) IsDefined() bool {
	return (nil != c.regexp)
}

// String returns a string representation of the Context
func (c Context) String() string {
	if c.IsDefined() {
		return c.input
	}
	return ""
}

// Equals checks for equality (with correct result for underlying regexps)
func (c Context) Equals(c2 Context) bool {
	if c.IsDefined() && !c2.IsDefined() {
		return false
	} else if c2.IsDefined() && !c.IsDefined() {
		return false
	} else if !c2.IsDefined() && !c.IsDefined() {
		return true
	}
	return c.input == c2.input
}

// Rule is a g2p rule representation
type Rule struct {
	Input        string
	Output       []string
	LeftContext  Context
	RightContext Context
}

// String returns a string representation of the Context
func (r Rule) String() string {
	return fmt.Sprintf("%s -> %s / %s _ %s", r.Input, r.Output, r.LeftContext, r.RightContext)
}

// Equals checks for equality (with correct result for underlying slices and regexps)
func (r Rule) Equals(r2 Rule) bool {
	return r.Input == r2.Input &&
		reflect.DeepEqual(r.Output, r2.Output) &&
		r.LeftContext.Equals(r2.LeftContext) &&
		r.RightContext.Equals(r2.RightContext)
}

// Test defines a rule test (input -> output)
type Test struct {
	Input  string
	Output []string
}

// Equals checks for equality (with correct result for underlying slices)
func (t1 Test) Equals(t2 Test) bool {
	return t1.Input == t2.Input && reflect.DeepEqual(t1.Output, t2.Output)
}

// RuleSet is a set of g2p rules, with variables and built-in tests
type RuleSet struct {
	Vars  map[string]string
	Rules []Rule
	Tests []Test
}

// Test runs the built-in tests. Returns an array of errors, if any.
func (rs RuleSet) Test() []error {
	var errs []error
	for _, test := range rs.Tests {
		input := test.Input
		expect := test.Output
		result0, err := rs.Apply(input)
		result := []string{}
		for _, trans := range result0 {
			result = append(result, strings.Join(trans.Phonemes, " "))
		}
		if err != nil {
			errs = append(errs, errors.New(fmt.Sprintf("%v", err)))
		}
		if !reflect.DeepEqual(expect, result) {
			errs = append(errs, errors.New(fmt.Sprintf("for '%s', expected %v, got %v", input, expect, result)))
		}
	}
	return errs
}

func expand(transes [][]string) []Trans {
	n := 1
	for _, arr := range transes {
		n = n * len(arr)
	}

	res := make([][]string, n, 2*n)

	k := 0
	for _, arr := range transes {
		for j := 0; j < n; j++ {
			k++
			if k == len(arr) {
				k = 0
			}

			res[j] = append(res[j], arr[k])
		}

	}
	var expanded = []Trans{}
	for _, phns := range res {
		expanded = append(expanded, Trans{Phonemes: phns})
	}
	return expanded
}

// Apply applies the rules to an input string, returns a slice of transcriptions
func (rs RuleSet) Apply(s0 string) ([]Trans, error) {
	var i = 0
	res := [][]string{}
	var couldntMap = []string{}
	for i < len(s0) {
		s := s0[i:len(s0)]
		left := s0[0:i]
		var matchFound = false
		for _, rule := range rs.Rules {
			if strings.HasPrefix(s, rule.Input) &&
				rule.LeftContext.Matches(left) {
				right := s0[i+len(rule.Input) : len(s0)]
				if rule.RightContext.Matches(right) {
					i = i + len(rule.Input)
					res = append(res, rule.Output)
					matchFound = true
					break
				}
			}
		}
		if !matchFound {
			res = append(res, []string{"_"})
			i = i + 1
			couldntMap = append(couldntMap, s[0:1])
		}
	}
	//fmt.Printf("%v\n", res)
	if len(couldntMap) > 0 {
		return expand(res), fmt.Errorf("Found unmappable symbol(s) in input string: %v in %s", couldntMap, s0)
	} else {
		return expand(res), nil
	}
}
