package rbg2p

import (
	"bufio"
	"fmt"
	"net/http"
	u "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// LoadSyllURL loads a syllabifier from an URL
func LoadSyllURL(url string) (Syllabifier, error) {
	urlP, err := u.Parse(url)
	if err != nil {
		return Syllabifier{}, err
	}
	resp, err := http.Get(urlP.String())
	if err != nil {
		return Syllabifier{}, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	return loadSyll(scanner, url)
}

// LoadSyllFile loads a syllabifier from the specified file
func LoadSyllFile(fName string) (Syllabifier, error) {
	fh, err := os.Open(filepath.Clean(fName))
	if err != nil {
		return Syllabifier{}, err
	}
	/* #nosec G307 */
	defer fh.Close()
	scanner := bufio.NewScanner(fh)
	return loadSyll(scanner, fName)
}

// LoadSyllFile loads a syllabifier from the specified file
func loadSyll(scanner *bufio.Scanner, inputPath string) (Syllabifier, error) {
	var err error
	syllDefLines := []string{}
	res := Syllabifier{}
	phonemeDelimiter := " "
	n := 0
	var phonemeSetLine string
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return res, err
		}
		n++
		l := trimComment(strings.TrimSpace(scanner.Text()))
		if isBlankLine(l) || isComment(l) {
		} else if isSyllTest(l) {
			t, err := newSyllTest(l)
			if err != nil {
				return res, err
			}
			res.Tests = append(res.Tests, t)
		} else if isSyllDefLine(l) {
			syllDefLines = append(syllDefLines, l)
		} else if isPhonemeDelimiter(l) {
			phonemeDelimiter, err = parsePhonemeDelimiter(l)
			if err != nil {
				return res, err
			}
		} else if isPhonemeSet(l) {
			phonemeSetLine = l
		} else if isG2PLine(l) {
			// do nothing
		} else {
			return res, fmt.Errorf("unknown input line: %s", l)
		}

	}
	if len(phonemeSetLine) == 0 {
		return res, fmt.Errorf("missing required phoneme set definition")
	}

	syllDef, stressPlacement, err := loadSyllDef(syllDefLines, phonemeDelimiter)
	if err != nil {
		return res, err
	}
	res.SyllDef = syllDef
	phnSet, err := parsePhonemeSet(phonemeSetLine, res.SyllDef, phonemeDelimiter)
	if err != nil {
		return res, err
	}
	res.StressPlacement = stressPlacement
	res.PhonemeSet = phnSet

	return res, nil
}

func loadSyllDef(syllDefLines []string, phnDelim string) (SyllDef, StressPlacement, error) {
	def := MOPSyllDef{} // TODO: Handle other sylldefs too?
	def.PhnDelim = phnDelim
	stressPlacement := Undefined

	for _, l := range syllDefLines {
		if isStressPlacement(l) {
			stress, err := newStressPlacement(l)
			if err != nil {
				return def, stressPlacement, err
			}
			stressPlacement = stress
			continue
		}
		err := parseMOPSyllDef(l, &def)
		if err != nil {
			return def, stressPlacement, err
		}
	}

	// fmt.Printf("DEBUG <%v>\n", def.Onsets[0])
	// for _, onset := range def.Onsets {
	// 	if onset == "f t" {
	// 		fmt.Printf("DEBUG <%v> <%v>\n", onset, len([]rune(onset)))
	// 	}
	// }

	if len(def.Stress) == 0 {
		return def, stressPlacement, fmt.Errorf("STRESS is required for the syllable definition")
	}
	if len(def.Onsets) == 0 {
		return def, stressPlacement, fmt.Errorf("ONSETS is required for the syllable definition")
	}
	if len(def.Syllabic) == 0 {
		return def, stressPlacement, fmt.Errorf("SYLLABIC is required for the syllable definition")
	}
	if len(def.SyllDelim) == 0 {
		return def, stressPlacement, fmt.Errorf("DELIMITER is required for the syllable definition")
	}

	return def, stressPlacement, nil
}

func isSyllTest(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF TEST ")
}

var syllTestRe = regexp.MustCompile("^SYLLDEF TEST +(.+) +-> +(.+)$")

func newSyllTest(s string) (SyllTest, error) {
	matchRes := syllTestRe.FindStringSubmatch(s)
	if matchRes == nil {
		return SyllTest{}, fmt.Errorf("invalid test definition: %s", s)
	}
	input := matchRes[1]
	output := matchRes[2]
	if strings.Contains(output, "->") {
		return SyllTest{}, fmt.Errorf("invalid test definition: %s", s)
	}
	return SyllTest{Input: input, Output: output}, nil
}

var stressPlacementRe = regexp.MustCompile("^SYLLDEF +STRESS_PLACEMENT +(FirstInSyllable|BeforeSyllabic|AfterSyllabic)$")

func isStressPlacement(s string) bool {
	return strings.HasPrefix(s, "SYLLDEF STRESS_PLACEMENT ")
}
func newStressPlacement(s string) (StressPlacement, error) {
	matchRes := stressPlacementRe.FindStringSubmatch(s)
	if matchRes == nil {
		matchRes = syllDefTypeRe.FindStringSubmatch(s)
		if matchRes == nil {
			return Undefined, fmt.Errorf("invalid stress placement definition: %s", s)
		}
	}
	value := matchRes[1]

	// TODO: generate _strings.go using 'stringer -type=StressPlacement' but this doesn't work right now for some reason (tried two different computers)
	if strings.ToLower(value) == "firstinsyllable" {
		return FirstInSyllable, nil
	} else if strings.ToLower(value) == "beforesyllabic" {
		return BeforeSyllabic, nil
	} else if strings.ToLower(value) == "aftersyllabic" {
		return AfterSyllabic, nil
	}
	return Undefined, fmt.Errorf("invalid stress placement: %s", s)
}

var syllDefRe = regexp.MustCompile("^SYLLDEF +(ONSETS|SYLLABIC|DELIMITER|STRESS) +\"(.+)\"$")
var syllDefBoolRe = regexp.MustCompile("^SYLLDEF +(INCLUDE_PHONEME_DELIMITER) (true|false)$")
var syllDefTypeRe = regexp.MustCompile("^SYLLDEF (TYPE) (MOP)$")

func parseMOPSyllDef(s string, syllDef *MOPSyllDef) error {
	// SYLLDEF (ONSETS|SYLLABIC|DELIMITER|...) "VALUE"
	// SYLLDEF TYPE VALUE
	matchRes := syllDefRe.FindStringSubmatch(s)
	if matchRes == nil {
		matchRes = syllDefTypeRe.FindStringSubmatch(s)
		if matchRes == nil {
			matchRes = syllDefBoolRe.FindStringSubmatch(s)
			if matchRes == nil {
				return fmt.Errorf("invalid sylldef definition: %s", s)
			}
		}
	}
	var includePhnDelimSet = false
	name := matchRes[1]
	value := strings.Replace(strings.TrimSpace(matchRes[2]), "\\\"", "\"", -1)
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
	} else if name == "INCLUDE_PHONEME_DELIMITER" {
		var bl, err = strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse boolean value for %s : %s", name, value)
		}
		syllDef.IncludePhnDelim = bl
		includePhnDelimSet = true
	} else {
		return fmt.Errorf("invalid sylldef variable : %s", s)
	}
	if !includePhnDelimSet {
		syllDef.IncludePhnDelim = true
	}
	return nil
}
