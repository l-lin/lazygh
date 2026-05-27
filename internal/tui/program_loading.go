package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	program.executeWorkflowPlan(gui, program.sessionLoadPlan())
}

func (program *Program) maybeLoadActivePullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, program.model.ActivePullRequestTab())
}

func (program *Program) maybeLoadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	program.executeWorkflowPlan(gui, program.pullRequestListLoadPlan(tab))
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	tab := program.model.ActivePullRequestTab()
	program.executeWorkflowPlan(gui, program.pullRequestListReloadPlan(tab))
	_ = program.afterStateChange(gui)
}

func (program *Program) reloadNotifications(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	program.executeWorkflowPlan(gui, program.notificationReloadPlan())
	_ = program.afterStateChange(gui)
}

func (program *Program) listPullRequests(tab PullRequestTab) ([]githubdomain.PullRequest, error) {
	return program.pullRequestListQueries.ListPullRequests(program.pullRequestSearch(tab).Command)
}

func (program *Program) pullRequestRowsForTab(tab PullRequestTab, pullRequests []githubdomain.PullRequest, err error) []PullRequestRow {
	if err == nil {
		pullRequests = program.pullRequestsWithOpenedPullRequestSummary(tab, pullRequests)
	}
	return pullRequestStateRows(program.pullRequestListState(tab), pullRequests, err)
}

func (store *pullRequestListStore) pullRequestsLoadStarted(tab PullRequestTab) bool {
	switch tab {
	case MyPullRequestsTab:
		return store.myPullRequestsLoadStarted
	case RequestedPullRequestsTab:
		return store.requestedPullRequestsLoadStarted
	default:
		return store.additionalPullRequestsLoadStarted[tab]
	}
}

func (store *pullRequestListStore) setPullRequestsLoadStarted(tab PullRequestTab, value bool) {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsLoadStarted = value
	case RequestedPullRequestsTab:
		store.requestedPullRequestsLoadStarted = value
	default:
		store.additionalPullRequestsLoadStarted[tab] = value
	}
}

func (store *pullRequestListStore) setPullRequestsLoading(tab PullRequestTab, value bool) {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsLoading = value
	case RequestedPullRequestsTab:
		store.requestedPullRequestsLoading = value
	default:
		store.additionalPullRequestsLoading[tab] = value
	}
}

func (store *pullRequestListStore) setPullRequestsCount(tab PullRequestTab, count int, known bool) {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsCount = count
		store.myPullRequestsCountKnown = known
	case RequestedPullRequestsTab:
		store.requestedPullRequestsCount = count
		store.requestedPullRequestsCountKnown = known
	default:
		store.additionalPullRequestsCounts[tab] = pullRequestCountState{count: count, known: known}
	}
}

func (store *pullRequestListStore) resetPullRequestListLoadState() {
	if store == nil {
		return
	}

	store.myPullRequestsLoadStarted = false
	store.requestedPullRequestsLoadStarted = false
	store.myPullRequestsLoading = false
	store.requestedPullRequestsLoading = false
	store.myPullRequestsCount = 0
	store.myPullRequestsCountKnown = false
	store.requestedPullRequestsCount = 0
	store.requestedPullRequestsCountKnown = false
	store.additionalPullRequestsLoadStarted = map[PullRequestTab]bool{}
	store.additionalPullRequestsLoading = map[PullRequestTab]bool{}
	store.additionalPullRequestsCounts = map[PullRequestTab]pullRequestCountState{}
}
