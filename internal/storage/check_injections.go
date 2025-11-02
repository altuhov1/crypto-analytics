package storage

import (
	"regexp"
)

var (
	dangerousCharsRegex = regexp.MustCompile(`[;'"\|\&\$><!\\]`)
	sqlInjectionRegex   = regexp.MustCompile(`(?i)(\bDROP\b|\bDELETE\b|\bUPDATE\b|\bINSERT\b|\bSELECT\b.*\bFROM\b|\bUNION\b.*\bSELECT\b|--|\/\*|\*\/|;)`)
)

func hasDangerousCharacters(input string) bool {
	return dangerousCharsRegex.MatchString(input) || sqlInjectionRegex.MatchString(input)
}
