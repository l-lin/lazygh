package tui

import (
	clip "github.com/l-lin/lazygh/internal/clipboard"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type SessionQueries interface {
	GetConnectedUser() (githubdomain.ConnectedUser, error)
}

type PullRequestListQueries interface {
	ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error)
}

type NotificationQueries interface {
	ListNotifications() ([]githubdomain.Notification, error)
	GetIssueDetail(repository string, number int) (githubdomain.IssueDetail, error)
	GetReleaseDetail(repository string, id int) (githubdomain.ReleaseDetail, error)
}

type DetailQueries interface {
	GetPullRequestDetail(repository string, number int) (githubdomain.PullRequestDetail, error)
	GetPullRequestDiff(repository string, number int) (githubdomain.PullRequestDiff, error)
	GetCommitDiff(repository string, commitOID string) (githubdomain.CommitDiff, error)
	GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error)
}

type PullRequestMutations interface {
	CommentOnPullRequest(repository string, number int, body string) error
	UpdatePullRequestComment(commentID string, body string) error
	DeletePullRequestComment(commentID string) error
	RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error
	OpenPullRequestInBrowser(repository string, number int) error
	ListAssignableUsers(repository string) ([]githubdomain.PullRequestAuthor, error)
	SearchAssignableUsers(repository string, query string) ([]githubdomain.PullRequestAuthor, error)
	UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error
	EditPullRequestTitle(repository string, number int, title string) error
	EditPullRequestDescription(repository string, number int, body string) error
	MarkPullRequestReadyForReview(repository string, number int) error
	ConvertPullRequestToDraft(repository string, number int) error
	ClosePullRequest(repository string, number int) error
	ReopenPullRequest(repository string, number int) error
	SquashMergePullRequest(repository string, number int) error
	MergePullRequestWhenReady(repository string, number int, pullRequestID string) error
	EnablePullRequestAutoMerge(repository string, number int) error
	DisablePullRequestAutoMerge(repository string, number int) error
	EnqueuePullRequest(pullRequestID string) error
	DequeuePullRequest(pullRequestID string) error
	UpdatePullRequestBranch(repository string, number int) error
}

type ReviewMutations interface {
	ApprovePullRequest(repository string, number int) error
	ReviewPullRequestWithComment(repository string, number int, body string) error
	RequestChangesOnPullRequest(repository string, number int, body string) error
	SubmitPullRequestReview(pullRequestReviewID string, event githubdomain.PullRequestReviewEvent, body string) error
	AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubdomain.PullRequestReviewThreadTarget) error
	AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error
	UpdatePullRequestReviewComment(commentID string, body string) error
	DeletePullRequestReviewComment(commentID string) error
	ResolvePullRequestReviewThread(threadID string) error
	UnresolvePullRequestReviewThread(threadID string) error
	StartPendingPullRequestReview(repository string, number int) (string, error)
	GetPendingPullRequestReviewID(repository string, number int) (string, bool, error)
	DeletePullRequestReview(pullRequestReviewID string) error
}

type NotificationMutations interface {
	MarkNotificationRead(threadID string) error
	MarkNotificationDone(threadID string) error
	MarkAllNotificationsRead() (githubdomain.NotificationBulkReadResult, error)
	MarkAllNotificationsDone(notifications []githubdomain.Notification) (int, error)
}

type ReactionMutations interface {
	AddReaction(subjectID string, content githubdomain.ReactionContent) error
	RemoveReaction(subjectID string, content githubdomain.ReactionContent) error
}

type BuildQueries interface {
	GetPullRequestBuildRun(repository string, check githubdomain.PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(repository string, check githubdomain.PullRequestStatusCheck) ([]githubdomain.PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.PullRequestBuildRunJob, string, error)
}

type MarkdownHTMLRenderer interface {
	RenderMarkdownHTML(repository string, markdown string) (string, error)
}

type AuthTokenProvider interface {
	GetAuthToken() (string, error)
}

type ClipboardReader = clip.Reader

type ClipboardWriter = clip.Writer

type ExternalEditor = externalEditor

type LinkOpener = linkOpener

type ThemePresetStore = themePresetStore

type AppDeps struct {
	SessionQueries        SessionQueries
	PullRequestList       PullRequestListQueries
	NotificationQueries   NotificationQueries
	DetailQueries         DetailQueries
	PullRequestMutations  PullRequestMutations
	ReviewMutations       ReviewMutations
	NotificationMutations NotificationMutations
	ReactionMutations     ReactionMutations
	BuildQueries          BuildQueries
	MarkdownHTMLRenderer  MarkdownHTMLRenderer
	AuthTokenProvider     AuthTokenProvider
	ClipboardReader       ClipboardReader
	ClipboardWriter       ClipboardWriter
	ExternalEditor        ExternalEditor
	LinkOpener            LinkOpener
	ThemePresetStore      ThemePresetStore
}

func (program *Program) hasSessionQueries() bool {
	return program != nil && program.sessionQueries != nil
}

func (program *Program) hasPullRequestListQueries() bool {
	return program != nil && program.pullRequestListQueries != nil
}

func (program *Program) hasNotificationQueries() bool {
	return program != nil && program.notificationQueries != nil
}

func (program *Program) hasDetailQueries() bool {
	return program != nil && program.detailQueries != nil
}

func (program *Program) hasPullRequestMutations() bool {
	return program != nil && program.pullRequestMutations != nil
}

func (program *Program) hasReviewMutations() bool {
	return program != nil && program.reviewMutations != nil
}

func (program *Program) hasNotificationMutations() bool {
	return program != nil && program.notificationMutations != nil
}

func (program *Program) hasReactionMutations() bool {
	return program != nil && program.reactionMutations != nil
}

func (program *Program) hasBuildQueries() bool {
	return program != nil && program.buildQueries != nil
}

func (program *Program) hasMarkdownHTMLRenderer() bool {
	return program != nil && program.markdownHTMLRenderer != nil
}

func (program *Program) hasAuthTokenProvider() bool {
	return program != nil && program.authTokenProvider != nil
}
