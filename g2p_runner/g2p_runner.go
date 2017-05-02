package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stts-se/rbg2p"
)

var l = log.New(os.Stderr, "", 0)

func print(orth string, transes []rbg2p.Trans, phnDelim string) {
	ts := []string{}
	for _, t := range transes {
		ts = append(ts, strings.Join(t.Phonemes, phnDelim))
	}
	fmt.Printf("%s\t%s\n", orth, strings.Join(ts, "\t"))
}

func transcribe(ruleSet rbg2p.RuleSet, orth string, force bool) bool {
	transes, err := ruleSet.Apply(orth)
	if err != nil {
		l.Printf("Couldn't transcribe '%s' : %s", orth, err)
		if force {
			print(orth, transes, ruleSet.PhonemeDelimiter)
		}
		return false
	}
	print(orth, transes, ruleSet.PhonemeDelimiter)
	return true
}

func main() {

	var f = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var force = f.Bool("force", false, "print transcriptions even if errors are found (default: false)")
	var ssFile = f.String("symbolset", "", "use specified symbol set file for validating the symbols in the g2p rule set (default: none; overrides the g2p rule file's symbolset, if any)")
	var help = f.Bool("help", false, "print help message")

	var usage = `go run g2p_runner.go <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)>

FLAGS:
   -force      bool    print transcriptions even if errors are found (default: false)
   -symbolset  string  use specified symbol set file for validating the symbols in the g2p rule set (default: none)
   -help       bool    print help message`

	f.Usage = func() {
		l.Printf(usage)
	}

	var args = os.Args
	if strings.HasSuffix(args[0], "g2p_runner") {
		args = args[1:] // remove first argument if it's the program name
	}
	err := f.Parse(args)
	if err != nil {
		os.Exit(1)
	}

	args = f.Args()

	if *help {
		l.Println(usage)
		os.Exit(1)
	}

	if len(args) < 2 {
		l.Println(usage)
		os.Exit(1)
	}

	g2pFile := args[0]
	ruleSet, err := rbg2p.LoadFile(g2pFile)
	if err != nil {
		l.Printf("couldn't load file %s : %s", g2pFile, err)
		os.Exit(1)
	}

	if *ssFile != "" {
		phonemeSet, err := rbg2p.LoadPhonemeSetFile(*ssFile, ruleSet.PhonemeDelimiter)
		if err != nil {
			l.Printf("couldn't load symbol set : %s", err)
			os.Exit(1)
		}
		ruleSet.PhonemeSet = phonemeSet
	}

	haltingError := false
	result := ruleSet.Test()
	for _, e := range result.Errors {
		l.Printf("ERROR: %v\n", e)
	}
	l.Printf("%d ERROR(S) FOR %s\n", len(result.Errors), g2pFile)
	for _, e := range result.Warnings {
		l.Printf("WARNING: %v\n", e)
	}
	l.Printf("%d WARNING(S) FOR %s\n", len(result.Warnings), g2pFile)
	if len(result.Errors) > 0 {
		haltingError = true
	}
	if len(result.FailedTests) > 0 {
		for _, e := range result.FailedTests {
			l.Printf("FAILED TEST: %v\n", e)
		}
		l.Printf("%d OF %d TESTS FAILED FOR %s\n", len(result.FailedTests), len(ruleSet.Tests), g2pFile)
		haltingError = true
	} else {
		l.Printf("ALL %d TESTS PASSED FOR %s\n", len(ruleSet.Tests), g2pFile)
	}

	if haltingError && !*force {
		os.Exit(1)
	}

	fmt.Println()

	nTotal := 0
	nErrs := 0
	nTrans := 0
	for i := 1; i < len(args); i++ {
		s := args[i]
		if _, err := os.Stat(s); os.IsNotExist(err) {
			nTotal = nTotal + 1
			if transcribe(ruleSet, strings.ToLower(s), *force) {
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
				if transcribe(ruleSet, line, *force) {
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
