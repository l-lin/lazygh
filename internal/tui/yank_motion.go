package tui

import "github.com/jesseduffield/gocui"

type detailYankMotionSelectionKind int

const (
	detailYankMotionCharacterInclusive detailYankMotionSelectionKind = iota
	detailYankMotionCharacterExclusive
	detailYankMotionLinewise
)

type detailYankSnapshot struct {
	cursor                    detailPosition
	originRow                 int
	preferredColumn           int
	manualViewportScroll      bool
	preserveViewportSyncCount int
}

func newDetailYankSnapshot(state detailViewState) detailYankSnapshot {
	return detailYankSnapshot{
		cursor:                    state.cursor,
		originRow:                 state.originRow,
		preferredColumn:           state.preferredColumn,
		manualViewportScroll:      state.manualViewportScroll,
		preserveViewportSyncCount: state.preserveViewportSyncCount,
	}
}

func (state *detailViewState) restoreYankSnapshot(snapshot detailYankSnapshot) {
	state.cursor = snapshot.cursor
	state.originRow = snapshot.originRow
	state.preferredColumn = snapshot.preferredColumn
	state.manualViewportScroll = snapshot.manualViewportScroll
	state.preserveViewportSyncCount = snapshot.preserveViewportSyncCount
}

func (state *detailViewState) hasPendingYank() bool {
	return state.pendingYank
}

func (state *detailViewState) armPendingYank() {
	state.pendingKeySequence.clear()
	state.pendingCharacterMotion = detailPendingCharacterMotion{}
	state.pendingYank = true
}

func (program *Program) startDetailYank(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailYankRequested{Target: detailMotionTargetDetail})
}

func (program *Program) startPullRequestBuildRunPopupYank(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailYankRequested{Target: detailMotionTargetBuildPopup})
}

func (program *Program) copySelectedText(state *detailViewState, document detailDocument) {
	selection, _ := detailSelectionForCurrentMode(*state, document)
	err := program.writeTextToClipboard(state.selectedText(document))
	state.exitVisualMode()
	if err == nil {
		program.activateYankHighlight(state, selection)
		program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: detailYankSuccessMessage})
	} else {
		program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: detailYankFailureMessage})
	}
}

func (program *Program) copySelectedPullRequestBuildRunPopupText(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopyPullRequestBuildRunPopupContentRequested{})
}

func (program *Program) writeTextToClipboard(text string) error {
	if program.clipboardWriter == nil {
		return ErrClipboardUnavailable
	}
	return program.clipboardWriter.WriteText(text)
}

func (program *Program) finishPendingYank(document detailDocument, state *detailViewState, snapshot detailYankSnapshot, selectionKind detailYankMotionSelectionKind) {
	selection, ok := detailSelectionForYankMotion(document, snapshot.cursor, state.cursor, selectionKind)
	state.restoreYankSnapshot(snapshot)
	state.pendingYank = false
	if !ok {
		return
	}
	if err := program.writeTextToClipboard(selection.text(document)); err == nil {
		program.activateYankHighlight(state, selection)
		program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: detailYankSuccessMessage})
	} else {
		program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: detailYankFailureMessage})
	}
}

func detailYankText(document detailDocument, anchor detailPosition, target detailPosition, selectionKind detailYankMotionSelectionKind) (string, bool) {
	selection, ok := detailSelectionForYankMotion(document, anchor, target, selectionKind)
	if !ok {
		return "", false
	}
	return selection.text(document), true
}
