package tui

import "github.com/jesseduffield/gocui"

func (program *Program) syncPullRequestBuildRunPopupViewState(document detailDocument, viewportHeight int) {
	if program == nil || program.pullRequestBuildRunPopup == nil {
		return
	}
	program.pullRequestBuildRunPopup.viewState.sync(document, viewportHeight)
}

func (program *Program) syncPullRequestBuildRunPopupRenderState(view *gocui.View) {
	if program == nil || view == nil || program.pullRequestBuildRunPopup == nil {
		return
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(view))
	program.pullRequestBuildRunPopup.viewState.syncSearch(document, program.pullRequestBuildRunPopup.searchQuery)
}

func (program *Program) currentPullRequestBuildRunPopupLinkSnapshot(view *gocui.View) (detailDocument, detailViewState, bool) {
	popup := program.pullRequestBuildRunPopup
	if program == nil || popup == nil {
		return detailDocument{}, detailViewState{}, false
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	state := popup.viewState
	state.sync(document, viewPageSize(view))
	return document, state, true
}
