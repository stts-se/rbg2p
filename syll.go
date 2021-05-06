package rbg2p

import (
	"fmt"
	"os"
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
	StressPlacement() StressPlacement
	IncludePhonemeDelimiter() bool
	SyllableDelimiter() string
}

// MOPSyllDef is a Maximum Onset Principle implementation of the SyllDef interface
type MOPSyllDef struct {
	Onsets          []string
	Syllabic        []string
	PhnDelim        string
	SyllDelim       string
	Stress          []string
	StressPlcmnt    StressPlacement
	IncludePhnDelim bool
}

// PhonemeDelimiter is the string used to separate phonemes (required by interface)
func (def MOPSyllDef) PhonemeDelimiter() string {
	return def.PhnDelim
}

// IncludePhonemeDelimiter defines whether the syllable boundaries should be surrounded by the phoneme delimiter
func (def MOPSyllDef) IncludePhonemeDelimiter() bool {
	return def.IncludePhnDelim
}

// SyllableDelimiter is the string used to separate syllables (required by interface)
func (def MOPSyllDef) SyllableDelimiter() string {
	return def.SyllDelim
}

// StressPlacement
func (def MOPSyllDef) StressPlacement() StressPlacement {
	return def.StressPlcmnt
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

	// if def.StressPlacement() != AfterSyllabic {
	// 	if len(left) > 0 && def.IsStress(left[len(left)-1]) {
	// 		return false
	// 	}
	// }

	// if def.StressPlacement() == AfterSyllabic {
	// 	if len(right) > 0 && def.IsStress(right[0]) {
	// 		right = right[1:]
	// 	}
	// }

	//fmt.Println("debug validsplit internal left/right", left, right)

	onset := []string{}
	keepCond := func(s string) bool {
		if def.StressPlacement() != AfterSyllabic {
			return !def.IsSyllabic(s)
		}
		return !def.IsSyllabic(s) // && !def.IsStress(s)
	}
	for i := 0; i < len(right) && keepCond(right[i]); i++ {
		if def.IsStress(right[i]) {
			if def.StressPlacement() == AfterSyllabic {
				onset = append(onset, right[i])
			}
		} else {
			onset = append(onset, right[i])
		}
	}
	//s := strings.Join(onset, def.PhonemeDelimiter())
	//fmt.Println("debug validsplit test onset1", s, def.validOnset(s))
	if !def.validOnset(strings.Join(onset, def.PhonemeDelimiter())) {
		return false
	}
	test := onset
	for i := len(left) - 1; i >= 0 && keepCond(left[i]); i-- {
		test = append([]string{left[i]}, test...)
		//s := strings.Join(test, def.PhonemeDelimiter())
		//fmt.Println("debug validsplit test onset2", s, def.validOnset(s))
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
	Debug           bool
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
	sylled := s.syllabify(t)
	res := s.stringWithStressPlacement(sylled)
	if s.Debug {
		fmt.Fprintf(os.Stderr, "%s\t%s\t%v\t%s\n", "SYLLABIFY", t, sylled, res)
	}
	return res
}

func (s Syllabifier) syllabify(t trans) sylledTrans {
	res := sylledTrans{trans: t}
	left := []string{}
	right := t.listPhonemes()
	for gi, g2p := range t.phonemes {
		for pi, p := range g2p.p {
			validSplit := s.SyllDef.ValidSplit(left, right)
			//fmt.Println("debug syllabify", gi, pi, p, left, right, validSplit, s.SyllDef.ContainsSyllabic(left), s.SyllDef.ContainsSyllabic(right))
			if len(left) > 0 && validSplit && s.SyllDef.ContainsSyllabic(left) && s.SyllDef.ContainsSyllabic(right) {
				index := boundary{g: gi, p: pi}
				res.boundaries = append(res.boundaries, index)
				left = []string{}
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
	if s.Debug {
		fmt.Fprintf(os.Stderr, "PARSED SYLLABLES\t%v\n", syllables)
	}
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
	if s.SyllDef.IncludePhonemeDelimiter() {
		return strings.Join(res, s.SyllDef.PhonemeDelimiter()+s.SyllDef.SyllableDelimiter()+s.SyllDef.PhonemeDelimiter())
	}
	return strings.Join(res, s.SyllDef.SyllableDelimiter())
}
