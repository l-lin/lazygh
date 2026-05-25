package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgConnectedUserLoadPlanned struct{}

type MsgPullRequestsLoadPlanned struct {
	Tab PullRequestTab
}

type MsgNotificationsLoadPlanned struct{}

type MsgPullRequestDetailLoadPlanned struct {
	Key string
}

type MsgPullRequestDiffLoadPlanned struct {
	Key string
}

type MsgIssueDetailLoadPlanned struct {
	Repository string
	Number     int
}

type MsgReleaseDetailLoadPlanned struct {
	Repository string
	ID         int
}

type MsgCurrentDetailImageHTMLLoadPlanned struct {
	SourceKey string
}

type MsgCurrentDetailImageLoadPlanned struct {
	ImageURL string
}

type MsgPullRequestDetailCacheHydrated struct {
	Summary githubdomain.PullRequest
	Result  pullRequestDetailResult
}

type MsgPullRequestDiffCacheHydrated struct {
	Summary githubdomain.PullRequest
	Result  pullRequestDiffResult
}

func (MsgConnectedUserLoadPlanned) isMsg()          {}
func (MsgPullRequestsLoadPlanned) isMsg()           {}
func (MsgNotificationsLoadPlanned) isMsg()          {}
func (MsgPullRequestDetailLoadPlanned) isMsg()      {}
func (MsgPullRequestDiffLoadPlanned) isMsg()        {}
func (MsgIssueDetailLoadPlanned) isMsg()            {}
func (MsgReleaseDetailLoadPlanned) isMsg()          {}
func (MsgCurrentDetailImageHTMLLoadPlanned) isMsg() {}
func (MsgCurrentDetailImageLoadPlanned) isMsg()     {}
func (MsgPullRequestDetailCacheHydrated) isMsg()    {}
func (MsgPullRequestDiffCacheHydrated) isMsg()      {}
