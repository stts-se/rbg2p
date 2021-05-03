# rbg2p
Utilities for rule based, manually written, grapheme to phoneme rules 

[![GoDoc](https://godoc.org/github.com/stts-se/rbg2p?status.svg)](https://godoc.org/github.com/stts-se/rbg2p) [![Go Report Card](https://goreportcard.com/badge/github.com/stts-se/rbg2p)](https://goreportcard.com/report/github.com/stts-se/rbg2p) [![Github actions workflow status](https://github.com/stts-se/rbg2p/workflows/Go/badge.svg)](https://github.com/stts-se/rbg2p/actions)

## Command line tools

### G2P

    g2p <FLAGS> <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)

    FLAGS:
      -column int
            only convert specified column (default: first field)
      -coverage
            run coverage check (rules applied/not applied) (default: false)
      -debug
            print extra debug info (default: false)
      -force
            print transcriptions even if errors are found (default: false)
      -help
            print help and exit
      -quiet
            inhibit warnings (default: false)
      -symbolset string
            use specified symbol set file for validating the symbols in the g2p rule set (default: none; overrides the g2p rule file's symbolset, if any)
      -test
            test g2p against input file; orth <tab> trans (default: false)
      -test:removestress
            remove stress when comparing using the -test switch (default: false)

<!--
### Syllabification

    $ syll <G2P/SYLL RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)
-->


### Microservice API/server

     $ server cmd/server/g2p_files
     
 Visit http://localhost:6771/ for info on available API calls
 

---

_This work was supported by the Swedish Post and Telecom Authority (PTS) through the grant "Wikispeech – en användargenererad talsyntes på Wikipedia" (2016–2017)._
