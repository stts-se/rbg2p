# rbg2p
Utilities for rule based, manually written, grapheme to phoneme rules

## Command line
    $ go run g2p_runner/g2p_runner.go <G2P RULE FILE> <WORDS (FILES OR LIST OF WORDS)> (optional)
    
    FLAGS:
       -force      bool    print transcriptions even if errors are found (default: false)
       -symbolset  string  use specified symbol set file for validating the symbols in
                           the g2p rule set (default: none)
       -help       bool    print help message

## Microservice API/server

     $ go run server/*.go server/g2p_files
     
 Visit http://localhost:6771/ for info on available API calls
 
## API docs

[![GoDoc](https://godoc.org/github.com/stts-se/rbg2p?status.svg)](https://godoc.org/github.com/stts-se/rbg2p)
