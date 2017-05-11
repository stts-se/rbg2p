package syll

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/stts-se/rbg2p/util"
)

// LoadFile loads a syllabifier from the specified file
func LoadFile(fName string) (Syllabifier, util.PhonemeSet, error) {
	syllDefLines := []string{}
	res := Syllabifier{}
	phonemeDelimiter := " "
	fh, err := os.Open(fName)
	defer fh.Close()
	if err != nil {
		return res, util.PhonemeSet{}, err
	}
	n := 0
	var phonemeSetLine string
	s := bufio.NewScanner(fh)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return res, util.PhonemeSet{}, err
		}
		n++
		l := util.TrimComment(strings.TrimSpace(s.Text()))
		if util.IsBlankLine(l) || util.IsComment(l) {
		} else if isTest(l) {
			t, err := newTest(l)
			if err != nil {
				return res, util.PhonemeSet{}, err
			}
			res.Tests = append(res.Tests, t)
		} else if util.IsSyllDefLine(l) {
			syllDefLines = append(syllDefLines, l)
		} else if util.IsPhonemeDelimiter(l) {
			phonemeDelimiter, err = util.ParsePhonemeDelimiter(l)
		} else if util.IsPhonemeSet(l) {
			phonemeSetLine = l
		} else {
			return res, util.PhonemeSet{}, fmt.Errorf("unknown input line: %s", l)
		}

	}
	if len(phonemeSetLine) == 0 {
		return res, util.PhonemeSet{}, fmt.Errorf("missing required phoneme set definition")
	}

	phnSet, err := util.ParsePhonemeSet(phonemeSetLine, phonemeDelimiter)
	if err != nil {
		return res, util.PhonemeSet{}, err
	}
	syllDef, err := LoadSyllDef(syllDefLines, phonemeDelimiter)
	if err != nil {
		return res, util.PhonemeSet{}, err
	}
	res.SyllDef = syllDef
	return res, phnSet, nil
}

// LoadSyllDef loads a syllable definition from a set of input lines, and an explicitly specified phoneme delimiter
func LoadSyllDef(syllDefLines []string, phnDelim string) (SyllDef, error) {
	def := MOPSyllDef{} // TODO: Handle other sylldefs too?
	def.PhnDelim = phnDelim

	for _, l := range syllDefLines {
		err := parseMOPSyllDef(l, &def)
		if err != nil {
			return def, err
		}
	}

	if len(def.Stress) == 0 {
		return def, fmt.Errorf("STRESS is required for the syllable definition")
	}
	if len(def.Onsets) == 0 {
		return def, fmt.Errorf("ONSETS is required for the syllable definition")
	}
	if len(def.Syllabic) == 0 {
		return def, fmt.Errorf("SYLLABIC is required for the syllable definition")
	}
	if len(def.SyllDelim) == 0 {
		return def, fmt.Errorf("DELIMITER is required for the syllable definition")
	}

	return def, nil
}

func isTest(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF TEST ")
}

var testRe = regexp.MustCompile("^SYLLDEF TEST +(.+) +-> +(.+)$")

func newTest(s string) (Test, error) {
	var matchRes []string
	matchRes = testRe.FindStringSubmatch(s)
	if matchRes == nil {
		return Test{}, fmt.Errorf("invalid test definition: " + s)
	}
	input := matchRes[1]
	output := matchRes[2]
	if strings.Contains(output, "->") {
		return Test{}, fmt.Errorf("invalid test definition: " + s)
	}
	return Test{Input: input, Output: output}, nil
}

var stressPlacementRe = regexp.MustCompile("^SYLLDEF +STRESS_PLACEMENT +(FirstInSyllable|BeforeSyllabic|AfterSyllabic)$")

func isStressPlacement(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF STRESS_PLACEMENT ")
}
func newStressPlacement(s string) (StressPlacement, error) {
	matchRes := syllDefRe.FindStringSubmatch(s)
	if matchRes == nil {
		matchRes = syllDefTypeRe.FindStringSubmatch(s)
		if matchRes == nil {
			return FirstInSyllable, fmt.Errorf("invalid stress placement definition: " + s)
		}
	}
	//value := matchRes[1]
	return FirstInSyllable, nil
}

var syllDefRe = regexp.MustCompile("^SYLLDEF +(ONSETS|SYLLABIC|DELIMITER|STRESS) +\"(.+)\"$")
var syllDefTypeRe = regexp.MustCompile("^SYLLDEF (TYPE) (MOP)$")
var commaSplit = regexp.MustCompile(" *, *")
var multiSpace = regexp.MustCompile(" +")

func parseMOPSyllDef(s string, syllDef *MOPSyllDef) error {
	// SYLLDEF (ONSETS|SYLLABIC|DELIMITER) "VALUE"
	// SYLLDEF TYPE VALUE
	matchRes := syllDefRe.FindStringSubmatch(s)
	if matchRes == nil {
		matchRes = syllDefTypeRe.FindStringSubmatch(s)
		if matchRes == nil {
			return fmt.Errorf("invalid sylldef definition: " + s)
		}
	}
	name := matchRes[1]
	value := strings.Replace(strings.TrimSpace(matchRes[2]), "\\\"", "\"", -1)
	_, err := regexp.Compile(value)
	if err != nil {
		return fmt.Errorf("invalid sylldef input (regular expression failed) for /%s/: %s", s, err)
	}
	if name == "TYPE" {
		if value != "MOP" {
			return fmt.Errorf("invalid sylldef type %s", value)
		}
	} else if name == "ONSETS" {
		syllDef.Onsets = commaSplit.Split(value, -1)
	} else if name == "SYLLABIC" {
		syllDef.Syllabic = multiSpace.Split(value, -1)
	} else if name == "STRESS" {
		syllDef.Stress = multiSpace.Split(value, -1)
	} else if name == "DELIMITER" {
		syllDef.SyllDelim = value
	} else {
		return fmt.Errorf("invalid sylldef variable : %s", s)
	}
	return nil
}
