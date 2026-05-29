package tui

func (program *Program) pendingKeySequenceTargetForView(viewName string) keySequenceTarget {
	if program == nil {
		return keySequenceTarget{}
	}

	switch viewName {
	case viewDetailName:
		return program.detailState.pendingKeySequenceTarget()
	case viewPullRequestBuildInfoName:
		if program.pullRequestBuildRunPopup != nil {
			return program.pullRequestBuildRunPopup.pendingKeySequenceTarget()
		}
	}

	return program.navigationState.pendingSelectionKeySequenceTarget()
}

func (program *Program) armPendingKeySequenceForView(viewName string, target keySequenceTarget) {
	switch viewName {
	case viewDetailName:
		program.updateDetailState(func(state detailStateModel) detailStateModel {
			return state.withPendingKeySequenceArmed(target)
		})
	case viewPullRequestBuildInfoName:
		if program.pullRequestBuildRunPopup != nil {
			program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
				return state.withPendingKeySequenceArmed(target)
			})
			return
		}
		fallthrough
	default:
		program.updateNavigationState(func(state navigationStateModel) navigationStateModel {
			return state.withPendingSelectionKeySequenceArmed(target)
		})
	}
}

func (program *Program) clearPendingKeySequenceForView(viewName string) {
	switch viewName {
	case viewDetailName:
		program.updateDetailState(func(state detailStateModel) detailStateModel {
			return state.withPendingKeySequenceCleared()
		})
	case viewPullRequestBuildInfoName:
		if program.pullRequestBuildRunPopup != nil {
			program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
				return state.withPendingKeySequenceCleared()
			})
			return
		}
		fallthrough
	default:
		program.clearPendingSelectionKeySequence()
	}
}

func (program *Program) clearPendingSelectionKeySequence() {
	program.updateNavigationState(func(state navigationStateModel) navigationStateModel {
		return state.withPendingSelectionKeySequenceCleared()
	})
}
