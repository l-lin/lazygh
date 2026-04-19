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
	actionsPopupTopPadding    = 2
)

func (program *Program) layoutActionsPopupViews(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	contentMaxY := program.layoutContentHeight(maxY)
	totalWidth := maxX / 2
	if totalWidth < actionsPopupMinWidth {
		totalWidth = actionsPopupFallbackWidth
	}
	if totalWidth > maxX-4 {
		totalWidth = max(10, maxX-4)
	}
	if totalWidth > maxX {
		totalWidth = maxX
	}
	if totalWidth < 1 {
		totalWidth = 1
	}

	totalHeight := max(actionsPopupMinHeight, len(program.currentActionsPopupActions())+4)
	if totalHeight > contentMaxY-2 {
		totalHeight = max(3, contentMaxY-2)
	}
	if totalHeight > contentMaxY {
		totalHeight = contentMaxY
	}

	x0 := clampCoordinate((maxX-totalWidth)/2, maxX)
	y0 := clampCoordinate((contentMaxY-totalHeight)/2, contentMaxY)
	x1 := x0 + totalWidth - 1
	y1 := y0 + totalHeight - 1
	if x1 >= maxX {
		x1 = maxX - 1
		x0 = clampCoordinate(x1-totalWidth+1, maxX)
	}
	if y1 >= contentMaxY {
		y1 = contentMaxY - 1
		y0 = clampCoordinate(y1-totalHeight+1, contentMaxY)
	}

	popupView, err := gui.SetView(viewActionsPopupName, x0, y0, x1, y1, 0)
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
	view.Title = program.actionsPopupTitle()
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocui.ColorDefault
	view.Wrap = false
	view.Editable = false
	view.Highlight = true
	view.HighlightInactive = true
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
	fmt.Fprintf(view, "%d of %d actions\n\n", len(filteredIndexes), len(actions))
	if len(filteredIndexes) == 0 {
		fmt.Fprintln(view, "No matching actions.")
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

	for _, index := range filteredIndexes {
		if index < 0 || index >= len(actions) {
			continue
		}
		fmt.Fprintln(view, actions[index].title)
	}

	selectedVisibleIndex := program.model.ActionsPopupSelectedVisibleIndex()
	if selectedVisibleIndex < 0 {
		selectedVisibleIndex = 0
	}
	view.SetOrigin(0, 0)
	view.SetCursor(0, actionsPopupTopPadding+selectedVisibleIndex)
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
