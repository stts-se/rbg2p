package syllabification

import (
	"fmt"
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

func TestBaqSyller(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	lines := []string{"SYLLDEF TYPE MOP",
		"SYLLDEF ONSETS \"p, b, t, d, k, g, c, gj, ts, ts`, tS, f, s, s`, S, x, jj, m, n, J, l, L, r, rr, j, w, B, D, G, T\"",
		`SYLLDEF SYLLABIC "a e i o u"`,
		`SYLLDEF STRESS "\" %"`,
		`SYLLDEF DELIMITER "."`,
	}
	def, err := LoadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{def}

	input := strings.Split("f rr a g rr \" a n s i a", " ")
	expect := "f rr a g . rr \" a n . s i . a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f rr a g rr a n s i a", " ")
	expect = "f rr a g . rr a n . s i . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

}

func TestSwsInputWithStress(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	lines := []string{"SYLLDEF TYPE MOP",
		`SYLLDEF ONSETS "p, b, t, rt, m, n, d, rd, k, g, rn, f, v, C, rs, r, l, s, x, S, h, rl, j, s, p, r, rs p r, s p l, rs p l, s p j, rs p j, s t r, rs rt r, s k r, rs k r, s k v, rs k v, p r, p j, p l, b r, b j, b l, t r, rt r, t v, rt v, d r, rd r, d v, rd v, k r, k l, k v, k n, g r, g l, g n, f r, f l, f j, f n, v r, s p, s t, s k, s v, s l, s m, s n, n j, rs p, rs rt, rs k, rs v, rs rl, rs m, rs rn, rn j, m j, rr"`,
		`SYLLDEF SYLLABIC "i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu"`,
		`SYLLDEF STRESS "\" \"\" %"`,
		`SYLLDEF DELIMITER "."`,
	}
	def, err := LoadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{def}

	input := strings.Split("\" d u0 S a", " ")
	expect := "\" d u0 . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d", " ")
	expect = "p a . \" r A: d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d % g r e: n", " ")
	expect = "p a . \" r A: d . % g r e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f r \" a g r a n s I a", " ")
	expect = "f r \" a . g r a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f rr \" a g rr a n s I a", " ")
	expect = "f rr \" a g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	fmt.Println("\n\n")

	input = strings.Split("f rr a g rr a n s I a", " ")
	expect = "f rr a g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d \"\" g r e: n", " ")
	expect = "p a . \" r A: d . \"\" g r e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

}
