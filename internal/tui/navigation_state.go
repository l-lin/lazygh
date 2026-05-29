package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (state navigationStateModel) withReviewSession(reviewSession reviewSessionState) navigationStateModel {
	state.reviewSession = reviewSession
	return state
}

func (state navigationStateModel) withOpenedPullRequestSummaryPinned(tab PullRequestTab, summary githubdomain.PullRequest) navigationStateModel {
	summaryCopy := summary
	state.openedPullRequestSummary = &summaryCopy
	state.openedPullRequestTab = tab
	return state
}

func (state navigationStateModel) withOpenedPullRequestSummaryCleared() navigationStateModel {
	state.openedPullRequestSummary = nil
	state.openedPullRequestTab = MyPullRequestsTab
	return state
}
