package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func gocuiColorOrDefault(hexColor string) gocui.Attribute {
	if strings.TrimSpace(hexColor) == "" {
		return gocui.ColorDefault
	}
	return gocui.GetColor(hexColor)
}
