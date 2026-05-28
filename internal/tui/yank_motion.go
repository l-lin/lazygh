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

func detailYankText(document detailDocument, anchor detailPosition, target detailPosition, selectionKind detailYankMotionSelectionKind) (string, bool) {
	selection, ok := detailSelectionForYankMotion(document, anchor, target, selectionKind)
	if !ok {
		return "", false
	}
	return selection.text(document), true
}
