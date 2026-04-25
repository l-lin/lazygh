package tui

import "strings"

const (
	roundedPillLeftSeparator  = ""
	roundedPillRightSeparator = ""
)

func renderRoundedPill(text string, foregroundHex string, backgroundHex string) string {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return ""
	}
	if strings.TrimSpace(foregroundHex) == "" || strings.TrimSpace(backgroundHex) == "" {
		return trimmedText
	}

	separatorPrefix := foregroundColorEscape(backgroundHex)
	textPrefix := foregroundColorEscape(foregroundHex) + backgroundColorEscape(backgroundHex)
	return styleText(roundedPillLeftSeparator, separatorPrefix) + styleText(" "+trimmedText+" ", textPrefix) + styleText(roundedPillRightSeparator, separatorPrefix)
}
