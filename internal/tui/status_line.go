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

	presenter := program.statusLinePresenter()
	if presenter.IsFailure() {
		view.FgColor = gocui.GetColor(theme.FailureHex)
	} else {
		view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	statusText := truncateStatusLineText(presenter.Text(), program.statusLineAvailableWidth(view))
	fmt.Fprint(view, statusText)
}

func (program *Program) statusLineAvailableWidth(view *gocui.View) int {
	if view == nil {
		return 0
	}

	availableWidth := view.InnerWidth()
	keyHintsWidth := 0
	if program != nil && program.gui != nil {
		keyHintsView, actualErr := program.gui.View(viewStatusLineKeyHintsName)
		if actualErr == nil && keyHintsView != nil {
			keyHintsWidth = keyHintsView.InnerWidth()
		}
	}
	return statusLineAvailableTextWidth(availableWidth, keyHintsWidth)
}

func statusLineAvailableTextWidth(statusLineWidth int, keyHintsWidth int) int {
	if keyHintsWidth > 0 {
		statusLineWidth -= keyHintsWidth + 1
	}
	return max(statusLineWidth, 0)
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
