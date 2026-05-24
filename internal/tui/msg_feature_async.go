package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgActionsPopupAsyncGHCommandFinished struct {
	Err     error
	Success actionsPopupAsyncSuccess
}

type MsgNotificationMutationStarted struct {
	OptimisticRows []NotificationRow
	LoadingMessage string
}

type MsgNotificationMutationFinished struct {
	Snapshot               notificationMutationSnapshot
	SuccessFeedbackMessage string
	Err                    error
}

type MsgStoryReviewPrepared struct {
	Prepared preparedStoryReview
	Err      error
}

type MsgAssigneePickerSearchLoadingStarted struct {
	RequestID int
	Query     string
}

type MsgAssigneePickerSearchLoaded struct {
	RequestID int
	Query     string
	Results   []githubdomain.PullRequestAuthor
	Err       error
}

type MsgPullRequestBuildRunLoaded struct {
	Target       pullRequestBuildRunTarget
	RawRunOutput string
	Jobs         []githubdomain.PullRequestBuildRunJob
	JobsErr      error
	Err          error
}

type MsgPullRequestBuildRunJobLogLoaded struct {
	Repository   string
	Job          githubdomain.PullRequestBuildRunJob
	RawLogOutput string
	Err          error
}

func (MsgActionsPopupAsyncGHCommandFinished) isMsg() {}
func (MsgNotificationMutationStarted) isMsg()        {}
func (MsgNotificationMutationFinished) isMsg()       {}
func (MsgStoryReviewPrepared) isMsg()                {}
func (MsgAssigneePickerSearchLoadingStarted) isMsg() {}
func (MsgAssigneePickerSearchLoaded) isMsg()         {}
func (MsgPullRequestBuildRunLoaded) isMsg()          {}
func (MsgPullRequestBuildRunJobLogLoaded) isMsg()    {}
