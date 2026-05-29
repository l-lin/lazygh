package tui

func (program *Program) updatePastedPullRequestTabState(transition func(pastedPullRequestTabState) pastedPullRequestTabState) {
	if program == nil {
		return
	}
	program.pastedPullRequests = transition(program.pastedPullRequests)
}

func (program *Program) pastedPullRequestTab() (PullRequestTab, bool) {
	if program == nil || program.model == nil {
		return 0, false
	}
	for _, tab := range program.model.PullRequestTabs() {
		if program.model.PullRequestTabLabel(tab) == pastedPullRequestsTabLabel {
			return tab, true
		}
	}
	return 0, false
}

func (program *Program) isPastedPullRequestTab(tab PullRequestTab) bool {
	if program == nil || program.model == nil {
		return false
	}
	return program.model.PullRequestTabLabel(tab) == pastedPullRequestsTabLabel
}
