package tui

import (
	persistcache "codeberg.org/l-lin/lazygh/internal/cache"
	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

type persistentPullRequestCache interface {
	PullRequests(search appconfig.PullRequestSearch) ([]githubcli.PullRequest, bool, error)
	SavePullRequests(search appconfig.PullRequestSearch, pullRequests []githubcli.PullRequest) error
	PullRequestDetail(repository string, number int) (persistcache.CachedPullRequestDetail, bool, error)
	SavePullRequestDetail(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) error
	PullRequestDiff(repository string, number int) (persistcache.CachedPullRequestDiff, bool, error)
	SavePullRequestDiff(summary githubcli.PullRequest, diff githubcli.PullRequestDiff) error
	InvalidatePullRequest(repository string, number int) error
	Clear() error
	Close() error
}
