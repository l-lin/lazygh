package tui

import (
	glamouransi "charm.land/glamour/v2/ansi"
	"github.com/l-lin/lazygh/internal/theme"
)

func syntaxColorAndStylePrefix(hexColor string, styles []string) string {
	return ansiTextStylePrefix(styles) + foregroundColorEscape(hexColor)
}

func ansiTextStylePrefix(styles []string) string {
	prefix := ""
	for _, style := range styles {
		switch style {
		case theme.TextStyleBold:
			prefix += ansiBold
		case theme.TextStyleItalic:
			prefix += ansiItalic
		case theme.TextStyleUnderline:
			prefix += underlineEscape
		}
	}
	return prefix
}

func applyTextStyles(style *glamouransi.StylePrimitive, styles []string) {
	if style == nil {
		return
	}

	style.Bold = boolPtr(textStyleEnabled(styles, theme.TextStyleBold))
	style.Italic = boolPtr(textStyleEnabled(styles, theme.TextStyleItalic))
	style.Underline = boolPtr(textStyleEnabled(styles, theme.TextStyleUnderline))
}

func textStyleEnabled(styles []string, expected string) bool {
	for _, style := range styles {
		if style == expected {
			return true
		}
	}
	return false
}
