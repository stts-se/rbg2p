package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/stts-se/rbg2p"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var l = log.New(os.Stderr, "", 0)

func print(input string, orth string, transes []string) {
	fmt.Printf("%s\t%s\n", input, strings.Join(transes, "  #  "))
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

var removeBoundariesRE = regexp.MustCompile(`[.!~] *`)
var removeStressRE = regexp.MustCompile(`[%"] *`)
var removeStress *bool
var transSplitRE = regexp.MustCompile(" +# +")

func cleanTransForDiff(t string) string {
	var res = t
	res = removeBoundariesRE.ReplaceAllString(res, "")
	if *removeStress {
		res = removeStressRE.ReplaceAllString(res, "")
	}
	//res = strings.Replace(res, "'", "", -1)
	return res
}

// func cleanTransForIJDiff(t string) string {
// 	var res = t
// 	res = strings.Replace(res, " i ", " j ", -1)
// 	return res
// }

func compareForDiff(old []string, new []string) (string, bool) {
	for i, s := range old {
		old[i] = cleanTransForDiff(s)
	}
	for i, s := range new {
		new[i] = cleanTransForDiff(s)
	}
	// var oldIJ = []string{}
	// var newIJ = []string{}
	// for _, s := range old {
	// 	oldIJ = append(oldIJ, cleanTransForIJDiff(s))
	// }
	// for _, s := range new {
	// 	newIJ = append(newIJ, cleanTransForIJDiff(s))
	// }
	if reflect.DeepEqual(old, new) {
		return "ALL EQ", true
	} else if old[0] == new[0] {
		return "#1 EQ", false
		// } else if reflect.DeepEqual(oldIJ, newIJ) {
		// 	return "ALL EQ IJ", false
		// } else if oldIJ[0] == newIJ[0] {
		// 	return "#1 EQ IJ", false
	} else {
		return "DIFF", false
	}
}

func main() {

	var f = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var debug = f.Bool("debug", false, "print extra debug info (default: false)")
	var force = f.Bool("force", false, "print transcriptions even if errors are found (default: false)")
	var column = f.Int("column", 0, "only convert specified column (default: first field)")
	var coverageCheck = f.Bool("coverage", false, "run coverage check (rules applied/not applied) (default: false)")
	var quiet = f.Bool("quiet", false, "inhibit warnings (default: false)")
	var test = f.Bool("test", false, "test g2p against input file; orth <tab> trans (default: false)")
	removeStress = f.Bool("test:removestress", false, "remove stress when comparing using the -test switch (default: false)")
	var ssFile = f.String("symbolset", "", "use specified symbol set file for validating the symbols in the g2p rule set, one symbol per line (default: none; overrides the g2p rule file's symbolset, if any)")
	var help = f.Bool("help", false, "print help and exit")

	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "g2p <FLAGS> <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)\n")
		fmt.Fprintf(os.Stderr, "\nFLAGS:\n")
		f.PrintDefaults()
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
		f.Usage()
		os.Exit(1)
	}

	if len(args) < 1 {
		f.Usage()
		os.Exit(1)
	}

	rbg2p.Debug = *debug

	g2pFile := args[0]
	ruleSet, err := rbg2p.LoadFile(g2pFile)
	if err != nil {
		l.Printf("couldn't load rule file %s : %s", g2pFile, err)
		os.Exit(1)
	}

	if *ssFile != "" {
		syllDelimIncludesPhnDelim := true
		if ruleSet.Syllabifier.SyllDef != nil {
			syllDelimIncludesPhnDelim = ruleSet.Syllabifier.SyllDef.IncludePhonemeDelimiter()
		}

		phonemeSet, err := rbg2p.LoadPhonemeSetFile(*ssFile, syllDelimIncludesPhnDelim, ruleSet.SyllableDelimiter, ruleSet.PhonemeDelimiter)
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
	if !*quiet {
		for _, e := range result.Warnings {
			l.Printf("WARNING: %v\n", e)
		}
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

	var rulesApplied, rulesNotApplied int
	if *coverageCheck {
		rulesApplied = 0
		rulesNotApplied = 0
		for _, r := range ruleSet.Rules {
			rs := r.String()
			ruleSet.RulesAppliedMutex.RLock()
			if n, ok := ruleSet.RulesApplied[rs]; ok {
				if !*quiet {
					l.Printf("TEST RULE APPLIED\t%s\tat input line %v\t%v", rs, r.LineNumber, n)
				}
				rulesApplied++
			} else {
				if !*quiet {
					l.Printf("TEST RULE NOT APPLIED\t%s\tat input line %v", rs, r.LineNumber)
				}
				rulesNotApplied++
			}
		}
		l.Printf("%-24s: % 7d", "TEST RULES APPLIED", rulesApplied)
		l.Printf("%-24s: % 7d", "TEST RULES NOT APPLIED", rulesNotApplied)
		rulesApplied = 0
		rulesNotApplied = 0
		ruleSet.RulesApplied = make(map[string]int)
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
		fmt.Println("ORTH\tG2P TRANSES\tREF TRANSES\tDIFFTAG\t(DIFF)?")
	}
	var processString = func(s string) {
		nTotal = nTotal + 1
		fs := strings.Split(s, "\t")
		o := fs[*column]
		res := transcribe(ruleSet, o)
		if res.result || *force {
			nTrans = nTrans + 1
			if *test {
				refTranses := []string{}
				for _, s := range fs[(*column + 1):] {
					refTranses = append(refTranses, transSplitRE.Split(s, -1)...)
					// for _, refT := range transSplitRE.Split(s, -1) {
					// 	refTranses = append(refTranses, refT)
					// }
				}
				nTests++
				info, _ := compareForDiff(res.transes, refTranses)
				testRes[info]++
				outFs := []string{res.orth, strings.Join(res.transes, " # "), strings.Join(refTranses, " # "), info}
				if info == "DIFF" {
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(outFs[1], outFs[2], false)
					diffsOnly := []diffmatchpatch.Diff{}
					diffsOnlyText := []string{}
					for _, d := range diffs {
						if d.Type != diffmatchpatch.DiffEqual {
							diffsOnly = append(diffsOnly, d)
							diffsOnlyText = append(diffsOnlyText, d.Text)
						}
					}
					outFs = append(outFs, dmp.DiffPrettyText(diffs))
					outFs = append(outFs, fmt.Sprintf("%v", diffsOnly))
					outFs = append(outFs, strings.Join(diffsOnlyText, "|"))
				}

				fmt.Println(strings.Join(outFs, "\t"))
			} else {
				print(s, res.orth, res.transes)
			}
		}
		if !res.result {
			nErrs = nErrs + 1
		}
	}

	if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			s := args[i]
			if _, err := os.Stat(s); os.IsNotExist(err) {
				processString(s)
				// nTotal = nTotal + 1
				// res := transcribe(ruleSet, s)
				// if res.result || *force {
				// 	nTrans = nTrans + 1
				// 	fmt.Printf("%s\t%s\n", s, strings.Join(res.transes, "\t"))
				// }
				// if !res.result {
				// 	nErrs = nErrs + 1
				// }
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
					line := sc.Text()
					if strings.TrimSpace(line) == "" {
						l.Println("Skipping empty line")
						continue
					}
					if strings.HasPrefix(strings.TrimSpace(line), "#") {
						l.Println("Skipping line " + line)
						continue
					}
					processString(line)
					// 	nTotal = nTotal + 1
					// 	fs := strings.Split(line, "\t")
					// 	o, refTranses := fs[0], fs[1:]
					// 	res := transcribe(ruleSet, o)
					// 	if res.result || *force {
					// 		nTrans = nTrans + 1
					// 		if *test {
					// 			nTests++
					// 			info, _ := compareForDiff(res.transes, refTranses)
					// 			testRes[info]++
					// 			outFs := []string{res.orth, strings.Join(res.transes, " # "), strings.Join(refTranses, "#"), info}
					// 			if info == "DIFF" {
					// 				dmp := diffmatchpatch.New()
					// 				diffs := dmp.DiffMain(outFs[1], outFs[2], false)
					// 				diffsOnly := []diffmatchpatch.Diff{}
					// 				diffsOnlyText := []string{}
					// 				for _, d := range diffs {
					// 					if d.Type != diffmatchpatch.DiffEqual {
					// 						diffsOnly = append(diffsOnly, d)
					// 						diffsOnlyText = append(diffsOnlyText, d.Text)
					// 					}
					// 				}
					// 				outFs = append(outFs, dmp.DiffPrettyText(diffs))
					// 				outFs = append(outFs, fmt.Sprintf("%v", diffsOnly))
					// 				outFs = append(outFs, strings.Join(diffsOnlyText, "|"))
					// 			}

					// 			fmt.Println(strings.Join(outFs, "\t"))
					// 		} else {
					// 			print(res.orth, res.transes)
					// 		}
					// 	}
					// 	if !res.result {
					// 		nErrs = nErrs + 1
					// 	}
				}
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Reading input from stdin...\n")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				l.Println("Skipping empty line")
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				l.Println("Skipping line " + line)
				continue
			}
			processString(line)
		}
	}
	if *coverageCheck {
		rulesApplied = 0
		rulesNotApplied = 0
		for _, r := range ruleSet.Rules {
			rs := r.String()
			if n, ok := ruleSet.RulesApplied[rs]; ok {
				if !*quiet {
					l.Printf("RULE APPLIED\t%s\tat input line %v\t%v", rs, r.LineNumber, n)
				}
				rulesApplied++
			} else {
				if !*quiet {
					l.Printf("RULE NOT APPLIED\t%s\tat input line %v", rs, r.LineNumber)
				}
				rulesNotApplied++
			}
		}
		ruleSet.RulesApplied = make(map[string]int)
	}

	l.Printf("%-21s: % 7d", "TOTAL INPUT", nTotal)
	l.Printf("%-21s: % 7d", "ERRORS", nErrs)
	l.Printf("%-21s: % 7d", "TRANSCRIBED", nTrans)
	if *coverageCheck {
		l.Printf("%-21s: % 7d", "RULES APPLIED", rulesApplied)
		l.Printf("%-21s: % 7d", "RULES NOT APPLIED", rulesNotApplied)
	}
	if *test {
		l.Printf("%-21s: % 7d", "TESTED", nTests)
		var keys []string
		for k := range testRes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, tag := range keys {
			freq := testRes[tag]
			s := " > TEST " + tag
			l.Printf("%-21s: % 7d", s, freq)
		}
	}
}
