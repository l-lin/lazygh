package tui

import "github.com/l-lin/lazygh/internal/githubcli"

type legacySessionQueries interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
}

type legacyPullRequestListQueries interface {
	ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error)
}

type legacyNotificationQueries interface {
	ListNotifications() ([]githubcli.Notification, error)
	GetIssueDetail(repository string, number int) (githubcli.IssueDetail, error)
	GetReleaseDetail(repository string, id int) (githubcli.ReleaseDetail, error)
}

type legacyDetailQueries interface {
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error)
	GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error)
}

type legacyPullRequestMutations interface {
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

type legacyReviewMutations interface {
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

type legacyNotificationMutations interface {
	MarkNotificationRead(threadID string) error
	MarkNotificationDone(threadID string) error
	MarkAllNotificationsRead() (githubcli.NotificationBulkReadResult, error)
	MarkAllNotificationsDone(notifications []githubcli.Notification) (int, error)
}

type legacyReactionMutations interface {
	AddReaction(subjectID string, content githubcli.ReactionContent) error
	RemoveReaction(subjectID string, content githubcli.ReactionContent) error
}

type legacyBuildQueries interface {
	GetPullRequestBuildRun(repository string, check githubcli.PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(repository string, check githubcli.PullRequestStatusCheck) ([]githubcli.PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(repository string, check githubcli.PullRequestStatusCheck) (githubcli.PullRequestBuildRunJob, string, error)
}

type legacySessionQueriesAdapter struct{ legacy legacySessionQueries }
type legacyPullRequestListQueriesAdapter struct{ legacy legacyPullRequestListQueries }
type legacyNotificationQueriesAdapter struct{ legacy legacyNotificationQueries }
type legacyDetailQueriesAdapter struct{ legacy legacyDetailQueries }
type legacyPullRequestMutationsAdapter struct{ legacy legacyPullRequestMutations }
type legacyReviewMutationsAdapter struct{ legacy legacyReviewMutations }
type legacyNotificationMutationsAdapter struct{ legacy legacyNotificationMutations }
type legacyReactionMutationsAdapter struct{ legacy legacyReactionMutations }
type legacyBuildQueriesAdapter struct{ legacy legacyBuildQueries }

func appDepsFromCompatibilityLoader(loader any) AppDeps {
	if loader == nil {
		return AppDeps{}
	}
	if deps, ok := loader.(AppDeps); ok {
		return deps
	}

	deps := AppDeps{}
	if actual, ok := loader.(SessionQueries); ok {
		deps.SessionQueries = actual
	} else if actual, ok := loader.(legacySessionQueries); ok {
		deps.SessionQueries = legacySessionQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(PullRequestListQueries); ok {
		deps.PullRequestList = actual
	} else if actual, ok := loader.(legacyPullRequestListQueries); ok {
		deps.PullRequestList = legacyPullRequestListQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(NotificationQueries); ok {
		deps.NotificationQueries = actual
	} else if actual, ok := loader.(legacyNotificationQueries); ok {
		deps.NotificationQueries = legacyNotificationQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(DetailQueries); ok {
		deps.DetailQueries = actual
	} else if actual, ok := loader.(legacyDetailQueries); ok {
		deps.DetailQueries = legacyDetailQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(PullRequestMutations); ok {
		deps.PullRequestMutations = actual
	} else if actual, ok := loader.(legacyPullRequestMutations); ok {
		deps.PullRequestMutations = legacyPullRequestMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(ReviewMutations); ok {
		deps.ReviewMutations = actual
	} else if actual, ok := loader.(legacyReviewMutations); ok {
		deps.ReviewMutations = legacyReviewMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(NotificationMutations); ok {
		deps.NotificationMutations = actual
	} else if actual, ok := loader.(legacyNotificationMutations); ok {
		deps.NotificationMutations = legacyNotificationMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(ReactionMutations); ok {
		deps.ReactionMutations = actual
	} else if actual, ok := loader.(legacyReactionMutations); ok {
		deps.ReactionMutations = legacyReactionMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(BuildQueries); ok {
		deps.BuildQueries = actual
	} else if actual, ok := loader.(legacyBuildQueries); ok {
		deps.BuildQueries = legacyBuildQueriesAdapter{legacy: actual}
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
