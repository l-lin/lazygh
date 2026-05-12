package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type persistentPullRequestCache interface {
	PullRequests(search appconfig.PullRequestSearch) ([]githubdomain.PullRequestSummary, bool, error)
	SavePullRequests(search appconfig.PullRequestSearch, pullRequests []githubdomain.PullRequestSummary) error
	Notifications() ([]githubdomain.Notification, bool, error)
	SaveNotifications(notifications []githubdomain.Notification) error
	PullRequestDetail(repository string, number int) (persistcache.CachedPullRequestDetail, bool, error)
	SavePullRequestDetail(summary githubdomain.PullRequestSummary, detail githubdomain.PullRequestDetail) error
	PullRequestDiff(repository string, number int) (persistcache.CachedPullRequestDiff, bool, error)
	SavePullRequestDiff(summary githubdomain.PullRequestSummary, diff githubdomain.PullRequestDiff) error
	InvalidatePullRequest(repository string, number int) error
	Clear() error
	Close() error
}
