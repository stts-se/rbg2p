package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stts-se/pronlex/symbolset"
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
			print(orth, transes, ruleSet.PhnDelimiter)
		}
		return false
	}
	print(orth, transes, ruleSet.PhnDelimiter)
	return true
}

func main() {

	var f = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var force = f.Bool("force", false, "print transcriptions even if errors are found (default: false)")
	var ssFile = f.String("symbolset", "", "use specified symbol set for validating the symbols in the g2p rule set (default: none)")
	var help = f.Bool("help", false, "print help message")

	var usage = `go run g2p_runner.go <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)>

FLAGS:
   -force      bool    print transcriptions even if errors are found (default: false)
   -symbolset  string  use specified symbol set for validating the symbols in the g2p rule set (default: none)
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
		symbolSet, err := symbolset.LoadSymbolSet(*ssFile)
		if err != nil {
			l.Printf("couldn't load symbol set : %s", err)
			os.Exit(1)
		}
		validation, err := rbg2p.CompareToSymbolSet(ruleSet, symbolSet)
		if err != nil {
			l.Printf("couldn't validate against symbol set : %s", err)
			os.Exit(1)
		}
		if len(validation.Warnings) > 0 {
			for _, err := range validation.Warnings {
				l.Printf("SYMBOL SET WARNING: %v\n", err)
			}
		}
		if len(validation.Errors) > 0 {
			for _, err := range validation.Errors {
				l.Printf("SYMBOL SET ERROR: %v\n", err)
			}
			os.Exit(1)
		}
	}

	errors := ruleSet.Test()
	if len(errors) > 0 {
		for _, err = range errors {
			l.Printf("%v\n", err)
		}
		l.Printf("%d OF %d TESTS FAILED FOR %s\n", len(errors), len(ruleSet.Tests), g2pFile)
		os.Exit(1)
	} else {
		l.Printf("ALL %d TESTS PASSED FOR %s\n", len(ruleSet.Tests), g2pFile)
	}

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
