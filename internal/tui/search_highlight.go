package tui

import (
	"fmt"
	"regexp"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
)

func highlightSearchMatches(text string, query string) (string, int) {
	return highlightSearchMatchesWithBasePrefix(text, query, "")
}

func highlightSearchMatchesOnSelectedLine(text string, query string) (string, int) {
	selectedLinePrefix := ansiBold + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	selectedMatchPrefix := ansiBold + backgroundColorEscape(theme.SearchHighlightHex)
	return highlightSearchMatchesWithPrefixes(text, query, selectedLinePrefix, selectedMatchPrefix)
}

func highlightSearchMatchesWithBasePrefix(text string, query string, basePrefix string) (string, int) {
	return highlightSearchMatchesWithPrefixes(text, query, basePrefix, backgroundColorEscape(theme.SearchHighlightHex))
}

func highlightSearchMatchesWithPrefixes(text string, query string, basePrefix string, matchPrefix string) (string, int) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return applyPrefix(text, basePrefix), 0
	}

	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(trimmedQuery))
	matches := pattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return applyPrefix(text, basePrefix), 0
	}

	var builder strings.Builder
	previousIndex := 0
	for _, match := range matches {
		builder.WriteString(applyPrefix(text[previousIndex:match[0]], basePrefix))
		builder.WriteString(applyPrefix(text[match[0]:match[1]], matchPrefix))
		previousIndex = match[1]
	}
	builder.WriteString(applyPrefix(text[previousIndex:], basePrefix))
	return builder.String(), len(matches)
}

func applyPrefix(text string, prefix string) string {
	if text == "" || prefix == "" {
		return text
	}

	return prefix + text + ansiReset
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
