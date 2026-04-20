package tui

import "fmt"

func foregroundColorEscape(hexColor string) string {
	red, green, blue, ok := parseHexColor(hexColor)
	if !ok {
		return ""
	}

	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue)
}
