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
			}
		},
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
