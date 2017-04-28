package rbg2p

import (
	"fmt"

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
func CompareToSymbolSet(ruleSet RuleSet, symbolSet symbolset.SymbolSet) (TestResult, error) {
	var validation = TestResult{}
	var usedSymbols = map[string]bool{}
	for _, rule := range ruleSet.Rules {
		for _, output := range rule.Output {
			invalid, err := validate(output, symbolSet, usedSymbols)
			if err != nil {
				return TestResult{}, fmt.Errorf("found error in rule output %s : v%", output, err)
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
				return TestResult{}, fmt.Errorf("found error in test output %s : v%", output, err)
			}
			for _, symbol := range invalid {
				validation.Errors = append(validation.Errors, fmt.Sprintf("invalid symbol in test output %s: %s", test, symbol))
			}

		}
	}
	for _, warn := range checkForUnusedSymbols(usedSymbols, symbolSet) {
		validation.Warnings = append(validation.Warnings, warn)
	}
	return validation, nil
}
