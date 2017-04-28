package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stts-se/rbg2p"
)

var l = log.New(os.Stderr, "", 0)

func print(orth string, transes []rbg2p.Trans) {
	ts := []string{}
	for _, t := range transes {
		ts = append(ts, strings.Join(t.Phonemes, PhnDelimiter))
	}
	fmt.Printf("%s\t%s\n", orth, strings.Join(ts, "\t"))
}

func transcribe(ruleSet rbg2p.RuleSet, orth string) bool {
	transes, err := ruleSet.Apply(orth)
	if err != nil {
		l.Printf("Couldn't transcribe '%s' : %s", orth, err)
		return false
	}
	print(orth, transes)
	return true
}

func main() {

	if len(os.Args) < 3 {
		log.Println("go run g2p_runner.go <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)>")
		os.Exit(1)
	}

	g2pFile := os.Args[1]
	ruleSet, err := rbg2p.LoadFile(g2pFile)
	if err != nil {
		log.Printf("couldn't load file %s : %s", g2pFile, err)
		os.Exit(1)
	}

	errors := ruleSet.Test()
	if len(errors) > 0 {
		for _, err = range errors {
			fmt.Printf("%v\n", err)
		}
		l.Printf("%d OF %d TESTS FAILED FOR %s\n", len(errors), len(ruleSet.Tests), g2pFile)
		os.Exit(1)
	} else {
		l.Printf("ALL %d TESTS PASSED FOR %s\n", len(ruleSet.Tests), g2pFile)
	}

	nTotal := 0
	nErrs := 0
	nTrans := 0
	for i := 2; i < len(os.Args); i++ {
		s := os.Args[i]
		if _, err := os.Stat(s); os.IsNotExist(err) {
			nTotal = nTotal + 1
			if transcribe(ruleSet, strings.ToLower(s)) {
				nTrans = nTrans + 1
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
				line := strings.ToLower(sc.Text())
				if transcribe(ruleSet, line) {
					nTrans = nTrans + 1
				} else {
					nErrs = nErrs + 1
				}
			}
		}
	}
	l.Printf("TOTAL WORDS: %d", nTotal)
	l.Printf("ERRORS: %d", nErrs)
	l.Printf("TRANSCRIBED: %d", nTrans)
}
