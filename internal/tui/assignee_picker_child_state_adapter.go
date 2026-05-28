package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) openAssigneePicker(target pullRequestAssigneePickerTarget) {
	if program == nil {
		return
	}
	program.actionsPopupWidget.assigneePicker = newAssigneePickerState(target, program.currentConnectedUserLogin(), program.currentConnectedUserName())
}

func (program *Program) updateAssigneePickerState(transition func(assigneePickerState) assigneePickerState) bool {
	if program == nil || program.actionsPopupWidget.assigneePicker == nil {
		return false
	}

	updatedState := transition(*program.actionsPopupWidget.assigneePicker)
	program.actionsPopupWidget.assigneePicker = &updatedState
	return true
}

func (program *Program) resetAssigneePickerSearch(query string) int {
	requestID := 0
	program.updateAssigneePickerState(func(state assigneePickerState) assigneePickerState {
		var updatedState assigneePickerState
		updatedState, requestID = state.withSearchReset(query)
		return updatedState
	})
	return requestID
}

func (program *Program) markAssigneePickerSearchLoading(query string) {
	program.updateAssigneePickerState(func(state assigneePickerState) assigneePickerState {
		return state.withSearchLoadingStarted(query)
	})
}

func (program *Program) applyAssigneePickerSearchLoadedState(query string, results []githubdomain.PullRequestAuthor) {
	program.updateAssigneePickerState(func(state assigneePickerState) assigneePickerState {
		return state.withSearchLoaded(query, results)
	})
}

func (program *Program) toggleAssigneePickerSelectionState(candidate githubdomain.PullRequestAuthor) {
	program.updateAssigneePickerState(func(state assigneePickerState) assigneePickerState {
		return state.withSelectionToggled(candidate)
	})
}
