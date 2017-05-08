package rbg2p

import "strings"

// sBound represent syllable boundaries
type sBound struct {
	g int
	p int
}

// SylledTrans is a syllabified transcription (containing a Trans instance and a slice of indices for syllable boundaries)
type SylledTrans struct {
	Trans      Trans
	boundaries []sBound
}

func (t SylledTrans) isBoundary(b sBound) bool {
	for _, bound := range t.boundaries {
		if bound == b {
			return true
		}
	}
	return false
}

// String returns a string representation of the SylledTrans, given the specified delimiters for phonemes and syllables
func (t SylledTrans) String(phnDelimiter string, syllDelimiter string) string {
	res := []string{}
	for gi, g2p := range t.Trans.Phonemes {
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

//ListPhonemes returns a slice of phonemes as strings
func (t SylledTrans) ListPhonemes() []string {
	return t.Trans.ListPhonemes()
}

// SyllDef is an interface for implementing custom made syllabification strategies
type SyllDef interface {
	validSplit(left []string, right []string) bool
}

// MOPSyllDef is a Maximum Onset Principle implementation of the SyllDef interface
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

// Syllabifier is a module to divide a transcription into syllables
type Syllabifier struct {
	SyllDef SyllDef
}

// Syllabify is used to divide a transcription into syllables
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
