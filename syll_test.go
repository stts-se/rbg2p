package rbg2p

import (
	"reflect"
	"strings"
	"testing"
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
	syller := Syllabifier{SyllDef: def}

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
			"?",
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
	testMOPValidOnset(t, def, "?", true)
}

func TestSylledTransString(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	//
	input := sylledTrans{
		trans: trans{
			phonemes: []g2p{
				{g: "t", p: []string{"t"}},
				{g: "o", p: []string{"O"}},
				{g: "ff", p: []string{"f"}},
				{g: "e", p: []string{"@"}},
				{g: "l", p: []string{"l"}},
			},
		},
		boundaries: []boundary{
			{g: 2, p: 0},
		},
	}
	res := input.string(" ", ".")
	expect := "t O . f @ l"
	if res != expect {
		t.Errorf(fsExpGot, input, expect, res)
	}

	//
	input = sylledTrans{
		trans: trans{
			phonemes: []g2p{
				{g: "t", p: []string{"t"}},
				{g: "o", p: []string{"O"}},
				{g: "x", p: []string{"k", "s"}},
				{g: "e", p: []string{"@"}},
				{g: "l", p: []string{"l"}},
			},
		},
		boundaries: []boundary{
			{g: 2, p: 1},
		},
	}
	res = input.string(" ", ".")
	expect = "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, input, expect, res)
	}
}

func testSyllabify(t *testing.T, syller Syllabifier, input string, expect string) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"
	inputT := trans{}
	for _, p := range strings.Split(input, " ") {
		inputT.phonemes = append(inputT.phonemes, g2p{"", []string{p}})
	}
	resT := syller.syllabify(inputT)
	res := resT.string(" ", ".")
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
	syller := Syllabifier{SyllDef: def}

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
	syller := Syllabifier{SyllDef: def}

	//
	inputT := trans{
		phonemes: []g2p{
			{g: "t", p: []string{"t"}},
			{g: "o", p: []string{"O"}},
			{g: "x", p: []string{"k", "s"}},
			{g: "el", p: []string{"@", "l"}},
		},
	}

	inputS := inputT.string(" ")
	res0 := syller.syllabify(inputT)
	res := res0.string(" ", ".")
	expect := "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}

	//
	inputT = trans{
		phonemes: []g2p{
			{g: "t", p: []string{"t"}},
			{g: "o", p: []string{"O"}},
			{g: "x", p: []string{"k", "s"}},
			{g: "e", p: []string{"@"}},
			{g: "l", p: []string{"l"}},
		},
	}

	inputS = inputT.string(" ")
	res0 = syller.syllabify(inputT)
	res = res0.string(" ", ".")
	expect = "t O k . s @ l"
	if res != expect {
		t.Errorf(fsExpGot, inputS, expect, res)
	}

	//
	inputT = trans{
		phonemes: []g2p{
			{g: "t", p: []string{"t"}},
			{g: "u", p: []string{"u0"}},
			{g: "ng", p: []string{"N"}},
			{g: "a", p: []string{"a"}},
			{g: "n", p: []string{"n"}},
		},
	}

	inputS = inputT.string(" ")
	res0 = syller.syllabify(inputT)
	res = res0.string(" ", ".")
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
	syller := Syllabifier{SyllDef: def}

	//
	inputT := trans{
		phonemes: []g2p{
			{g: "b", p: []string{"b"}},
			{g: "a", p: []string{"a"}},
			{g: "rr", p: []string{"rr"}},
			{g: "t", p: []string{"t"}},
			{g: "r", p: []string{"r"}},
			{g: "ä", p: []string{"{:"}},
			{g: "d", p: []string{"d"}},
			{g: "e", p: []string{"@"}},
			{g: "n", p: []string{"n"}},
		},
	}

	inputS := inputT.string(" ")
	res0 := syller.syllabify(inputT)
	res := res0.string(" ", ".")
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
	syller := Syllabifier{SyllDef: def}

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
	res := res0.string(" ", ".")
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
	def, stressP, err := loadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{SyllDef: def, StressPlacement: stressP}

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
	def, stressP, err := loadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{SyllDef: def, StressPlacement: stressP}

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

func TestSwsInputWithStressPlacement_FirstInSyllable(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	lines := []string{"SYLLDEF TYPE MOP",
		`SYLLDEF ONSETS "p, b, t, rt, m, n, d, rd, k, g, rn, f, v, C, rs, r, l, s, x, S, h, rl, j, s, p, r, rs p r, s p l, rs p l, s p j, rs p j, s t r, rs rt r, s k r, rs k r, s k v, rs k v, p r, p j, p l, b r, b j, b l, t r, rt r, t v, rt v, d r, rd r, d v, rd v, k r, k l, k v, k n, g r, g l, g n, f r, f l, f j, f n, v r, s p, s t, s k, s v, s l, s m, s n, n j, rs p, rs rt, rs k, rs v, rs rl, rs m, rs rn, rn j, m j, rr"`,
		`SYLLDEF SYLLABIC "i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu"`,
		`SYLLDEF STRESS "\" \"\" %"`,
		`SYLLDEF DELIMITER "."`,
		`SYLLDEF STRESS_PLACEMENT FirstInSyllable`,
	}
	def, stressP, err := loadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{SyllDef: def, StressPlacement: stressP}

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
	expect = "\" f r a . g r a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f rr \" a g rr a n s I a", " ")
	expect = "\" f rr a g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

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

func TestSwsInputWithStressPlacement_AfterSyllabic(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	lines := []string{"SYLLDEF TYPE MOP",
		`SYLLDEF ONSETS "p, b, t, rt, m, n, d, rd, k, g, rn, f, v, C, rs, r, l, s, x, S, h, rl, j, s, p, r, rs p r, s p l, rs p l, s p j, rs p j, s t r, rs rt r, s k r, rs k r, s k v, rs k v, p r, p j, p l, b r, b j, b l, t r, rt r, t v, rt v, d r, rd r, d v, rd v, k r, k l, k v, k n, g r, g l, g n, f r, f l, f j, f n, v r, s p, s t, s k, s v, s l, s m, s n, n j, rs p, rs rt, rs k, rs v, rs rl, rs m, rs rn, rn j, m j, rr"`,
		`SYLLDEF SYLLABIC "i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu"`,
		`SYLLDEF STRESS "\" \"\" %"`,
		`SYLLDEF DELIMITER "."`,
		`SYLLDEF STRESS_PLACEMENT AfterSyllabic`,
	}
	def, stressP, err := loadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{SyllDef: def, StressPlacement: stressP}

	input := strings.Split("\" d u0 S a", " ")
	expect := "d u0 \" . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d", " ")
	expect = "p a . r A: \" d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d % g r e: n", " ")
	expect = "p a . r A: \" d . g r e: % n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f r \" a g r a n s I a", " ")
	expect = "f r a \" . g r a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f rr \" a g rr a n s I a", " ")
	expect = "f rr a \" g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("f rr a g rr a n s I a", " ")
	expect = "f rr a g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d \"\" g r e: n", " ")
	expect = "p a . r A: \" d . g r e: \"\" n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

}

func TestSwsInputWithStressPlacement_BeforeSyllabic(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	lines := []string{"SYLLDEF TYPE MOP",
		`SYLLDEF ONSETS "p, b, t, rt, m, n, d, rd, k, g, rn, f, v, C, rs, r, l, s, x, S, h, rl, j, s, p, r, rs p r, s p l, rs p l, s p j, rs p j, s t r, rs rt r, s k r, rs k r, s k v, rs k v, p r, p j, p l, b r, b j, b l, t r, rt r, t v, rt v, d r, rd r, d v, rd v, k r, k l, k v, k n, g r, g l, g n, f r, f l, f j, f n, v r, s p, s t, s k, s v, s l, s m, s n, n j, rs p, rs rt, rs k, rs v, rs rl, rs m, rs rn, rn j, m j, rr"`,
		`SYLLDEF SYLLABIC "i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu"`,
		`SYLLDEF STRESS "\" \"\" %"`,
		`SYLLDEF DELIMITER "."`,
		`SYLLDEF STRESS_PLACEMENT BeforeSyllabic`,
	}
	def, stressP, err := loadSyllDef(lines, " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	syll := Syllabifier{SyllDef: def, StressPlacement: stressP}

	input := strings.Split("\" d u0 S a", " ")
	expect := "d \" u0 . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d", " ")
	expect = "p a . r \" A: d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d % g r e: n", " ")
	expect = "p a . r \" A: d . g r % e: n"
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

	input = strings.Split("f rr a g rr a n s I a", " ")
	expect = "f rr a g . rr a n . s I . a"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a \" r A: d \"\" g r e: n", " ")
	expect = "p a . r \" A: d . g r \"\" e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("a \" r A: d \"\" g r e: n", " ")
	expect = "a . r \" A: d . g r \"\" e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

}
