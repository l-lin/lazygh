package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) mutatePullRequestBuildRunPopupViewState(gui *gocui.Gui, view *gocui.View, mutate func(*detailViewState, detailDocument, int)) error {
	if err := program.mutatePullRequestBuildRunPopupViewStateWithoutRefresh(gui, view, mutate); err != nil {
		return err
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) mutatePullRequestBuildRunPopupViewStateWithoutRefresh(gui *gocui.Gui, view *gocui.View, mutate func(*detailViewState, detailDocument, int)) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	mutate(&popup.viewState, document, viewportHeight)
	popup.viewState.sync(document, viewportHeight)
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveLeft(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveRight(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorDown(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveDown(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorUp(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveUp(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToRowStart(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToRowEnd(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.handleGoToTopPrefix(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToBottom(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToNextWord(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToWordEnd(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToNextBigWord(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToBigWordEnd(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToPreviousWord(document, viewportHeight)
	})
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.moveToPreviousBigWord(document, viewportHeight)
	})
}

func (program *Program) enterPullRequestBuildRunPopupVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.enterVisualMode()
		state.sync(document, viewportHeight)
	})
}

func (program *Program) enterPullRequestBuildRunPopupLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.enterLineVisualMode(document)
		state.sync(document, viewportHeight)
	})
}

func (program *Program) pagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.pageDown(document, viewportHeight)
	})
}

func (program *Program) pagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.pageUp(document, viewportHeight)
	})
}

func (program *Program) fullPagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.fullPageDown(document, viewportHeight)
	})
}

func (program *Program) fullPagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.fullPageUp(document, viewportHeight)
	})
}

func (program *Program) copyPullRequestBuildRunPopupContent(gui *gocui.Gui, view *gocui.View) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)

	if popup.viewState.mode.isVisual() {
		selectedText := popup.viewState.selectedText(document)
		var err error
		switch {
		case program.clipboardWriter == nil:
			err = ErrClipboardUnavailable
		default:
			err = program.clipboardWriter.WriteText(selectedText)
		}
		popup.viewState.exitVisualMode()
		if err == nil {
			program.setFeedback(program.model.Focus(), detailYankSuccessMessage)
		} else {
			program.setFeedback(program.model.Focus(), detailYankFailureMessage)
		}
		return program.refreshViewsIfGUI(gui)
	}

	trimmedRunURL := strings.TrimSpace(popup.runURL)
	var err error
	switch {
	case trimmedRunURL == "":
		err = ErrNoPullRequestURL
	case program.clipboardWriter == nil:
		err = ErrClipboardUnavailable
	default:
		err = program.clipboardWriter.WriteText(trimmedRunURL)
	}

	switch {
	case err == nil:
		program.setFeedback(program.model.Focus(), yankSuccessMessage)
	case errors.Is(err, ErrNoPullRequestURL):
		program.setFeedback(program.model.Focus(), yankUnavailableMessage)
	default:
		program.setFeedback(program.model.Focus(), yankFailureMessage)
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) openPullRequestBuildRunPopupLinkUnderCursor(gui *gocui.Gui, view *gocui.View) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}
	if !popup.viewState.consumeGoToTopPrefix() {
		popup.viewState.clearPendingPrefix()
		return nil
	}
	if program.linkOpener == nil {
		program.setFeedback(program.model.Focus(), openLinkOpenerUnavailableMessage)
		return program.refreshViewsIfGUI(gui)
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	url, ok := program.currentPullRequestBuildRunPopupLink(actualView)
	switch {
	case !ok:
		program.setFeedback(program.model.Focus(), openLinkUnavailableMessage)
	case program.linkOpener.Open(url) == nil:
		program.setFeedback(program.model.Focus(), openLinkSuccessMessage)
	default:
		program.setFeedback(program.model.Focus(), openLinkFailureMessage)
	}
	return program.refreshViewsIfGUI(gui)
}
