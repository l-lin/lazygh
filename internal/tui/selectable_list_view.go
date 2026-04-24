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
		program.renderHighlightedLine(view, program.displayItemTitle(item), state.query, showSelectedLine && visibleIndex == state.selectedVisibleLine)
	}

	program.selectListLine(view, state.selectedVisibleLine, len(state.items))
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

func (program *Program) selectListLine(view *gocui.View, selectedIndex int, lineCount int) {
	if view == nil || lineCount < 1 {
		return
	}

	_, currentOriginY := view.Origin()
	originY, cursorY := visibleListLinePosition(selectedIndex, currentOriginY, view.InnerHeight(), lineCount)
	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}

func (program *Program) centerListLine(view *gocui.View, selectedIndex int, lineCount int) {
	if view == nil || lineCount < 1 {
		return
	}

	originY, cursorY := centeredListLinePosition(selectedIndex, view.InnerHeight(), lineCount)
	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}

func visibleListLinePosition(selectedIndex int, currentOriginY int, visibleHeight int, lineCount int) (int, int) {
	visibleHeight = maxInt(1, visibleHeight)
	lineCount = maxInt(1, lineCount)
	selectedIndex = clampIndex(selectedIndex, lineCount)
	maxOriginY := maxInt(0, lineCount-visibleHeight)
	currentOriginY = clampInt(currentOriginY, 0, maxOriginY)

	originY := currentOriginY
	if selectedIndex < originY {
		originY = selectedIndex
	}
	if selectedIndex >= originY+visibleHeight {
		originY = selectedIndex - visibleHeight + 1
	}
	originY = clampInt(originY, 0, maxOriginY)
	return originY, selectedIndex - originY
}

func centeredListLinePosition(selectedIndex int, visibleHeight int, lineCount int) (int, int) {
	visibleHeight = maxInt(1, visibleHeight)
	lineCount = maxInt(1, lineCount)
	selectedIndex = clampIndex(selectedIndex, lineCount)
	originY := centeredViewportOrigin(selectedIndex, visibleHeight, lineCount)
	return originY, selectedIndex - originY
}

func centeredViewportOrigin(selectedRow int, visibleHeight int, rowCount int) int {
	visibleHeight = maxInt(1, visibleHeight)
	rowCount = maxInt(1, rowCount)
	selectedRow = clampIndex(selectedRow, rowCount)
	maxOriginY := maxInt(0, rowCount-visibleHeight)
	originY := selectedRow - visibleHeight/2
	return clampInt(originY, 0, maxOriginY)
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
