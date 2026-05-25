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
	actionsPopupSearchPromptPrefix = "> "
)

func (program *Program) layoutActionsPopupViews(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	presenter := program.actionsPopupPresenter()
	contentMaxY := program.layoutContentHeight(maxY)
	chromeFrame := presenter.frame(maxX, contentMaxY)

	chromeView, err := gui.SetView(viewActionsPopupChromeName, chromeFrame.x0, chromeFrame.y0, chromeFrame.x1, chromeFrame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureActionsPopupChromeView(chromeView)
	program.renderActionsPopupChromeView(chromeView)

	listFrame := presenter.listFrame(maxX, contentMaxY)
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
	presenter := program.actionsPopupPresenter()
	frame := presenter.searchFrame(maxX, program.layoutContentHeight(maxY))
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

func (program *Program) configureActionsPopupChromeView(view *gocui.View) {
	presenter := program.actionsPopupPresenter()
	configureFramedOverlayView(view, presenter.title(), presenter.footer())
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
	presenter := program.actionsPopupPresenter()
	program.configureBottomPromptView(view, nil, false)
	view.Title = presenter.title()
	view.Footer = ""
	view.Highlight = true
	view.HighlightInactive = true
	if program.usesManualSelectedLineRendering(presenter.searchQuery) {
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

	presenter := program.actionsPopupPresenter()
	query := presenter.searchQuery
	visibleLines := program.currentActionsPopupVisibleLines()
	if len(visibleLines) == 0 {
		fmt.Fprintln(view, presenter.emptyMessage())
		view.SetOrigin(0, 0)
		view.SetCursor(0, 0)
		return
	}

	selectedRenderedLine := program.currentActionsPopupSelectedRenderedLine()
	showSelectedLine := program.usesManualSelectedLineRendering(query)
	for visibleIndex, line := range visibleLines {
		if line.separator {
			originalForegroundColor := view.FgColor
			view.FgColor = gocui.GetColor(theme.InactiveBorderHex)
			fmt.Fprintln(view, strings.Repeat("─", maxInt(view.InnerWidth(), 1)))
			view.FgColor = originalForegroundColor
			continue
		}
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
	presenter := program.actionsPopupPresenter()
	prompt := actionsPopupSearchPromptPrefix + presenter.promptText()
	fmt.Fprint(view, prompt)
	program.setInputCursor(view, prompt, presenter.promptCursor()+utf8.RuneCountInString(actionsPopupSearchPromptPrefix))
}
