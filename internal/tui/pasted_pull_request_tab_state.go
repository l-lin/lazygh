package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

const pastedPullRequestsTabLabel = "Pasted"

type pastedPullRequestTabState struct {
	pullRequests []githubdomain.PullRequest
}

func (state pastedPullRequestTabState) withPullRequestAdded(summary githubdomain.PullRequest) pastedPullRequestTabState {
	if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return state
	}

	updatedPullRequests := make([]githubdomain.PullRequest, 0, len(state.pullRequests)+1)
	updatedPullRequests = append(updatedPullRequests, summary)
	for _, existing := range state.pullRequests {
		if samePullRequestIdentity(existing, summary) {
			continue
		}
		updatedPullRequests = append(updatedPullRequests, existing)
	}
	state.pullRequests = updatedPullRequests
	return state
}

func (state pastedPullRequestTabState) rows() []PullRequestRow {
	rows := make([]PullRequestRow, 0, len(state.pullRequests))
	for _, pullRequest := range state.pullRequests {
		rows = append(rows, pullRequestRow(pullRequest))
	}
	return rows
}

func (state pastedPullRequestTabState) rowCount() int {
	return len(state.pullRequests)
}

func (state pastedPullRequestTabState) tabSeed() (PullRequestTabSeed, bool) {
	rows := state.rows()
	if len(rows) == 0 {
		return PullRequestTabSeed{}, false
	}
	return PullRequestTabSeed{Label: pastedPullRequestsTabLabel, PullRequests: pullRequestItems(rows)}, true
}
