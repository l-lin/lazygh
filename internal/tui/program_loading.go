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

func (program *Program) maybeLoadMyPullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, MyPullRequestsTab)
}

func (program *Program) maybeLoadRequestedPullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, RequestedPullRequestsTab)
}

func (program *Program) maybeLoadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	if gui == nil || program.githubLoader == nil || program.pullRequestsLoadStarted(tab) || program.model.ActivePullRequestTab() != tab {
		return
	}

	program.setPullRequestsLoadStarted(tab, true)
	program.setPullRequestsLoading(tab, true)
	program.asyncRunner.Go(func() {
		program.loadPullRequests(gui, tab)
	})
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil {
		return
	}

	tab := program.model.ActivePullRequestTab()
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

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.setPullRequestsLoading(tab, false)
		program.setPullRequestsCount(tab, len(pullRequests), err == nil)
		program.model.SetPullRequestRows(tab, program.pullRequestRowsForTab(tab, pullRequests, err))
		return program.refreshViews(gui)
	})
}

func (program *Program) listPullRequests(tab PullRequestTab) ([]githubcli.PullRequest, error) {
	switch tab {
	case RequestedPullRequestsTab:
		return program.githubLoader.ListRequestedPullRequests()
	default:
		return program.githubLoader.ListMyPullRequests()
	}
}

func (program *Program) pullRequestRowsForTab(tab PullRequestTab, pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
	switch tab {
	case RequestedPullRequestsTab:
		return requestedPullRequestsStateRows(pullRequests, err)
	default:
		return myPullRequestsStateRows(pullRequests, err)
	}
}

func (program *Program) pullRequestsLoadStarted(tab PullRequestTab) bool {
	switch tab {
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsLoadStarted
	default:
		return program.myPullRequestsLoadStarted
	}
}

func (program *Program) setPullRequestsLoadStarted(tab PullRequestTab, value bool) {
	switch tab {
	case RequestedPullRequestsTab:
		program.requestedPullRequestsLoadStarted = value
	default:
		program.myPullRequestsLoadStarted = value
	}
}

func (program *Program) setPullRequestsLoading(tab PullRequestTab, value bool) {
	switch tab {
	case RequestedPullRequestsTab:
		program.requestedPullRequestsLoading = value
	default:
		program.myPullRequestsLoading = value
	}
}

func (program *Program) setPullRequestsCount(tab PullRequestTab, count int, known bool) {
	switch tab {
	case RequestedPullRequestsTab:
		program.requestedPullRequestsCount = count
		program.requestedPullRequestsCountKnown = known
	default:
		program.myPullRequestsCount = count
		program.myPullRequestsCountKnown = known
	}
}
