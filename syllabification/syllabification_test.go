package syllabification

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stts-se/rbg2p"
)

func testMOPValidSplit(t *testing.T, syller Syllabifier, left string, right string, expect bool) {
	var fsExpGot = "/%s - %s/. Expected: %v got: %v"
	res := syller.SyllDef.ValidSplit(strings.Split(left, " "), strings.Split(right, " "))
	if res != expect {
		t.Errorf(fsExpGot, left, right, expect, res)
	}
}

func TestMOPValidSplit(t *testing.T) {
	def := MOPSyllDef{
		Onsets: []string{
			"p",
			"t",
			"k",
			"r",
			"p r",
			"p r O",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"O"},
	}
	syller := Syllabifier{def}

	testMOPValidSplit(t, syller, "p", "k", true)

	testMOPValidSplit(t, syller, "p", "r", false)

	testMOPValidSplit(t, syller, "p", "r O", false)

	testMOPValidSplit(t, syller, "k", "p r O", true)
	testMOPValidSplit(t, syller, "k", "p r A", false) // since /A/ is not defined as syllabic here
}

func testMOPValidOnset(t *testing.T, def MOPSyllDef, onset string, expect bool) {
	var fsExpGot = "/%s/. Expected: %v got: %v"
	res := def.validOnset(onset)
	if res != expect {
		t.Errorf(fsExpGot, onset, expect, res)
	}
}

func TestMOPValidOnset(t *testing.T) {
	def := MOPSyllDef{
		Onsets: []string{
			"p",
			"t",
			"k",
			"r",
			"p r",
			"p r O",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"O"},
	}
	testMOPValidOnset(t, def, "p r", true)
	testMOPValidOnset(t, def, "", true)
}

func TestSylledTransString(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	//
	input := SylledTrans{
		Trans: rbg2p.Trans{
			Phonemes: []rbg2p.G2P{
				rbg2p.G2P{G: "t", P: []string{"t"}},
				rbg2p.G2P{G: "o", P: []string{"O"}},
				rbg2p.G2P{G: "ff", P: []string{"f"}},
				rbg2p.G2P{G: "e", P: []string{"@"}},
				rbg2p.G2P{G: "l", P: []string{"l"}},
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
			Phonemes: []rbg2p.G2P{
				rbg2p.G2P{G: "t", P: []string{"t"}},
				rbg2p.G2P{G: "o", P: []string{"O"}},
				rbg2p.G2P{G: "x", P: []string{"k", "s"}},
				rbg2p.G2P{G: "e", P: []string{"@"}},
				rbg2p.G2P{G: "l", P: []string{"l"}},
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
		Onsets: []string{
			"p",
			"t",
			"k",
			"r",
			"p r",
			"p r O",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"O", "e", "a"},
	}
	syller := Syllabifier{def}

	testSyllabify(t, syller, "k O t e", "k O . t e")
	testSyllabify(t, syller, "k O t e p r O g r a m", "k O . t e . p r O g . r a m")
}

func TestSyllabify2(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	def := MOPSyllDef{
		Onsets: []string{
			"p",
			"t",
			"k",
			"s",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"O", "e", "a", "@", "u0"},
	}
	syller := Syllabifier{def}

	//
	inputT := rbg2p.Trans{
		Phonemes: []rbg2p.G2P{
			rbg2p.G2P{G: "t", P: []string{"t"}},
			rbg2p.G2P{G: "o", P: []string{"O"}},
			rbg2p.G2P{G: "x", P: []string{"k", "s"}},
			rbg2p.G2P{G: "el", P: []string{"@", "l"}},
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
		Phonemes: []rbg2p.G2P{
			rbg2p.G2P{G: "t", P: []string{"t"}},
			rbg2p.G2P{G: "o", P: []string{"O"}},
			rbg2p.G2P{G: "x", P: []string{"k", "s"}},
			rbg2p.G2P{G: "e", P: []string{"@"}},
			rbg2p.G2P{G: "l", P: []string{"l"}},
		},
	}

	inputS = inputT.String(" ")
	res0 = syller.Syllabify(inputT)
	res = res0.String(" ", ".")
	expect = "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}

	//
	inputT = rbg2p.Trans{
		Phonemes: []rbg2p.G2P{
			rbg2p.G2P{G: "t", P: []string{"t"}},
			rbg2p.G2P{G: "u", P: []string{"u0"}},
			rbg2p.G2P{G: "ng", P: []string{"N"}},
			rbg2p.G2P{G: "a", P: []string{"a"}},
			rbg2p.G2P{G: "n", P: []string{"n"}},
		},
	}

	inputS = inputT.String(" ")
	res0 = syller.Syllabify(inputT)
	res = res0.String(" ", ".")
	expect = "t u0 N . a n"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}
}

func TestSyllabify3(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	def := MOPSyllDef{
		Onsets: []string{
			"b",
			"d",
			"t r",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"a", "@", "{:"},
	}
	syller := Syllabifier{def}

	//
	inputT := rbg2p.Trans{
		Phonemes: []rbg2p.G2P{
			rbg2p.G2P{G: "b", P: []string{"b"}},
			rbg2p.G2P{G: "a", P: []string{"a"}},
			rbg2p.G2P{G: "rr", P: []string{"rr"}},
			rbg2p.G2P{G: "t", P: []string{"t"}},
			rbg2p.G2P{G: "r", P: []string{"r"}},
			rbg2p.G2P{G: "ä", P: []string{"{:"}},
			rbg2p.G2P{G: "d", P: []string{"d"}},
			rbg2p.G2P{G: "e", P: []string{"@"}},
			rbg2p.G2P{G: "n", P: []string{"n"}},
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

func TestSyllabify4(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	def := MOPSyllDef{
		Onsets:    strings.Split("p b t rt m n d rd k g N rn f v C rs r l s x S h rl j", " "),
		Syllabic:  strings.Split("i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu", " "),
		PhnDelim:  " ",
		SyllDelim: ".",
	}
	syller := Syllabifier{def}

	//
	inputT := rbg2p.Trans{
		Phonemes: []rbg2p.G2P{
			rbg2p.G2P{G: "b", P: []string{"b"}},
			rbg2p.G2P{G: "o", P: []string{"O"}},
			rbg2p.G2P{G: "rt", P: []string{"rt"}},
			rbg2p.G2P{G: "a", P: []string{"a"}},
			rbg2p.G2P{G: "d", P: []string{"d"}},
			rbg2p.G2P{G: "u", P: []string{"u0"}},
			rbg2p.G2P{G: "sch", P: []string{"S"}},
		},
	}

	inputS := inputT.String(" ")
	res0 := syller.Syllabify(inputT)
	res := res0.String(" ", ".")
	expect := "b O . rt a . d u0 S"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}
}
