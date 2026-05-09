package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) nextPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatPullRequestBuildRunPopupSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
		return program.pullRequestBuildRunPopup.viewState.followNextSearchMatch(document, program.pullRequestBuildRunPopup.searchQuery, viewportHeight)
	})
}

func (program *Program) previousPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatPullRequestBuildRunPopupSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
		return program.pullRequestBuildRunPopup.viewState.followPreviousSearchMatch(document, program.pullRequestBuildRunPopup.searchQuery, viewportHeight)
	})
}

func (program *Program) repeatPullRequestBuildRunPopupSearch(gui *gocui.Gui, view *gocui.View, repeat func(detailDocument, int) bool) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil || popup.searchActive || popup.viewState.mode != detailNormalMode {
		return nil
	}
	if strings.TrimSpace(popup.searchQuery) == "" {
		return nil
	}

	actualMoved := false
	if err := program.mutatePullRequestBuildRunPopupViewStateWithoutRefresh(gui, view, func(_ *detailViewState, document detailDocument, viewportHeight int) {
		actualMoved = repeat(document, viewportHeight)
	}); err != nil {
		return err
	}
	if !actualMoved {
		return nil
	}

	return program.refreshViewsIfGUI(gui)
}
