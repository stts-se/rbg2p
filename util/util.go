package util

import (
	"fmt"
	"regexp"
	"strings"
)

// TestResult is a container for test results (errors, warnings, and failed tests from tests speficied in the g2p rule file)
type TestResult struct {
	Errors      []string
	Warnings    []string
	FailedTests []string
}

// IsSyllDefLine is used to tell if an input line in a g2p file is part of syllabification definition
func IsSyllDefLine(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF ")
}

var multiSpace = regexp.MustCompile(" +")

var commentAtEndRe = regexp.MustCompile("^(.*[^/]+)//+.*$")

// TrimComment removes trailing comment, if any, from an input line
func TrimComment(s string) string {
	return strings.TrimSpace(commentAtEndRe.ReplaceAllString(s, "$1"))
}

// IsComment is used to check if an input line is a comment line
func IsComment(s string) bool {
	return strings.HasPrefix(s, "//")
}

// IsBlankLine is used to check if an input line is blank
func IsBlankLine(s string) bool {
	return len(s) == 0
}

// IsPhonemeDelimiter is used to check if an input line defines a phoneme delimiter
func IsPhonemeDelimiter(s string) bool {
	return strings.HasPrefix(s, "PHONEME_DELIMITER ")
}

var phnDelimRe = regexp.MustCompile("^(PHONEME_DELIMITER) +\"(.*)\"$")

// ParsePhonemeDelimiter is to used parse an input line defining a phoneme delimiter
func ParsePhonemeDelimiter(s string) (string, error) {
	var matchRes []string
	matchRes = phnDelimRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", fmt.Errorf("invalid phoneme delimiter definition: " + s)
	}
	return matchRes[2], nil
}

// IsPhonemeSet is used to check if an input line defines a phoneme set
func IsPhonemeSet(s string) bool {
	return strings.HasPrefix(s, "PHONEME_SET ")
}

var phnSetRe = regexp.MustCompile("^(PHONEME_SET) +\"(.*)\"$")

// ParsePhonemeSet is used to parse an input line defining a phoneme set
func ParsePhonemeSet(line string, phnDelim string) (PhonemeSet, error) {
	var matchRes []string
	matchRes = phnSetRe.FindStringSubmatch(line)
	if matchRes == nil {
		return PhonemeSet{}, fmt.Errorf("invalid phoneme set definition: " + line)
	}
	value := matchRes[2]
	phonemes := multiSpace.Split(value, -1)
	phonemeSet, err := NewPhonemeSet(phonemes, phnDelim)
	if err != nil {
		return PhonemeSet{}, fmt.Errorf("couldn't create phoneme set : %s", err)
	}

	return phonemeSet, nil
}
