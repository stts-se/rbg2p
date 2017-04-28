/*Utilities for rule based, manually written, grapheme to phoneme rules.

Each g2p rule set is defined in a .g2p file with the following content:

VARIABLES

Regexp variables prefixed by VAR:
     VAR <NAME> <VALUE>

Examples:
     VAR VOWEL [aeyuio]
     VAR AFFRICATE (tS|dZ)


RULES

Grapheme to phoneme rules written in a format loosely based on phonotactic rules. The rules are ordered, and typically the rule order is of great importance.

     <INPUT> -> <OUTPUT>
     <INPUT> -> <OUTPUT> / <CONTEXT>
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>)
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>) / <CONTEXT>

Context:
     <LEFT CONTEXT> _ <RIGHT CONTEXT>

# is used for anchoring (marks the start/end of the input string).

Examples:
     a -> ? a / # _
     a -> a
     e -> e
     skt -> (s t, s k t) / _
     ck -> k


COMMENTS

Comments are prefixed by //


TESTS

Test examples prefixed by TEST:
     TEST <INPUT> -> <OUTPUT>

or with variants:
     TEST <INPUT> -> (<OUTPUT1>, <OUTPUT2>)

Examples:
     TEST hit -> h i t
     TEST kex -> (k e ks, C e ks)


For more examples (used for unit tests), see the test_data folder: https://github.com/stts-se/rbg2p/tree/master/test_data

*/
package rbg2p
