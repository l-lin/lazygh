package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	program.executeCmds(gui, program.sessionStore.planLoad(program, gui))
}

func (program *Program) maybeLoadActivePullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, program.model.ActivePullRequestTab())
}

func (program *Program) maybeLoadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	program.executeCmds(gui, program.pullRequestListStore.planLoad(program, gui, tab))
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	tab := program.model.ActivePullRequestTab()
	program.executeCmds(gui, program.pullRequestListStore.planReload(program, gui, tab))
	_ = program.afterStateChange(gui)
}

func (program *Program) reloadNotifications(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	program.executeCmds(gui, program.notificationStore.planReload(program, gui))
	_ = program.afterStateChange(gui)
}

func (program *Program) loadConnectedUser(gui *gocui.Gui) {
	user, err := program.sessionQueries.GetConnectedUser()
	program.dispatchAsync(gui, MsgConnectedUserLoaded{User: user, Err: err})
}

func (program *Program) loadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	pullRequests, err := program.listPullRequests(tab)
	program.dispatchAsync(gui, MsgPullRequestsLoaded{Tab: tab, PullRequests: pullRequests, Err: err})
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
