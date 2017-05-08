package rbg2p

import "strings"

// Trans is a container for phonemes in a transcription
type Trans struct {
	Phonemes []G2P
}

/*G2P is a container for one-to-many grapheme-phoneme mapping received from the G2P ruleset. Primarily package internal use. Examples (IPA symbols):
  x -> k, s
  sch -> ʃ
  ff -> f
  au -> a‿u
  rt -> ʈ
**/
type G2P struct {
	G string
	P []string
}

//ListPhonemes returns a slice of phonemes as strings
func (t Trans) ListPhonemes() []string {
	var phns []string
	for _, g2p := range t.Phonemes {
		for _, p := range g2p.P {
			phns = append(phns, p)
		}
	}
	return phns
}

func (t Trans) String(phnDelimiter string) string {
	var phns []string
	for _, p := range t.ListPhonemes() {
		if len(p) > 0 {
			phns = append(phns, p)
		}
	}
	return strings.Join(phns, phnDelimiter)
}
