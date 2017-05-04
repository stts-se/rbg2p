package rbg2p

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Trans is a container for phonemes in a transcription
type Trans struct {
	Phonemes []string
}

func (t Trans) String(phnDelimiter string) string {
	var phns []string
	for _, p := range t.Phonemes {
		if len(p) > 0 {
			phns = append(phns, p)
		}
	}
	return strings.Join(phns, phnDelimiter)
}

// Context in which the rule applies (left hand/right hand context specified by a regular expression)
type Context struct {
	// Input is the regexp as written in the input string
	Input string

	// Regexp is the input string converted to a regular expression for internal use (with variables expanded, and adapted anchoring)
	Regexp *regexp.Regexp
}

// Matches checks if the input string matches the context rule
func (c Context) Matches(s string) bool {
	if c.IsDefined() {
		return c.Regexp.MatchString(s)
	}
	return true
}

// IsDefined returns true if the contained regexp is defined
func (c Context) IsDefined() bool {
	return (nil != c.Regexp)
}

// String returns a string representation of the Context
func (c Context) String() string {
	if c.IsDefined() {
		return c.Input
	}
	return ""
}

// equals checks for equality (including underlying regexps); used for unit tests
func (c Context) equals(c2 Context) bool {
	if c.IsDefined() && !c2.IsDefined() {
		return false
	} else if c2.IsDefined() && !c.IsDefined() {
		return false
	} else if !c2.IsDefined() && !c.IsDefined() {
		return true
	}
	return c.Input == c2.Input
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

// equals checks for equality (including underlying underlying slices and regexps); used for unit tests
func (r Rule) equals(r2 Rule) bool {
	return r.Input == r2.Input &&
		reflect.DeepEqual(r.Output, r2.Output) &&
		r.LeftContext.equals(r2.LeftContext) &&
		r.RightContext.equals(r2.RightContext)
}

// Test defines a rule test (input -> output)
type Test struct {
	Input  string
	Output []string
}

// equals checks for equality (including underlying slices); used for unit tests
func (t1 Test) equals(t2 Test) bool {
	return t1.Input == t2.Input && reflect.DeepEqual(t1.Output, t2.Output)
}

// RuleSet is a set of g2p rules, with variables and built-in tests
type RuleSet struct {
	CharacterSet     []string
	PhonemeSet       PhonemeSet
	PhonemeDelimiter string
	DefaultPhoneme   string
	Vars             map[string]string
	Rules            []Rule
	Tests            []Test
}

// TestResult is a container for test results (errors, warnings, and failed tests from tests speficied in the g2p rule file)
type TestResult struct {
	Errors      []string
	Warnings    []string
	FailedTests []string
}

func (rs RuleSet) checkForUnusedChars(coveredChars map[string]bool, individualChars map[string]bool, validation *TestResult) {
	// var errors = []string{}
	// for _, char := range rs.CharacterSet {
	// 	if _, ok := coveredChars[char]; !ok {
	// 		errors = append(errors, char)
	// 	}
	// }
	// if len(errors) > 0 {
	// 	validation.Warnings = append(validation.Warnings, fmt.Sprintf("no rules exist for character(s): %s", strings.Join(errors, ",")))
	// }

	var errors = []string{}
	for _, char := range rs.CharacterSet {
		if _, ok := individualChars[char]; !ok {
			errors = append(errors, char)
		}
	}
	if len(errors) > 0 {
		validation.Errors = append(validation.Errors, fmt.Sprintf("no default rule for character(s): %s", strings.Join(errors, ",")))
	}
}

func (rs RuleSet) hasPhonemeSet() bool {
	return len(rs.PhonemeSet.Symbols) > 0
}

// Test runs the built-in tests. Returns a test result with errors and warnings, if any.
func (rs RuleSet) Test() TestResult {
	var result = TestResult{}
	var coveredChars = map[string]bool{}
	var individualChars = map[string]bool{}
	for _, rule := range rs.Rules {
		for _, char := range strings.Split(rule.Input, "") {
			coveredChars[char] = true
		}
		coveredChars[rule.Input] = true
		if !rule.LeftContext.IsDefined() && !rule.RightContext.IsDefined() {
			individualChars[rule.Input] = true
		}
	}
	rs.checkForUnusedChars(coveredChars, individualChars, &result)

	if rs.hasPhonemeSet() {
		validation, err := compareToPhonemeSet(rs)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%v", err))
		}
		for _, w := range validation.Warnings {
			result.Warnings = append(result.Warnings, w)
		}
		for _, e := range validation.Errors {
			result.Errors = append(result.Errors, e)
		}
	}

	for _, test := range rs.Tests {
		input := test.Input
		expect := test.Output
		res0, err := rs.Apply(strings.ToLower(input))
		res := []string{}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%v", err))
		}
		for _, trans := range res0 {
			res = append(res, trans.String(rs.PhonemeDelimiter))
		}
		if !reflect.DeepEqual(expect, res) {
			result.FailedTests = append(result.FailedTests, fmt.Sprintf("for '%s', expected %v, got %v", input, expect, res))
		}
	}
	return result
}

func expand(transes [][]string) []Trans {
	n := 1
	for _, arr := range transes {
		n = n * len(arr)
	}

	res := make([]Trans, n, 2*n)

	k := 0
	for _, arr := range transes {
		k = 0
		for j := 0; j < n; j++ {
			if k == len(arr) {
				k = 0
			}
			res[j] = Trans{Phonemes: append(res[j].Phonemes, arr[k])}
			k++

		}
	}
	return res
}

// Apply applies the rules to an input string, returns a slice of transcriptions. If unknown input characters are found, an error will be created, and an underscore will be appended to the transcription. Even if an error is returned, the loop will continue until the end of the input string.
func (rs RuleSet) Apply(s string) ([]Trans, error) {
	var i = 0
	var s0 = []rune(s)
	res := [][]string{}
	var couldntMap = []string{}
	for i < len(s0) {
		ss := string(s0[i:len(s0)])
		thisChar := string(s0[i : i+1])
		left := string(s0[0:i])
		var matchFound = false
		for _, rule := range rs.Rules {
			if strings.HasPrefix(ss, rule.Input) &&
				rule.LeftContext.Matches(left) {
				ruleInputLen := len([]rune(rule.Input))
				right := string(s0[i+ruleInputLen : len(s0)])
				if rule.RightContext.Matches(right) {
					i = i + ruleInputLen
					res = append(res, rule.Output)
					matchFound = true
					break
				}
			}
		}
		if !matchFound {
			res = append(res, []string{rs.DefaultPhoneme})
			i = i + 1
			couldntMap = append(couldntMap, thisChar)
		}
	}
	if len(couldntMap) > 0 {
		return expand(res), fmt.Errorf("Found unmappable symbol(s) in input string: %v in %s", couldntMap, s)
	}
	return expand(res), nil
}
