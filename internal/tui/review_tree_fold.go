package tui

import "github.com/jesseduffield/gocui"

func (program *Program) togglePullRequestFold(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgToggleReviewTreeRowVisibility{})
}

func (program *Program) closeAllReviewTreeFolds(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgSetAllReviewTreeFolds{Collapsed: true})
}

func (program *Program) openAllReviewTreeFolds(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgSetAllReviewTreeFolds{Collapsed: false})
}

func (program *Program) reviewTreeFoldBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible()
}
