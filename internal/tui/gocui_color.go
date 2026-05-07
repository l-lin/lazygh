package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func gocuiColorOrDefault(hexColor string) gocui.Attribute {
	if strings.TrimSpace(hexColor) == "" {
		return gocui.ColorDefault
	}
	return gocui.GetColor(hexColor)
}

func foregroundColorEscapeForAttribute(attribute gocui.Attribute) string {
	color := attribute & gocui.AttrColorBits
	red, green, blue := color.RGB()
	if red < 0 || green < 0 || blue < 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%sm", trueColorANSIParameters(38, int(red), int(green), int(blue)))
}
