package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

var g2pLineRe = regexp.MustCompile("^(CHARACTER_SET|TEST|DEFAULT_PHONEME|FILTER|VAR|) .*")

func isG2PLine(s string) bool {
	return g2pLineRe.MatchString(s) || ruleRe.MatchString(s)
}

// LoadFile loads a g2p rule set from the specified file
func LoadFile(fName string) (RuleSet, error) {
	ruleSet := RuleSet{Vars: map[string]string{}}
	ruleSet.DefaultPhoneme = "_"
	ruleSet.PhonemeDelimiter = " "
	syllDefLines := []string{}
	fh, err := os.Open(fName)
	defer fh.Close()
	if err != nil {
		return ruleSet, err
	}
	n := 0
	s := bufio.NewScanner(fh)
	var inputLines []string
	var ruleLines []string
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
			t, err := newFilter(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.Filters = append(ruleSet.Filters, t)
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

	for _, l := range ruleLines {
		r, err := newRule(l, ruleSet.Vars)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.Rules = append(ruleSet.Rules, r)
	}
	if ruleSet.CharacterSet == nil || len(ruleSet.CharacterSet) == 0 {
		return ruleSet, fmt.Errorf("No character set defined for input file %s", fName)
	}
	ruleSet.Content = strings.Join(inputLines, "\n")
	return ruleSet, nil
}

var constRe = regexp.MustCompile("^(CHARACTER_SET|DEFAULT_PHONEME) +\"(.*)\"$")
var isConstRe = regexp.MustCompile("^(CHARACTER_SET|DEFAULT_PHONEME) .*")

func isConst(s string) bool {
	return isConstRe.MatchString(s)
}

func parseConst(s string, ruleSet *RuleSet) error {
	var matchRes []string
	matchRes = constRe.FindStringSubmatch(s)
	if matchRes != nil {
		name := matchRes[1]
		value := matchRes[2]
		if name == "CHARACTER_SET" {
			ruleSet.CharacterSet = strings.Split(value, "")
		} else if name == "DEFAULT_PHONEME" {
			ruleSet.DefaultPhoneme = value
		} else {
			return fmt.Errorf("invalid const definition: " + s)
		}
	} else {
		return fmt.Errorf("invalid const definition: " + s)
	}
	return nil
}

var varRe = regexp.MustCompile("^VAR +([^ \"]+) +([^ \"]+)$")

func newVar(s string) (string, string, error) {
	// VAR NAME VALUE
	matchRes := varRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", "", fmt.Errorf("invalid var definition: " + s)
	}
	name := matchRes[1]
	value := matchRes[2]
	_, err := regexp2.Compile(value, regexp2.None)
	if err != nil {
		return "", "", fmt.Errorf("invalid var in input (regular expression failed) for /%s/: %s", s, err)
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
			return Test{}, fmt.Errorf("invalid test definition: " + s)
		}
		outputS = matchRes[2]
	}
	if strings.Contains(outputS, "->") {
		return Test{}, fmt.Errorf("invalid test definition: " + s)
	}
	input := matchRes[1]
	output := commaSplit.Split(outputS, -1)
	return Test{Input: input, Output: output}, nil
}

var filterRe = regexp.MustCompile("^FILTER +\"(.+)\" +-> +\"(.+)\"$")

func newFilter(s string) (Filter, error) {
	matchRes := filterRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Filter{}, fmt.Errorf("invalid filter definition: " + s)
	}
	input := matchRes[1]
	output := strings.Replace(matchRes[2], "\\\"", "\"", -1)
	if strings.Contains(output, "->") {
		return Filter{}, fmt.Errorf("invalid filter definition: " + s)
	}
	re, err := regexp2.Compile(input, regexp2.None)
	if err != nil {
		return Filter{}, fmt.Errorf("invalid regexp in filter definition input /%s/ : %s", s, err)
	}
	return Filter{Regexp: re, Output: output}, nil
}

func expandVars(s0 string, isLeft bool, vars map[string]string) (*regexp2.Regexp, error) {
	if isLeft {
		s0 = strings.Replace(s0, "#", "^", -1)
	} else {
		s0 = strings.Replace(s0, "#", "$", -1)
	}
	splitted := strings.Split(s0, " ")
	for i, s := range splitted {
		if val, ok := vars[strings.TrimSpace(s)]; ok {
			splitted[i] = val
		}
	}
	if isLeft {
		return regexp2.Compile(strings.Join(splitted, "")+"$", regexp2.None)
	}
	return regexp2.Compile("^"+strings.Join(splitted, ""), regexp2.None)
}

var contextRe = regexp.MustCompile("^ +/ +((?:[^_>]+)?) *_ *((?:[^_>]+)?)$")

func newContext(s string, vars map[string]string) (Context, Context, error) {
	if len(strings.TrimSpace(s)) == 0 {
		return Context{}, Context{}, nil
	}
	matchRes := contextRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Context{}, Context{}, fmt.Errorf("invalid context definition: " + s)
	}
	left := Context{}
	right := Context{}
	leftS := strings.TrimSpace(matchRes[1])
	if len(leftS) > 0 {
		re, err := expandVars(leftS, true, vars)
		if err != nil {
			return Context{}, Context{}, fmt.Errorf("invalid context definition: %s", err)
		}
		left.Regexp = re
		left.Input = leftS
	}
	rightS := strings.TrimSpace(matchRes[2])
	if len(rightS) > 0 {
		re, err := expandVars(rightS, false, vars)
		if err != nil {
			return Context{}, Context{}, fmt.Errorf("invalid context definition: %s", err)
		}
		right.Regexp = re
		right.Input = rightS
	}
	return left, right, nil
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

func newRule(s string, vars map[string]string) (Rule, error) {
	// INPUT -> OUTPUT
	// INPUT -> OUTPUT / LEFTCONTEXT _ RIGHTCONTEXT
	matchRes := ruleRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Rule{}, fmt.Errorf("invalid rule definition: " + s)
	}
	input := matchRes[1]
	output, err := newRuleOutput(matchRes[2], s)
	if err != nil {
		return Rule{}, err
	}
	left, right, err := newContext(matchRes[3], vars)
	if err != nil {
		return Rule{}, err
	}
	return Rule{Input: input, Output: output, LeftContext: left, RightContext: right}, nil
}
