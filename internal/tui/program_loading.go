package tui

import (
	"sort"
	"strings"
	"time"

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
		sortPullRequestsByMostRecentUpdate(pullRequests)
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

func sortPullRequestsByMostRecentUpdate(pullRequests []githubcli.PullRequest) {
	sort.SliceStable(pullRequests, func(left int, right int) bool {
		leftUpdatedAt, leftOK := pullRequestUpdatedAtTime(pullRequests[left])
		rightUpdatedAt, rightOK := pullRequestUpdatedAtTime(pullRequests[right])
		switch {
		case leftOK && rightOK:
			if leftUpdatedAt.Equal(rightUpdatedAt) {
				return false
			}
			return leftUpdatedAt.After(rightUpdatedAt)
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return false
		}
	})
}

func pullRequestUpdatedAtTime(pullRequest githubcli.PullRequest) (time.Time, bool) {
	updatedAt, actualErr := time.Parse(time.RFC3339, strings.TrimSpace(pullRequest.UpdatedAt))
	if actualErr != nil {
		return time.Time{}, false
	}
	return updatedAt, true
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
