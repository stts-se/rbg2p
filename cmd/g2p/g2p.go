package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/stts-se/rbg2p"
)

var l = log.New(os.Stderr, "", 0)

func print(orth string, transes []string) {
	fmt.Printf("%s\t%s\n", orth, strings.Join(transes, "\t"))
}

type transResult struct {
	orth    string
	transes []string
	result  bool
}

func transcribe(ruleSet rbg2p.RuleSet, orth string) transResult {
	transes, err := ruleSet.Apply(orth)
	if err != nil {
		l.Printf("Couldn't transcribe '%s' : %s", orth, err)
		return transResult{orth: orth, transes: transes, result: false}
	}
	return transResult{orth: orth, transes: transes, result: true}
}

var removeStressAndBoundaries = regexp.MustCompile("[.\"%!~] *")

func cleanTransForDiff(t string) string {
	var res = t
	res = removeStressAndBoundaries.ReplaceAllString(res, "")
	res = strings.Replace(res, "'", "", -1)
	return res
}

func compareForDiff(old []string, new []string) (string, bool) {
	for i, s := range old {
		old[i] = cleanTransForDiff(s)
	}
	for i, s := range new {
		new[i] = cleanTransForDiff(s)
	}
	if reflect.DeepEqual(old, new) {
		return "ALL EQ", true
	} else if old[0] == new[0] {
		return "#1 EQ", false
	} else if len(new) > 1 && old[0] == new[1] {
		return "ONE EQ", false
	} else if len(old) > 1 && old[1] == new[0] {
		return "ONE EQ", false
	} else if len(old) > 1 && len(new) > 1 && old[1] == new[1] {
		return "ONE EQ", false
	} else {
		return "REALDIFF", false
	}
}

func main() {

	var f = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var force = f.Bool("force", false, "print transcriptions even if errors are found (default: false)")
	var test = f.Bool("test", false, "test g2p against input file; orth <tab> trans (default: false)")
	var ssFile = f.String("symbolset", "", "use specified symbol set file for validating the symbols in the g2p rule set (default: none; overrides the g2p rule file's symbolset, if any)")
	var help = f.Bool("help", false, "print help message")

	var usage = `go run g2p.go <FLAGS> <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)

FLAGS:
   -force      bool    print transcriptions even if errors are found (default: false)
   -test       bool    test g2p against input file; orth <tab> trans (default: false)
   -symbolset  string  use specified symbol set file for validating the symbols in the g2p rule set (default: none)
   -help       bool    print help message`

	f.Usage = func() {
		l.Printf(usage)
	}

	var args = os.Args
	if strings.HasSuffix(args[0], "g2p") {
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

	g2pFile := args[0]
	ruleSet, err := rbg2p.LoadFile(g2pFile)
	if err != nil {
		l.Printf("couldn't load rule file %s : %s", g2pFile, err)
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

	nTotal := 0
	nErrs := 0
	nTrans := 0
	nTests := 0
	testRes := make(map[string]int)
	if *test {
		fmt.Println("ORTH\tDIFFTAG\tNEW TRANSES\tOLD TRANSES")
	}
	for i := 1; i < len(args); i++ {
		s := args[i]
		if _, err := os.Stat(s); os.IsNotExist(err) {
			nTotal = nTotal + 1
			res := transcribe(ruleSet, strings.ToLower(s))
			if res.result {
				nTrans = nTrans + 1
				print(res.orth, res.transes)
			} else {
				nErrs = nErrs + 1
				if *force {
					print(res.orth, res.transes)
				}
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
				fs := strings.Split(line, "\t")
				o, refTranses := fs[0], fs[1:]
				res := transcribe(ruleSet, strings.ToLower(o))
				if res.result || *force {
					nTrans = nTrans + 1
					if *test {
						nTests++
						info, eq := compareForDiff(res.transes, refTranses)
						testRes[info]++
						if !eq && info != "REALDIFF" {
							testRes["SMALLDIFF"]++
						}
						outFs := []string{res.orth, info, strings.Join(res.transes, " # "), strings.Join(refTranses, "#")}
						fmt.Println(strings.Join(outFs, "\t"))
					} else {
						print(res.orth, res.transes)
					}
				}
				if !res.result {
					nErrs = nErrs + 1
				}
			}
		}
	}
	l.Printf("%-18s: % 7d", "TOTAL WORDS", nTotal)
	l.Printf("%-18s: % 7d", "ERRORS", nErrs)
	l.Printf("%-18s: % 7d", "TRANSCRIBED", nTrans)
	if *test {
		l.Printf("%-18s: % 7d", "TESTED", nTests)
		var keys []string
		for k := range testRes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, tag := range keys {
			freq := testRes[tag]
			s := " > TEST " + tag
			l.Printf("%-18s: % 7d", s, freq)
		}
	}
}
