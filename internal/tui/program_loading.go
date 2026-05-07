package tui

import (
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || program.connectedUserLoadStarted {
		return
	}

	program.connectedUserLoadStarted = true
	program.asyncRunner.Go(func() {
		program.loadConnectedUser(gui)
	})
}

func (program *Program) maybeLoadActivePullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, program.model.ActivePullRequestTab())
}

func (program *Program) maybeLoadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	if gui == nil || program.pullRequestsLoadStarted(tab) || program.model.ActivePullRequestTab() != tab {
		return
	}

	program.hydratePullRequestsFromCache(tab)
	if program.githubLoader == nil {
		return
	}

	program.setPullRequestsLoadStarted(tab, true)
	program.setPullRequestsLoading(tab, true)
	program.asyncRunner.Go(func() {
		program.loadPullRequests(gui, tab)
	})
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	tab := program.model.ActivePullRequestTab()
	program.hydratePullRequestsFromCache(tab)
	if program.githubLoader == nil {
		return
	}

	program.setPullRequestsLoadStarted(tab, true)
	program.setPullRequestsLoading(tab, true)
	program.asyncRunner.Go(func() {
		program.loadPullRequests(gui, tab)
	})

	_ = program.refreshViews(gui)
}

func (program *Program) loadConnectedUser(gui *gocui.Gui) {
	user, err := program.githubLoader.GetConnectedUser()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.model.SetUsers([]Item{connectedUserStateItem(user, err)})
		return program.refreshViews(gui)
	})
}

func (program *Program) loadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	pullRequests, err := program.listPullRequests(tab)
	if err == nil {
		program.cachePullRequests(tab, pullRequests)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.setPullRequestsLoading(tab, false)
		if err == nil {
			program.setPullRequestsCount(tab, len(pullRequests), true)
			program.model.SetPullRequestRows(tab, program.pullRequestRowsForTab(tab, pullRequests, nil))
			return program.refreshViews(gui)
		}

		if !program.shouldPreservePullRequestRowsOnRefreshError(tab) {
			program.setPullRequestsCount(tab, 0, false)
			program.model.SetPullRequestRows(tab, program.pullRequestRowsForTab(tab, nil, err))
		}
		return program.refreshViews(gui)
	})
}

func (program *Program) listPullRequests(tab PullRequestTab) ([]githubcli.PullRequest, error) {
	return program.githubLoader.ListPullRequests(program.pullRequestSearch(tab).Command)
}

func (program *Program) pullRequestRowsForTab(tab PullRequestTab, pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
	return pullRequestStateRows(program.pullRequestListState(tab), pullRequests, err)
}

func (program *Program) pullRequestsLoadStarted(tab PullRequestTab) bool {
	switch tab {
	case MyPullRequestsTab:
		return program.myPullRequestsLoadStarted
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsLoadStarted
	default:
		return program.additionalPullRequestsLoadStarted[tab]
	}
}

func (program *Program) setPullRequestsLoadStarted(tab PullRequestTab, value bool) {
	switch tab {
	case MyPullRequestsTab:
		program.myPullRequestsLoadStarted = value
	case RequestedPullRequestsTab:
		program.requestedPullRequestsLoadStarted = value
	default:
		program.additionalPullRequestsLoadStarted[tab] = value
	}
}

func (program *Program) setPullRequestsLoading(tab PullRequestTab, value bool) {
	switch tab {
	case MyPullRequestsTab:
		program.myPullRequestsLoading = value
	case RequestedPullRequestsTab:
		program.requestedPullRequestsLoading = value
	default:
		program.additionalPullRequestsLoading[tab] = value
	}
}

func (program *Program) setPullRequestsCount(tab PullRequestTab, count int, known bool) {
	switch tab {
	case MyPullRequestsTab:
		program.myPullRequestsCount = count
		program.myPullRequestsCountKnown = known
	case RequestedPullRequestsTab:
		program.requestedPullRequestsCount = count
		program.requestedPullRequestsCountKnown = known
	default:
		program.additionalPullRequestsCounts[tab] = pullRequestCountState{count: count, known: known}
	}
}
