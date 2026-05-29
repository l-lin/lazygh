package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) listPullRequests(tab PullRequestTab) ([]githubdomain.PullRequest, error) {
	search, ok := program.searchBackedPullRequestSearch(tab)
	if !ok {
		return nil, nil
	}
	return program.pullRequestListQueries.ListPullRequests(search.Command)
}
