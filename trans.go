package rbg2p

import "strings"

// Trans is a container for phonemes in a transcription. Primarily for package internal use.
type trans struct {
	phonemes []g2p
}

/*G2P is a container for one-to-many grapheme-phoneme mapping received from the G2P ruleset. Primarily for package internal use. Examples (IPA symbols):
  x -> k, s
  sch -> ʃ
  ff -> f
  au -> a‿u
  rt -> ʈ
*/
type g2p struct {
	g string
	p []string
}

//ListPhonemes returns a slice of phonemes as strings
func (t trans) listPhonemes() []string {
	var phns []string
	for _, g2p := range t.phonemes {
		for _, p := range g2p.p {
			phns = append(phns, p)
		}
	}
	return phns
}

func (t trans) string(phnDelimiter string) string {
	var phns []string
	for _, p := range t.listPhonemes() {
		if len(p) > 0 {
			phns = append(phns, p)
		}
	}
	return strings.Join(phns, phnDelimiter)
}

// boundary represents syllable boundaries. Primarily for package internal use.
type boundary struct {
	g int
	p int
}

// sylledTrans is a syllabified transcription (containing a Trans instance and a slice of indices for syllable boundaries)
type sylledTrans struct {
	trans      trans
	boundaries []boundary
	stress     []string
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
func (t sylledTrans) string(phnDelimiter string, syllDelimiter string) string {
	res := []string{}
	for gi, g2p := range t.trans.phonemes {
		for pi, p := range g2p.p {
			index := boundary{g: gi, p: pi}
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
	for gi, g2p := range t.trans.phonemes {
		for pi, p := range g2p.p {
			index := boundary{g: gi, p: pi}
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
func (t sylledTrans) listPhonemes() []string {
	return t.trans.listPhonemes()
}
