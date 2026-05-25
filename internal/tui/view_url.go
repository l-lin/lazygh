package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) OpenPullRequestByURL(rawURL string) error {
	if !program.hasDetailQueries() && !program.hasPullRequestListQueries() {
		return errors.New("github loader is unavailable")
	}

	summary, err := githubdomain.ParsePullRequestURL(rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestInBrowser(summary)
}

func (program *Program) openPullRequestInBrowser(summary githubdomain.PullRequest) error {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return errors.New("missing pull request identity")
	}

	summary.Repository.NameWithOwner = repository
	Update(program, MsgOpenPullRequestInBrowserView{Summary: summary})
	return nil
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

func (program *Program) pullRequestsWithOpenedPullRequestSummary(tab PullRequestTab, pullRequests []githubdomain.PullRequest) []githubdomain.PullRequest {
	openedSummary, ok := program.openedPullRequestSummaryForTab(tab)
	if !ok {
		return pullRequests
	}

	updatedPullRequests := append([]githubdomain.PullRequest(nil), pullRequests...)
	for index, pullRequest := range updatedPullRequests {
		if !samePullRequestIdentity(pullRequest, openedSummary) {
			continue
		}
		program.pinOpenedPullRequestSummary(tab, pullRequest)
		updatedPullRequests[index] = pullRequest
		return updatedPullRequests
	}

	return append([]githubdomain.PullRequest{openedSummary}, updatedPullRequests...)
}

func pullRequestSummaryRowCount(rows []PullRequestRow) int {
	count := 0
	for _, row := range rows {
		if row.Summary != nil {
			count++
		}
	}
	return count
}

func samePullRequestIdentity(left any, right any) bool {
	leftSummary, ok := toDomainPullRequestSummary(left)
	if !ok {
		return false
	}
	rightSummary, ok := toDomainPullRequestSummary(right)
	if !ok {
		return false
	}
	leftKey := pullRequestDetailKey(leftSummary.Repository, leftSummary.Number)
	rightKey := pullRequestDetailKey(rightSummary.Repository, rightSummary.Number)
	return leftKey != "" && leftKey == rightKey
}
