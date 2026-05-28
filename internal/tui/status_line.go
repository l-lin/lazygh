package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/theme"
)

func (program *Program) configureStatusLineView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
}

func (program *Program) configureStatusLineKeyHintsView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
	view.FgColor = gocui.GetColor(theme.InactiveTitleHex)
}

func (program *Program) renderStatusLineView(view *gocui.View) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(program.statusLinePresenter().Text()))
}

func (program *Program) renderStatusLineKeyHintsView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}
