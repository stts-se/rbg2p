package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type PhonemeSet struct {
	Symbols     []string
	DelimiterRe *regexp.Regexp
}

func NewPhonemeSet(symbols []string, delimiter string) (PhonemeSet, error) {
	delimRe, err := regexp.Compile(delimiter)
	if err != nil {
		return PhonemeSet{}, fmt.Errorf("couldn't create delimiter regexp from string /%s/ : %s", delimiter, err)
	}
	return PhonemeSet{
		Symbols:     symbols,
		DelimiterRe: delimRe,
	}, nil
}

func LoadPhonemeSetFile(fName string, delimiter string) (PhonemeSet, error) {
	symbols := []string{}
	fh, err := os.Open(fName)
	defer fh.Close()
	if err != nil {
		return PhonemeSet{}, err
	}
	n := 0
	s := bufio.NewScanner(fh)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return PhonemeSet{}, err
		}
		n++
		l := strings.TrimSpace(s.Text())
		if len(l) > 0 && !strings.HasPrefix(l, "//") {
			symbols = append(symbols, l)
		}
	}
	delimRe, err := regexp.Compile(delimiter)
	if err != nil {
		return PhonemeSet{}, fmt.Errorf("couldn't create delimiter regexp from string /%s/ : %s", delimiter, err)
	}
	return PhonemeSet{
		Symbols:     symbols,
		DelimiterRe: delimRe,
	}, nil
}

func (ss PhonemeSet) ValidPhoneme(symbol string) bool {
	for _, s := range ss.Symbols {
		if s == symbol {
			return true
		}
	}
	return false
}

func (ss PhonemeSet) SplitTranscription(trans string) ([]string, error) {
	if ss.DelimiterRe.MatchString("") {
		splitted, unknown, err := splitIntoPhonemes(ss.Symbols, trans)
		if err != nil {
			return []string{}, err
		}
		if len(unknown) > 0 {
			return []string{}, fmt.Errorf("found unknown phonemes in transcription /%v/: %s\n", trans, unknown)
		}
		return splitted, nil
	} else {
		return ss.DelimiterRe.Split(trans, -1), nil
	}
}

func validate(input string, phonemeSet PhonemeSet, usedSymbols map[string]bool) ([]string, error) {
	var invalid = []string{}
	splitted, err := phonemeSet.SplitTranscription(input)
	if err != nil {
		return nil, err
	}
	for _, symbol := range splitted {
		if !phonemeSet.ValidPhoneme(symbol) {
			invalid = append(invalid, symbol)
		}
		usedSymbols[symbol] = true
	}
	return invalid, nil
}

func checkForUnusedSymbols(symbols map[string]bool, phonemeSet PhonemeSet) []string {
	warnings := []string{}
	for _, symbol := range phonemeSet.Symbols {
		if _, ok := symbols[symbol]; !ok {
			warnings = append(warnings, fmt.Sprintf("symbol /%s/ not used in g2p rule file", symbol))
		}
	}
	return warnings
}

// CompareToPhonemeSet validates the phonemes in the g2p rule set against the specified phonemeset. Returns an array of invalid phonemes, if any; or if errors are found, this is returned instead.
func CompareToPhonemeSet(ruleSet RuleSet) (TestResult, error) {
	var validation = TestResult{}
	var usedSymbols = map[string]bool{}
	for _, rule := range ruleSet.Rules {
		for _, output := range rule.Output {
			invalid, err := validate(output, ruleSet.PhonemeSet, usedSymbols)
			if err != nil {
				return TestResult{}, fmt.Errorf("found error in rule output /%s/ : %s", output, err)
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in rule output %s: %s", rule, symbol))
				usedSymbols[symbol] = true
			}
		}
	}
	for _, test := range ruleSet.Tests {
		for _, output := range test.Output {
			invalid, err := validate(output, ruleSet.PhonemeSet, usedSymbols)
			if err != nil {
				return TestResult{}, fmt.Errorf("found error in test output /%s/ : %s", output, err)
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in test output %s: %s", test, symbol))
			}

		}
	}
	for _, warn := range checkForUnusedSymbols(usedSymbols, ruleSet.PhonemeSet) {
		validation.Warnings = append(validation.Warnings, warn)
	}
	return validation, nil
}
