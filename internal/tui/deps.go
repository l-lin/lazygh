package tui

import (
	clip "github.com/l-lin/lazygh/internal/clipboard"
	"github.com/l-lin/lazygh/internal/githubcli"
)

type SessionQueries interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
}

type PullRequestListQueries interface {
	ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error)
}

type NotificationQueries interface {
	ListNotifications() ([]githubcli.Notification, error)
	GetIssueDetail(repository string, number int) (githubcli.IssueDetail, error)
	GetReleaseDetail(repository string, id int) (githubcli.ReleaseDetail, error)
}

type DetailQueries interface {
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error)
	GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error)
}

type PullRequestMutations interface {
	CommentOnPullRequest(repository string, number int, body string) error
	UpdatePullRequestComment(commentID string, body string) error
	DeletePullRequestComment(commentID string) error
	RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error
	OpenPullRequestInBrowser(repository string, number int) error
	ListAssignableUsers(repository string) ([]githubcli.PullRequestAuthor, error)
	UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error
	EditPullRequestTitle(repository string, number int, title string) error
	EditPullRequestDescription(repository string, number int, body string) error
	MarkPullRequestReadyForReview(repository string, number int) error
	ConvertPullRequestToDraft(repository string, number int) error
	ClosePullRequest(repository string, number int) error
	ReopenPullRequest(repository string, number int) error
	SquashMergePullRequest(repository string, number int) error
}

type ReviewMutations interface {
	ApprovePullRequest(repository string, number int) error
	ReviewPullRequestWithComment(repository string, number int, body string) error
	RequestChangesOnPullRequest(repository string, number int, body string) error
	SubmitPullRequestReview(pullRequestReviewID string, event githubcli.PullRequestReviewEvent, body string) error
	AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubcli.PullRequestReviewThreadTarget) error
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
	MarkAllNotificationsRead() (githubcli.NotificationBulkReadResult, error)
	MarkAllNotificationsDone(notifications []githubcli.Notification) (int, error)
}

type ReactionMutations interface {
	AddReaction(subjectID string, content githubcli.ReactionContent) error
	RemoveReaction(subjectID string, content githubcli.ReactionContent) error
}

type BuildQueries interface {
	GetPullRequestBuildRun(repository string, check githubcli.PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(repository string, check githubcli.PullRequestStatusCheck) ([]githubcli.PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(repository string, check githubcli.PullRequestStatusCheck) (githubcli.PullRequestBuildRunJob, string, error)
}

type MarkdownHTMLRenderer interface {
	RenderMarkdownHTML(repository string, markdown string) (string, error)
}

type AuthTokenProvider interface {
	GetAuthToken() (string, error)
}

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

func appDepsFromCompatibilityLoader(loader any) AppDeps {
	if loader == nil {
		return AppDeps{}
	}

	deps := AppDeps{}
	if actual, ok := loader.(SessionQueries); ok {
		deps.SessionQueries = actual
	}
	if actual, ok := loader.(PullRequestListQueries); ok {
		deps.PullRequestList = actual
	}
	if actual, ok := loader.(NotificationQueries); ok {
		deps.NotificationQueries = actual
	}
	if actual, ok := loader.(DetailQueries); ok {
		deps.DetailQueries = actual
	}
	if actual, ok := loader.(PullRequestMutations); ok {
		deps.PullRequestMutations = actual
	}
	if actual, ok := loader.(ReviewMutations); ok {
		deps.ReviewMutations = actual
	}
	if actual, ok := loader.(NotificationMutations); ok {
		deps.NotificationMutations = actual
	}
	if actual, ok := loader.(ReactionMutations); ok {
		deps.ReactionMutations = actual
	}
	if actual, ok := loader.(BuildQueries); ok {
		deps.BuildQueries = actual
	}
	if actual, ok := loader.(MarkdownHTMLRenderer); ok {
		deps.MarkdownHTMLRenderer = actual
	}
	if actual, ok := loader.(AuthTokenProvider); ok {
		deps.AuthTokenProvider = actual
	}
	if actual, ok := loader.(ClipboardWriter); ok {
		deps.ClipboardWriter = actual
	}
	if actual, ok := loader.(ExternalEditor); ok {
		deps.ExternalEditor = actual
	}
	if actual, ok := loader.(LinkOpener); ok {
		deps.LinkOpener = actual
	}
	if actual, ok := loader.(ThemePresetStore); ok {
		deps.ThemePresetStore = actual
	}
	return deps
}
