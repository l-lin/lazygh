package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
)

const (
	actionsPopupFallbackWidth      = 20
	actionsPopupMinWidth           = 20
	actionsPopupMinHeight          = 6
	actionsPopupCompactMinHeight   = 4
	actionsPopupSearchPromptPrefix = "> "
)

func (program *Program) layoutActionsPopupViews(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	chromeFrame := program.actionsPopupFrame(maxX, maxY)

	chromeView, err := gui.SetView(viewActionsPopupChromeName, chromeFrame.x0, chromeFrame.y0, chromeFrame.x1, chromeFrame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureActionsPopupChromeView(chromeView)
	program.renderActionsPopupChromeView(chromeView)

	listFrame := program.actionsPopupListFrame(maxX, maxY)
	popupView, err := gui.SetView(viewActionsPopupName, listFrame.x0, listFrame.y0, listFrame.x1, listFrame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureActionsPopupView(popupView)
	program.renderActionsPopupView(popupView)

	if _, err = gui.SetViewOnTop(viewActionsPopupChromeName); err != nil && !isUnknownViewError(err) {
		return err
	}
	_, err = gui.SetViewOnTop(viewActionsPopupName)
	if isUnknownViewError(err) {
		return nil
	}
	return err
}

func (program *Program) layoutActionsPopupSearchView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	frame := program.actionsPopupSearchFrame(maxX, maxY)
	view, err := gui.SetView(viewActionsPopupSearchName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
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

func (program *Program) actionsPopupFrame(maxX int, maxY int) paneFrame {
	contentMaxY := program.layoutContentHeight(maxY)
	totalWidth := boundedQuarterWidth(maxX, actionsPopupMinWidth, actionsPopupFallbackWidth)
	totalHeight := program.actionsPopupHeight(contentMaxY)
	return centeredOverlayFrame(maxX, contentMaxY, totalWidth, totalHeight)
}

func (program *Program) actionsPopupSearchFrame(maxX int, maxY int) paneFrame {
	popupFrame := program.actionsPopupFrame(maxX, maxY)
	return paneFrame{x0: popupFrame.x0, y0: popupFrame.y0, x1: popupFrame.x1, y1: popupFrame.y0 + 2}
}

func (program *Program) actionsPopupListFrame(maxX int, maxY int) paneFrame {
	popupFrame := program.actionsPopupFrame(maxX, maxY)
	return paneFrame{x0: popupFrame.x0, y0: popupFrame.y0 + 1, x1: popupFrame.x1, y1: popupFrame.y1}
}

func (program *Program) actionsPopupHeight(contentMaxY int) int {
	totalHeight := maxInt(actionsPopupMinHeight, program.currentActionsPopupRenderedLineCount()+3)
	if program.assigneePickerVisible() {
		totalHeight = maxInt(actionsPopupCompactMinHeight, actionsPopupCompactHeight(program.currentActionsPopupRenderedLineCount()))
		if program.assigneePickerLoading() {
			totalHeight = maxInt(totalHeight, actionsPopupCompactMinHeight+2)
		}
	}
	if totalHeight > contentMaxY-2 {
		totalHeight = maxInt(3, contentMaxY-2)
	}
	return totalHeight
}

func actionsPopupCompactHeight(renderedLineCount int) int {
	if renderedLineCount < 1 {
		return actionsPopupCompactMinHeight
	}
	return (renderedLineCount+1)/2 + 3
}

func (program *Program) configureActionsPopupChromeView(view *gocui.View) {
	configureFramedOverlayView(view, program.actionsPopupTitle(), program.actionsPopupFooter())
	view.Wrap = false
	view.Editable = false
	view.Highlight = false
	view.HighlightInactive = false
}

func (program *Program) renderActionsPopupChromeView(view *gocui.View) {
	if view == nil || !program.model.ActionsPopupVisible() {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprintln(view)
}

func (program *Program) configureActionsPopupView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Title = program.actionsPopupTitle()
	view.Footer = ""
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

	query := program.model.ActionsPopupSearchQuery()
	visibleLines := program.currentActionsPopupVisibleLines()
	if len(visibleLines) == 0 {
		fmt.Fprintln(view, program.emptyActionsPopupMessage())
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

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
	if view == nil || !program.model.ActionsPopupSearchActive() {
		return
	}

	view.Clear()
	prompt := actionsPopupSearchPromptPrefix + program.currentActionsPopupSearchText()
	fmt.Fprint(view, prompt)
	program.setInputCursor(view, prompt, program.currentActionsPopupSearchCursor()+utf8.RuneCountInString(actionsPopupSearchPromptPrefix))
}

func (program *Program) actionsPopupTitle() string {
	title := "Actions"
	if program.assigneePickerVisible() {
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
	if program.assigneePickerVisible() {
		return assigneePickerSearchFooterHint
	}

	query := strings.TrimSpace(program.model.ActionsPopupSearchQuery())
	itemLabel := "actions"
	if program.themePickerVisible() {
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
		if strings.TrimSpace(program.model.ActionsPopupSearchQuery()) == "" {
			return assigneePickerSearchFooterHint
		}
		return "No matching assignees."
	}
	return "No matching actions."
}
