package syllabification

import (
	"strings"

	"github.com/stts-se/rbg2p"
)

// Boundary represents syllable boundaries. Primarily for package internal use.
type Boundary struct {
	G int
	P int
}

// SylledTrans is a syllabified transcription (containing a Trans instance and a slice of indices for syllable boundaries)
type SylledTrans struct {
	Trans      rbg2p.Trans
	Boundaries []Boundary
}

func (t SylledTrans) isBoundary(b Boundary) bool {
	for _, bound := range t.Boundaries {
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
		for pi, p := range g2p.P {
			index := Boundary{G: gi, P: pi}
			if t.isBoundary(index) {
				res = append(res, syllDelimiter)
			}
			if len(p) > 0 {
				res = append(res, p)
			}
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
	ValidSplit(left []string, right []string) bool
	ContainsSyllabic(phonemes []string) bool
	IsDefined() bool
	PhonemeDelimiter() string
	SyllableDelimiter() string
}

// MOPSyllDef is a Maximum Onset Principle implementation of the SyllDef interface
type MOPSyllDef struct {
	Onsets    []string
	Syllabic  []string
	PhnDelim  string
	SyllDelim string
}

// PhonemeDelimiter is the string used to separate phonemes (required by interface)
func (def MOPSyllDef) PhonemeDelimiter() string {
	return def.PhnDelim
}

// SyllableDelimiter is the string used to separate syllables (required by interface)
func (def MOPSyllDef) SyllableDelimiter() string {
	return def.SyllDelim
}

// IsDefined is used to determine if there is a syllabifier defined or not (required by interface)
func (def MOPSyllDef) IsDefined() bool {
	return len(def.Onsets) > 0
}

func (def MOPSyllDef) isSyllabic(phoneme string) bool {
	for _, s := range def.Syllabic {
		if s == phoneme {
			return true
		}
	}
	return false
}

// ContainsSyllabic tells if the input phoneme slice contains any syllabic phonemes (required by interface)
func (def MOPSyllDef) ContainsSyllabic(phonemes []string) bool {
	for _, p := range phonemes {
		if def.isSyllabic(p) {
			return true
		}
	}
	return false

}

func (def MOPSyllDef) validOnset(onset string) bool {
	if len(onset) == 0 {
		return true
	}
	for _, s := range def.Onsets {
		if s == onset {
			return true
		}
	}
	return false
}

// ValidSplit is called by Syllabifier.Syllabify to test where to put the boundaries
func (def MOPSyllDef) ValidSplit(left []string, right []string) bool {
	onset := []string{}
	for i := 0; i < len(right) && !def.isSyllabic(right[i]); i++ {
		onset = append(onset, right[i])
	}
	if !def.validOnset(strings.Join(onset, def.PhonemeDelimiter())) {
		return false
	}
	test := onset
	for i := len(left) - 1; i >= 0 && !def.isSyllabic(left[i]); i-- {
		test = append([]string{left[i]}, test...)
		if def.validOnset(strings.Join(test, def.PhonemeDelimiter())) {
			return false
		}
	}
	return true
}

// Syllabifier is a module to divide a transcription into syllables
type Syllabifier struct {
	SyllDef SyllDef
}

// IsDefined is used to determine if there is a syllabifier defined or not
func (s Syllabifier) IsDefined() bool {
	return s.SyllDef.IsDefined()
}

// Syllabify is used to divide a transcription into syllables
func (s Syllabifier) Syllabify(t rbg2p.Trans) SylledTrans {
	res := SylledTrans{Trans: t}
	left := []string{}
	right := t.ListPhonemes()
	for gi, g2p := range t.Phonemes {
		for pi, p := range g2p.P {
			if len(left) > 0 && s.SyllDef.ValidSplit(left, right) && s.SyllDef.ContainsSyllabic(left) && s.SyllDef.ContainsSyllabic(right) {
				index := Boundary{G: gi, P: pi}
				res.Boundaries = append(res.Boundaries, index)
			}
			//fmt.Printf("Syllabify.debug\t%s %s %v %v\n", left, right, s.SyllDef.ValidSplit(left, right), res.Boundaries)
			left = append(left, p)
			right = right[1:len(right)]
		}
	}
	return res
}
