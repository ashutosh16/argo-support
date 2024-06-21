package genai

import (
	"regexp"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/forPelevin/gomoji"
)

type Sanitizer struct {
	nonASCIIWithTicksAndCrosses *regexp.Regexp
	packSpace                   *regexp.Regexp
}

const (
	nonASCIIWithTicksAndCrossesPattern = "[^\x20-\x7F\u2713\u2717\u00D7]"
	packSpacePattern                   = `\s+`
)

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		nonASCIIWithTicksAndCrosses: regexp.MustCompile(nonASCIIWithTicksAndCrossesPattern),
		packSpace:                   regexp.MustCompile(packSpacePattern),
	}
}

// replaceUnicodeEscape replaces the literal string "\u001b" with the actual ESC character "\x1b".
// This is necessary because we're processing the output of a Java program that uses the Java Unicode escape and
// we need to convert it to the Go escape character.
// TODO: Do we need to handle other Unicode escape sequences from other languages?
func replaceUnicodeEscape(input string) string {
	return strings.Replace(input, "\\u001b", "\x1b", -1)
}

func removeEmptyLines(str string) string {
	var lines []string
	for _, line := range strings.Split(str, `\n`) {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, `\n`)
}

func (s *Sanitizer) Sanitize(content string) string {
	// Remove emojis
	content = gomoji.RemoveEmojis(content)

	// Remove ANSI escape codes
	content = replaceUnicodeEscape(content)
	content = stripansi.Strip(content)

	// Remove non-ASCII characters but keep ticks and crosses
	content = s.nonASCIIWithTicksAndCrosses.ReplaceAllString(content, "")

	// Remove multiple spaces
	content = s.packSpace.ReplaceAllString(content, " ")

	// Remove empty lines
	content = removeEmptyLines(content)

	return content
}
