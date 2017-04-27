package rbg2p

import (
	"fmt"
	"reflect"
	"regexp"
)

// Context in which the rule applies (left hand/right hand context specified by a regular expression)
type Context struct {
	regexp *regexp.Regexp
}

// IsDefined return true if the contained regexp is defined
func (c Context) IsDefined() bool {
	return (nil != c.regexp)
}

// String returns a string representation of the Context
func (c Context) String() string {
	if c.IsDefined() {
		return c.regexp.String()
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
	return c.regexp.String() == c2.regexp.String()
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
