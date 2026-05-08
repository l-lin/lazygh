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
	totalHeight := maxInt(actionsPopupMinHeight, program.currentActionsPopupRenderedLineCount()+2)
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
	if program.assigneePickerLoading() {
		fmt.Fprintln(view, strings.TrimSpace(program.loadingSpinnerFrame()))
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

	query := program.model.ActionsPopupSearchQuery()
	if len(program.model.ActionsPopupFilteredActionIndexes()) == 0 {
		fmt.Fprintln(view, program.emptyActionsPopupMessage())
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

	visibleLines := program.currentActionsPopupVisibleLines()
	selectedRenderedLine := program.currentActionsPopupSelectedRenderedLine()
	showSelectedLine := program.usesManualSelectedLineRendering(query)
	for visibleIndex, line := range visibleLines {
		program.renderItemLine(view, line.item(view.InnerWidth()), query, showSelectedLine && line.selectable && visibleIndex == selectedRenderedLine)
	}

	if selectedRenderedLine < 0 {
		selectedRenderedLine = 0
	}
	program.selectListLine(view, selectedRenderedLine, len(visibleLines))
}

func (program *Program) configureActionsPopupSearchView(view *gocui.View) {
	program.configureBottomPromptView(view, gocui.EditorFunc(program.editActionsPopupSearch), true)
}

func (program *Program) renderActionsPopupSearchView(view *gocui.View) {
	program.renderBottomPromptView(view, program.currentActionsPopupSearchText(), program.currentActionsPopupSearchCursor())
}

func (program *Program) actionsPopupTitle() string {
	title := "Actions"
	if program.assigneePickerVisible() || program.assigneePickerLoading() {
		title = assigneePickerTitle
	} else if program.themePickerVisible() {
		title = themePickerTitle
	} else if program.reactionPickerVisible() {
		title = reactionPickerTitle
	}

	message := strings.TrimSpace(program.actionsPopupErrorMessage)
	if message == "" {
		message = strings.TrimSpace(program.actionsPopupConfirmationMessage())
	}
	if message == "" {
		return title
	}
	return fmt.Sprintf("%s · %s", title, message)
}

func (program *Program) actionsPopupFooter() string {
	query := strings.TrimSpace(program.model.ActionsPopupSearchQuery())
	itemLabel := "actions"
	if program.assigneePickerVisible() || program.assigneePickerLoading() {
		itemLabel = "assignees"
	} else if program.themePickerVisible() {
		itemLabel = "themes"
	} else if program.reactionPickerVisible() {
		itemLabel = "reactions"
	}

	actions := program.currentActionsPopupActions()
	filteredIndexes := program.model.ActionsPopupFilteredActionIndexes()
	countSummary := ""
	if query != "" {
		countSummary = fmt.Sprintf("%d of %d %s", len(filteredIndexes), len(actions), itemLabel)
	}
	return countSummary
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

func (program *Program) emptyActionsPopupMessage() string {
	if program.assigneePickerVisible() {
		return "No matching assignees."
	}
	return "No matching actions."
}
