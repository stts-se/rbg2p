package util

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// PhonemeSet is a package internal container for the phoneme set definition
type PhonemeSet struct {
	Symbols     []string
	DelimiterRe *regexp.Regexp
}

// NewPhonemeSet creates a phoneme set from a slice of symbols, and a phoneme delimiter string
func NewPhonemeSet(symbols []string, delimiter string) (PhonemeSet, error) {
	reString := delimiter
	if len(delimiter) > 0 {
		reString = delimiter + "+"
	}
	delimRe, err := regexp.Compile(reString)

	if err != nil {
		return PhonemeSet{}, fmt.Errorf("couldn't create delimiter regexp from string /%s/ : %s", delimiter, err)
	}
	return PhonemeSet{
		Symbols:     symbols,
		DelimiterRe: delimRe,
	}, nil
}

// LoadPhonemeSetFile loads a phoneme set definitionf from file (one phoneme per line, // for comments)
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

// ValidPhoneme returns true if the input symbol is a valid phoneme, otherwise false
func (ss PhonemeSet) ValidPhoneme(symbol string) bool {
	for _, s := range ss.Symbols {
		if s == symbol {
			return true
		}
	}
	return false
}

// SplitTranscription splits the input transcription into a slice of phonemes, based on the pre-defined phoneme delimiter
func (ss PhonemeSet) SplitTranscription(trans string) ([]string, error) {
	if len(trans) == 0 {
		return []string{}, nil
	}
	if ss.DelimiterRe.MatchString("") {
		splitted, unknown, err := splitIntoPhonemes(ss.Symbols, trans)
		if err != nil {
			return []string{}, err
		}
		if len(unknown) > 0 {
			return []string{}, fmt.Errorf("found unknown phonemes in transcription /%v/: %s\n", trans, unknown)
		}
		return splitted, nil
	}
	return ss.DelimiterRe.Split(trans, -1), nil
}

// Validate an input string using a specified phoneme set
func Validate(input string, phonemeSet PhonemeSet) ([]string, error) {
	var invalid = []string{}
	splitted, err := phonemeSet.SplitTranscription(input)
	if err != nil {
		return nil, err
	}
	for _, symbol := range splitted {
		if !phonemeSet.ValidPhoneme(symbol) {
			invalid = append(invalid, symbol)
		}
	}
	return invalid, nil
}

// CheckForUnusedSymbols compares the phoneme set to a map of used symbols, to tell what symbols in the phoneme set hasn't been used. Mainly for package internal use.
func CheckForUnusedSymbols(usedSymbols map[string]bool, phonemeSet PhonemeSet) []string {
	warnings := []string{}
	for _, symbol := range phonemeSet.Symbols {
		if _, ok := usedSymbols[symbol]; !ok {
			warnings = append(warnings, fmt.Sprintf("symbol /%s/ not used in g2p rule file", symbol))
		}
	}
	return warnings
}
