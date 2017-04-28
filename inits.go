package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var commentAtEndRe = regexp.MustCompile("^(.+)//.*$")

func trimComment(s string) string {
	return commentAtEndRe.ReplaceAllString(s, "$1")
}

func isComment(s string) bool {
	return strings.HasPrefix(s, "//")
}

func isVar(s string) bool {
	return strings.HasPrefix(s, "VAR ")
}

func isTest(s string) bool {
	return strings.HasPrefix(s, "TEST ")
}

func isBlankLine(s string) bool {
	return len(s) == 0
}

// LoadFile loads a g2p rule set from the specified file
func LoadFile(fName string) (RuleSet, error) {
	ruleSet := RuleSet{Vars: map[string]string{}}
	ruleSet.FallbackSymbol = "_"
	ruleSet.PhnDelimiter = " "
	fh, err := os.Open(fName)
	defer fh.Close()
	if err != nil {
		return ruleSet, err
	}
	n := 0
	s := bufio.NewScanner(fh)
	var ruleLines []string
	for s.Scan() {
		if err := s.Err(); err != nil {
			return ruleSet, err
		}
		n++
		l := trimComment(strings.TrimSpace(s.Text()))
		if isBlankLine(l) || isComment(l) {
		} else if isFallbackSymbol(l) {
			s, err := newFallback(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.FallbackSymbol = s
		} else if isPhnSeparator(l) {
			s, err := newPhnSeparator(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.PhnDelimiter = s
		} else if isVar(l) {
			name, value, err := newVar(l)
			if err != nil {
				return ruleSet, err
			}
			ruleSet.Vars[name] = value
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
	for _, l := range ruleLines {
		r, err := newRule(l, ruleSet.Vars)
		if err != nil {
			return ruleSet, err
		}
		ruleSet.Rules = append(ruleSet.Rules, r)
	}
	return ruleSet, nil
}

func isFallbackSymbol(s string) bool {
	return strings.HasPrefix(s, "DEFAULT_PHONEME ")
}

var fallbackRe = regexp.MustCompile("^DEFAULT_PHONEME +\"(.*)\"$")

func newFallback(s string) (string, error) {
	// SET FALLBACK VALUE
	matchRes := fallbackRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", fmt.Errorf("invalid fallback definition: " + s)
	}
	value := matchRes[1]
	_, err := regexp.Compile(value)
	if err != nil {
		return "", fmt.Errorf("invalid fallback in input (regular expression failed) for /%s/: %s", s, err)
	}
	return value, nil
}

func isPhnSeparator(s string) bool {
	return strings.HasPrefix(s, "PHONEME_DELIMITER ")
}

var phnSepRe = regexp.MustCompile("^PHONEME_DELIMITER +\"(.*)\"$")

func newPhnSeparator(s string) (string, error) {
	// SET PHN DELIM VALUE
	matchRes := phnSepRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", fmt.Errorf("invalid phoneme separator definition: " + s)
	}
	value := matchRes[1]
	_, err := regexp.Compile(value)
	if err != nil {
		return "", fmt.Errorf("invalid phoneme separator in input (regular expression failed) for /%s/: %s", s, err)
	}
	return value, nil
}

var varRe = regexp.MustCompile("^VAR +([^ ]+) +([^ ]+)$")

func newVar(s string) (string, string, error) {
	// VAR NAME VALUE
	matchRes := varRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", "", fmt.Errorf("invalid var definition: " + s)
	}
	name := matchRes[1]
	value := matchRes[2]
	_, err := regexp.Compile(value)
	if err != nil {
		return "", "", fmt.Errorf("invalid var in input (regular expression failed) for /%s/: %s", s, err)
	}
	return name, value, nil
}

var testReSimple = regexp.MustCompile("^TEST +([^ ]+) +-> +([^,()]+)$")
var testReVariants = regexp.MustCompile("^TEST +([^ ]+) +-> +[(](.+,.+)[)]$")
var commaSplit = regexp.MustCompile(" *, *")

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

func expandVars(s0 string, isLeft bool, vars map[string]string) (*regexp.Regexp, error) {
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
		return regexp.Compile(strings.Join(splitted, "") + "$")
	}
	return regexp.Compile("^" + strings.Join(splitted, ""))
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
