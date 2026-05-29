package tui

import "github.com/jesseduffield/gocui"

func (program *Program) syncPullRequestBuildRunPopupViewState(document detailDocument, viewportHeight int) {
	if program == nil || program.pullRequestBuildRunPopup == nil {
		return
	}
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withViewStateSynced(document, viewportHeight)
	})
}

func (program *Program) syncPullRequestBuildRunPopupShellState(view *gocui.View) {
	if program == nil || view == nil || program.pullRequestBuildRunPopup == nil {
		return
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withRenderStateSynced(document, viewPageSize(view))
	})
}

func (program *Program) currentPullRequestBuildRunPopupLinkSnapshot(view *gocui.View) (detailDocument, detailViewState, bool) {
	popup := program.pullRequestBuildRunPopup
	if program == nil || popup == nil {
		return detailDocument{}, detailViewState{}, false
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	state := popup.withViewStateSynced(document, viewPageSize(view))
	return document, state.viewState, true
}
