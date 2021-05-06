package rbg2p

import (
	"fmt"
	"reflect"
	"regexp"
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

//автопило́т 'автопило́т', expected /V . f t 7 . p' I . " l o t/, got /V f . t 7 . p' I . " l o t/
//ка́льмар 'ка́льмар', expected /" k a l' . m 7 r/, got /" k a l' . . m 7 r/
//партсъе́зд 'партсъе́зд', expected /p V r ts . " j e s t/, got /p V r . " ts j e s t/

func dontTestMOPValidOnsetIssueMay2020(t *testing.T) {
	def := MOPSyllDef{
		Onsets: []string{
			"b", "b l", "b r", "b Z", "d", "d j", "d l", "d m", "d n", "d r", "d v", "d z", "dz", "dZ", "f", "f j", "f l", "f r", "f p", "f s", "f s k", "f s p", "f s t r", "f t", "f x", "f Z", "g", "g d", "g l", "g m", "g n", "g r", "g v", "g Z", "j", "k", "k l", "k n", "k p", "k r", "k r Z", "k s", "k s t", "k ts", "k v", "l", "l b", "l d", "l g", "l j", "l v", "l Z", "m", "m g", "m g l", "m k", "m r", "m S", "m s t", "m t s", "m ts", "m tS", "n", "n j", "p", "p l", "p j", "p n", "p r", "p r Z", "p s", "p S", "p s k", "p t", "p tS", "p x", "r", "s", "S", "s f", "s j", "S j", "s k", "S k", "s k l", "s k r", "s k v", "s l", "S l", "s m", "S m", "s n", "S n", "s p", "S p", "s p l", "s p r", "s r", "s t", "s ts", "S t", "s t r", "S t r", "s t v", "S ts", "s v", "S v", "s x", "t", "t k", "t l", "t m", "t p", "t r", "ts", "tS", "t s", "tS", "tS j", "tS k", "tS n", "ts j", "ts m", "ts v", "ts k", "tS m", "ts v", "t v", "v", "v d", "v g r", "v j", "v l", "v n", "v r", "v x", "v z", "v z l", "v z m", "v z v", "v Z", "x", "x l", "x m", "x n", "x r", "x v", "x p", "x t", "x k", "z", "Z", "z b", "z b r", "z d", "Z d", "z d r", "Z g", "Z g l", "Z j", "z l", "Z l", "z m", "Z m", "Z n", "z n", "z r", "z v", "b'", "b l'", "b r'", "b Z'", "d'", "d j'", "d l'", "d m'", "d n'", "d r'", "d v'", "d z'", "dz'", "dZ'", "f'", "f j'", "f l'", "f r'", "g'", "g d'", "g l'", "g m'", "g n'", "g r'", "g v'", "g Z'", "j'", "k'", "k l'", "k n'", "k p'", "k r'", "k r Z'", "k s'", "k s t'", "k ts'", "k v'", "l'", "l b'", "l d'", "l g'", "l j'", "l v'", "l Z'", "m'", "m g'", "m g l'", "m k'", "m r'", "m S'", "m s t'", "m t s'", "m ts'", "m tS'", "n'", "n j'", "p'", "p l'", "p j'", "p n'", "p r'", "p r Z'", "p s'", "p S'", "p s k'", "p t'", "p tS'", "p x'", "r'", "s'", "S'", "s f'", "s j'", "S j'", "s k'", "S k'", "s k l'", "s k r'", "s k v'", "s l'", "S l'", "s m'", "S m'", "s n'", "S n'", "s p'", "S p'", "s p l'", "s p r'", "s r'", "s t'", "s ts'", "S t'", "s t r'", "S t r'", "s t v'", "S ts'", "s v'", "S v'", "s x'", "t'", "t k'", "t l'", "t m'", "t p'", "t r'", "ts'", "tS'", "t s'", "tS'", "tS j'", "tS k'", "tS n'", "ts j'", "ts m'", "ts v'", "ts k'", "tS m'", "ts v'", "t v'", "v'", "v l'", "v n'", "v r'", "v x'", "v z'", "v z l'", "v z m'", "v z v'", "v Z'", "x'", "x l'", "x m'", "x n'", "x r'", "x v'", "z'", "Z'", "z b'", "z b r'", "z d'", "Z d'", "z d r'", "Z g'", "Z g l'", "Z j'", "z l'", "Z l'", "z m'", "Z m'", "Z n'", "z n'", "z r'", "z v'",
		},
		PhnDelim:  " ",
		SyllDelim: ".",
		Syllabic:  []string{"a", "e", "E", "i", "I", "o", "u", "@", "V", "1", "7"},
	}
	testMOPValidOnset(t, def, "f t", true)
	phnSet := PhonemeSet{
		Symbols:   []string{"a", "e", "E", "i", "I", "o", "u", "@", "V", "1", "7", "j", "r", "p", "b", "t", "d", "k", "g", "f", "v", "s", "z", "S", "Z", "x", "m", "n", "l", "p'", "b'", "t'", "d'", "k'", "g'", "f'", "v'", "s'", "z'", "S'", "x'", "m'", "n'", "l'", "r'", "tS", "dZ", "ts", "dz", "tS'", "dZ'", "\"", "%", ".", "-"},
		SyllDelim: Regexp{RE: regexp.MustCompile(" "), Source: " "},
	}
	syller := Syllabifier{
		SyllDef:         def,
		Tests:           []SyllTest{},
		StressPlacement: FirstInSyllable,
		PhonemeSet:      phnSet,
	}

	myExpGotFmt := "For input <%v>, expected <%v>, got <%v>"

	//
	input := "V f t 7 p' I l \" o t"
	expect := "V . f t 7 . p' I . \" l o t"
	res, err := syller.SyllabifyFromString(input)
	if err != nil {
		t.Errorf("Got error from SyllabifyFromString: %v", err)
	}
	if res != expect {
		fmt.Printf("%#v\n", res)
		fmt.Printf("%#v\n", expect)
		t.Errorf(myExpGotFmt, input, expect, res)
	}
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

	input := strings.Split("d \" u0 S a", " ")
	expect := "\" d u0 . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a r \" A: d", " ")
	expect = "p a . \" r A: d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a r \" A: d g r % e: n", " ")
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

func TestStressPlacements(t *testing.T) {
	var fsExpGot = "Input: %s; Expected: %v got: %v"

	var err error
	var def SyllDef
	var stressP StressPlacement
	var expect, result string
	var input []string

	var baseLines = []string{"SYLLDEF TYPE MOP",
		`SYLLDEF ONSETS "r, t, p, s, d, f, g, h, j, k, l, v, b, n, m, p r"`,
		`SYLLDEF SYLLABIC "a o u e i"`,
		`SYLLDEF STRESS "1"`,
		`SYLLDEF DELIMITER "."`,
	}
	def, stressP, err = loadSyllDef(append(baseLines, "SYLLDEF STRESS_PLACEMENT AfterSyllabic"), " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	syllAfterSyllabic := Syllabifier{SyllDef: def, StressPlacement: stressP}

	def, stressP, err = loadSyllDef(append(baseLines, "SYLLDEF STRESS_PLACEMENT BeforeSyllabic"), " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	syllBeforeSyllabic := Syllabifier{SyllDef: def, StressPlacement: stressP}

	def, stressP, err = loadSyllDef(append(baseLines, "SYLLDEF STRESS_PLACEMENT FirstInSyllable"), " ")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	syllFirstInSyllable := Syllabifier{SyllDef: def, StressPlacement: stressP}

	//

	input = strings.Split("d u 1 k a", " ")
	expect = "d u 1 . k a"
	result = syllAfterSyllabic.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}
	input = strings.Split("p a r a 1 d", " ")
	expect = "p a . r a 1 d"
	result = syllAfterSyllabic.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}

	//
	input = strings.Split("d 1 u k a", " ")
	expect = "d 1 u . k a"
	result = syllBeforeSyllabic.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}
	input = strings.Split("p a r 1 a d", " ")
	expect = "p a . r 1 a d"
	result = syllBeforeSyllabic.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}

	//
	input = strings.Split("d 1 u k a", " ")
	expect = "1 d u . k a"
	result = syllFirstInSyllable.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}
	input = strings.Split("p a r 1 a d", " ")
	expect = "p a . 1 r a d"
	result = syllFirstInSyllable.SyllabifyFromPhonemes(input)
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

	input := strings.Split("d u0 \" S a", " ")
	expect := "d u0 \" . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)
	}

	input = strings.Split("p a r A: \" d", " ")
	expect = "p a . r A: \" d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a r A: \" d g r e: % n", " ")
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

	input = strings.Split("p a r A: \" d g r e: \"\" n", " ")
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

	input := strings.Split("d \" u0 S a", " ")
	expect := "d \" u0 . S a"
	result := syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a r \" A: d", " ")
	expect = "p a . r \" A: d"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("p a r \" A: d g r % e: n", " ")
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

	input = strings.Split("p a r \" A: d g r \"\" e: n", " ")
	expect = "p a . r \" A: d . g r \"\" e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

	input = strings.Split("a r \" A: d g r \"\" e: n", " ")
	expect = "a . r \" A: d . g r \"\" e: n"
	result = syll.SyllabifyFromPhonemes(input)
	if result != expect {
		t.Errorf(fsExpGot, input, expect, result)

	}

}
