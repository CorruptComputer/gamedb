package helpers

import (
	"regexp"
)

//noinspection GoUnusedGlobalVariable
var (
	RegexInts                  = regexp.MustCompile(`[0-9]+`)
	RegexIntsOnly              = regexp.MustCompile(`^[0-9]+$`)
	RegexMD5Only               = regexp.MustCompile(`^[a-f0-9]{32}$`)
	RegexSha1Only              = regexp.MustCompile(`^[a-f0-9]{40}$`)
	RegexTimestamps            = regexp.MustCompile(`1[0-9]{9}`)
	RegexNonInts               = regexp.MustCompile(`[^0-9]+`)
	RegexNonNumbers            = regexp.MustCompile(`[^0-9-]+`)
	RegexNonAlphaNumeric       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	RegexNonAlphaNumericSpace  = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	RegexMultipleNewLines      = regexp.MustCompile(`[\n]{3,}`)
	RegexNewLine               = regexp.MustCompile(`[\n\r]`)
	RegexSpaces                = regexp.MustCompile(`[\s]`)
	RegexFilterEmptyCharacters = regexp.MustCompile(`[\p{Cf}\p{Co}\p{Cs}\p{Cc}\p{C}\p{Braille}]`)
	RegexSmallRomanOnly        = regexp.MustCompile(`^[IVX]+$`)
)
