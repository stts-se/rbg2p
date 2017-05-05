package rbg2p

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

var fsExpGot = "Expected: %v got: %v"

// container to compare variables in testing
type tVar struct {
	Name  string
	Value string
}

func TestNewVar(t *testing.T) {
	validLines := map[string]tVar{
		"VAR VOICED_PLOSIVE [dgb]":      tVar{Name: "VOICED_PLOSIVE", Value: "[dgb]"},
		"VAR VOWEL [aoiuye]":            tVar{Name: "VOWEL", Value: "[aoiuye]"},
		"VAR VOICELESS [p|k|t|f|s|h|c]": tVar{Name: "VOICELESS", Value: "[p|k|t|f|s|h|c]"},
	}
	invalidLines := map[string]tVar{
		"VAR VOICED_PLOSIVE [dgb]": tVar{Name: "VOICED_PLOSIVE", Value: "dgb"},
		"VAR VOWEL [aoiuye]":       tVar{Name: "VOWEL", Value: "[aoiuye"},
	}
	failLines := []string{
		"VAR VOICED_PLOSIVE",
		"VAR VOICED_PLOSIVE [dgb] anka",
	}

	for l, expect := range validLines {
		name, val, err := newVar(l)
		result := tVar{Name: name, Value: val}
		if err != nil {
			t.Errorf("didn't expect error for input var line %s : %s", l, err)
		} else if expect != result {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		name, val, err := newVar(l)
		result := tVar{Name: name, Value: val}
		if err != nil {
			t.Errorf("didn't expect error for input var line %s : %s", l, err)
		} else if expect == result {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, _, err := newVar(l)
		if err == nil {
			t.Errorf("expected error for input var line %s", l)
		}
	}

}

func TestNewTest(t *testing.T) {
	validLines := map[string]Test{
		"TEST anka -> AnkA":            Test{Input: "anka", Output: []string{"AnkA"}},
		"TEST banka -> (bAnkA, bANkA)": Test{Input: "banka", Output: []string{"bAnkA", "bANkA"}},
	}
	invalidLines := map[string]Test{
		"TEST anka -> AnkA":            Test{Input: "anka", Output: []string{"anka"}},
		"TEST banka -> (bAnkA, bANkA)": Test{Input: "banka", Output: []string{"bAnkA", "bANKkA"}},
	}
	failLines := []string{
		"TEST anka",
		"TEST anka AnkA",
		"TEST anka -> AnkA -> ANkA",
		"TEST anka -> (AnkA)",
		"TEST banka -> bAnkA, bANkA",
	}

	for l, expect := range validLines {
		result, err := newTest(l)
		if err != nil {
			t.Errorf("didn't expect error for input test line %s : %s", l, err)
		} else if !expect.equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newTest(l)
		if err != nil {
			t.Errorf("didn't expect error for input test line %s : %s", l, err)
		} else if expect.equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, err := newTest(l)
		if err == nil {
			t.Errorf("expected error for input test line %s", l)
		}
	}

}

func TestNewRule(t *testing.T) {
	vars := map[string]string{
		"VOICED": "[dgjlvbnm]",
	}
	validLines := map[string]Rule{
		"sch -> (x, S) / _ #": Rule{Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{"#", regexp.MustCompile("$")}},
		"sch -> (x, S)": Rule{Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A / _ VOICED": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{"VOICED", regexp.MustCompile("[dgjlvbnm]")}},
		"a -> A / _ VOICED #": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{"VOICED #", regexp.MustCompile("[dgjlvbnm]$")}},
	}
	invalidLines := map[string]Rule{}
	failLines := []string{
		"sch -> x, S",
		"sch -> (x)",
	}

	for l, expect := range validLines {
		result, err := newRule(l, vars)
		if err != nil {
			t.Errorf("didn't expect error for input rule line %s : %s", l, err)
		} else if !expect.equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newRule(l, vars)
		if err != nil {
			t.Errorf("didn't expect error for input rule line %s : %s", l, err)
		} else if expect.equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, err := newRule(l, vars)
		if err == nil {
			t.Errorf("expected error for input rule line %s", l)
		}
	}

}
func TestLoadFile1(t *testing.T) {
	fName := "test_data/test.g2p"
	_, err := LoadFile(fName)
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
	}
}
func TestLoadFile2(t *testing.T) {
	fName := "test_data/test_err.g2p"
	_, err := LoadFile(fName)
	if err == nil {
		t.Errorf("expected error for input file %s :", fName)
	}
}

func loadAndTest(t *testing.T, fName string) (RuleSet, error) {
	rs, err := LoadFile(fName)
	if err != nil {
		return rs, fmt.Errorf("didn't expect error for input file %s : %s", fName, err)
	}

	result := rs.Test()
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %v\n", e)
		}
		fmt.Printf("%d ERROR(S) FOR %s\n", len(result.Errors), fName)
	}
	if len(result.Warnings) > 0 {
		for _, e := range result.Warnings {
			fmt.Printf("WARNING: %v\n", e)
		}
		fmt.Printf("%d WARNING(S) FOR %s\n", len(result.Warnings), fName)
	}
	if len(result.FailedTests) > 0 {
		for _, e := range result.FailedTests {
			fmt.Printf("FAILED TEST: %v\n", e)
		}
		fmt.Printf("%d OF %d TESTS FAILED FOR %s\n", len(result.FailedTests), len(rs.Tests), fName)
	} else {
		fmt.Printf("ALL %d TESTS PASSED FOR %s\n", len(rs.Tests), fName)
	}
	if len(result.Errors) > 0 || len(result.FailedTests) > 0 {
		return rs, fmt.Errorf("Init/tests failed for %s", fName)
	}
	return rs, nil
}
func TestSws(t *testing.T) {
	_, err := loadAndTest(t, "test_data/sws_test.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestApply(t *testing.T) {
	fName := "test_data/test.g2p"
	rs, err := loadAndTest(t, fName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	_, err = rs.Apply("hiß")
	if err == nil {
		t.Errorf("expected error for input file %s", fName)
		return
	}

	_, err = rs.Apply("hit")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

	_, err = rs.Apply("dusch")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

	_, err = rs.Apply("duscha")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

}

func TestWithPhnDelim(t *testing.T) {
	fName := "test_data/test_specs.g2p"
	rs, err := loadAndTest(t, fName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}
	_, err = rs.Apply("hi§")
	if err == nil {
		t.Errorf("expected error for input file %s", fName)
		return
	}

	_, err = rs.Apply("hit")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

	_, err = rs.Apply("dusch")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

	_, err = rs.Apply("duscha")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
		return
	}

}

func TestIPA(t *testing.T) {
	_, err := loadAndTest(t, "test_data/ipa_test.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func xxxTestHun(t *testing.T) {
	_, err := loadAndTest(t, "test_data/hun.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func xxxTestMkd(t *testing.T) {
	_, err := loadAndTest(t, "test_data/mkd.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func xxxTestCze(t *testing.T) {
	_, err := loadAndTest(t, "test_data/czc.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestExpansionAlgorithm(t *testing.T) {
	rs := RuleSet{PhonemeDelimiter: " "}
	input := [][]string{[]string{"1a", "1b"}, []string{"2a", "2b"}}
	expect := []Trans{
		Trans{[]string{"1a", "2a"}},
		Trans{[]string{"1a", "2b"}},
		Trans{[]string{"1b", "2a"}},
		Trans{[]string{"1b", "2b"}},
	}

	result := rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

	input = [][]string{[]string{"1a", "1b"}, []string{"2"}, []string{"3a", "3b"}}
	expect = []Trans{
		Trans{[]string{"1a", "2", "3a"}},
		Trans{[]string{"1a", "2", "3b"}},
		Trans{[]string{"1b", "2", "3a"}},
		Trans{[]string{"1b", "2", "3b"}},
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

	input = [][]string{[]string{"1a", "1b"}, []string{"2a", "2b"}, []string{"3a", "3b"}}
	expect = []Trans{
		Trans{[]string{"1a", "2a", "3a"}},
		Trans{[]string{"1a", "2a", "3b"}},
		Trans{[]string{"1a", "2b", "3a"}},
		Trans{[]string{"1a", "2b", "3b"}},
		Trans{[]string{"1b", "2a", "3a"}},
		Trans{[]string{"1b", "2a", "3b"}},
		Trans{[]string{"1b", "2b", "3a"}},
		Trans{[]string{"1b", "2b", "3b"}},
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

	input = [][]string{[]string{"1a", "1b"}, []string{"2a", "2b", "2c"}, []string{"3a", "3b"}}
	expect = []Trans{
		Trans{[]string{"1a", "2a", "3a"}},
		Trans{[]string{"1a", "2a", "3b"}},
		Trans{[]string{"1a", "2b", "3a"}},
		Trans{[]string{"1a", "2b", "3b"}},
		Trans{[]string{"1a", "2c", "3a"}},
		Trans{[]string{"1a", "2c", "3b"}},
		Trans{[]string{"1b", "2a", "3a"}},
		Trans{[]string{"1b", "2a", "3b"}},
		Trans{[]string{"1b", "2b", "3a"}},
		Trans{[]string{"1b", "2b", "3b"}},
		Trans{[]string{"1b", "2c", "3a"}},
		Trans{[]string{"1b", "2c", "3b"}},
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

	input = [][]string{[]string{"b"}, []string{"O"}, []string{"rt", "r t"}, []string{"a"}, []string{"d"}, []string{"u0"}, []string{"S", "x"}}
	expect = []Trans{
		Trans{[]string{"b", "O", "rt", "a", "d", "u0", "S"}},
		Trans{[]string{"b", "O", "rt", "a", "d", "u0", "x"}},
		Trans{[]string{"b", "O", "r", "t", "a", "d", "u0", "S"}},
		Trans{[]string{"b", "O", "r", "t", "a", "d", "u0", "x"}},
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %#v\nFound    %#v", expect, result)
	}

	input = [][]string{[]string{"1"}, []string{"2"}, []string{"3a", "3b"}, []string{"4"}, []string{"5"}, []string{"6"}, []string{"7a", "7b"}}
	expect = []Trans{
		Trans{[]string{"1", "2", "3a", "4", "5", "6", "7a"}},
		Trans{[]string{"1", "2", "3a", "4", "5", "6", "7b"}},
		Trans{[]string{"1", "2", "3b", "4", "5", "6", "7a"}},
		Trans{[]string{"1", "2", "3b", "4", "5", "6", "7b"}},
	}
	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

	input = [][]string{[]string{"1"}, []string{"2"}, []string{"3a", "3b"}, []string{"4"}, []string{"5"}, []string{"6"}, []string{"7a", "7b"}, []string{"8"}}
	expect = []Trans{
		Trans{[]string{"1", "2", "3a", "4", "5", "6", "7a", "8"}},
		Trans{[]string{"1", "2", "3a", "4", "5", "6", "7b", "8"}},
		Trans{[]string{"1", "2", "3b", "4", "5", "6", "7a", "8"}},
		Trans{[]string{"1", "2", "3b", "4", "5", "6", "7b", "8"}},
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", expect, result)
	}

}
