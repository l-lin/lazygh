package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

type detailMotionTarget int

const (
	detailMotionTargetDetail detailMotionTarget = iota
	detailMotionTargetBuildPopup
)

type detailMotionOperation int

const (
	detailMotionOperationArmCharacter detailMotionOperation = iota
	detailMotionOperationConsumePendingCharacter
	detailMotionOperationRepeatCharacter
	detailMotionOperationArmPendingYank
	detailMotionOperationFinishPendingYank
	detailMotionOperationMoveLeft
	detailMotionOperationMoveRight
	detailMotionOperationMoveDown
	detailMotionOperationMoveUp
	detailMotionOperationMoveToRowStart
	detailMotionOperationMoveToRowEnd
	detailMotionOperationMoveToTop
	detailMotionOperationMoveToBottom
	detailMotionOperationMoveToNextWord
	detailMotionOperationMoveToWordEnd
	detailMotionOperationMoveToNextBigWord
	detailMotionOperationMoveToBigWordEnd
	detailMotionOperationMoveToPreviousWord
	detailMotionOperationMoveToPreviousBigWord
	detailMotionOperationEnterVisualMode
	detailMotionOperationEnterLineVisualMode
	detailMotionOperationPageDown
	detailMotionOperationPageUp
	detailMotionOperationFullPageDown
	detailMotionOperationFullPageUp
	detailMotionOperationFollowSubmittedSearch
	detailMotionOperationRepeatSearch
)

type detailMotionCmd struct {
	Target        detailMotionTarget
	Operation     detailMotionOperation
	View          *gocui.View
	Direction     detailCharacterMotionDirection
	Mode          detailCharacterMotionMode
	Reverse       bool
	SelectionKind detailYankMotionSelectionKind
	Rune          rune
}

type detailMotionCommandRuntime struct {
	executeDetail     func(*gocui.Gui, detailMotionCmd)
	executeBuildPopup func(*gocui.Gui, detailMotionCmd)
}

func newDetailMotionCommandRuntime(program *Program) detailMotionCommandRuntime {
	if program == nil {
		return detailMotionCommandRuntime{}
	}
	return detailMotionCommandRuntime{
		executeDetail: func(gui *gocui.Gui, command detailMotionCmd) {
			switch command.Operation {
			case detailMotionOperationArmCharacter:
				_ = program.mutateDetailViewState(gui, command.View, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.armCharacterMotion(command.Direction, command.Mode)
				})
			case detailMotionOperationConsumePendingCharacter:
				_ = program.mutateDetailViewStateForYankMotion(gui, command.View, command.SelectionKind, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.consumePendingCharacterMotion(document, viewportHeight, command.Rune)
				})
			case detailMotionOperationRepeatCharacter:
				if !program.detailState.viewState.hasLastCharacterMotion {
					return
				}
				_ = program.mutateDetailViewStateForYankMotion(gui, command.View, command.SelectionKind, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.repeatCharacterMotion(document, viewportHeight, command.Reverse)
				})
			case detailMotionOperationArmPendingYank:
				_ = program.mutateDetailViewState(gui, command.View, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.armPendingYank()
				})
			case detailMotionOperationFinishPendingYank:
				_ = program.mutateDetailViewStateForYankMotion(gui, command.View, command.SelectionKind, func(detailDocument, int) {})
			case detailMotionOperationEnterVisualMode:
				_ = program.mutateDetailViewState(gui, command.View, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.enterVisualMode()
					program.syncCurrentDetailViewport(document, viewportHeight)
				})
			case detailMotionOperationEnterLineVisualMode:
				_ = program.mutateDetailViewState(gui, command.View, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.enterLineVisualMode(document)
					program.syncCurrentDetailViewport(document, viewportHeight)
				})
			default:
				selectionKind, ok := detailMotionSelectionKindForOperation(command.Operation)
				if !ok {
					return
				}
				_ = program.mutateDetailViewStateForYankMotion(gui, command.View, selectionKind, func(document detailDocument, viewportHeight int) {
					applyDetailMotionStateOperation(&program.detailState.viewState, document, viewportHeight, command.Operation)
				})
			}
		},
		executeBuildPopup: func(gui *gocui.Gui, command detailMotionCmd) {
			popup := program.pullRequestBuildRunPopup
			if popup == nil {
				return
			}

			switch command.Operation {
			case detailMotionOperationArmCharacter:
				_ = program.mutatePullRequestBuildRunPopupViewState(gui, command.View, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.armCharacterMotion(command.Direction, command.Mode)
				})
			case detailMotionOperationConsumePendingCharacter:
				_ = program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, command.View, command.SelectionKind, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.consumePendingCharacterMotion(document, viewportHeight, command.Rune)
				})
			case detailMotionOperationRepeatCharacter:
				if !popup.viewState.hasLastCharacterMotion {
					return
				}
				_ = program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, command.View, command.SelectionKind, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.repeatCharacterMotion(document, viewportHeight, command.Reverse)
				})
			case detailMotionOperationArmPendingYank:
				_ = program.mutatePullRequestBuildRunPopupViewState(gui, command.View, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.armPendingYank()
				})
			case detailMotionOperationFinishPendingYank:
				_ = program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, command.View, command.SelectionKind, func(*detailViewState, detailDocument, int) {})
			case detailMotionOperationEnterVisualMode:
				_ = program.mutatePullRequestBuildRunPopupViewState(gui, command.View, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.enterVisualMode()
					state.sync(document, viewportHeight)
				})
			case detailMotionOperationEnterLineVisualMode:
				_ = program.mutatePullRequestBuildRunPopupViewState(gui, command.View, func(state *detailViewState, document detailDocument, viewportHeight int) {
					state.enterLineVisualMode(document)
					state.sync(document, viewportHeight)
				})
			case detailMotionOperationFollowSubmittedSearch:
				_ = program.mutatePullRequestBuildRunPopupViewState(gui, command.View, func(state *detailViewState, document detailDocument, viewportHeight int) {
					if strings.TrimSpace(popup.searchQuery) == "" {
						state.syncSearch(document, "")
						return
					}
					state.followSubmittedSearch(document, popup.searchQuery, viewportHeight)
				})
			case detailMotionOperationRepeatSearch:
				if popup.searchActive || popup.viewState.mode != detailNormalMode {
					return
				}
				if strings.TrimSpace(popup.searchQuery) == "" {
					return
				}
				_ = program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, command.View, detailYankMotionCharacterInclusive, func(state *detailViewState, document detailDocument, viewportHeight int) {
					if command.Reverse {
						state.followPreviousSearchMatch(document, popup.searchQuery, viewportHeight)
						return
					}
					state.followNextSearchMatch(document, popup.searchQuery, viewportHeight)
				})
			default:
				selectionKind, ok := detailMotionSelectionKindForOperation(command.Operation)
				if !ok {
					return
				}
				_ = program.mutatePullRequestBuildRunPopupViewStateForYankMotion(gui, command.View, selectionKind, func(state *detailViewState, document detailDocument, viewportHeight int) {
					applyDetailMotionStateOperation(state, document, viewportHeight, command.Operation)
				})
			}
		},
	}
}

func detailMotionSelectionKindForOperation(operation detailMotionOperation) (detailYankMotionSelectionKind, bool) {
	switch operation {
	case detailMotionOperationMoveLeft,
		detailMotionOperationMoveRight,
		detailMotionOperationMoveToRowStart,
		detailMotionOperationMoveToRowEnd,
		detailMotionOperationMoveToWordEnd,
		detailMotionOperationMoveToBigWordEnd,
		detailMotionOperationMoveToPreviousWord,
		detailMotionOperationMoveToPreviousBigWord:
		return detailYankMotionCharacterInclusive, true
	case detailMotionOperationMoveToNextWord,
		detailMotionOperationMoveToNextBigWord:
		return detailYankMotionCharacterExclusive, true
	case detailMotionOperationMoveDown,
		detailMotionOperationMoveUp,
		detailMotionOperationMoveToTop,
		detailMotionOperationMoveToBottom,
		detailMotionOperationPageDown,
		detailMotionOperationPageUp,
		detailMotionOperationFullPageDown,
		detailMotionOperationFullPageUp:
		return detailYankMotionLinewise, true
	default:
		return detailYankMotionCharacterInclusive, false
	}
}

func applyDetailMotionStateOperation(state *detailViewState, document detailDocument, viewportHeight int, operation detailMotionOperation) {
	switch operation {
	case detailMotionOperationMoveLeft:
		state.moveLeft(document, viewportHeight)
	case detailMotionOperationMoveRight:
		state.moveRight(document, viewportHeight)
	case detailMotionOperationMoveDown:
		state.moveDown(document, viewportHeight)
	case detailMotionOperationMoveUp:
		state.moveUp(document, viewportHeight)
	case detailMotionOperationMoveToRowStart:
		state.moveToRowStart(document, viewportHeight)
	case detailMotionOperationMoveToRowEnd:
		state.moveToRowEnd(document, viewportHeight)
	case detailMotionOperationMoveToTop:
		state.moveToTop(document, viewportHeight)
	case detailMotionOperationMoveToBottom:
		state.moveToBottom(document, viewportHeight)
	case detailMotionOperationMoveToNextWord:
		state.moveToNextWord(document, viewportHeight)
	case detailMotionOperationMoveToWordEnd:
		state.moveToWordEnd(document, viewportHeight)
	case detailMotionOperationMoveToNextBigWord:
		state.moveToNextBigWord(document, viewportHeight)
	case detailMotionOperationMoveToBigWordEnd:
		state.moveToBigWordEnd(document, viewportHeight)
	case detailMotionOperationMoveToPreviousWord:
		state.moveToPreviousWord(document, viewportHeight)
	case detailMotionOperationMoveToPreviousBigWord:
		state.moveToPreviousBigWord(document, viewportHeight)
	case detailMotionOperationPageDown:
		state.pageDown(document, viewportHeight)
	case detailMotionOperationPageUp:
		state.pageUp(document, viewportHeight)
	case detailMotionOperationFullPageDown:
		state.fullPageDown(document, viewportHeight)
	case detailMotionOperationFullPageUp:
		state.fullPageUp(document, viewportHeight)
	}
}

func (command detailMotionCmd) execute(program *Program, gui *gocui.Gui) {
	executeDetailMotionCommand(newDetailMotionCommandRuntime(program), gui, command)
}

func executeDetailMotionCommand(runtime detailMotionCommandRuntime, gui *gocui.Gui, command detailMotionCmd) {
	switch command.Target {
	case detailMotionTargetBuildPopup:
		if runtime.executeBuildPopup != nil {
			runtime.executeBuildPopup(gui, command)
		}
	default:
		if runtime.executeDetail != nil {
			runtime.executeDetail(gui, command)
		}
	}
}
