package rbg2p

// StressPlacement is used to define where in a syllable the stress should be put in an output string
type StressPlacement int

const (
	// Undefined - position not defined
	Undefined StressPlacement = iota

	// FirstInSyllable -- before the syllable's first phoneme
	FirstInSyllable

	// BeforeSyllabic -- before the first syllabic phoneme
	BeforeSyllabic

	// AfterSyllabic -- after the first syllabic phoneme
	AfterSyllabic
)

type syllable struct {
	phonemes []string
	stress   string
}

func (s Syllabifier) parse(t sylledTrans) []syllable {
	syllables := t.syllables()
	res := []syllable{}
	for _, syll := range syllables {
		newSyll := syllable{}
		for _, p := range syll {
			if s.SyllDef.IsStress(p) {
				newSyll.stress = p
			} else {
				newSyll.phonemes = append(newSyll.phonemes, p)
			}
		}
		res = append(res, newSyll)
	}
	return res
}
