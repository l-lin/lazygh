package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgConnectedUserLoaded struct {
	User githubdomain.ConnectedUser
	Err  error
}

type MsgPullRequestsLoaded struct {
	Tab          PullRequestTab
	PullRequests []githubdomain.PullRequest
	Err          error
}

type MsgNotificationsLoaded struct {
	Notifications []githubdomain.Notification
	Err           error
}

type MsgPullRequestDetailLoaded struct {
	Summary                 githubdomain.PullRequest
	Detail                  githubdomain.PullRequestDetail
	Err                     error
	PendingReviewState      pendingPullRequestReviewState
	PendingReviewStateKnown bool
}

type MsgPullRequestDiffLoaded struct {
	Summary githubdomain.PullRequest
	RawDiff githubdomain.PullRequestDiff
	Err     error
}

type MsgIssueDetailLoaded struct {
	Repository string
	Number     int
	Detail     githubdomain.IssueDetail
	Err        error
}

type MsgReleaseDetailLoaded struct {
	Repository string
	ID         int
	Detail     githubdomain.ReleaseDetail
	Err        error
}

type MsgCurrentDetailImageHTMLLoaded struct {
	Source       detailImageHTMLSource
	RenderedHTML string
	Err          error
}

type MsgCurrentDetailImageLoaded struct {
	ImageURL string
	Image    loadedDetailImage
	Err      error
}

type MsgLoadingSpinnerTick struct{}

type MsgTransientErrorPopupExpired struct {
	Generation uint64
}

func (MsgConnectedUserLoaded) isMsg()          {}
func (MsgPullRequestsLoaded) isMsg()           {}
func (MsgNotificationsLoaded) isMsg()          {}
func (MsgPullRequestDetailLoaded) isMsg()      {}
func (MsgPullRequestDiffLoaded) isMsg()        {}
func (MsgIssueDetailLoaded) isMsg()            {}
func (MsgReleaseDetailLoaded) isMsg()          {}
func (MsgCurrentDetailImageHTMLLoaded) isMsg() {}
func (MsgCurrentDetailImageLoaded) isMsg()     {}
func (MsgLoadingSpinnerTick) isMsg()           {}
func (MsgTransientErrorPopupExpired) isMsg()   {}
