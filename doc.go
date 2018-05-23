/*Package rbg2p contains utilities for rule based, manually written, grapheme to phoneme rules.

Each g2p rule set is defined in a .g2p file with the following content:

    * specific variables
      - used to define constant variables such as character set and phoneme delimiter
    * variables
      - any variables for use in the context of the actual rules
    * sylldef - definitions for dividing transcriptions into syllables
    * rules - g2p rules
    * filters - transcription filters applied after the rules
    * tests - input/output tests
    * comments

SPECIFIC VARIABLES

Defines a set of constant variables, such as character set and phoneme delimiter. Please note that quotes are required around the value, since space and the empty string can be used as a value.
     <NAME> "<VALUE>"

Available variables (* means required):
     CHARACTER_SET*     (default: none)
      - used to check that each character in the character set has at least one rule
     PHONEME_SET        (default: none)
      - space separated symbol set, used to validate the phonemes in the g2p rules
     DEFAULT_PHONEME    (default: "_")
      - used for input input (orthographic) symbols
     PHONEME_DELIMITER  (default: " ")
      - used to concatenate phonemes into a transcriptions
     DOWNCASE_INPUT     (default: true)

Examples:
     CHARACTER_SET "abcdefghijklmnopqrstuvwxyzåäö"
     PHONEME_SET "a au o u i y e eu p t k b d g r s f h j l v w m n S tS"
     DEFAULT_PHONEME "_"
     PHONEME_DELIMITER " "


VARIABLES

Regexp variables prefixed by VAR, that can be used in the rule context as exemplified below. The variable names must not contain underscore (_).
     VAR <NAME> <VALUE>

Examples:
     VAR VOWEL [aeyuio]
     VAR AFFRICATE (tS|dZ)
     VAR VOICELESS [ptksf]


SYLLDEF

An set of variables prefixed by SYLLDEF, used for syllabification (not required).
     SYLLDEF <NAME> "<VALUE>"

Currently, only maximum onset (MOP) syllabification can be used.
Variables currently available:

     TYPE    (default: MOP)
      - currently, the only value allowed here is MOP
     ONSETS
      - a comma separated list of valid syllable onsets (typically consonant clusters)
     SYLLABIC
      - a space separated list of syllabic phonemes (typically vowels)
     STRESS
      - a space separated list of stress symbols
     DELIMITER
      - syllable delimiter symbol

Examples:
     SYLLDEF TYPE MOP
     SYLLDEF ONSETS "p, b, t, rt, m, n, d, rd, k, g, rn, f, v, C, rs, r, l, s, x, S, h, rl, j, s, p, r, rs p r, s p l, rs p l, s p j, rs p j, s t r, rs rt r, s k r, rs k r, s k v, rs k v, p r, p j, p l, b r, b j, b l, t r, rt r, t v, rt v, d r, rd r, d v, rd v, k r, k l, k v, k n, g r, g l, g n, f r, f l, f j, f n, v r, s p, s t, s k, s v, s l, s m, s n, n j, rs p, rs rt, rs k, rs v, rs rl, rs m, rs rn, rn j, m j"
     SYLLDEF SYLLABIC "i: I u0 }: a A: u: U E: {: E { au y: Y e: e 2: 9: 2 9 o: O @ eu"
     SYLLDEF STRESS "\" %"
     SYLLDEF DELIMITER "."


RULES

Grapheme to phoneme rules written in a format loosely based on phonotactic rules. The rules are ordered, and typically the rule order is of great importance.

     <INPUT> -> <OUTPUT>
     <INPUT> -> <OUTPUT> / <CONTEXT>
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>)
     <INPUT> -> (<OUTPUT1>, <OUTPUT2>) / <CONTEXT>

Context:
     <LEFT CONTEXT> _ <RIGHT CONTEXT>


<INPUT> is a string of one or more input characters. <OUTPUT> is a string representing the output (separated by the pre-defined phoneme delimiter, above). For empty output, i.e., when a character should not be pronounced, use the empty set symbol "∅" (U+2205).

<CONTEXT> is the context in which the <INPUT> should occur for the rule to apply. Pre-defined variables (above) can be use in the context specs. # is used for anchoring (marks the start/end of the input string).

Examples:
     a -> ? a / # _
     a -> a
     e -> e
     skt -> (s t, s k t) / _
     ck -> k
     b -> p / _ VOICELESS
     h -> ∅ / # _


FILTERS

Regexp replacement filters for transcriptions. The filters are applied after the g2p rules. Pre-defined variables (see above) cannot be used in the filters for now.
     FILTER "<FROM RE>" -> "<TO RE>"

Example:
     FILTER "^" -> "\" " // place stress first in transcription


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


---

SEPARATE SYLLABIFICATION RULE FILE

A .syll file for syllabification contains a subset of the items used for a proper g2p.

Example (for the CMU lexicon):

   PHONEME_SET "AA AE AH AX AO AW AY B CH D DH EH ER EY F G HH IH IY JH K L M N NG OW OY P R S SH T TH UH UW V W Y Z ZH 1 2"
   PHONEME_DELIMITER " "

   SYLLDEF TYPE MOP
   SYLLDEF ONSETS "P, T, K, B, D, G, CH, JH, F, V, T, D, S, Z, S, Z, H, L, M, N, N, R, W, J, P R, T R, B R, G R, S T R, S P R, S K R, P L, T L, B L, G L, S T L, S P L, S K L, S P, S T, S K"
   SYLLDEF SYLLABIC "AA AE AH AX AO AW AY EH ER EY IH IY OW OY UH UW"
   SYLLDEF STRESS "1 2"
   SYLLDEF DELIMITER "$"

   SYLLDEF TEST AX P R 1 AA K S AX M AX T -> AX $ P R 1 AA K $ S AX $ M AX T
   SYLLDEF TEST W 1 UH D S T R 2 IY M -> W 1 UH D $ S T R 2 IY M


For details on the .g2p file format, check docs for the root folder of this package.


For more examples (used for unit tests), see the test_data folder: https://github.com/stts-se/rbg2p/tree/master/test_data


To test a single g2p file from the command line, use cmd/g2p.

To import and use the rbg2p rule package in another go program:

    import (
           "github.com/stts-se/rbg2p"
    )

    func main() {
            var g2pFile, orth
            // initialize g2pFile and orth

            ruleSet, err := rbg2p.LoadFile(g2pFile)
            // check for error in err

            testRes := ruleSet.Test()
            // check for error in testRes

            transes, err := ruleSet.Apply(orth)
    }



*/
package rbg2p
