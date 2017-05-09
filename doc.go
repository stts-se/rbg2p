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

Examples:
     CHARACTER_SET "abcdefghijklmnopqrstuvwxyzåäö"
     PHONEME_SET "a au o u i y e eu p t k b d g r s f h j l v w m n S tS"
     DEFAULT_PHONEME "_"
     PHONEME_DELIMITER " "


VARIABLES

Regexp variables prefixed by VAR, that can be used in the rule context as exemplified below.
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


For more examples (used for unit tests), see the test_data folder: https://github.com/stts-se/rbg2p/tree/master/test_data

*/
package rbg2p
