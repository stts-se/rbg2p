package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/stts-se/rbg2p"
)

func syllabify(syller rbg2p.Syllabifier, trans string) (string, bool) {
	phonemes, err := syller.PhonemeSet.SplitTranscription(trans)
	if err != nil {
		l.Printf("%s", err)
		return "", false
	}
	sylled := syller.SyllabifyFromPhonemes(phonemes)
	//fmt.Printf("%s\t%s\n", trans, sylled)
	return sylled, true
}

var l = log.New(os.Stderr, "", 0)

func main() {
	var f = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var force = f.Bool("force", false, "print transcriptions even if errors are found (default: false)")
	var column = f.Int("column", 0, "only convert specified column (default: first field)")
	var help = f.Bool("help", false, "print help message")

	var usage = `syll <FLAGS> <G2P/SYLL RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)

FLAGS:
   -force      bool    print transcriptions even if errors are found (default: false)
   -column     string  only convert specified column (default: first field)
   -help       bool    print help message`

	f.Usage = func() {
		l.Printf(usage)
	}

	var args = os.Args
	if strings.HasSuffix(args[0], "syll") {
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

	if len(args) < 1 {
		l.Println(usage)
		os.Exit(1)
	}

	ruleFile := args[0]
	syller, err := rbg2p.LoadSyllFile(ruleFile)
	if err != nil {
		l.Printf("couldn't load rule file %s : %s", ruleFile, err)
		os.Exit(1)
	}

	haltingError := false
	result := syller.Test()
	for _, e := range result.Errors {
		l.Printf("ERROR: %v\n", e)
	}
	l.Printf("%d ERROR(S) FOR %s\n", len(result.Errors), ruleFile)
	for _, e := range result.Warnings {
		l.Printf("WARNING: %v\n", e)
	}
	l.Printf("%d WARNING(S) FOR %s\n", len(result.Warnings), ruleFile)
	if len(result.Errors) > 0 {
		haltingError = true
	}
	if len(result.FailedTests) > 0 {
		for _, e := range result.FailedTests {
			l.Printf("FAILED TEST: %v\n", e)
		}
		l.Printf("%d OF %d TESTS FAILED FOR %s\n", len(result.FailedTests), len(syller.Tests), ruleFile)
		haltingError = true
	} else {
		l.Printf("ALL %d TESTS PASSED FOR %s\n", len(syller.Tests), ruleFile)
	}

	if haltingError && !*force {
		os.Exit(1)
	}

	nTotal := 0
	nErrs := 0
	nOK := 0
	for i := 1; i < len(args); i++ {
		s := args[i]
		if _, err := os.Stat(s); os.IsNotExist(err) {
			nTotal = nTotal + 1
			fs := strings.Split(s, "\t")
			o := fs[*column]
			if res, ok := syllabify(syller, o); ok {
				fmt.Printf("%s\t%s\n", s, res)
				nOK = nOK + 1
			} else {
				nErrs = nErrs + 1
			}
		} else {
			fh, err := os.Open(filepath.Clean(s))
			if err != nil {
				l.Println(err)
				os.Exit(1)
			}
			/* #nosec G307 */
			defer fh.Close()
			sc := bufio.NewScanner(fh)
			for sc.Scan() {
				if err := sc.Err(); err != nil {
					l.Println(err)
					os.Exit(1)
				}
				nTotal = nTotal + 1
				line := sc.Text()
				fs := strings.Split(line, "\t")
				o := fs[*column]
				if res, ok := syllabify(syller, o); ok {
					fmt.Printf("%s\t%s\n", line, res)
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
