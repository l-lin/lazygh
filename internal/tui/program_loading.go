package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	program.executeWorkflowCommands(gui, program.sessionStore.planLoad(program, gui))
}

func (program *Program) maybeLoadActivePullRequests(gui *gocui.Gui) {
	program.maybeLoadPullRequests(gui, program.model.ActivePullRequestTab())
}

func (program *Program) maybeLoadPullRequests(gui *gocui.Gui, tab PullRequestTab) {
	program.executeWorkflowCommands(gui, program.pullRequestListStore.planLoad(program, gui, tab))
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	tab := program.model.ActivePullRequestTab()
	program.executeWorkflowCommands(gui, program.pullRequestListStore.planReload(program, gui, tab))
	_ = program.refreshViews(gui)
}

func (program *Program) loadConnectedUser(gui *gocui.Gui) {
	user, err := program.sessionQueries.GetConnectedUser()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		connectedUserLogin := ""
		if err == nil {
			connectedUserLogin = strings.TrimSpace(user.Login)
		}
		if program.connectedUserLogin != connectedUserLogin {
			program.connectedUserLogin = connectedUserLogin
			program.invalidatePullRequestDetailDocumentCache()
			program.invalidateReviewDiffRenderCache()
		}
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
			rows := program.pullRequestRowsForTab(tab, pullRequests, nil)
			program.setPullRequestsCount(tab, pullRequestSummaryRowCount(rows), true)
			program.model.SetPullRequestRows(tab, rows)
			program.selectOpenedPullRequestRow(tab)
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
	return program.pullRequestListQueries.ListPullRequests(program.pullRequestSearch(tab).Command)
}

func (program *Program) pullRequestRowsForTab(tab PullRequestTab, pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
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
