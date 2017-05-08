package rbg2p

import (
	"reflect"
	"strings"
	"testing"
)

func testMOPValidSplit(t *testing.T, syller Syllabifier, left string, right string, expect bool) {
	var fsExpGot = "/%s - %s/. Expected: %v got: %v"
	res := syller.SyllDef.validSplit(strings.Split(left, " "), strings.Split(right, " "))
	if res != expect {
		t.Errorf(fsExpGot, left, right, expect, res)
	}
}

func TestMOPValidSplit(t *testing.T) {
	def := MOPSyllDef{
		onsets: []string{
			"p",
			"t",
			"k",
			"r",
			"p r",
			"p r O",
		},
		phonemeDelimiter: " ",
		syllabic:         []string{"O"},
	}
	syller := Syllabifier{def}

	testMOPValidSplit(t, syller, "p", "k", true)

	testMOPValidSplit(t, syller, "p", "r", false)

	testMOPValidSplit(t, syller, "p", "r O", false)

	testMOPValidSplit(t, syller, "k", "p r O", true)
	testMOPValidSplit(t, syller, "k", "p r A", false) // since /A/ is not defined as syllabic here
}

func TestSylledTransString(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	//
	input := SylledTrans{
		Trans: Trans{
			[]g2p{
				g2p{"t", []string{"t"}},
				g2p{"o", []string{"O"}},
				g2p{"ff", []string{"f"}},
				g2p{"e", []string{"@"}},
				g2p{"l", []string{"l"}},
			},
		},
		boundaries: []sBound{
			sBound{g: 2, p: 0},
		},
	}
	res := input.String(" ", ".")
	expect := "t O . f @ l"
	if res != expect {
		t.Errorf(fsExpGot, input, expect, res)
	}

	//
	input = SylledTrans{
		Trans: Trans{
			[]g2p{
				g2p{"t", []string{"t"}},
				g2p{"o", []string{"O"}},
				g2p{"x", []string{"k", "s"}},
				g2p{"e", []string{"@"}},
				g2p{"l", []string{"l"}},
			},
		},
		boundaries: []sBound{
			sBound{g: 2, p: 1},
		},
	}
	res = input.String(" ", ".")
	expect = "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, input, expect, res)
	}
}

func testSyllabify(t *testing.T, syller Syllabifier, input string, expect string) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"
	inputT := Trans{}
	for _, p := range strings.Split(input, " ") {
		inputT.Phonemes = append(inputT.Phonemes, g2p{"", []string{p}})
	}
	resT := syller.Syllabify(inputT)
	res := resT.String(" ", ".")
	if !reflect.DeepEqual(res, expect) {
		t.Errorf(fsExpGot, input, expect, res)
	}
}

func TestSyllabify1(t *testing.T) {
	def := MOPSyllDef{
		onsets: []string{
			"p",
			"t",
			"k",
			"r",
			"p r",
			"p r O",
		},
		phonemeDelimiter: " ",
		syllabic:         []string{"O", "e", "a"},
	}
	syller := Syllabifier{def}

	testSyllabify(t, syller, "k O t e", "k O . t e")
	testSyllabify(t, syller, "k O t e p r O g r a m", "k O . t e . p r O g . r a m")
}

func TestSyllabify2(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	def := MOPSyllDef{
		onsets: []string{
			"p",
			"t",
			"k",
			"s",
		},
		phonemeDelimiter: " ",
		syllabic:         []string{"O", "e", "a", "@"},
	}
	syller := Syllabifier{def}

	//
	inputT := Trans{
		[]g2p{
			g2p{"t", []string{"t"}},
			g2p{"o", []string{"O"}},
			g2p{"x", []string{"k", "s"}},
			g2p{"el", []string{"@", "l"}},
		},
	}

	inputS := inputT.String(" ")
	res0 := syller.Syllabify(inputT)
	res := res0.String(" ", ".")
	expect := "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}

	//
	inputT = Trans{
		[]g2p{
			g2p{"t", []string{"t"}},
			g2p{"o", []string{"O"}},
			g2p{"x", []string{"k", "s"}},
			g2p{"e", []string{"@"}},
			g2p{"l", []string{"l"}},
		},
	}

	inputS = inputT.String(" ")
	res0 = syller.Syllabify(inputT)
	res = res0.String(" ", ".")
	expect = "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}
}

func TestSyllabify3(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	def := MOPSyllDef{
		onsets: []string{
			"b",
			"d",
			"t r",
		},
		phonemeDelimiter: " ",
		syllabic:         []string{"a", "@", "{:"},
	}
	syller := Syllabifier{def}

	//
	inputT := Trans{
		[]g2p{
			g2p{"b", []string{"b"}},
			g2p{"a", []string{"a"}},
			g2p{"rr", []string{"rr"}},
			g2p{"t", []string{"t"}},
			g2p{"r", []string{"r"}},
			g2p{"ä", []string{"{:"}},
			g2p{"d", []string{"d"}},
			g2p{"e", []string{"@"}},
			g2p{"n", []string{"n"}},
		},
	}

	inputS := inputT.String(" ")
	res0 := syller.Syllabify(inputT)
	res := res0.String(" ", ".")
	expect := "b a rr . t r {: . d @ n"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}
}
