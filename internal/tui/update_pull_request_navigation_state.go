package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type openedPullRequestNormalization struct {
	pullRequests              []githubdomain.PullRequest
	matchedOpenedSummary      githubdomain.PullRequest
	matchedOpenedSummaryKnown bool
}

func (program *Program) pinOpenedPullRequestSummary(tab PullRequestTab, summary githubdomain.PullRequest) {
	program.pinOpenedPullRequestSummaryState(tab, summary)
}

func (program *Program) openedPullRequestSummaryForTab(tab PullRequestTab) (githubdomain.PullRequest, bool) {
	if program.navigationState.openedPullRequestSummary == nil || program.navigationState.openedPullRequestTab != tab {
		return githubdomain.PullRequest{}, false
	}
	return *program.navigationState.openedPullRequestSummary, true
}

func (program *Program) applyLoadedPullRequestRows(tab PullRequestTab, pullRequests []githubdomain.PullRequest) {
	selectedKey := program.selectedPullRequestKey(tab)
	normalized := program.normalizeLoadedPullRequests(tab, pullRequests)
	normalized.pullRequests = sortPullRequests(normalized.pullRequests)
	if normalized.matchedOpenedSummaryKnown {
		program.pinOpenedPullRequestSummary(tab, normalized.matchedOpenedSummary)
	}

	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		store = store.withPullRequestFreshnessTrackingEnabled()
		store = store.withPullRequestTabMembership(tab, normalized.pullRequests)
		return store.withoutPullRequestFreshnessAbsentFromTabs(configuredPullRequestTabs(program.model))
	})
	rows := pullRequestStateRowsWithRepositoryStyle(program.runtimeConfig.displayConfig.RepositoryStyle, program.pullRequestListState(tab), normalized.pullRequests, nil)
	rows = program.decoratePullRequestRows(rows)
	program.setPullRequestsCount(tab, pullRequestSummaryRowCount(rows), true)
	program.model.SetPullRequestRows(tab, rows)
	program.selectPullRequestKeyOrFirst(tab, selectedKey)
	program.markCurrentPullRequestSeen()
}

func (program *Program) selectedPullRequestKey(tab PullRequestTab) string {
	if program == nil || program.model == nil {
		return ""
	}
	rows := program.model.PullRequestRows(tab)
	index := program.model.SelectedPullRequestIndex(tab)
	if index < 0 || index >= len(rows) || rows[index].Summary == nil {
		return ""
	}
	return pullRequestDetailKey(rows[index].Summary.Repository, rows[index].Summary.Number)
}

func (program *Program) refreshPullRequestUnreadMarkers() {
	if program == nil || program.model == nil {
		return
	}
	for _, tab := range program.model.PullRequestTabs() {
		program.model.SetPullRequestRows(tab, program.decoratePullRequestRows(program.model.PullRequestRows(tab)))
	}
}

func (program *Program) selectPullRequestKeyOrFirst(tab PullRequestTab, key string) {
	rows := program.model.PullRequestRows(tab)
	if strings.TrimSpace(key) != "" {
		for index, row := range rows {
			if row.Summary == nil {
				continue
			}
			actualKey := pullRequestDetailKey(row.Summary.Repository, row.Summary.Number)
			if actualKey == key {
				program.model.SelectPullRequestIndex(tab, index)
				return
			}
		}
	}
	program.model.SelectPullRequestIndex(tab, 0)
}

func (program *Program) normalizeLoadedPullRequests(tab PullRequestTab, pullRequests []githubdomain.PullRequest) openedPullRequestNormalization {
	openedSummary, ok := program.openedPullRequestSummaryForTab(tab)
	return normalizePullRequestsWithOpenedSummary(pullRequests, openedSummary, ok)
}

func normalizePullRequestsWithOpenedSummary(pullRequests []githubdomain.PullRequest, openedSummary githubdomain.PullRequest, hasOpenedSummary bool) openedPullRequestNormalization {
	if !hasOpenedSummary {
		return openedPullRequestNormalization{pullRequests: pullRequests}
	}

	updatedPullRequests := append([]githubdomain.PullRequest(nil), pullRequests...)
	for _, pullRequest := range updatedPullRequests {
		if !samePullRequestIdentity(pullRequest, openedSummary) {
			continue
		}
		return openedPullRequestNormalization{
			pullRequests:              updatedPullRequests,
			matchedOpenedSummary:      pullRequest,
			matchedOpenedSummaryKnown: true,
		}
	}

	return openedPullRequestNormalization{pullRequests: append([]githubdomain.PullRequest{openedSummary}, updatedPullRequests...)}
}
