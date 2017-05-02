/*Package rbg2p contains utilities for rule based, manually written, grapheme to phoneme rules.

Each g2p rule set is defined in a .g2p file with the following content:

    * specific variables
      - used to define constant variables such as character set and phoneme delimiter
    * variables (optional)
      - any variables for use in the context of the actual rules
    * rules - g2p rules
    * tests - input/output tests
    * comments (optional)

SPECIFIC VARIABLES

Defines a set of constant variables, such as character set and phoneme delimiter. Please note that quotes are required around the value, since space and the empty string can be used as a value.
     <NAME> "<VALUE>"

Available variables (* is required):
     CHARACTER_SET*     (default: none)
      - used to check that each character in the character set has at least one rule
     PHONEME_SET        (default: none)
      - space separated symbol set, used to validate the phonemes in the g2p rules
     DEFAULT_PHONEME    (default: "_")
      - used for input input (orthographic) symbols
     PHONEME_DELIMITER  (default: " ")
      - used to concatenate phonemes into a transcriptions

Examples:
     CHARACTER_SET "abcdefghijklmnopqrstuvwxyzåäö"
     PHONEME_SET "a o u i y e p t k b d g r s f h j l v w m n"
     DEFAULT_PHONEME "_"
     PHONEME_DELIMITER " "


VARIABLES

Regexp variables prefixed by VAR, that can be used in the rule context as exemplified below.
     VAR <NAME> <VALUE>

Examples:
     VAR VOWEL [aeyuio]
     VAR AFFRICATE (tS|dZ)
     VAR VOICELESS [ptksf]


RULES

Grapheme to phoneme rules written in a format loosely based on phonotactic rules. The rules are ordered, and typically the rule order is of great importance.

     <INPUT> -> <OUTPUT>
     <INPUT> -> <OUTPUT> / <CONTEXT>
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>)
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>) / <CONTEXT>

Context:
     <LEFT CONTEXT> _ <RIGHT CONTEXT>

Pre-defined variables (above) can be use in the contex specs. # is used for anchoring (marks the start/end of the input string).

Examples:
     a -> ? a / # _
     a -> a
     e -> e
     skt -> (s t, s k t) / _
     ck -> k
     b -> p / _ VOICELESS


COMMENTS

Comments are prefixed by //


TESTS

Test examples prefixed by TEST:
     TEST <INPUT> -> <OUTPUT>

or with variants:
     TEST <INPUT> -> (<OUTPUT1>, <OUTPUT2>)

Examples:
     TEST hit -> h i t
     TEST kex -> (k e k s, C e k s)


For more examples (used for unit tests), see the test_data folder: https://github.com/stts-se/rbg2p/tree/master/test_data

*/
package rbg2p
