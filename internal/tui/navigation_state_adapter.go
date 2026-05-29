package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) updateNavigationState(transition func(navigationStateModel) navigationStateModel) {
	if program == nil {
		return
	}
	program.navigationState = transition(program.navigationState)
}

func (program *Program) pinOpenedPullRequestSummaryState(tab PullRequestTab, summary githubdomain.PullRequest) {
	program.updateNavigationState(func(state navigationStateModel) navigationStateModel {
		return state.withOpenedPullRequestSummaryPinned(tab, summary)
	})
}

func (program *Program) clearOpenedPullRequestSummaryState() {
	program.updateNavigationState(func(state navigationStateModel) navigationStateModel {
		return state.withOpenedPullRequestSummaryCleared()
	})
}
