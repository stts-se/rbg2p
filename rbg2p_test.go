package rbg2p

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dlclark/regexp2"
)

var fsExpGot = "Expected: %v got: %v"

// container to compare variables in testing
type tVar struct {
	name  string
	value string
}

func TestNewVar(t *testing.T) {
	validLines := map[string]tVar{
		"VAR VOICEDPLOSIVE [dgb]":       {name: "VOICEDPLOSIVE", value: "[dgb]"},
		"VAR VOWEL [aoiuye]":            {name: "VOWEL", value: "[aoiuye]"},
		"VAR VOICELESS [p|k|t|f|s|h|c]": {name: "VOICELESS", value: "[p|k|t|f|s|h|c]"},
	}
	invalidLines := map[string]tVar{
		"VAR VOICEDPLOSIVE [dgb]": {name: "VOICEDPLOSIVE", value: "dgb"},
		"VAR VOWEL [aoiuye]":      {name: "VOWEL", value: "[aoiuye"},
	}
	failLines := []string{
		"VAR VOICED_PLOSIVE",
		"VAR VOICED_PLOSIVE [dgb] anka",
	}

	for l, expect := range validLines {
		name, val, err := newVar(l)
		result := tVar{name: name, value: val}
		if err != nil {
			t.Errorf("didn't expect error for input var line %s : %s", l, err)
		} else if expect != result {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		name, val, err := newVar(l)
		result := tVar{name: name, value: val}
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
		"TEST anka -> AnkA":            {Input: "anka", Output: []string{"AnkA"}},
		"TEST banka -> (bAnkA, bANkA)": {Input: "banka", Output: []string{"bAnkA", "bANkA"}},
	}
	invalidLines := map[string]Test{
		"TEST anka -> AnkA":            {Input: "anka", Output: []string{"anka"}},
		"TEST banka -> (bAnkA, bANkA)": {Input: "banka", Output: []string{"bAnkA", "bANKkA"}},
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

func testCompileRegexp(s string) *regexp2.Regexp {
	return regexp2.MustCompile("$", regexp2.None)
}

func TestNewRule(t *testing.T) {
	vars := map[string]string{
		"VOICED": "[dgjlvbnm]",
	}
	validLines := map[string]Rule{
		"sch -> (x, S) / _ #": {Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{"#", testCompileRegexp("$")}},
		"sch -> (x, S)": {Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A": {Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A / _ VOICED": {Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{"VOICED", testCompileRegexp("[dgjlvbnm]")}},
		"a -> A / _ VOICED #": {Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{"VOICED #", testCompileRegexp("[dgjlvbnm]$")}},
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

func TestLoadFile3(t *testing.T) {
	fName := "test_data/enu_cmu.syll"
	_, err := LoadSyllFile(fName)
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
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

func g2pFromSlice(variants []string) g2p {
	return g2p{g: "", p: variants}
}
func g2pTransFromSlice(phns []string) trans {
	res := []g2p{}
	for _, ps := range phns {
		res = append(res, g2p{g: "", p: strings.Split(ps, ", ")}) // split on comma to make it easier to create transcriptions for unit tests

	}
	return trans{phonemes: res}
}

// to make it easier to compare results from unit tests
func trans2string(t trans) string {
	var phns []string
	for _, g2p := range t.phonemes {
		if len(g2p.p) > 0 {
			phns = append(phns, strings.Join(g2p.p, ", "))
		}
	}
	return strings.Join(phns, " ")
}
func transes2string(transes []trans) []string {
	res := []string{}
	for _, t := range transes {
		res = append(res, trans2string(t))
	}
	return res
}

func TestExpansionAlgorithm(t *testing.T) {
	rs := RuleSet{PhonemeDelimiter: " "}

	//
	input := []g2p{g2pFromSlice([]string{"1a", "1b"}), g2pFromSlice([]string{"2a", "2b"})}
	expect := []trans{
		g2pTransFromSlice([]string{"1a", "2a"}),
		g2pTransFromSlice([]string{"1a", "2b"}),
		g2pTransFromSlice([]string{"1b", "2a"}),
		g2pTransFromSlice([]string{"1b", "2b"}),
	}

	result := rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"1a", "1b"}), g2pFromSlice([]string{"2"}), g2pFromSlice([]string{"3a", "3b"})}
	expect = []trans{
		g2pTransFromSlice([]string{"1a", "2", "3a"}),
		g2pTransFromSlice([]string{"1a", "2", "3b"}),
		g2pTransFromSlice([]string{"1b", "2", "3a"}),
		g2pTransFromSlice([]string{"1b", "2", "3b"}),
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"1a", "1b"}), g2pFromSlice([]string{"2a", "2b"}), g2pFromSlice([]string{"3a", "3b"})}
	expect = []trans{
		g2pTransFromSlice([]string{"1a", "2a", "3a"}),
		g2pTransFromSlice([]string{"1a", "2a", "3b"}),
		g2pTransFromSlice([]string{"1a", "2b", "3a"}),
		g2pTransFromSlice([]string{"1a", "2b", "3b"}),
		g2pTransFromSlice([]string{"1b", "2a", "3a"}),
		g2pTransFromSlice([]string{"1b", "2a", "3b"}),
		g2pTransFromSlice([]string{"1b", "2b", "3a"}),
		g2pTransFromSlice([]string{"1b", "2b", "3b"}),
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"1a", "1b"}), g2pFromSlice([]string{"2a", "2b", "2c"}), g2pFromSlice([]string{"3a", "3b"})}
	expect = []trans{
		g2pTransFromSlice([]string{"1a", "2a", "3a"}),
		g2pTransFromSlice([]string{"1a", "2a", "3b"}),
		g2pTransFromSlice([]string{"1a", "2b", "3a"}),
		g2pTransFromSlice([]string{"1a", "2b", "3b"}),
		g2pTransFromSlice([]string{"1a", "2c", "3a"}),
		g2pTransFromSlice([]string{"1a", "2c", "3b"}),
		g2pTransFromSlice([]string{"1b", "2a", "3a"}),
		g2pTransFromSlice([]string{"1b", "2a", "3b"}),
		g2pTransFromSlice([]string{"1b", "2b", "3a"}),
		g2pTransFromSlice([]string{"1b", "2b", "3b"}),
		g2pTransFromSlice([]string{"1b", "2c", "3a"}),
		g2pTransFromSlice([]string{"1b", "2c", "3b"}),
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"b"}), g2pFromSlice([]string{"O"}), g2pFromSlice([]string{"rt", "r t"}), g2pFromSlice([]string{"a"}), g2pFromSlice([]string{"d"}), g2pFromSlice([]string{"u0"}), g2pFromSlice([]string{"S", "x"})}
	expect = []trans{
		g2pTransFromSlice([]string{"b", "O", "rt", "a", "d", "u0", "S"}),
		g2pTransFromSlice([]string{"b", "O", "rt", "a", "d", "u0", "x"}),
		g2pTransFromSlice([]string{"b", "O", "r, t", "a", "d", "u0", "S"}),
		g2pTransFromSlice([]string{"b", "O", "r, t", "a", "d", "u0", "x"}),
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %#v\nFound    %#v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"1"}), g2pFromSlice([]string{"2"}), g2pFromSlice([]string{"3a", "3b"}), g2pFromSlice([]string{"4"}), g2pFromSlice([]string{"5"}), g2pFromSlice([]string{"6"}), g2pFromSlice([]string{"7a", "7b"})}
	expect = []trans{
		g2pTransFromSlice([]string{"1", "2", "3a", "4", "5", "6", "7a"}),
		g2pTransFromSlice([]string{"1", "2", "3a", "4", "5", "6", "7b"}),
		g2pTransFromSlice([]string{"1", "2", "3b", "4", "5", "6", "7a"}),
		g2pTransFromSlice([]string{"1", "2", "3b", "4", "5", "6", "7b"}),
	}
	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

	//
	input = []g2p{g2pFromSlice([]string{"1"}), g2pFromSlice([]string{"2"}), g2pFromSlice([]string{"3a", "3b"}), g2pFromSlice([]string{"4"}), g2pFromSlice([]string{"5"}), g2pFromSlice([]string{"6"}), g2pFromSlice([]string{"7a", "7b"}), g2pFromSlice([]string{"8"})}
	expect = []trans{
		g2pTransFromSlice([]string{"1", "2", "3a", "4", "5", "6", "7a", "8"}),
		g2pTransFromSlice([]string{"1", "2", "3a", "4", "5", "6", "7b", "8"}),
		g2pTransFromSlice([]string{"1", "2", "3b", "4", "5", "6", "7a", "8"}),
		g2pTransFromSlice([]string{"1", "2", "3b", "4", "5", "6", "7b", "8"}),
	}

	result = rs.expand(input)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("\nExpected %v\nFound    %v", transes2string(expect), transes2string(result))
	}

}

func TestSyllabifySwsTestFile(t *testing.T) {
	rs, err := loadAndTest(t, "test_data/sws_test.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}

	var fsExpGot = "Input: %s; Expected: %v got: %v"

	syller := rs.Syllabifier

	//
	inputT := trans{
		phonemes: []g2p{
			{g: "b", p: []string{"b"}},
			{g: "o", p: []string{"O"}},
			{g: "rt", p: []string{"rt"}},
			{g: "a", p: []string{"a"}},
			{g: "d", p: []string{"d"}},
			{g: "u", p: []string{"u0"}},
			{g: "sch", p: []string{"S"}},
		},
	}

	inputS := inputT.string(" ")
	res0 := syller.syllabify(inputT)
	res := res0.string(rs.PhonemeDelimiter, rs.SyllableDelimiter)
	expect := "b O . rt a . d u0 S"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
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

func xxxTestBaq(t *testing.T) {
	_, err := loadAndTest(t, "server/g2p_files/basque_sampa.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}
func loadAndTestSyll(t *testing.T, fName string) (Syllabifier, error) {
	syller, err := LoadSyllFile(fName)
	if err != nil {
		return syller, fmt.Errorf("didn't expect error for input file %s : %s", fName, err)
	}

	result := syller.Test()
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
		fmt.Printf("%d OF %d TESTS FAILED FOR %s\n", len(result.FailedTests), len(syller.Tests), fName)
	} else {
		fmt.Printf("ALL %d TESTS PASSED FOR %s\n", len(syller.Tests), fName)
	}
	if len(result.Errors) > 0 || len(result.FailedTests) > 0 {
		return syller, fmt.Errorf("Init/tests failed for %s", fName)
	}
	return syller, nil
}

func TestSwsSyll(t *testing.T) {
	_, err := loadAndTestSyll(t, "test_data/sws_test_syll.g2p")
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestSwsSyllFail(t *testing.T) {
	_, err := loadAndTestSyll(t, "test_data/sws_test_syll_fail.g2p")
	if err == nil {
		t.Errorf("expected error here")
	}
}

func TestSwsFail(t *testing.T) {
	fName := "test_data/sws_test_fail.g2p"
	_, err := LoadFile(fName)
	if err == nil {
		t.Errorf("expected error here")
	}
}
