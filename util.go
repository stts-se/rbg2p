package rbg2p

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

func isSyllDefLine(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF ")
}

var multiSpace = regexp.MustCompile(" +")

var commentAtEndRe = regexp.MustCompile("^(.*[^/]+)//+.*$")

func trimComment(s string) string {
	return strings.TrimSpace(commentAtEndRe.ReplaceAllString(s, "$1"))
}

func isComment(s string) bool {
	return strings.HasPrefix(s, "//")
}

func isBlankLine(s string) bool {
	return len(s) == 0
}

func isPhonemeDelimiter(s string) bool {
	return strings.HasPrefix(s, "PHONEME_DELIMITER ")
}

var phnDelimRe = regexp.MustCompile("^(PHONEME_DELIMITER) +\"(.*)\"$")

func parsePhonemeDelimiter(s string) (string, error) {
	var matchRes []string
	matchRes = phnDelimRe.FindStringSubmatch(s)
	if matchRes == nil {
		return "", fmt.Errorf("invalid phoneme delimiter definition: " + s)
	}
	return matchRes[2], nil
}

func isPhonemeSet(s string) bool {
	return strings.HasPrefix(s, "PHONEME_SET ")
}

var phnSetRe = regexp.MustCompile("^(PHONEME_SET) +\"(.*)\"$")

func parsePhonemeSet(line string, phnDelim string) (PhonemeSet, error) {
	var matchRes []string
	matchRes = phnSetRe.FindStringSubmatch(line)
	if matchRes == nil {
		return PhonemeSet{}, fmt.Errorf("invalid phoneme set definition: " + line)
	}
	value := matchRes[2]
	phonemes := multiSpace.Split(value, -1)
	phonemeSet, err := newPhonemeSet(phonemes, phnDelim)
	if err != nil {
		return PhonemeSet{}, fmt.Errorf("couldn't create phoneme set : %s", err)
	}

	return phonemeSet, nil
}

var commaSplit = regexp.MustCompile(" *, *")
