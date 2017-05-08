package syllabification

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stts-se/rbg2p"
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
		Trans: rbg2p.Trans{
			[]rbg2p.G2P{
				rbg2p.G2P{"t", []string{"t"}},
				rbg2p.G2P{"o", []string{"O"}},
				rbg2p.G2P{"ff", []string{"f"}},
				rbg2p.G2P{"e", []string{"@"}},
				rbg2p.G2P{"l", []string{"l"}},
			},
		},
		Boundaries: []Boundary{
			Boundary{G: 2, P: 0},
		},
	}
	res := input.String(" ", ".")
	expect := "t O . f @ l"
	if res != expect {
		t.Errorf(fsExpGot, input, expect, res)
	}

	//
	input = SylledTrans{
		Trans: rbg2p.Trans{
			[]rbg2p.G2P{
				rbg2p.G2P{"t", []string{"t"}},
				rbg2p.G2P{"o", []string{"O"}},
				rbg2p.G2P{"x", []string{"k", "s"}},
				rbg2p.G2P{"e", []string{"@"}},
				rbg2p.G2P{"l", []string{"l"}},
			},
		},
		Boundaries: []Boundary{
			Boundary{G: 2, P: 1},
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
	inputT := rbg2p.Trans{}
	for _, p := range strings.Split(input, " ") {
		inputT.Phonemes = append(inputT.Phonemes, rbg2p.G2P{"", []string{p}})
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
	inputT := rbg2p.Trans{
		[]rbg2p.G2P{
			rbg2p.G2P{"t", []string{"t"}},
			rbg2p.G2P{"o", []string{"O"}},
			rbg2p.G2P{"x", []string{"k", "s"}},
			rbg2p.G2P{"el", []string{"@", "l"}},
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
	inputT = rbg2p.Trans{
		[]rbg2p.G2P{
			rbg2p.G2P{"t", []string{"t"}},
			rbg2p.G2P{"o", []string{"O"}},
			rbg2p.G2P{"x", []string{"k", "s"}},
			rbg2p.G2P{"e", []string{"@"}},
			rbg2p.G2P{"l", []string{"l"}},
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
	inputT := rbg2p.Trans{
		[]rbg2p.G2P{
			rbg2p.G2P{"b", []string{"b"}},
			rbg2p.G2P{"a", []string{"a"}},
			rbg2p.G2P{"rr", []string{"rr"}},
			rbg2p.G2P{"t", []string{"t"}},
			rbg2p.G2P{"r", []string{"r"}},
			rbg2p.G2P{"ä", []string{"{:"}},
			rbg2p.G2P{"d", []string{"d"}},
			rbg2p.G2P{"e", []string{"@"}},
			rbg2p.G2P{"n", []string{"n"}},
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
