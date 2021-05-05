package rbg2p

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
)

var Debug = false

// Context in which the rule applies (left hand/right hand context specified by a regular expression)
type Context struct {
	// Input is the regexp as written in the input string
	Input string

	// Regexp is the input string converted to a regular expression for internal use (with variables expanded, and adapted anchoring)
	Regexp *regexp2.Regexp
}

// Matches checks if the input string matches the context rule
func (c Context) Matches(s string) (bool, error) {
	if c.IsDefined() {
		res, err := c.Regexp.MatchString(s)
		if err != nil {
			return false, err
		}
		return res, nil
	}
	return true, nil
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

// Filter is a regexp filter for rules that cannot be expressed using the standard rule systme
type Filter struct {
	Regexp *regexp2.Regexp
	Output string
}

// Apply is used to apply the filter to an input string
func (f Filter) Apply(s string) (string, error) {
	//return f.Regexp.ReplaceAllString(s, f.Output)
	return f.Regexp.Replace(s, f.Output, -1, -1)
}

// Prefilter is a regexp filter
type Prefilter struct {
	Regexp *regexp2.Regexp
	Output string
}

// Apply is used to apply the prefilter to an input string
func (pf Prefilter) Apply(s string) (string, error) {
	//return f.Regexp.ReplaceAllString(s, f.Output)
	return pf.Regexp.Replace(s, pf.Output, -1, -1)
}

// Rule is a g2p rule representation
type Rule struct {
	Input        string
	Output       []string
	LeftContext  Context
	RightContext Context
	LineNumber   int // for debugging
}

// String returns a string representation of the Rule
func (r Rule) String() string {
	var output string
	if len(r.Output) == 1 {
		output = r.Output[0]
	} else {
		output = fmt.Sprintf("(%s)", strings.Join(r.Output, ", "))
	}
	return fmt.Sprintf("%s -> %s / %s _ %s", r.Input, output, r.LeftContext, r.RightContext)
}

// equals: checks for equality (including underlying slices and regexps); used for unit tests
func (r Rule) equals(r2 Rule) bool {
	return r.Input == r2.Input &&
		reflect.DeepEqual(r.Output, r2.Output) &&
		r.LeftContext.equals(r2.LeftContext) &&
		r.RightContext.equals(r2.RightContext)
}

// equalsExceptOutput: checks for equality except for output (including underlying slices and regexps); used for unit tests
func (r Rule) equalsExceptOutput(r2 Rule) bool {
	return r.Input == r2.Input &&
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
	CharacterSet      []string
	PhonemeSet        PhonemeSet
	PhonemeDelimiter  string
	SyllableDelimiter string
	DefaultPhoneme    string
	DowncaseInput     bool
	Vars              map[string]string
	Rules             []Rule
	RulesAppliedMutex *sync.RWMutex
	RulesApplied      map[string]int // for coverage checks
	Tests             []Test
	Filters           []Filter
	Prefilters        []Prefilter
	Syllabifier       Syllabifier
	Content           string
	Debug             bool
}

func (rs RuleSet) isInitialized() bool {
	return len(rs.Rules) > 0
}

func (rs RuleSet) checkForUnusedChars(coveredChars map[string]bool, individualChars map[string]bool, validation *TestResult) {
	var errors = []string{}
	for _, char := range rs.CharacterSet {
		if _, ok := individualChars[char]; !ok {
			errors = append(errors, char)
		}
	}
	sort.Strings(errors)
	if len(errors) > 0 {
		validation.Errors = append(validation.Errors, fmt.Sprintf("no default rule for character(s): %s", strings.Join(errors, ",")))
	}
}

func (rs RuleSet) checkForUndefinedChars(coveredChars map[string]bool, individualChars map[string]bool, validation *TestResult) {
	var definedChars = make(map[string]bool)
	var errors = []string{}
	for _, char := range rs.CharacterSet {
		definedChars[char] = true
	}
	for char := range individualChars {
		for _, ch := range strings.Split(char, "") {
			if _, ok := definedChars[ch]; !ok {
				errors = append(errors, ch)
			}
		}
	}
	sort.Strings(errors)
	if len(errors) > 0 {
		validation.Errors = append(validation.Errors, fmt.Sprintf("undefined character(s) used in rule set: %s", strings.Join(errors, ",")))
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
	rs.checkForUndefinedChars(coveredChars, individualChars, &result)

	if rs.hasPhonemeSet() {
		validation, err := compareToPhonemeSet(rs)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%v", err))
		}
		result.Warnings = append(result.Warnings, validation.Warnings...)
		result.Errors = append(result.Errors, validation.Errors...)
	}

	for _, test := range rs.Tests {
		input := test.Input
		expect := test.Output
		if rs.DowncaseInput {
			input = strings.ToLower(input)
		}
		res, err := rs.Apply(input)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%v", err))
		}
		//delim := rs.PhonemeDelimiter
		if !reflect.DeepEqual(expect, res) {
			result.FailedTests = append(result.FailedTests, fmt.Sprintf("for '%s', expected /%s/, got /%s/", input, strings.Join(expect, "/ + /"), strings.Join(res, "/ + /")))
		}
	}
	return result
}

func (rs RuleSet) expandLoop(head g2p, tail []g2p, acc []trans) []trans {
	res := []trans{}
	for i := 0; i < len(acc); i++ {
		for _, add := range head.p {
			appendRange := []g2p{}
			// build prefix from previous rounds
			appendRange = append(appendRange, acc[i].phonemes...)
			// append current phonemes
			g2p := g2p{g: head.g, p: strings.Split(add, rs.PhonemeDelimiter)}
			appendRange = append(appendRange, g2p)
			res = append(res, trans{phonemes: appendRange})
		}

	}
	if len(tail) == 0 {
		return res
	}
	return rs.expandLoop(tail[0], tail[1:], res)
}

func (rs RuleSet) expand(phonemes []g2p) []trans {
	if len(phonemes) > 1 {
		return rs.expandLoop(phonemes[0], phonemes[1:], []trans{{}})
	} else if len(phonemes) == 1 {
		return rs.expandLoop(phonemes[0], []g2p{}, []trans{{}})
	}
	// empty trans
	return []trans{}
}

func (rs RuleSet) applyFilters(trans string) (string, error) {
	res := trans
	var err error
	for _, f := range rs.Filters {
		input := res
		res, err = f.Apply(res)
		if err != nil {
			return res, fmt.Errorf("couldn't execute regexp : %v", err)
		}
		if rs.Debug {
			fmt.Fprintf(os.Stderr, "FILTER\t%s\t%s\t%s\n", f, input, res)
		}
	}
	return res, nil
}

func (rs RuleSet) applyPrefilters(trans string) (string, error) {
	res := trans
	var err error
	for _, pf := range rs.Prefilters {
		res, err = pf.Apply(res)
		if err != nil {
			return res, fmt.Errorf("couldn't execute regexp : %v", err)
		}
	}
	return res, nil
}

// Apply applies the rules to an input string, returns a slice of transcriptions. If unknown input characters are found, an error will be created, and an underscore will be appended to the transcription. Even if an error is returned, the loop will continue until the end of the input string.
func (rs RuleSet) Apply(s string) ([]string, error) {
	if !rs.isInitialized() {
		return []string{}, fmt.Errorf("RuleSet is not initialized")
	}

	var i = 0
	if rs.DowncaseInput {
		s = strings.ToLower(s)
	}
	var prefiltered string
	pfted, pferr := rs.applyPrefilters(s)
	if pferr != nil {
		return []string{}, fmt.Errorf("couldn't apply prefilter: %s", s)
	}
	prefiltered = pfted
	var s0 = []rune(prefiltered)
	res := []g2p{}
	var couldntMap = []string{}
	for i < len(s0) {
		ss := string(s0[i:])
		thisChar := string(s0[i : i+1])
		left := string(s0[0:i])
		var matchFound = false
		for _, rule := range rs.Rules {
			leftMatch, err := rule.LeftContext.Matches(left)
			if err != nil {
				return []string{}, fmt.Errorf("couldn't execute regexp /%s/ : %s", rule.LeftContext.Regexp, err)
			}
			if strings.HasPrefix(ss, rule.Input) && leftMatch {
				ruleInputLen := len([]rune(rule.Input))
				right := string(s0[i+ruleInputLen:])
				rightMatch, err := rule.RightContext.Matches(right)
				if err != nil {
					return []string{}, fmt.Errorf("couldn't execute regexp /%s/ : %s", rule.RightContext.Regexp, err)
				}
				if rightMatch {
					i = i + ruleInputLen
					res = append(res, g2p{g: rule.Input, p: rule.Output})
					matchFound = true
					ruleString := rule.String()
					rs.RulesAppliedMutex.Lock()
					rs.RulesApplied[ruleString]++
					rs.RulesAppliedMutex.Unlock()
					if Debug {
						fmt.Fprintf(os.Stderr, "%s\t%v\t%v\t%v\t%v\n", "RULE APPLIED", rule, s, ss, res)
					}
					break
				}
			}
		}
		if !matchFound {
			res = append(res, g2p{g: thisChar, p: []string{rs.DefaultPhoneme}})
			i = i + 1
			couldntMap = append(couldntMap, thisChar)
		}
	}
	expanded := rs.expand(res)

	transes := []string{}
	for _, t := range expanded {
		if rs.Syllabifier.IsDefined() {
			s := rs.Syllabifier.syllabifyToString(t)
			transes = append(transes, s)
		} else {
			transes = append(transes, t.string(rs.PhonemeDelimiter))
		}
	}
	var filtered []string
	for _, t := range transes {
		fted, err := rs.applyFilters(t)
		if err != nil {
			return filtered, err
		}
		filtered = append(filtered, fted)
	}
	if len(couldntMap) > 0 {
		return filtered, fmt.Errorf("found unmappable symbol(s) in input string: %v in %s", couldntMap, s)
	}
	return filtered, nil
}

// compareToPhonemeSet validates the phonemes in the g2p rule set against the specified phonemeset. Returns an array of invalid phonemes, if any; or if errors are found, this is returned instead.
func compareToPhonemeSet(ruleSet RuleSet) (TestResult, error) {
	var validation = TestResult{}
	var usedSymbols = map[string]bool{}
	for _, rule := range ruleSet.Rules {
		for _, output := range rule.Output {
			invalid, err := ruleSet.PhonemeSet.validate(output)
			if err != nil {
				return TestResult{}, fmt.Errorf("found error in rule output /%s/ : %s", output, err)
			}
			splitted, err := ruleSet.PhonemeSet.SplitTranscription(output)
			if err != nil {
				return TestResult{}, err
			}
			for _, symbol := range splitted {
				usedSymbols[symbol] = true
			}

			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in rule output %s: %s", rule, symbol))
			}
		}
	}
	for _, test := range ruleSet.Tests {
		for _, output := range test.Output {
			invalid, err := ruleSet.PhonemeSet.validate(output)
			if err != nil {
				return TestResult{}, fmt.Errorf("found error in test output /%s/ : %s", output, err)
			}
			splitted, err := ruleSet.PhonemeSet.SplitTranscription(output)
			if err != nil {
				return TestResult{}, err
			}
			for _, symbol := range splitted {
				usedSymbols[symbol] = true
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in test output %s: %s", test, symbol))
			}

		}
	}
	validation.Warnings = append(validation.Warnings, checkForUnusedSymbols(usedSymbols, ruleSet.PhonemeSet)...)
	return validation, nil
}
