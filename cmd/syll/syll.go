package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stts-se/rbg2p/syll"
	"github.com/stts-se/rbg2p/util"
)

func syllabify(syller syll.Syllabifier, phnSet util.PhonemeSet, trans string) bool {
	phonemes, err := phnSet.SplitTranscription(trans)
	if err != nil {
		l.Printf("%s", err)
		return false
	}
	sylled := syller.SyllabifyFromPhonemes(phonemes)
	fmt.Printf("%s\t%s\n", trans, sylled)
	return true
}

var l = log.New(os.Stderr, "", 0)

func main() {
	var usage = `go run syll.go <G2P/SYLL RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)`

	var args = os.Args
	if strings.HasSuffix(args[0], "syll") {
		args = args[1:] // remove first argument if it's the program name
	}

	if len(args) < 1 {
		l.Println(usage)
		os.Exit(1)
	}

	ruleFile := args[0]
	syller, phnSet, err := syll.LoadFile(ruleFile)
	if err != nil {
		l.Printf("couldn't load rule file %s : %s", ruleFile, err)
		os.Exit(1)
	}

	nTotal := 0
	nErrs := 0
	nOK := 0
	for i := 1; i < len(args); i++ {
		s := args[i]
		if _, err := os.Stat(s); os.IsNotExist(err) {
			nTotal = nTotal + 1
			if syllabify(syller, phnSet, s) {
				nOK = nOK + 1
			} else {
				nErrs = nErrs + 1
			}
		} else {
			fh, err := os.Open(s)
			defer fh.Close()
			if err != nil {
				l.Println(err)
				os.Exit(1)
			}
			sc := bufio.NewScanner(fh)
			for sc.Scan() {
				if err := sc.Err(); err != nil {
					l.Println(err)
					os.Exit(1)
				}
				nTotal = nTotal + 1
				line := sc.Text()
				if syllabify(syller, phnSet, line) {
					nOK = nOK + 1
				} else {
					nErrs = nErrs + 1
				}
			}
		}
	}
	l.Printf("TOTAL WORDS: %d", nTotal)
	l.Printf("ERRORS: %d", nErrs)
	l.Printf("SYLLABIFIED: %d", nOK)
}
