package rbg2p

import (
	"fmt"
	"strings"
)

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

// IsStress is used to check if the input symbol is a stress symbol
func (def MOPSyllDef) IsStress(symbol string) bool {
	for _, s := range def.Stress {
		if s == symbol {
			return true
		}
	}
	return false
}

// IsSyllabic is used to check if the input phoneme is syllabic
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
		//fmt.Printf("DEBUG <%v> <%v> %v\n", s, onset, s == onset)
		if s == onset {
			return true
		}
	}
	return false
}

// ValidSplit is called by Syllabifier.Syllabify to test where to put the boundaries
func (def MOPSyllDef) ValidSplit(left0 []string, right0 []string) bool {
	left := left0
	right := right0
	// left := []string{}
	// for _, s := range left0 {
	// 	if s != "" {
	// 		left = append(left, s)
	// 	}
	// }
	// right := []string{}
	// for _, s := range right0 {
	// 	if s != "" {
	// 		right = append(right, s)
	// 	}
	// }

	if len(left) > 0 && def.IsStress(left[len(left)-1]) {
		return false
	}

	if len(right) > 0 && def.IsStress(right[0]) {
		right = right[1:]
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

// SyllTest defines a rule test (input -> output)
type SyllTest struct {
	Input  string
	Output string
}

// Syllabifier is a module to divide a transcription into syllables
type Syllabifier struct {
	SyllDef         SyllDef
	Tests           []SyllTest
	StressPlacement StressPlacement
	PhonemeSet      PhonemeSet
}

// IsDefined is used to determine if there is a syllabifier defined or not
func (s Syllabifier) IsDefined() bool {
	return s.SyllDef != nil && s.SyllDef.IsDefined()
}

// SyllabifyFromPhonemes is used to divide a range of phonemes into syllables and create an output string
func (s Syllabifier) SyllabifyFromPhonemes(phns []string) string {
	t := trans{}
	for _, phn := range phns {
		t.phonemes = append(t.phonemes, g2p{g: "", p: []string{phn}})
	}
	return s.syllabifyToString(t)
}

// SyllabifyFromString is used to divide a transcription string into syllables and create an output string
func (s Syllabifier) SyllabifyFromString(trans string) (string, error) {
	phns, err := s.PhonemeSet.SplitTranscription(trans)
	if err != nil {
		return "", err
	}
	return s.SyllabifyFromPhonemes(phns), nil
}

// syllabifyToString is used to divide a transcription into syllables and create an output string
func (s Syllabifier) syllabifyToString(t trans) string {
	res := s.syllabify(t)
	return s.stringWithStressPlacement(res)
}

func (s Syllabifier) syllabify(t trans) sylledTrans {
	res := sylledTrans{trans: t}
	left := []string{}
	right := t.listPhonemes()
	for gi, g2p := range t.phonemes {
		for pi, p := range g2p.p {
			if len(left) > 0 && s.SyllDef.ValidSplit(left, right) && s.SyllDef.ContainsSyllabic(left) && s.SyllDef.ContainsSyllabic(right) {
				index := boundary{g: gi, p: pi}
				res.boundaries = append(res.boundaries, index)
			}
			left = append(left, p)
			right = right[1:]
		}
	}
	return res
}

//Test to test the input syllabifier definition using tests in the input data or file
func (s Syllabifier) Test() TestResult {
	var result = TestResult{}
	for _, test := range s.Tests {
		res, err := s.SyllabifyFromString(test.Input)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("found error in test input (couldn't split) /%s/ : %s", test.Input, err))
		}
		if res != test.Output {
			result.Errors = append(result.Errors, fmt.Sprintf("from /%s/ expected /%s/, found /%s/", test.Input, test.Output, res))
		}
	}

	return result
}

func (s Syllabifier) stringWithStressPlacement(t sylledTrans) string {
	if s.StressPlacement == Undefined {
		return t.string(s.SyllDef.PhonemeDelimiter(), s.SyllDef.SyllableDelimiter())
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
