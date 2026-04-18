package tui

import (
	"fmt"
	"regexp"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const ansiReset = "\x1b[0m"

func highlightSearchMatches(text string, query string) (string, int) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return text, 0
	}

	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(trimmedQuery))
	matches := pattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text, 0
	}

	var builder strings.Builder
	previousIndex := 0
	for _, match := range matches {
		builder.WriteString(text[previousIndex:match[0]])
		builder.WriteString(backgroundColorEscape(theme.SearchHighlightHex))
		builder.WriteString(text[match[0]:match[1]])
		builder.WriteString(ansiReset)
		previousIndex = match[1]
	}
	builder.WriteString(text[previousIndex:])
	return builder.String(), len(matches)
}

func countSearchMatches(text string, query string) int {
	_, matchCount := highlightSearchMatches(text, query)
	return matchCount
}

func backgroundColorEscape(hexColor string) string {
	red, green, blue, ok := parseHexColor(hexColor)
	if !ok {
		return ""
	}

	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", red, green, blue)
}

func parseHexColor(hexColor string) (int, int, int, bool) {
	trimmedColor := strings.TrimPrefix(strings.TrimSpace(hexColor), "#")
	if len(trimmedColor) != 6 {
		return 0, 0, 0, false
	}

	var red, green, blue int
	_, err := fmt.Sscanf(trimmedColor, "%02x%02x%02x", &red, &green, &blue)
	if err != nil {
		return 0, 0, 0, false
	}

	return red, green, blue, true
}
