package tui

import (
	"strings"
	"time"
)

const underlineEscape = "\x1b[4m"

func styleText(text string, prefixes ...string) string {
	if text == "" {
		return ""
	}

	prefix := strings.Join(filterEmptyStrings(prefixes), "")
	if prefix == "" {
		return text
	}
	return prefix + text + ansiReset
}

func hyperlinkText(target string, text string, prefixes ...string) string {
	trimmedTarget := strings.TrimSpace(target)
	if trimmedTarget == "" {
		return styleText(text, prefixes...)
	}
	return openHyperlinkEscape(trimmedTarget) + styleText(text, prefixes...) + closeHyperlinkEscape()
}

func openHyperlinkEscape(target string) string {
	trimmedTarget := strings.TrimSpace(target)
	if trimmedTarget == "" {
		return ""
	}
	return "\x1b]8;;" + trimmedTarget + "\x1b\\"
}

func closeHyperlinkEscape() string {
	return "\x1b]8;;\x1b\\"
}

func filterEmptyStrings(values []string) []string {
	filteredValues := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filteredValues = append(filteredValues, value)
	}
	return filteredValues
}

func formatTimestamp(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "-"
	}

	parsedTime, err := time.Parse(time.RFC3339, trimmedValue)
	if err != nil {
		return trimmedValue
	}

	return parsedTime.UTC().Format("2006-01-02 15:04 UTC")
}

func runeCountInt(value string) int {
	return len([]rune(value))
}

func centeredText(title string, viewWidth int) string {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" || viewWidth <= len([]rune(trimmedTitle)) {
		return trimmedTitle
	}

	padding := viewWidth - len([]rune(trimmedTitle))
	leftPadding := padding / 2
	rightPadding := padding - leftPadding
	return strings.Repeat(" ", leftPadding) + trimmedTitle + strings.Repeat(" ", rightPadding)
}
