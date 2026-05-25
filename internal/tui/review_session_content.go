package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func renderReadOnlyTextView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}
