package tui

import "strings"

type detailMotionClipboardResult struct {
	selection detailSelectionRange
	text      string
}

type detailMotionApplyResult struct {
	state     detailViewState
	clipboard *detailMotionClipboardResult
}

func (program *Program) applyDetailMotionRequested(message MsgDetailMotionRequested) []Cmd {
	state, searchQuery, searchActive, ok := program.detailMotionRequestState(message.Target)
	if !ok {
		return nil
	}

	switch message.Operation {
	case detailMotionOperationConsumePendingCharacter:
		if !state.hasPendingCharacterMotion() {
			return nil
		}
	case detailMotionOperationRepeatCharacter:
		if !state.hasLastCharacterMotion {
			return nil
		}
	case detailMotionOperationMoveToOtherSelectionEnd:
		if !state.mode.isVisual() {
			return nil
		}
	case detailMotionOperationRepeatSearch:
		if state.mode != detailNormalMode || searchActive || strings.TrimSpace(searchQuery) == "" {
			return nil
		}
	}

	return []Cmd{detailMotionCmd{
		Target:        message.Target,
		Operation:     message.Operation,
		Direction:     message.Direction,
		Mode:          message.Mode,
		Reverse:       message.Reverse,
		SelectionKind: message.SelectionKind,
		Rune:          message.Rune,
	}}
}

func (program *Program) applyDetailYankRequested(message MsgDetailYankRequested) []Cmd {
	switch message.Target {
	case detailMotionTargetBuildPopup:
		if program.pullRequestBuildRunPopup == nil {
			return nil
		}
		if program.pullRequestBuildRunPopup.viewState.mode.isVisual() {
			return program.applyCopyPullRequestBuildRunPopupContentRequested(MsgCopyPullRequestBuildRunPopupContentRequested{})
		}
		operation := detailMotionOperationArmPendingYank
		if program.pullRequestBuildRunPopup.viewState.hasPendingYank() {
			operation = detailMotionOperationFinishPendingYank
		}
		return []Cmd{detailMotionCmd{Target: message.Target, Operation: operation, SelectionKind: detailYankMotionLinewise}}
	default:
		if program.model == nil || program.model.Focus() != FocusDetailView || !program.model.PaneVisible(FocusDetailView) {
			return nil
		}
		if program.detailState.viewState.mode.isVisual() {
			return []Cmd{prepareSelectedDetailClipboardWriteCmd{Target: program.model.Focus()}}
		}
		operation := detailMotionOperationArmPendingYank
		if program.detailState.viewState.hasPendingYank() {
			operation = detailMotionOperationFinishPendingYank
		}
		return []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: operation, SelectionKind: detailYankMotionLinewise}}
	}
}

func (program *Program) detailMotionRequestState(target detailMotionTarget) (detailViewState, string, bool, bool) {
	switch target {
	case detailMotionTargetBuildPopup:
		if program.pullRequestBuildRunPopup == nil {
			return detailViewState{}, "", false, false
		}
		return program.pullRequestBuildRunPopup.viewState, program.pullRequestBuildRunPopup.searchQuery, program.pullRequestBuildRunPopup.searchActive, true
	default:
		if program.model == nil || program.model.Focus() != FocusDetailView || !program.model.PaneVisible(FocusDetailView) {
			return detailViewState{}, "", false, false
		}
		return program.detailState.viewState, program.model.DetailSearchQuery(), false, true
	}
}

func (program *Program) applyFocusDetailRenderedLineResolved(message MsgFocusDetailRenderedLineResolved) {
	searchQuery := program.model.DetailSearchQuery()
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		state = state.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
		return state.withFocusedLineAndSearchSynced(message.Document, message.ViewportHeight, message.RenderedLine, searchQuery)
	})
}

func (program *Program) applyDetailViewportResolved(message MsgDetailViewportResolved) {
	searchQuery := program.model.DetailSearchQuery()
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		state = state.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
		state = state.withViewportOperation(message.Document, message.ViewportHeight, message.Operation)
		return state.withSearchSynced(message.Document, searchQuery)
	})
}

func (program *Program) applyDetailMotionResolved(message MsgDetailMotionResolved) []Cmd {
	switch message.Target {
	case detailMotionTargetBuildPopup:
		popup := program.pullRequestBuildRunPopup
		if popup == nil {
			return nil
		}
		actual := applyDetailMotionToViewState(detailMotionApplyInput{
			state:          popup.viewState,
			document:       message.Document,
			viewportHeight: message.ViewportHeight,
			operation:      message.Operation,
			direction:      message.Direction,
			mode:           message.Mode,
			reverse:        message.Reverse,
			selectionKind:  message.SelectionKind,
			runeValue:      message.Rune,
			searchQuery:    popup.searchQuery,
			searchActive:   popup.searchActive,
		})
		program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
			return state.withViewState(actual.state)
		})
		return program.detailMotionClipboardCommands(message.Target, actual.clipboard)
	default:
		searchQuery := program.model.DetailSearchQuery()
		actual := detailMotionApplyResult{}
		program.updateDetailState(func(state detailStateModel) detailStateModel {
			state = state.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
			actual = applyDetailMotionToViewState(detailMotionApplyInput{
				state:          state.viewState,
				document:       message.Document,
				viewportHeight: message.ViewportHeight,
				operation:      message.Operation,
				direction:      message.Direction,
				mode:           message.Mode,
				reverse:        message.Reverse,
				selectionKind:  message.SelectionKind,
				runeValue:      message.Rune,
				searchQuery:    searchQuery,
			})
			return state.withViewState(actual.state)
		})
		return program.detailMotionClipboardCommands(message.Target, actual.clipboard)
	}
}

func (program *Program) applyDetailViewSyncPlanResolved(message MsgDetailViewSyncPlanResolved) {
	searchQuery := program.model.DetailSearchQuery()
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		state = state.synced(program.currentDetailIdentity(), message.Plan.document, message.ViewportHeight, searchQuery)
		if !message.Plan.focusLineKnown {
			return state
		}
		return state.withFocusedLineAndSearchSynced(message.Plan.document, message.ViewportHeight, message.Plan.focusLine, searchQuery)
	})
}

func (program *Program) detailMotionClipboardCommands(target detailMotionTarget, clipboard *detailMotionClipboardResult) []Cmd {
	if clipboard == nil {
		return nil
	}

	focus := FocusDetailView
	if program != nil && program.model != nil {
		focus = program.model.Focus()
	}

	selectionTarget := clipboardWriteSelectionDetail
	if target == detailMotionTargetBuildPopup {
		selectionTarget = clipboardWriteSelectionBuildPopup
	}
	return []Cmd{writeClipboardCmd{
		Text:            clipboard.text,
		SuccessMessage:  detailYankSuccessMessage,
		FailureMessage:  detailYankFailureMessage,
		Target:          focus,
		Selection:       clipboard.selection,
		SelectionTarget: selectionTarget,
	}}
}

type detailMotionApplyInput struct {
	state          detailViewState
	document       detailDocument
	viewportHeight int
	operation      detailMotionOperation
	direction      detailCharacterMotionDirection
	mode           detailCharacterMotionMode
	reverse        bool
	selectionKind  detailYankMotionSelectionKind
	runeValue      rune
	searchQuery    string
	searchActive   bool
}

func applyDetailMotionToViewState(input detailMotionApplyInput) detailMotionApplyResult {
	state := input.state
	state.sync(input.document, input.viewportHeight)
	state.syncSearch(input.document, input.searchQuery)

	if shouldFinishPendingYankForDetailMotion(input.operation) {
		if input.operation == detailMotionOperationRepeatCharacter && !state.hasLastCharacterMotion {
			return detailMotionApplyResult{state: state}
		}
		if input.operation == detailMotionOperationRepeatSearch {
			if input.searchActive || state.mode != detailNormalMode || strings.TrimSpace(input.searchQuery) == "" {
				return detailMotionApplyResult{state: state}
			}
		}
	}

	snapshot := detailYankSnapshot{}
	hadPendingYank := shouldFinishPendingYankForDetailMotion(input.operation) && state.hasPendingYank()
	if hadPendingYank {
		snapshot = newDetailYankSnapshot(state)
	}

	switch input.operation {
	case detailMotionOperationArmCharacter:
		state.armCharacterMotion(input.direction, input.mode)
	case detailMotionOperationConsumePendingCharacter:
		state.consumePendingCharacterMotion(input.document, input.viewportHeight, input.runeValue)
	case detailMotionOperationRepeatCharacter:
		state.repeatCharacterMotion(input.document, input.viewportHeight, input.reverse)
	case detailMotionOperationArmPendingYank:
		state.armPendingYank()
	case detailMotionOperationFinishPendingYank:
		// No-op: the clipboard request is derived after the state stays synced.
	case detailMotionOperationEnterVisualMode:
		state.enterVisualMode()
	case detailMotionOperationEnterLineVisualMode:
		state.enterLineVisualMode(input.document)
	case detailMotionOperationMoveToOtherSelectionEnd:
		state.moveToOtherSelectionEnd(input.document, input.viewportHeight)
	case detailMotionOperationFollowSubmittedSearch:
		trimmedQuery := strings.TrimSpace(input.searchQuery)
		if trimmedQuery == "" {
			state.syncSearch(input.document, "")
			return detailMotionApplyResult{state: state}
		}
		if input.reverse {
			state.followPreviousSearchMatch(input.document, trimmedQuery, input.viewportHeight)
			break
		}
		state.followSubmittedSearch(input.document, trimmedQuery, input.viewportHeight)
	case detailMotionOperationRepeatSearch:
		trimmedQuery := strings.TrimSpace(input.searchQuery)
		if trimmedQuery == "" || input.searchActive || state.mode != detailNormalMode {
			return detailMotionApplyResult{state: state}
		}
		if input.reverse {
			state.followPreviousSearchMatch(input.document, trimmedQuery, input.viewportHeight)
			break
		}
		state.followNextSearchMatch(input.document, trimmedQuery, input.viewportHeight)
	default:
		applyDetailMotionStateOperation(&state, input.document, input.viewportHeight, input.operation)
	}

	state.sync(input.document, input.viewportHeight)
	state.syncSearch(input.document, input.searchQuery)
	if !hadPendingYank {
		return detailMotionApplyResult{state: state}
	}

	selection, ok := detailSelectionForYankMotion(input.document, snapshot.cursor, state.cursor, input.selectionKind)
	state.restoreYankSnapshot(snapshot)
	state.pendingYank = false
	if !ok {
		return detailMotionApplyResult{state: state}
	}
	return detailMotionApplyResult{
		state: state,
		clipboard: &detailMotionClipboardResult{
			selection: selection,
			text:      selection.text(input.document),
		},
	}
}

func shouldFinishPendingYankForDetailMotion(operation detailMotionOperation) bool {
	switch operation {
	case detailMotionOperationConsumePendingCharacter,
		detailMotionOperationRepeatCharacter,
		detailMotionOperationFinishPendingYank,
		detailMotionOperationMoveLeft,
		detailMotionOperationMoveRight,
		detailMotionOperationMoveDown,
		detailMotionOperationMoveUp,
		detailMotionOperationMoveToRowStart,
		detailMotionOperationMoveToRowEnd,
		detailMotionOperationMoveToTop,
		detailMotionOperationMoveToBottom,
		detailMotionOperationMoveToNextWord,
		detailMotionOperationMoveToWordEnd,
		detailMotionOperationMoveToNextBigWord,
		detailMotionOperationMoveToBigWordEnd,
		detailMotionOperationMoveToPreviousWord,
		detailMotionOperationMoveToPreviousBigWord,
		detailMotionOperationPageDown,
		detailMotionOperationPageUp,
		detailMotionOperationFullPageDown,
		detailMotionOperationFullPageUp,
		detailMotionOperationRepeatSearch:
		return true
	default:
		return false
	}
}
