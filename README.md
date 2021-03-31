# rbg2p
Utilities for rule based, manually written, grapheme to phoneme rules 

[![GoDoc](https://godoc.org/github.com/stts-se/rbg2p?status.svg)](https://godoc.org/github.com/stts-se/rbg2p) [![Go Report Card](https://goreportcard.com/badge/github.com/stts-se/rbg2p)](https://goreportcard.com/report/github.com/stts-se/rbg2p) [![Github actions workflow status](https://github.com/stts-se/rbg2p/workflows/Go/badge.svg)](https://github.com/stts-se/rbg2p/actions)

## Command line tools

### G2P

    $ g2p <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)
    
    FLAGS:
       -force      bool    print transcriptions even if errors are found (default: false)
       -column     string  only convert specified column (default: first field)
       -quiet      bool    inhibit warnings (default: false)
       -test       bool    test g2p against input file; orth <tab> trans (default: false)
       -help       bool    print help message


<!--
### Syllabification

    $ syll <G2P/SYLL RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)
-->


### Microservice API/server

     $ server cmd/server/g2p_files
     
 Visit http://localhost:6771/ for info on available API calls
 
