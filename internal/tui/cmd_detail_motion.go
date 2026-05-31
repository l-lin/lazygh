package tui

import "github.com/jesseduffield/gocui"

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
	detailMotionOperationMoveToOtherSelectionEnd
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
	Direction     detailCharacterMotionDirection
	Mode          detailCharacterMotionMode
	Reverse       bool
	SelectionKind detailYankMotionSelectionKind
	Rune          rune
}

type detailMotionCommandRuntime struct {
	executeMessage                          func(*gocui.Gui, Msg) error
	resolveView                             func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument                   func(*gocui.View) detailDocument
	currentPullRequestBuildRunPopupDocument func(*gocui.View) detailDocument
}

func newDetailMotionCommandRuntime(program *Program) detailMotionCommandRuntime {
	if program == nil {
		return detailMotionCommandRuntime{}
	}
	return detailMotionCommandRuntime{
		executeMessage:                          program.executeRuntimeMessage,
		resolveView:                             program.resolveView,
		currentDetailDocument:                   program.currentDetailDocument,
		currentPullRequestBuildRunPopupDocument: program.currentPullRequestBuildRunPopupDocument,
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
	case detailMotionOperationMoveToOtherSelectionEnd:
		state.moveToOtherSelectionEnd(document, viewportHeight)
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
	if runtime.executeMessage == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	document := detailDocument{}
	switch command.Target {
	case detailMotionTargetBuildPopup:
		if runtime.currentPullRequestBuildRunPopupDocument == nil {
			return
		}
		if runtime.resolveView != nil {
			actualView = runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
		}
		document = runtime.currentPullRequestBuildRunPopupDocument(actualView)
	default:
		if runtime.currentDetailDocument == nil {
			return
		}
		if runtime.resolveView != nil {
			actualView = runtime.resolveView(gui, nil, viewDetailName)
		}
		document = runtime.currentDetailDocument(actualView)
	}

	selectionKind := command.SelectionKind
	if actualSelectionKind, ok := detailMotionSelectionKindForOperation(command.Operation); ok {
		selectionKind = actualSelectionKind
	}

	_ = runtime.executeMessage(gui, MsgDetailMotionResolved{
		Target:         command.Target,
		Operation:      command.Operation,
		Direction:      command.Direction,
		Mode:           command.Mode,
		Reverse:        command.Reverse,
		SelectionKind:  selectionKind,
		Rune:           command.Rune,
		Document:       document,
		ViewportHeight: viewPageSize(actualView),
	})
}
