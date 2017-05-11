package syll

type StressPlacement int

const (
	Undefined StressPlacement = iota
	FirstInSyllable
	BeforeSyllabic
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
