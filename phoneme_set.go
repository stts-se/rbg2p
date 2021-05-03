package rbg2p

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
		reString = "[" + delimiter + "]"
		if delimiter == " " {
			reString = reString + "+"
		}
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

// LoadPhonemeSetFile loads a phoneme set definition from file (one phoneme per line, // for comments)
func LoadPhonemeSetFile(fName string, delimiter string) (PhonemeSet, error) {
	symbols := []string{}
	fh, err := os.Open(filepath.Clean(fName))
	if err != nil {
		return PhonemeSet{}, err
	}
	/* #nosec G307 */
	defer fh.Close()
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

// validPhoneme returns true if the input symbol is a valid phoneme, otherwise false
func (ps PhonemeSet) validPhoneme(symbol string) bool {
	for _, s := range ps.Symbols {
		if s == symbol {
			return true
		}
	}
	return false
}

// SplitTranscription splits the input transcription into a slice of phonemes, based on the pre-defined phoneme delimiter
func (ps PhonemeSet) SplitTranscription(trans string) ([]string, error) {
	if len(trans) == 0 {
		return []string{}, nil
	}
	if ps.DelimiterRe.MatchString("") {
		splitted, unknown, err := splitIntoPhonemes(ps.Symbols, trans)
		if err != nil {
			return []string{}, err
		}
		if len(unknown) > 0 {
			return []string{}, fmt.Errorf("found unknown phonemes in transcription /%v/: %s", trans, unknown)
		}
		return splitted, nil
	}
	return ps.DelimiterRe.Split(trans, -1), nil
}

func (ps PhonemeSet) validate(input string) ([]string, error) {
	var invalid = []string{}
	splitted, err := ps.SplitTranscription(input)
	if err != nil {
		return nil, err
	}
	for _, symbol := range splitted {
		if !ps.validPhoneme(symbol) {
			invalid = append(invalid, symbol)
		}
	}
	return invalid, nil
}

func checkForUnusedSymbols(usedSymbols map[string]bool, phonemeSet PhonemeSet) []string {
	warnings := []string{}
	for _, symbol := range phonemeSet.Symbols {
		if _, ok := usedSymbols[symbol]; !ok {
			warnings = append(warnings, fmt.Sprintf("symbol /%s/ not used in g2p rule file", symbol))
		}
	}
	return warnings
}
