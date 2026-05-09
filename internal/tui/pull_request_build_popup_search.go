package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) pullRequestBuildRunPopupSearchActive() bool {
	return program != nil && program.pullRequestBuildRunPopup != nil && program.pullRequestBuildRunPopup.searchActive
}

func (program *Program) searchPromptVisible() bool {
	return program.model.SearchActive() || program.pullRequestBuildRunPopupSearchActive()
}

func (program *Program) startPullRequestBuildRunPopupSearch() {
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		popup.searchActive = true
		popup.viewState.clearPendingPrefix()
	}
}

func (program *Program) submitPullRequestBuildRunPopupSearch(gui *gocui.Gui) error {
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		popup.searchActive = false
		popup.searchQuery = program.currentSearchText()
	}
	program.searchEditor = nil
	if err := program.followSubmittedPullRequestBuildRunPopupSearch(gui); err != nil {
		return err
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) cancelPullRequestBuildRunPopupSearch(gui *gocui.Gui) error {
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		popup.searchActive = false
	}
	return program.closeSearch(gui)
}

func (program *Program) followSubmittedPullRequestBuildRunPopupSearch(gui *gocui.Gui) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, nil, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	if strings.TrimSpace(popup.searchQuery) == "" {
		popup.viewState.syncSearch(document, "")
		return nil
	}

	popup.viewState.followSubmittedSearch(document, popup.searchQuery, viewportHeight)
	return nil
}
