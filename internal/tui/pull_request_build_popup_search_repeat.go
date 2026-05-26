package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatSearch, View: view}})
	return nil
}

func (program *Program) previousPullRequestBuildRunPopupSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatSearch, View: view, Reverse: true}})
	return nil
}
