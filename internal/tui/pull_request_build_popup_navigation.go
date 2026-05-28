package tui

import "github.com/jesseduffield/gocui"

func (program *Program) movePullRequestBuildRunPopupCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveLeft})
}

func (program *Program) movePullRequestBuildRunPopupCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveRight})
}

func (program *Program) movePullRequestBuildRunPopupCursorDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveDown})
}

func (program *Program) movePullRequestBuildRunPopupCursorUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveUp})
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowStart})
}

func (program *Program) movePullRequestBuildRunPopupCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToRowEnd})
}

func (program *Program) movePullRequestBuildRunPopupCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToTop})
}

func (program *Program) movePullRequestBuildRunPopupCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBottom})
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextWord})
}

func (program *Program) movePullRequestBuildRunPopupCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToWordEnd})
}

func (program *Program) movePullRequestBuildRunPopupCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToNextBigWord})
}

func (program *Program) movePullRequestBuildRunPopupCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToBigWordEnd})
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousWord})
}

func (program *Program) movePullRequestBuildRunPopupCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveToPreviousBigWord})
}

func (program *Program) enterPullRequestBuildRunPopupVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterVisualMode})
}

func (program *Program) enterPullRequestBuildRunPopupLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationEnterLineVisualMode})
}

func (program *Program) pagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageDown})
}

func (program *Program) pagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationPageUp})
}

func (program *Program) fullPagePullRequestBuildRunPopupDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageDown})
}

func (program *Program) fullPagePullRequestBuildRunPopupUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFullPageUp})
}

func (program *Program) copyPullRequestBuildRunPopupContent(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopyPullRequestBuildRunPopupContentRequested{})
}

func (program *Program) openPullRequestBuildRunPopupLinkUnderCursor(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgOpenPullRequestBuildRunPopupLinkRequested{})
}
