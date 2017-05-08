package rbg2p

import "strings"

// sBound represent syllable boundaries
type sBound struct {
	g int
	p int
}

type SylledTrans struct {
	Trans      Trans
	boundaries []sBound
}

func (t SylledTrans) Phonemes() []g2p {
	return t.Trans.Phonemes
}

func (t SylledTrans) isBoundary(b sBound) bool {
	for _, bound := range t.boundaries {
		if bound == b {
			return true
		}
	}
	return false
}

func (t SylledTrans) String(phnDelimiter string, syllDelimiter string) string {
	res := []string{}
	for gi, g2p := range t.Phonemes() {
		for pi, p := range g2p.p {
			index := sBound{g: gi, p: pi}
			if t.isBoundary(index) {
				res = append(res, syllDelimiter)
			}
			res = append(res, p)
		}
	}
	return strings.Join(res, phnDelimiter)
}

func (t SylledTrans) ListPhonemes() []string {
	return t.Trans.ListPhonemes()
}

type SyllDef interface {
	validSplit(left []string, right []string) bool
}

type MOPSyllDef struct {
	onsets           []string
	syllabic         []string
	phonemeDelimiter string
}

func (def MOPSyllDef) isSyllabic(phoneme string) bool {
	for _, s := range def.syllabic {
		if s == phoneme {
			return true
		}
	}
	return false
}

func (def MOPSyllDef) validOnset(onset string) bool {
	for _, s := range def.onsets {
		if s == onset {
			return true
		}
	}
	return false
}

func (def MOPSyllDef) validSplit(left []string, right []string) bool {
	onset := []string{}
	for i := 0; i < len(right) && !def.isSyllabic(right[i]); i++ {
		onset = append(onset, right[i])
	}
	if !def.validOnset(strings.Join(onset, def.phonemeDelimiter)) {
		return false
	}
	test := onset
	for i := len(left) - 1; i >= 0 && !def.isSyllabic(left[i]); i-- {
		test = append([]string{left[i]}, test...)
		if def.validOnset(strings.Join(test, def.phonemeDelimiter)) {
			return false
		}
	}
	return true
}

type Syllabifier struct {
	SyllDef SyllDef
}

func (s Syllabifier) Syllabify(t Trans) SylledTrans {
	res := SylledTrans{Trans: t}
	left := []string{}
	right := t.ListPhonemes()
	for gi, g2p := range t.Phonemes {
		for pi, p := range g2p.p {
			//fmt.Printf("%s %s %v\n", left, right, s.SyllDef.validSplit(left, right))
			if len(left) > 0 && s.SyllDef.validSplit(left, right) {
				index := sBound{g: gi, p: pi}
				res.boundaries = append(res.boundaries, index)
			}
			left = append(left, p)
			right = right[1:len(right)]
		}
	}
	return res
}
