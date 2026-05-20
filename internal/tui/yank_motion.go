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
	if program.model.Focus() == FocusDetailView && program.detailViewState.mode.isVisual() {
		return program.copySelectedDetailText(gui, view)
	}
	if program.detailViewState.hasPendingYank() {
		return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(detailDocument, int) {})
	}
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.armPendingYank()
	})
}

func (program *Program) startPullRequestBuildRunPopupYank(gui *gocui.Gui, view *gocui.View) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}
	if popup.viewState.mode.isVisual() {
		return program.copySelectedPullRequestBuildRunPopupText(gui, view)
	}
	if popup.viewState.hasPendingYank() {
		return program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(*detailViewState, detailDocument, int) {})
	}
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.armPendingYank()
	})
}

func (program *Program) copySelectedText(state *detailViewState, document detailDocument) {
	err := program.writeTextToClipboard(state.selectedText(document))
	state.exitVisualMode()
	if err == nil {
		program.setFeedback(program.model.Focus(), detailYankSuccessMessage)
	} else {
		program.setFeedback(program.model.Focus(), detailYankFailureMessage)
	}
}

func (program *Program) copySelectedPullRequestBuildRunPopupText(gui *gocui.Gui, view *gocui.View) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	program.copySelectedText(&popup.viewState, document)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) writeTextToClipboard(text string) error {
	if program.clipboardWriter == nil {
		return ErrClipboardUnavailable
	}
	return program.clipboardWriter.WriteText(text)
}

func (program *Program) mutateDetailViewStateForYankMotion(gui *gocui.Gui, view *gocui.View, selectionKind detailYankMotionSelectionKind, mutate func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	actualView := program.resolveView(gui, view, viewDetailName)
	viewportHeight := viewPageSize(actualView)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewportHeight)
	snapshot := newDetailYankSnapshot(program.detailViewState)
	pendingYank := program.detailViewState.hasPendingYank()
	mutate(document, viewportHeight)
	program.syncDetailViewState(document, viewportHeight)
	if pendingYank {
		program.finishPendingYank(document, &program.detailViewState, snapshot, selectionKind)
	}
	program.syncActionsPopupSearch()
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) mutatePullRequestBuildRunPopupViewStateForYankMotion(gui *gocui.Gui, view *gocui.View, selectionKind detailYankMotionSelectionKind, mutate func(*detailViewState, detailDocument, int)) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	snapshot := newDetailYankSnapshot(popup.viewState)
	pendingYank := popup.viewState.hasPendingYank()
	mutate(&popup.viewState, document, viewportHeight)
	popup.viewState.sync(document, viewportHeight)
	if pendingYank {
		program.finishPendingYank(document, &popup.viewState, snapshot, selectionKind)
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) finishPendingYank(document detailDocument, state *detailViewState, snapshot detailYankSnapshot, selectionKind detailYankMotionSelectionKind) {
	selectedText, ok := detailYankText(document, snapshot.cursor, state.cursor, selectionKind)
	state.restoreYankSnapshot(snapshot)
	state.pendingYank = false
	if !ok {
		return
	}
	if err := program.writeTextToClipboard(selectedText); err == nil {
		program.setFeedback(program.model.Focus(), detailYankSuccessMessage)
	} else {
		program.setFeedback(program.model.Focus(), detailYankFailureMessage)
	}
}

func detailYankText(document detailDocument, anchor detailPosition, target detailPosition, selectionKind detailYankMotionSelectionKind) (string, bool) {
	anchor = document.clampPosition(anchor)
	target = document.clampPosition(target)

	switch selectionKind {
	case detailYankMotionLinewise:
		startRow := document.rowIndexForPosition(anchor)
		endRow := document.rowIndexForPosition(target)
		if startRow > endRow {
			startRow, endRow = endRow, startRow
		}
		return document.rowSelectionText(startRow, endRow), true
	case detailYankMotionCharacterExclusive:
		if document.comparePositions(anchor, target) == 0 {
			return "", false
		}
		start, end := anchor, target
		if document.comparePositions(start, end) > 0 {
			start, end = end, start
		}
		end = document.moveLeft(end)
		if document.comparePositions(start, end) > 0 {
			return "", false
		}
		return document.selectionText(start, end), true
	default:
		if document.comparePositions(anchor, target) == 0 {
			return "", false
		}
		start, end := anchor, target
		if document.comparePositions(start, end) > 0 {
			start, end = end, start
		}
		return document.selectionText(start, end), true
	}
}
