package tui

import "github.com/jesseduffield/gocui"

func (program *Program) movePullRequestBuildRunPopupCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveLeft, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorRight(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveRight, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveDown, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveUp, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowStart, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowEnd, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToTop, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBottom, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextWord, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToWordEnd, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextBigWord, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBigWordEnd, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousWord, View: view}})
	return nil
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousBigWord, View: view}})
	return nil
}

func (program *Program) enterPullRequestBuildRunPopupVisualMode(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterVisualMode, View: view}})
	return nil
}

func (program *Program) enterPullRequestBuildRunPopupLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterLineVisualMode, View: view}})
	return nil
}

func (program *Program) pagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageDown, View: view}})
	return nil
}

func (program *Program) pagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageUp, View: view}})
	return nil
}

func (program *Program) fullPagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageDown, View: view}})
	return nil
}

func (program *Program) fullPagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageUp, View: view}})
	return nil
}

func (program *Program) copyPullRequestBuildRunPopupContent(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopyPullRequestBuildRunPopupContentRequested{View: view})
}

func (program *Program) openPullRequestBuildRunPopupLinkUnderCursor(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgOpenPullRequestBuildRunPopupLinkRequested{View: view})
}
