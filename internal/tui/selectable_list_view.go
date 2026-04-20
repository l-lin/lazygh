package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

type selectableListViewState struct {
	focus               Focus
	query               string
	items               []Item
	selectedVisibleLine int
}

func (program *Program) configureSelectableListView(view *gocui.View, focus Focus, title string, query string) {
	program.applyViewStyle(view, focus, title, true)
	if program.usesManualSelectedLineRendering(query) {
		view.Highlight = false
		view.HighlightInactive = false
	}
	view.Wrap = false
}

func (program *Program) renderSelectableListView(view *gocui.View, state selectableListViewState) {
	view.Clear()

	if len(state.items) == 0 && strings.TrimSpace(state.query) != "" {
		fmt.Fprintln(view, searchNoMatchesMessage(state.query))
		return
	}

	showSelectedLine := program.usesManualSelectedLineRendering(state.query) && program.shouldHighlightSelection(state.focus, true)
	for visibleIndex, item := range state.items {
		program.renderHighlightedLine(view, item.Title, state.query, showSelectedLine && visibleIndex == state.selectedVisibleLine)
	}

	program.selectListLine(view, state.selectedVisibleLine)
}

func (program *Program) renderHighlightedLine(view *gocui.View, text string, query string, selected bool) {
	var highlightedText string
	if selected {
		highlightedText, _ = highlightSearchMatchesOnSelectedLine(text, query)
	} else {
		highlightedText, _ = highlightSearchMatches(text, query)
	}
	fmt.Fprintln(view, highlightedText)
}

func (program *Program) usesManualSelectedLineRendering(query string) bool {
	return strings.TrimSpace(query) != ""
}

func (program *Program) selectListLine(view *gocui.View, selectedIndex int) {
	if view == nil || len(view.BufferLines()) == 0 {
		return
	}

	originY, cursorY := selectedListLinePosition(selectedIndex, view.InnerHeight())
	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}

func selectedListLinePosition(selectedIndex int, visibleHeight int) (int, int) {
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	originY := 0
	cursorY := selectedIndex
	if cursorY >= visibleHeight {
		originY = cursorY - visibleHeight + 1
		cursorY = visibleHeight - 1
	}
	if cursorY < 0 {
		cursorY = 0
	}

	return originY, cursorY
}
