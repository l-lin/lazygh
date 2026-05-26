package tui

import "github.com/jesseduffield/gocui"

func (program *Program) movePullRequestBuildRunPopupCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveLeft}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorRight(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveRight}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveDown}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveUp}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowStart}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowEnd}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToTop}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBottom}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextWord}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToWordEnd}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextBigWord}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBigWordEnd}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousWord}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousBigWord}})
	return nil
}

func (program *Program) enterPullRequestBuildRunPopupVisualMode(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterVisualMode}})
	return nil
}

func (program *Program) enterPullRequestBuildRunPopupLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterLineVisualMode}})
	return nil
}

func (program *Program) pagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageDown}})
	return nil
}

func (program *Program) pagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageUp}})
	return nil
}

func (program *Program) fullPagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageDown}})
	return nil
}

func (program *Program) fullPagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageUp}})
	return nil
}

func (program *Program) copyPullRequestBuildRunPopupContent(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopyPullRequestBuildRunPopupContentRequested{})
}

func (program *Program) openPullRequestBuildRunPopupLinkUnderCursor(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgOpenPullRequestBuildRunPopupLinkRequested{})
}
