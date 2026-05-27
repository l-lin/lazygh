package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type openedPullRequestNormalization struct {
	pullRequests              []githubdomain.PullRequest
	matchedOpenedSummary      githubdomain.PullRequest
	matchedOpenedSummaryKnown bool
}

func (program *Program) pinOpenedPullRequestSummary(tab PullRequestTab, summary githubdomain.PullRequest) {
	summaryCopy := summary
	program.navigationState.openedPullRequestSummary = &summaryCopy
	program.navigationState.openedPullRequestTab = tab
}

func (program *Program) openedPullRequestSummaryForTab(tab PullRequestTab) (githubdomain.PullRequest, bool) {
	if program.navigationState.openedPullRequestSummary == nil || program.navigationState.openedPullRequestTab != tab {
		return githubdomain.PullRequest{}, false
	}
	return *program.navigationState.openedPullRequestSummary, true
}

func (program *Program) applyLoadedPullRequestRows(tab PullRequestTab, pullRequests []githubdomain.PullRequest) {
	normalized := program.normalizeLoadedPullRequests(tab, pullRequests)
	if normalized.matchedOpenedSummaryKnown {
		program.pinOpenedPullRequestSummary(tab, normalized.matchedOpenedSummary)
	}

	rows := pullRequestStateRows(program.pullRequestListState(tab), normalized.pullRequests, nil)
	program.setPullRequestsCount(tab, pullRequestSummaryRowCount(rows), true)
	program.model.SetPullRequestRows(tab, rows)
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
