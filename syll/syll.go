package syll

import (
	"fmt"
	"strings"

	"github.com/stts-se/rbg2p/util"
)

// boundary represents syllable boundaries. Primarily for package internal use.
type boundary struct {
	G int
	P int
}

// sylledTrans is a syllabified transcription (containing a Trans instance and a slice of indices for syllable boundaries)
type sylledTrans struct {
	Trans      util.Trans
	boundaries []boundary
	Stress     []string
}

func (t sylledTrans) isboundary(b boundary) bool {
	for _, bound := range t.boundaries {
		if bound == b {
			return true
		}
	}
	return false
}

// String returns a string representation of the sylledTrans, given the specified delimiters for phonemes and syllables
func (t sylledTrans) String(phnDelimiter string, syllDelimiter string) string {
	res := []string{}
	for gi, g2p := range t.Trans.Phonemes {
		for pi, p := range g2p.P {
			index := boundary{G: gi, P: pi}
			if t.isboundary(index) {
				res = append(res, syllDelimiter)
			}
			if len(p) > 0 {
				res = append(res, p)
			}
		}
	}
	return strings.Join(res, phnDelimiter)
}

// Syllables returns a slice of syllables consisting of (a slice of) phonemes
func (t sylledTrans) syllables() [][]string {
	res := [][]string{}
	thisSyllable := []string{}
	for gi, g2p := range t.Trans.Phonemes {
		for pi, p := range g2p.P {
			index := boundary{G: gi, P: pi}
			if t.isboundary(index) {
				res = append(res, thisSyllable)
				thisSyllable = []string{}
			}
			if len(p) > 0 {
				thisSyllable = append(thisSyllable, p)
			}
		}
	}
	res = append(res, thisSyllable)
	return res
}

//ListPhonemes returns a slice of phonemes as strings
func (t sylledTrans) ListPhonemes() []string {
	return t.Trans.ListPhonemes()
}

// SyllDef is an interface for implementing custom made syllabification strategies
type SyllDef interface {
	ValidSplit(left []string, right []string) bool
	ContainsSyllabic(phonemes []string) bool
	IsDefined() bool
	IsStress(symbol string) bool
	IsSyllabic(symbol string) bool
	PhonemeDelimiter() string
	SyllableDelimiter() string
}

// MOPSyllDef is a Maximum Onset Principle implementation of the SyllDef interface
type MOPSyllDef struct {
	Onsets    []string
	Syllabic  []string
	PhnDelim  string
	SyllDelim string
	Stress    []string
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

func (def MOPSyllDef) IsStress(symbol string) bool {
	for _, s := range def.Stress {
		if s == symbol {
			return true
		}
	}
	return false
}

func (def MOPSyllDef) IsSyllabic(phoneme string) bool {
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
		if def.IsSyllabic(p) {
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

	if len(left) > 0 && def.IsStress(left[len(left)-1]) {
		return false
	}

	if len(right) > 0 && def.IsStress(right[0]) {
		right = right[1:len(right)]
	}

	onset := []string{}
	for i := 0; i < len(right) && !def.IsSyllabic(right[i]); i++ {
		if !def.IsStress(right[i]) {
			onset = append(onset, right[i])
		}
	}
	if !def.validOnset(strings.Join(onset, def.PhonemeDelimiter())) {
		return false
	}
	test := onset
	for i := len(left) - 1; i >= 0 && !def.IsSyllabic(left[i]); i-- {
		test = append([]string{left[i]}, test...)
		if def.validOnset(strings.Join(test, def.PhonemeDelimiter())) {
			return false
		}
	}
	return true
}

// Test defines a rule test (input -> output)
type Test struct {
	Input  string
	Output string
}

// Syllabifier is a module to divide a transcription into syllables
type Syllabifier struct {
	SyllDef         SyllDef
	Tests           []Test
	StressPlacement StressPlacement
}

// IsDefined is used to determine if there is a syllabifier defined or not
func (s Syllabifier) IsDefined() bool {
	return s.SyllDef.IsDefined()
}

// SyllabifyFromPhonemes is used to divide a range of phonemes into syllables and create an output string
func (s Syllabifier) SyllabifyFromPhonemes(phns []string) string {
	t := util.Trans{}
	for _, phn := range phns {
		t.Phonemes = append(t.Phonemes, util.G2P{G: "", P: []string{phn}})
	}
	return s.SyllabifyToString(t)
}

// SyllabifyFromStromg is used to divide a transcription string into syllables and create an output string
func (s Syllabifier) SyllabifyFromString(phnSet util.PhonemeSet, trans string) (string, error) {
	phns, err := phnSet.SplitTranscription(trans)
	if err != nil {
		return "", err
	}
	return s.SyllabifyFromPhonemes(phns), nil
}

// SyllabifyToString is used to divide a transcription into syllables and create an output string
func (s Syllabifier) SyllabifyToString(t util.Trans) string {
	res := s.Syllabify(t)
	return s.StringWithStressPlacement(res)
}

// Syllabify is used to divide a transcription into syllables
func (s Syllabifier) Syllabify(t util.Trans) sylledTrans {
	res := sylledTrans{Trans: t}
	left := []string{}
	right := t.ListPhonemes()
	for gi, g2p := range t.Phonemes {
		for pi, p := range g2p.P {
			if len(left) > 0 && s.SyllDef.ValidSplit(left, right) && s.SyllDef.ContainsSyllabic(left) && s.SyllDef.ContainsSyllabic(right) {
				index := boundary{G: gi, P: pi}
				res.boundaries = append(res.boundaries, index)
			}
			//fmt.Printf("Syllabify.debug\t%s %s %v %v %v %v\n", left, right, s.SyllDef.ValidSplit(left, right), s.SyllDef.ContainsSyllabic(left), s.SyllDef.ContainsSyllabic(right), res.boundaries)
			left = append(left, p)
			right = right[1:len(right)]
		}
	}
	return res
}

func (s Syllabifier) Test(phnSet util.PhonemeSet) util.TestResult {
	var result = util.TestResult{}
	for _, test := range s.Tests {
		res, err := s.SyllabifyFromString(phnSet, test.Input)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("found error in test input (couldn't split) /%s/ : %s", test.Input, err))
		}
		if res != test.Output {
			result.Errors = append(result.Errors, fmt.Sprintf("from /%s/ expected /%s/, found /%s/", test.Input, test.Output, res))
		}
	}

	return result
}

func (s Syllabifier) StringWithStressPlacement(t sylledTrans) string {
	if s.StressPlacement == Undefined {
		return t.String(s.SyllDef.PhonemeDelimiter(), s.SyllDef.SyllableDelimiter())
	}
	syllables := s.parse(t)
	res := []string{}
	for _, syll := range syllables {
		newSyll := []string{}
		if (s.StressPlacement == FirstInSyllable) && syll.stress != "" {
			newSyll = append(newSyll, syll.stress)
		}
		for _, phn := range syll.phonemes {
			if s.SyllDef.IsSyllabic(phn) && syll.stress != "" && s.StressPlacement == BeforeSyllabic {
				newSyll = append(newSyll, syll.stress)
			}
			newSyll = append(newSyll, phn)
			if s.SyllDef.IsSyllabic(phn) && syll.stress != "" && s.StressPlacement == AfterSyllabic {
				newSyll = append(newSyll, syll.stress)
			}
		}
		res = append(res, strings.Join(newSyll, s.SyllDef.PhonemeDelimiter()))
	}
	return strings.Join(res, s.SyllDef.PhonemeDelimiter()+s.SyllDef.SyllableDelimiter()+s.SyllDef.PhonemeDelimiter())
}
