package rbg2p

import (
	"sort"
	"strings"
)

// package internal util for splitting transcriptions where there is no explicit phoneme delimiter

// Sort slice of strings according to len, longest string first
type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func SplitIntoPhonemes(knownPhonemes []string, transcription string) (phonemes []string, unknown []string, error error) {

	var known []string
	// start by discarding any phoneme strings not substrings of transcription
	for _, ph := range knownPhonemes {
		if len(ph) > 0 && strings.Index(transcription, ph) > -1 {
			known = append(known, ph)
		}
	}

	sort.Sort(byLength(known))
	ps, uk := splitIntoPhonemes0(&known, transcription, []string{}, []string{})
	return ps, uk, nil
}

// the recursive loop
func splitIntoPhonemes0(srted *[]string, trans string, phs []string, unk []string) ([]string, []string) {

	if len(trans) > 0 {
		pre, rest, ok := consume(srted, trans)
		if ok { // known phoneme is prefix if trans
			return splitIntoPhonemes0(srted, rest, append(phs, pre), unk)
		}
		// unknown prefix, chopped off first rune
		return splitIntoPhonemes0(srted, rest, append(phs, pre), append(unk, pre))
	}
	return phs, unk
}

func consume(srtd *[]string, trans string) (string, string, bool) {
	var resPref string
	var resSuffix string

	var prefixFound bool

	notInTrans := make(map[string]bool)
	for _, ph := range *srtd {

		ind := strings.Index(trans, ph)
		if ind == 0 { // bingo
			resPref = ph
			resSuffix = trans[len(ph):]
			prefixFound = true

			break
		}

		if ind < 0 {
			notInTrans[ph] = true
		}
	}

	// Discard phonemes we know are not in trans
	if len(notInTrans) > 0 {
		var tmp []string

		for _, ph := range *srtd {
			if !notInTrans[ph] {
				tmp = append(tmp, ph)
			}
		}

		*srtd = tmp
	}
	// no known phoneme prefix, separate first rune
	if prefixFound {
		return resPref, resSuffix, prefixFound
	}
	t := []rune(trans)
	return string(t[0]), string(t[1:]), false
}
