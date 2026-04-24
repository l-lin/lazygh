package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	actionsPopupFallbackWidth = 60
	actionsPopupMinWidth      = 40
	actionsPopupMinHeight     = 6
)

func (program *Program) layoutActionsPopupViews(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	contentMaxY := program.layoutContentHeight(maxY)
	totalWidth := boundedHalfWidth(maxX, actionsPopupMinWidth, actionsPopupFallbackWidth)
	totalHeight := maxInt(actionsPopupMinHeight, len(program.currentActionsPopupActions())+2)
	if totalHeight > contentMaxY-2 {
		totalHeight = maxInt(3, contentMaxY-2)
	}
	frame := centeredOverlayFrame(maxX, contentMaxY, totalWidth, totalHeight)

	popupView, err := gui.SetView(viewActionsPopupName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureActionsPopupView(popupView)
	program.renderActionsPopupView(popupView)
	_, err = gui.SetViewOnTop(viewActionsPopupName)
	if isUnknownViewError(err) {
		return nil
	}
	return err
}

func (program *Program) layoutActionsPopupSearchView(gui *gocui.Gui) error {
	view, err := program.layoutBottomPromptView(gui, viewActionsPopupSearchName)
	if err != nil {
		return err
	}

	program.configureActionsPopupSearchView(view)
	program.renderActionsPopupSearchView(view)
	_, err = gui.SetViewOnTop(viewActionsPopupSearchName)
	if isUnknownViewError(err) {
		return nil
	}
	return err
}

func (program *Program) configureActionsPopupView(view *gocui.View) {
	configureFramedOverlayView(view, program.actionsPopupTitle(), program.actionsPopupFooter())
	view.Wrap = false
	view.Editable = false
	view.Highlight = true
	view.HighlightInactive = true
	if program.usesManualSelectedLineRendering(program.model.ActionsPopupSearchQuery()) {
		view.Highlight = false
		view.HighlightInactive = false
	}
	view.SelBgColor = gocui.GetColor(theme.SelectedLineBackgroundHex)
	view.SelFgColor = gocui.GetColor(theme.ActiveTextHex)
	view.InactiveViewSelBgColor = gocui.GetColor(theme.SelectedLineBackgroundHex)
}

func (program *Program) renderActionsPopupView(view *gocui.View) {
	if view == nil || !program.model.ActionsPopupVisible() {
		return
	}

	view.Clear()
	actions := program.currentActionsPopupActions()
	filteredIndexes := program.model.ActionsPopupFilteredActionIndexes()
	query := program.model.ActionsPopupSearchQuery()
	if len(filteredIndexes) == 0 {
		fmt.Fprintln(view, "No matching actions.")
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

	selectedVisibleIndex := program.model.ActionsPopupSelectedVisibleIndex()
	showSelectedLine := program.usesManualSelectedLineRendering(query)
	for visibleIndex, index := range filteredIndexes {
		if index < 0 || index >= len(actions) {
			continue
		}
		program.renderHighlightedLine(view, actions[index].label(), query, showSelectedLine && visibleIndex == selectedVisibleIndex)
	}

	if selectedVisibleIndex < 0 {
		selectedVisibleIndex = 0
	}
	program.selectListLine(view, selectedVisibleIndex, len(filteredIndexes))
}

func (program *Program) configureActionsPopupSearchView(view *gocui.View) {
	program.configureBottomPromptView(view, gocui.EditorFunc(program.editActionsPopupSearch), true)
}

func (program *Program) renderActionsPopupSearchView(view *gocui.View) {
	program.renderBottomPromptView(view, program.currentActionsPopupSearchText(), program.currentActionsPopupSearchCursor())
}

func (program *Program) actionsPopupTitle() string {
	message := strings.TrimSpace(program.actionsPopupErrorMessage)
	if message == "" {
		return "Actions"
	}
	return fmt.Sprintf("Actions · %s", message)
}

func (program *Program) actionsPopupFooter() string {
	query := strings.TrimSpace(program.model.ActionsPopupSearchQuery())
	if query == "" {
		return ""
	}

	actions := program.currentActionsPopupActions()
	filteredIndexes := program.model.ActionsPopupFilteredActionIndexes()
	return fmt.Sprintf("%d of %d actions", len(filteredIndexes), len(actions))
}

func (program *Program) currentActionsPopupSearchText() string {
	if program.actionsPopupSearchEditor != nil {
		return program.actionsPopupSearchEditor.Text()
	}
	return program.model.ActionsPopupSearchQuery()
}

func (program *Program) currentActionsPopupSearchCursor() int {
	if program.actionsPopupSearchEditor != nil {
		return program.actionsPopupSearchEditor.Cursor()
	}
	return utf8.RuneCountInString(program.model.ActionsPopupSearchQuery())
}
