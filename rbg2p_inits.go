package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
)

func isVar(s string) bool {
	return strings.HasPrefix(s, "VAR ")
}

func isTest(s string) bool {
	return strings.HasPrefix(s, "TEST ")
}

func isFilter(s string) bool {
	return strings.HasPrefix(s, "FILTER ")
}

//var g2pLineRe = regexp.MustCompile("^(CHARACTER_SET|TEST|DEFAULT_PHONEME|FILTER|VAR|) .*")
var g2pLineRe = regexp.MustCompile("^(CHARACTER_SET|TEST|DEFAULT_PHONEME|FILTER|VAR|DOWNCASE_INPUT) .*")

func isG2PLine(s string) bool {
	return g2pLineRe.MatchString(s) || ruleRe.MatchString(s)
}

type usedVars map[string]int

// LoadFile loads a g2p rule set from the specified file
func LoadFile(fName string) (RuleSet, error) {
	usedVars := usedVars{}
	ruleSet := RuleSet{Vars: map[string]string{}}
	ruleSet.DefaultPhoneme = "_"
	ruleSet.PhonemeDelimiter = " "
	syllDefLines := []string{}
	fh, err := os.Open(filepath.Clean(fName))
	if err != nil {
		return ruleSet, err
	}
	defer fh.Close()
	n := 0
	s := bufio.NewScanner(fh)
	var inputLines []string
	var ruleLines []string
	var filterLines []string
	var phonemeSetLine string
	for s.Scan() {
		if err := s.Err(); err != nil {
			return ruleSet, err
		}
		n++
		lOrig := strings.TrimSpace(s.Text())
		l := trimComment(lOrig)
		inputLines = append(inputLines, lOrig)
		if isBlankLine(l) || isComment(l) {
		} else if isPhonemeDelimiter(l) {
			delim, err := parsePhonemeDelimiter(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.PhonemeDelimiter = delim
		} else if isPhonemeSet(l) {
			phonemeSetLine = l
		} else if isConst(l) {
			err := parseConst(l, &ruleSet)
			if err != nil {
				return ruleSet, err
			}
		} else if isVar(l) {
			name, value, err := newVar(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.Vars[name] = value
		} else if isSyllDefLine(l) {
			syllDefLines = append(syllDefLines, l)
		} else if isFilter(l) {
			filterLines = append(filterLines, l)
		} else if isTest(l) {
			t, err := newTest(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.Tests = append(ruleSet.Tests, t)
		} else { // is a rule
			ruleLines = append(ruleLines, l)
		}

	}
	for k, v := range ruleSet.Vars {
		v, _, err = expandVarsWithBrackets(v, ruleSet.Vars)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.Vars[k] = v
	}
	if len(phonemeSetLine) > 0 {
		phnSet, err := parsePhonemeSet(phonemeSetLine, ruleSet.PhonemeDelimiter)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.PhonemeSet = phnSet
	}
	if len(syllDefLines) > 0 {
		syllDef, stressPlacement, err := loadSyllDef(syllDefLines, ruleSet.PhonemeDelimiter)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.SyllableDelimiter = syllDef.SyllableDelimiter()
		ruleSet.Syllabifier = Syllabifier{SyllDef: syllDef, StressPlacement: stressPlacement, PhonemeSet: ruleSet.PhonemeSet}
	}

	for _, l := range filterLines {
		t, usedVarsTmp, err := newFilter(l, ruleSet.Vars)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.Filters = append(ruleSet.Filters, t)
		for k, v := range usedVarsTmp {
			usedVars[k] += v
		}
	}
	//ruleSet.Rules = append(ruleSet.Rules, Rule{Input: " ", Output: []string{" "}})
	for _, l := range ruleLines {
		r, usedVarsTmp, err := newRule(l, ruleSet.Vars)
		if err != nil {
			return ruleSet, err
		}
		for _, r0 := range ruleSet.Rules {
			if r0.equalsExceptOutput(r) {
				return ruleSet, fmt.Errorf("duplicate rules for input file %s: %s vs. %s", fName, r0, r)
			}
		}
		for k, v := range usedVarsTmp {
			usedVars[k] += v
		}
		ruleSet.Rules = append(ruleSet.Rules, r)
	}
	if ruleSet.CharacterSet == nil || len(ruleSet.CharacterSet) == 0 {
		return ruleSet, fmt.Errorf("no character set defined for input file %s", fName)
	}
	ruleSet.Content = strings.Join(inputLines, "\n")

	unusedVars := []string{}
	for vName := range ruleSet.Vars {
		if _, ok := usedVars[vName]; !ok {
			unusedVars = append(unusedVars, vName)
		}
	}
	if len(unusedVars) > 0 {
		sort.Strings(unusedVars)
		return ruleSet, fmt.Errorf("Unused variable(s) %s in %s", strings.Join(unusedVars, ", "), fName)
	}

	return ruleSet, nil
}

var constRe = regexp.MustCompile("^(CHARACTER_SET|DEFAULT_PHONEME|DOWNCASE_INPUT) (?:\"(.+)\"|([^\"]+))$")
var isConstRe = regexp.MustCompile("^(CHARACTER_SET|DEFAULT_PHONEME|DOWNCASE_INPUT) .*")
var isTrueRe = regexp.MustCompile("^(true|TRUE|1)$")
var isFalseRe = regexp.MustCompile("^(false|FALSE|0)$")

func isConst(s string) bool {
	return isConstRe.MatchString(s)
}

func parseConst(s string, ruleSet *RuleSet) error {
	matchRes := constRe.FindStringSubmatch(s)
	var downcaseInputIsSet = false
	if matchRes != nil {
		name := matchRes[1]
		value := matchRes[2]
		if value == "" {
			value = matchRes[3]
		}
		if name == "CHARACTER_SET" {
			ruleSet.CharacterSet = strings.Split(value, "")
		} else if name == "DEFAULT_PHONEME" {
			ruleSet.DefaultPhoneme = value
		} else if name == "DOWNCASE_INPUT" {
			if isTrueRe.MatchString(value) {
				ruleSet.DowncaseInput = true
				downcaseInputIsSet = true
			} else if isFalseRe.MatchString(value) {
				ruleSet.DowncaseInput = false
				downcaseInputIsSet = true
			} else {
				return fmt.Errorf("invalid boolean value for " + name + ": " + value)
			}
		} else {
			return fmt.Errorf("invalid const definition: " + s)
		}
	} else {
		return fmt.Errorf("invalid const definition: " + s)
	}
	if !downcaseInputIsSet {
		ruleSet.DowncaseInput = true
	}
	return nil
}

var varRe = regexp.MustCompile("^VAR +([^ \"]+) +([^ \"]+|[^ ].*[^ ])$")
var varReQuoteFix = regexp.MustCompile(`^"(.*)"$`)

func newVar(s string) (string, string, error) {
	// VAR NAME VALUE
	matchRes := varRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", "", fmt.Errorf("invalid VAR definition: " + s)
	}
	name := matchRes[1]
	value := matchRes[2]
	_, err := regexp2.Compile(value, regexp2.None)
	if err != nil {
		msg := fmt.Sprintf("invalid VAR input - regular expression failed: %s : %v", s, err)
		return "", "", fmt.Errorf(msg)
	}
	if strings.Contains(name, "_") {
		return "", "", fmt.Errorf("invalid VAR input - var names cannot contain underscore: %s", s)
	}
	if strings.HasPrefix(strings.TrimSpace(value), "=") {
		return "", "", fmt.Errorf("invalid VAR input - var names cannot start with equal sign: %s", s)
	}
	quoteFix := varReQuoteFix.FindStringSubmatch(value)
	if quoteFix != nil {
		value = quoteFix[1]
	}
	return name, value, nil
}

var testReSimple = regexp.MustCompile("^TEST +([^ ]+) +-> +([^,()]+)$")
var testReVariants = regexp.MustCompile("^TEST +([^ ]+) +-> +[(](.+,.+)[)]$")

func newTest(s string) (Test, error) {
	var outputS string
	var matchRes []string
	matchRes = testReSimple.FindStringSubmatch(s)
	if matchRes != nil {
		outputS = matchRes[2]
	} else {
		matchRes = testReVariants.FindStringSubmatch(s)
		if matchRes == nil {
			return Test{}, fmt.Errorf("invalid TEST definition: " + s)
		}
		outputS = matchRes[2]
	}
	if strings.Contains(outputS, "->") {
		return Test{}, fmt.Errorf("invalid TEST definition: " + s)
	}
	input := matchRes[1]
	output := commaSplit.Split(outputS, -1)
	return Test{Input: input, Output: output}, nil
}

var filterRe = regexp.MustCompile("^FILTER +\"(.+)\" +-> +\"(.*)\"$")

func newFilter(s string, vars map[string]string) (Filter, usedVars, error) {
	matchRes := filterRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Filter{}, usedVars{}, fmt.Errorf("invalid FILTER definition: " + s)
	}
	input := matchRes[1]
	output := strings.Replace(matchRes[2], "\\\"", "\"", -1)
	if strings.Contains(output, "->") {
		return Filter{}, usedVars{}, fmt.Errorf("invalid FILTER definition: " + s)
	}
	input, usedVars, err := expandVarsWithBrackets(input, vars)
	if err != nil {
		return Filter{}, usedVars, fmt.Errorf("invalid FILTER definition %s : %v", s, err)
	}
	re, err := regexp2.Compile(input, regexp2.None)
	if err != nil {
		return Filter{}, usedVars, fmt.Errorf("invalid FILTER definition (invalid regexp /%s/): %v", re, err)
	}
	return Filter{Regexp: re, Output: output}, usedVars, nil
}

var unexpandedBracketVar = regexp.MustCompile(`(?:^|[^\\]){([^},\\]+)}`)

func expandVarsWithBrackets(re0 string, vars map[string]string) (string, usedVars, error) {
	usedVars := usedVars{}
	re := re0
	for k, v := range vars {
		k0 := k
		k = fmt.Sprintf("{%s}", k)
		reTmp := re
		re = strings.Replace(re, k, v, -1)
		if re != reTmp {
			usedVars[k0]++
		}
	}
	unexpandedMatch := unexpandedBracketVar.FindStringSubmatch(re)
	if len(unexpandedMatch) > 1 {
		return "", usedVars, fmt.Errorf("Undefined variable %s", unexpandedMatch[1])
	}
	return re, usedVars, nil
}

var unexpandedContextVar1 = regexp.MustCompile("^[A-Z0-9]{2,}$")
var unexpandedContextVar2 = regexp.MustCompile("[A-Z]{2,}")

func expandContextVars(s0 string, isLeft bool, vars map[string]string) (*regexp2.Regexp, usedVars, error) {
	usedVars := usedVars{}
	if isLeft {
		s0 = strings.Replace(s0, "#", "^", -1)
	} else {
		s0 = strings.Replace(s0, "#", "$", -1)
	}
	splitted := strings.Split(s0, " ")
	for i, s := range splitted {
		if val, ok := vars[strings.TrimSpace(s)]; ok {
			splitted[i] = val
			usedVars[s]++
		} else { // if it's not a VAR, it should be valid orthographic
			if unexpandedContextVar1.MatchString(s) && unexpandedContextVar1.MatchString(s) {
				return &regexp2.Regexp{}, usedVars, fmt.Errorf("Undefined variable %s", s)
			}
		}
	}
	if isLeft {
		res, err := regexp2.Compile(strings.Join(splitted, "")+"$", regexp2.None)
		return res, usedVars, err
	}
	res, err := regexp2.Compile("^"+strings.Join(splitted, ""), regexp2.None)
	return res, usedVars, err
}

var contextRe = regexp.MustCompile("^ +/ +((?:[^_>]+)?) *_ *((?:[^_>]+)?)$")

func newContext(s string, vars map[string]string) (Context, Context, usedVars, error) {
	usedVars := usedVars{}
	if len(strings.TrimSpace(s)) == 0 {
		return Context{}, Context{}, usedVars, nil
	}
	matchRes := contextRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Context{}, Context{}, usedVars, fmt.Errorf("invalid context definition: " + s)
	}
	left := Context{}
	right := Context{}
	leftS := strings.TrimSpace(matchRes[1])
	if len(leftS) > 0 {
		re, usedVarsTmp, err := expandContextVars(leftS, true, vars)
		if err != nil {
			return Context{}, Context{}, usedVars, fmt.Errorf("invalid context definition: %s", err)
		}
		left.Regexp = re
		left.Input = leftS
		for k, v := range usedVarsTmp {
			usedVars[k] += v
		}
	}
	rightS := strings.TrimSpace(matchRes[2])
	if len(rightS) > 0 {
		re, usedVarsTmp, err := expandContextVars(rightS, false, vars)
		if err != nil {
			return Context{}, Context{}, usedVars, fmt.Errorf("invalid context definition: %s", err)
		}
		right.Regexp = re
		right.Input = rightS
		for k, v := range usedVarsTmp {
			usedVars[k] += v
		}
	}
	return left, right, usedVars, nil
}

var ruleRe = regexp.MustCompile("^([^ ]+) +-> +([^/]+)( +/.*$|$)")
var ruleOutputReSimple = regexp.MustCompile("^([^,()]+)$")
var ruleOutputReVariants = regexp.MustCompile("^[(](.+,.+)[)]$")
var emptyOutput = "∅"

func newRuleOutput(s string, l string) ([]string, error) {
	s = strings.TrimSpace(s)
	var outputS string
	var matchRes []string
	matchRes = ruleOutputReSimple.FindStringSubmatch(s)
	if matchRes != nil {
		outputS = matchRes[1]
	} else {
		matchRes = ruleOutputReVariants.FindStringSubmatch(s)
		if matchRes == nil {
			return []string{}, fmt.Errorf("invalid rule output definition: " + l)
		}
		outputS = matchRes[1]
	}
	if strings.Contains(outputS, "->") {
		return []string{}, fmt.Errorf("invalid rule output definition: " + l)
	}
	outputS = strings.Replace(outputS, emptyOutput, "", -1)
	return commaSplit.Split(outputS, -1), nil
}

func newRule(s string, vars map[string]string) (Rule, usedVars, error) {
	usedVars := usedVars{}
	// INPUT -> OUTPUT
	// INPUT -> OUTPUT / LEFTCONTEXT _ RIGHTCONTEXT
	matchRes := ruleRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Rule{}, usedVars, fmt.Errorf("invalid rule definition: " + s)
	}
	input := matchRes[1]
	if input == "\u00a0" { // nbsp
		input = " "
	}
	output, err := newRuleOutput(matchRes[2], s)
	if err != nil {
		return Rule{}, usedVars, err
	}
	left, right, usedVars, err := newContext(matchRes[3], vars)
	if err != nil {
		return Rule{}, usedVars, err
	}
	return Rule{Input: input, Output: output, LeftContext: left, RightContext: right}, usedVars, nil
}
