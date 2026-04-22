package tui

import (
	"strings"
	"time"
)

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
