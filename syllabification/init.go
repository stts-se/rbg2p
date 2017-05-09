package syllabification

import (
	"fmt"
	"regexp"
	"strings"
)

func LoadSyllDef(syllDefLines []string, phnDelim string) (SyllDef, error) {
	syllDef := MOPSyllDef{} // TODO: Handle other sylldefs too?
	syllDef.PhnDelim = phnDelim

	for _, l := range syllDefLines {
		err := parseMOPSyllDef(l, &syllDef)
		if err != nil {
			return syllDef, err
		}
	}
	return syllDef, nil
}

var syllDefRe = regexp.MustCompile("^SYLLDEF +(ONSETS|SYLLABIC|DELIMITER) +\"(.+)\"$")
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
	} else if name == "DELIMITER" {
		syllDef.SyllDelim = value
	} else {
		return fmt.Errorf("invalid sylldef variable : %s", s)
	}
	return nil
}
