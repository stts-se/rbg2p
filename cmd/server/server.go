package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/stts-se/rbg2p"
)

type g2pMutex struct {
	g2ps  map[string]rbg2p.RuleSet
	sylls map[string]rbg2p.Syllabifier
	mutex *sync.RWMutex
}

var g2pM = g2pMutex{
	g2ps:  make(map[string]rbg2p.RuleSet),
	sylls: make(map[string]rbg2p.Syllabifier),
	mutex: &sync.RWMutex{},
}

func g2pMain_Handler(w http.ResponseWriter, r *http.Request) {
	_, cmdFileName, _, _ := runtime.Caller(0)
	fString := path.Join(path.Dir(cmdFileName), "static/g2p_demo.html")
	http.ServeFile(w, r, fString)
}

func ping_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "rbg2p")
}

var wSplitRe = regexp.MustCompile(" *, *")

// Word internal struct for json
type Word struct {
	Orth    string   `json:"orth"`
	Transes []string `json:"transes"`
}

func syllabify(lang string, trans string) (string, int, error) {
	g2pM.mutex.RLock()
	defer g2pM.mutex.RUnlock()
	syller, ok := g2pM.sylls[lang]
	if !ok {
		ruleSet, ok := g2pM.g2ps[lang]
		if !ok {
			msg := "unknown 'lang': " + lang
			langs, err := listSyllLanguages()
			if err != nil {
				return "", http.StatusInternalServerError, err
			}
			msg = fmt.Sprintf("%s. Known 'lang' values: %s", msg, strings.Join(langs, ", "))
			return "", http.StatusBadRequest, fmt.Errorf(msg)
		}

		if !ruleSet.Syllabifier.IsDefined() {
			msg := fmt.Sprintf("no syllabifier defined for language %s", lang)
			return "", http.StatusInternalServerError, fmt.Errorf(msg)
		}
		syller = ruleSet.Syllabifier
	}
	phns, err := syller.PhonemeSet.SplitTranscription(trans)
	if err != nil {
		msg := fmt.Sprintf("couldn't split input transcription /%s/ : %s", trans, err)
		return "", http.StatusInternalServerError, fmt.Errorf(msg)
	}
	sylled := syller.SyllabifyFromPhonemes(phns)
	return sylled, http.StatusOK, nil
}

func syllabify_Handler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	trans := vars["trans"]
	if "" == trans {
		msg := "no value for the expected 'trans' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	res, status, err := syllabify(lang, trans)
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}

	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// j, err := json.Marshal(res)
	// if err != nil {
	// 	msg := fmt.Sprintf("failed json marshalling : %v", err)
	// 	log.Println(msg)
	// 	http.Error(w, msg, http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Fprintf(w, string(j))
	fmt.Fprintf(w, string(res))
}

func transcribe(lang string, word string) (Word, int, error) {
	g2pM.mutex.RLock()
	defer g2pM.mutex.RUnlock()
	ruleSet, ok := g2pM.g2ps[lang]
	if !ok {
		msg := "unknown 'lang': " + lang
		langs := listG2PLanguages()
		msg = fmt.Sprintf("%s. Known 'lang' values: %s", msg, strings.Join(langs, ", "))
		return Word{}, http.StatusBadRequest, fmt.Errorf(msg)
	}

	transes, err := ruleSet.Apply(word)
	if err != nil {
		msg := fmt.Sprintf("couldn't transcribe word : %v", err)
		return Word{}, http.StatusInternalServerError, fmt.Errorf(msg)
	}
	tRes := []string{}
	for _, trans := range transes {
		tRes = append(tRes, trans)
	}
	res := Word{word, tRes}
	return res, http.StatusOK, nil
}

func ruleContent(lang string) (string, int, error) {
	g2pM.mutex.RLock()
	defer g2pM.mutex.RUnlock()
	ruleSet, ok := g2pM.g2ps[lang]
	if !ok {
		msg := "unknown 'lang': " + lang
		langs := listG2PLanguages()
		msg = fmt.Sprintf("%s. Known 'lang' values: %s", msg, strings.Join(langs, ", "))
		return "", http.StatusBadRequest, fmt.Errorf(msg)
	}
	return ruleSet.Content, http.StatusOK, nil
}

func transcribe_Handler(w http.ResponseWriter, r *http.Request) {

	format := r.FormValue("format")
	if "xml" == format {
		transcribe_AsXml_Handler(w, r)
		return
	}

	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	word := vars["word"]
	if "" == word {
		msg := "no value for the expected 'word' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word = strings.ToLower(word)

	res, status, err := transcribe(lang, word)
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}

	if "text" == format || "txt" == format {
		res := strings.Join(res.Transes, "\n")
		fmt.Fprintf(w, res)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		j, err := json.Marshal(res)
		if err != nil {
			msg := fmt.Sprintf("failed json marshalling : %v", err)
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(j))
	}
}

// XMLWords container go generate xml from http request, for legacy calls from ltool/yalt
type XMLWords struct {
	XMLName xml.Name `xml:"words"`
	Words   []XMLWord
}

// XMLWord container go generate xml from http request, for legacy calls from ltool/yalt
type XMLWord struct {
	XMLName xml.Name `xml:"word"`
	Orth    string   `xml:"orth,attr"`
	Trans   string   `xml:"trans"`
}

// for legacy calls from ltool/yalt
func transcribe_AsXml_Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word := vars["word"]
	if "" == word {
		msg := "no value for the expected 'word' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word = strings.ToLower(word)
	res, status, err := transcribe(lang, word)
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}
	//<words>
	//<word orth='apa' word_lang='mk' trans_lang='mk' >" a p a</word>
	//</words>

	// words := XMLWords{
	// 	Words: []XMLWord{
	// 		XMLWord{Orth: word, Trans: res.Transes[0]},
	// 	},
	// }
	words := XMLWords{}
	for _, t := range res.Transes {
		words.Words = append(words.Words, XMLWord{Orth: word, Trans: t})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	xml, err := xml.Marshal(words)
	if err != nil {
		msg := fmt.Sprintf("failed xml marshalling : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(xml))
	//fmt.Fprintf(w, string(res.Transes[0]))
}

func listG2PLanguages() []string {
	var res []string
	for name := range g2pM.g2ps {
		res = append(res, name)
	}
	return res
}

func listSyllLanguages() ([]string, error) {
	var res []string
	for name, g2p := range g2pM.g2ps {
		if g2p.Syllabifier.IsDefined() {
			res = append(res, name)
		}
	}
	for name := range g2pM.sylls {
		if rbg2p.Contains(res, name) {
			return res, fmt.Errorf("name conflict for '%s' (g2p and syllabifier cannot share the same name)", name)
		}
		res = append(res, name)
	}
	return res, nil
}

func g2pList_Handler(w http.ResponseWriter, r *http.Request) {
	g2pM.mutex.RLock()
	res := listG2PLanguages()
	g2pM.mutex.RUnlock()

	sort.Strings(res)
	j, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed json marshalling : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(j))
}

func g2pRules_Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	res, status, err := ruleContent(lang)
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}

	fmt.Fprintf(w, res)
}

func syllList_Handler(w http.ResponseWriter, r *http.Request) {
	g2pM.mutex.RLock()
	res, err := listSyllLanguages()
	g2pM.mutex.RUnlock()
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	sort.Strings(res)
	j, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed json marshalling : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(j))
}

// langFromFilePath returns the base file name stripped from any '.g2p' extension
func langFromFilePath(p string) string {
	b := filepath.Base(p)
	if strings.HasSuffix(b, ".g2p") {
		b = b[0 : len(b)-4]
	} else if strings.HasSuffix(b, ".syll") {
		b = b[0 : len(b)-5]
	}
	return b
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "server <G2P FILES DIR>\n")
		os.Exit(0)
	}

	// g2p file dir. Each file in dir with .g2p extension
	// is treated as a g2p file
	var dir = os.Args[1]

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	// populate map of g2p rules from files.
	// The base file name minus '.g2p' is the language name.
	var fn string
	for _, f := range files {
		fn = filepath.Join(dir, f.Name())
		if strings.HasSuffix(fn, ".g2p") {

			ruleSet, err := rbg2p.LoadFile(fn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				fmt.Fprintf(os.Stderr, "server: skipping file: '%s'\n", fn)
				continue
			}

			lang := langFromFilePath(fn)
			g2pM.mutex.Lock()
			g2pM.g2ps[lang] = ruleSet
			g2pM.mutex.Unlock()
			fmt.Fprintf(os.Stderr, "server: loaded file '%s'\n", fn)

		} else if strings.HasSuffix(fn, ".syll") {

			syll, err := rbg2p.LoadSyllFile(fn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				fmt.Fprintf(os.Stderr, "server: skipping file: '%s'\n", fn)
				continue
			}

			lang := langFromFilePath(fn)
			g2pM.mutex.Lock()
			g2pM.sylls[lang] = syll
			g2pM.mutex.Unlock()
			fmt.Fprintf(os.Stderr, "server: loaded file '%s'\n", fn)

		} else {
			fmt.Fprintf(os.Stderr, "server: skipping file: '%s'\n", fn)
			continue
		}

	}

	if _, err := listSyllLanguages(); err != nil {
		log.Fatalf("server init error : %s", err)
	}

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", g2pMain_Handler)

	r.HandleFunc("/ping", ping_Handler)

	r.HandleFunc("/g2p/list", g2pList_Handler)
	r.HandleFunc("/syll/list", syllList_Handler)

	r.HandleFunc("/g2p/rules/{lang}", g2pRules_Handler)

	r.HandleFunc("/transcribe/{lang}/{word}", transcribe_Handler)
	r.HandleFunc("/syllabify/{lang}/{trans}", syllabify_Handler)

	// for legacy calls from ltool/yalt
	r.HandleFunc("/xmltranscribe/{lang}/{word}", transcribe_AsXml_Handler)

	fmt.Println("Serving urls:")
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		fmt.Println(t)
		return nil
	})

	port := ":6771"
	log.Printf("starting g2p server at port %s\n", port)
	err = http.ListenAndServe(port, r)
	if err != nil {

		log.Fatalf("no fun: %v\n", err)
	}

}
