# rbg2p
Utilities for rule based, manually written, grapheme to phoneme rules

[![GoDoc](https://godoc.org/github.com/stts-se/rbg2p?status.svg)](https://godoc.org/github.com/stts-se/rbg2p) [![Go Report Card](https://goreportcard.com/badge/github.com/stts-se/rbg2p)](https://goreportcard.com/report/github.com/stts-se/rbg2p)

## Command line tools

### G2P

    $ go run cmd/g2p/g2p.go <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)
    
    FLAGS:
       -force      bool    print transcriptions even if errors are found (default: false)
       -symbolset  string  use specified symbol set file for validating the symbols in
                           the g2p rule set (default: none)
       -help       bool    print help message


### Syllabification

    $ go run cmd/syll/syll.go <G2P/SYLL RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)


### Microservice API/server

     $ go run cmd/server/*.go cmd/server/g2p_files
     
 Visit http://localhost:6771/ for info on available API calls
 
