package tui

import "fmt"

func foregroundColorEscape(hexColor string) string {
	red, green, blue, ok := parseHexColor(hexColor)
	if !ok {
		return ""
	}

	return fmt.Sprintf("\x1b[%sm", trueColorANSIParameters(38, red, green, blue))
}

func trueColorANSIParameterSequence(colorMode int, hexColor string) string {
	red, green, blue, ok := parseHexColor(hexColor)
	if !ok {
		return ""
	}

	return trueColorANSIParameters(colorMode, red, green, blue)
}

func trueColorANSIParameters(colorMode int, red int, green int, blue int) string {
	return fmt.Sprintf("%d;2;%d;%d;%d", colorMode, red, green, blue)
}
