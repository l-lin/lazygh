package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatSearch})
}

func (program *Program) previousPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatSearch, Reverse: true})
}
