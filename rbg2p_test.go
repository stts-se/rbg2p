package rbg2p

import (
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
		} else if !expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newTest(l)
		if err != nil {
			t.Errorf("didn't expect error for input test line %s : %s", l, err)
		} else if expect.Equals(result) {
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
			RightContext: Context{regexp.MustCompile("$")}},
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
			RightContext: Context{regexp.MustCompile("[dgjlvbnm]")}},
		"a -> A / _ VOICED #": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{regexp.MustCompile("[dgjlvbnm]$")}},
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
		} else if !expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newRule(l, vars)
		if err != nil {
			t.Errorf("didn't expect error for input rule line %s : %s", l, err)
		} else if expect.Equals(result) {
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

func TestApply(t *testing.T) {
	fName := "test_data/test.g2p"
	rs, err := LoadFile(fName)
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
	}
	_, err = rs.Apply("hix")
	if err == nil {
		t.Errorf("expected error for input file %s : %s", fName)
	}

	_, err = rs.Apply("hit")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
	}

	_, err = rs.Apply("dusch")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
	}

	_, err = rs.Apply("duscha")
	if err != nil {
		t.Errorf("didn't expect error for input file %s : %s", fName, err)
	}

	// errors := rs.Test()
	// if len(errors) > 0 {
	// 	fmt.Printf("%d OF %d TESTS FAILED:\n", len(errors), len(rs.Tests))
	// 	for _, err = range errors {
	// 		fmt.Printf("%v\n", err)
	// 	}
	// } else {
	// 	fmt.Printf("ALL %d TESTS PASSED\n", len(rs.Tests))
	// }
}
