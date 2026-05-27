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

func (program *Program) applyFocusDetailRenderedLineResolved(message MsgFocusDetailRenderedLineResolved) {
	searchQuery := program.model.DetailSearchQuery()
	detailState := program.detailState.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
	detailState = detailState.withFocusedLine(message.Document, message.ViewportHeight, message.RenderedLine)
	detailState.viewState.syncSearch(message.Document, searchQuery)
	program.detailState = detailState
}

func (program *Program) applyDetailViewportResolved(message MsgDetailViewportResolved) {
	searchQuery := program.model.DetailSearchQuery()
	detailState := program.detailState.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
	switch message.Operation {
	case detailViewportOperationScrollDown:
		detailState.viewState.scrollDown(message.Document, message.ViewportHeight)
	case detailViewportOperationScrollUp:
		detailState.viewState.scrollUp(message.Document, message.ViewportHeight)
	case detailViewportOperationRecenter:
		detailState.viewState.recenter(message.Document, message.ViewportHeight)
		detailState.viewState.preserveViewportSyncCount++
	case detailViewportOperationPlaceTop:
		detailState.viewState.placeCursorAtViewportTop(message.Document, message.ViewportHeight)
		detailState.viewState.preserveViewportSyncCount++
	case detailViewportOperationPlaceBottom:
		detailState.viewState.placeCursorAtViewportBottom(message.Document, message.ViewportHeight)
		detailState.viewState.preserveViewportSyncCount++
	}
	detailState.viewState.syncSearch(message.Document, searchQuery)
	program.detailState = detailState
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
		popup.viewState = actual.state
		return program.detailMotionClipboardCommands(message.Target, actual.clipboard)
	default:
		searchQuery := program.model.DetailSearchQuery()
		detailState := program.detailState.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
		actual := applyDetailMotionToViewState(detailMotionApplyInput{
			state:          detailState.viewState,
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
		detailState.viewState = actual.state
		program.detailState = detailState
		return program.detailMotionClipboardCommands(message.Target, actual.clipboard)
	}
}

func (program *Program) applyDetailViewSyncPlanResolved(message MsgDetailViewSyncPlanResolved) {
	searchQuery := program.model.DetailSearchQuery()
	detailState := program.detailState.synced(program.currentDetailIdentity(), message.Plan.document, message.ViewportHeight, searchQuery)
	if message.Plan.focusLineKnown {
		detailState = detailState.withFocusedLine(message.Plan.document, message.ViewportHeight, message.Plan.focusLine)
		detailState.viewState.syncSearch(message.Plan.document, searchQuery)
	}
	program.detailState = detailState
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
