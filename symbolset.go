package rbg2p

import (
	"fmt"
	"strings"

	"github.com/stts-se/pronlex/symbolset"
)

func validate(input string, symbolSet symbolset.SymbolSet, usedSymbols map[string]bool) ([]string, error) {
	var invalid = []string{}
	splitted, err := symbolSet.SplitTranscription(input)
	if err != nil {
		return nil, err
	}
	for _, symbol := range splitted {
		if !symbolSet.ValidSymbol(symbol) {
			invalid = append(invalid, symbol)
		}
		usedSymbols[symbol] = true
	}
	return invalid, nil
}

type CompareResult struct {
	Errors   []string
	Warnings []string
}

func checkForUnusedChars(coveredChars map[string]bool, individualChars map[string]bool, characterSet []string, validation *CompareResult) {
	var errors = []string{}
	for _, char := range characterSet {
		if _, ok := coveredChars[char]; !ok {
			errors = append(errors, char)
		}
	}
	validation.Warnings = append(validation.Warnings, fmt.Sprintf("no rules exist for characters: %s", strings.Join(errors, ",")))

	errors = []string{}
	for _, char := range characterSet {
		if _, ok := individualChars[char]; !ok {
			errors = append(errors, char)
		}
	}
	validation.Errors = append(validation.Errors, fmt.Sprintf("no default rules for characters: %s", strings.Join(errors, ",")))
}

func checkForUnusedSymbols(symbols map[string]bool, symbolSet symbolset.SymbolSet) []string {
	warnings := []string{}
	for _, symbol := range symbolSet.PhoneticSymbols {
		if _, ok := symbols[symbol.String]; !ok {
			warnings = append(warnings, fmt.Sprintf("symbol /%s/ not used in g2p rule file", symbol.String))
		}
	}
	return warnings
}

// CompareToSymbolSet validates the phonemes in the g2p rule set against the specified symbolset. Returns an array of invalid symbols, if any; or if errors are found, this is returned instead.
func CompareToSymbolSet(ruleSet RuleSet, symbolSet symbolset.SymbolSet) (CompareResult, error) {
	var validation = CompareResult{}
	var usedSymbols = map[string]bool{}
	var coveredChars = map[string]bool{}
	var individualChars = map[string]bool{}
	for _, rule := range ruleSet.Rules {
		for _, char := range strings.Split(rule.Input, "") {
			coveredChars[char] = true
		}
		individualChars[rule.Input] = true
		for _, output := range rule.Output {
			invalid, err := validate(output, symbolSet, usedSymbols)
			if err != nil {
				return CompareResult{}, fmt.Errorf("found error in rule output %s : s%", output, err)
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in rule output %s: %s", rule, symbol))
				usedSymbols[symbol] = true
			}
		}
	}
	for _, test := range ruleSet.Tests {
		for _, output := range test.Output {
			invalid, err := validate(output, symbolSet, usedSymbols)
			if err != nil {
				return CompareResult{}, fmt.Errorf("found error in test output %v : s%", output, err)
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in test output %s: %s", test, symbol))
			}

		}
	}
	for _, warn := range checkForUnusedSymbols(usedSymbols, symbolSet) {
		validation.Warnings = append(validation.Warnings, warn)
	}
	checkForUnusedChars(coveredChars, individualChars, ruleSet.CharacterSet, &validation)
	return validation, nil
}
