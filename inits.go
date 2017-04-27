package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func isComment(s string) bool {
	return strings.HasPrefix(s, "#")
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
	ruleSet := RuleSet{}
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
		l := strings.TrimSpace(s.Text())
		if isBlankLine(l) || isComment(l) {
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

var varRe = regexp.MustCompile("^VAR +([^ ]+) +([^ ]+)$")

func newVar(s string) (string, string, error) {
	// VAR NAME VALUE
	matches := varRe.FindStringSubmatch(s)
	if matches == nil {
		return "", "", fmt.Errorf("invalid var definition in input: " + s)
	}
	name := matches[1]
	value := matches[2]
	return name, value, nil
}

var testRe = regexp.MustCompile("^TEST +([^ ]+) +-> +[{]?([^{}>]+)[}]?$")
var commaSplit = regexp.MustCompile(" *, *")

func newTest(s string) (Test, error) {
	// TEST ORTH -> TRANS
	// TEST ORTH -> {TRANS1, TRANS2}
	matches := testRe.FindStringSubmatch(s)
	if matches == nil {
		return Test{}, fmt.Errorf("invalid test definition in input: " + s)
	}
	input := matches[1]
	output := commaSplit.Split(matches[2], -1)
	return Test{Input: input, Output: output}, nil
}

func expandVars(s0 string, vars map[string]string) string {
	splitted := strings.Split(s0, " ")
	for i, s := range splitted {
		if val, ok := vars[strings.TrimSpace(s)]; ok {
			splitted[i] = val
		}
	}
	return strings.Join(splitted, " ")
}

var contextRe = regexp.MustCompile("^ +/ +((?:[^_ >]+)?) *_ *((?:[^_ >]+)?)$")

func newContext(s string, vars map[string]string) (Context, Context, error) {
	if len(strings.TrimSpace(s)) == 0 {
		return Context{}, Context{}, nil
	}
	matches := contextRe.FindStringSubmatch(s)
	if matches == nil {
		return Context{}, Context{}, fmt.Errorf("invalid context definition in input: " + s)
	}
	left := Context{}
	right := Context{}
	leftS := strings.TrimSpace(matches[1])
	if len(leftS) > 0 {
		re, err := regexp.Compile(strings.Replace(expandVars(leftS, vars), "#", "^", -1))
		if err != nil {
			return Context{}, Context{}, fmt.Errorf("invalid context definition in input: %s", err)
		}
		left.regexp = re
	}
	rightS := strings.TrimSpace(matches[2])
	if len(rightS) > 0 {
		re, err := regexp.Compile(strings.Replace(expandVars(rightS, vars), "#", "$", -1))
		if err != nil {
			return Context{}, Context{}, fmt.Errorf("invalid context definition in input: %s", err)
		}
		right.regexp = re
	}
	return left, right, nil
}

var ruleRe = regexp.MustCompile("^([^ ]+) +-> +[{]?([^{}/>]+)[}]?( +/.*$|$)")

func newRule(s string, vars map[string]string) (Rule, error) {
	// INPUT -> OUTPUT
	// INPUT -> OUTPUT / LEFTCONTEXT _ RIGHTCONTEXT
	matches := ruleRe.FindStringSubmatch(s)
	if matches == nil {
		return Rule{}, fmt.Errorf("invalid rule definition in input: " + s)
	}
	input := matches[1]
	output := commaSplit.Split(matches[2], -1)
	left, right, err := newContext(matches[3], vars)
	if err != nil {
		return Rule{}, err
	}
	return Rule{Input: input, Output: output, LeftContext: left, RightContext: right}, nil
}
