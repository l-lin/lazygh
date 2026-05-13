package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

type testLegacySessionQueries interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
}

type testLegacyPullRequestListQueries interface {
	ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error)
}

type testLegacyNotificationQueries interface {
	ListNotifications() ([]githubcli.Notification, error)
	GetIssueDetail(repository string, number int) (githubcli.IssueDetail, error)
	GetReleaseDetail(repository string, id int) (githubcli.ReleaseDetail, error)
}

type testLegacyDetailQueries interface {
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error)
	GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error)
}

type testLegacyPullRequestMutations interface {
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

type testLegacyReviewMutations interface {
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

type testLegacyNotificationMutations interface {
	MarkNotificationRead(threadID string) error
	MarkNotificationDone(threadID string) error
	MarkAllNotificationsRead() (githubcli.NotificationBulkReadResult, error)
	MarkAllNotificationsDone(notifications []githubcli.Notification) (int, error)
}

type testLegacyReactionMutations interface {
	AddReaction(subjectID string, content githubcli.ReactionContent) error
	RemoveReaction(subjectID string, content githubcli.ReactionContent) error
}

type testLegacyBuildQueries interface {
	GetPullRequestBuildRun(repository string, check githubcli.PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(repository string, check githubcli.PullRequestStatusCheck) ([]githubcli.PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(repository string, check githubcli.PullRequestStatusCheck) (githubcli.PullRequestBuildRunJob, string, error)
}

type testLegacySessionQueriesAdapter struct{ legacy testLegacySessionQueries }
type testLegacyPullRequestListQueriesAdapter struct {
	legacy testLegacyPullRequestListQueries
}
type testLegacyNotificationQueriesAdapter struct{ legacy testLegacyNotificationQueries }
type testLegacyDetailQueriesAdapter struct{ legacy testLegacyDetailQueries }
type testLegacyPullRequestMutationsAdapter struct {
	legacy testLegacyPullRequestMutations
}
type testLegacyReviewMutationsAdapter struct{ legacy testLegacyReviewMutations }
type testLegacyNotificationMutationsAdapter struct {
	legacy testLegacyNotificationMutations
}
type testLegacyReactionMutationsAdapter struct{ legacy testLegacyReactionMutations }
type testLegacyBuildQueriesAdapter struct{ legacy testLegacyBuildQueries }

func given_programWithTestGitHubDeps(model *Model, githubLoader any) *Program {
	return NewProgramWithModelAndDeps(model, given_testAppDeps(githubLoader))
}

func given_testAppDeps(loader any) AppDeps {
	if loader == nil {
		return AppDeps{}
	}
	if deps, ok := loader.(AppDeps); ok {
		return deps
	}

	deps := AppDeps{}
	if actual, ok := loader.(SessionQueries); ok {
		deps.SessionQueries = actual
	} else if actual, ok := loader.(testLegacySessionQueries); ok {
		deps.SessionQueries = testLegacySessionQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(PullRequestListQueries); ok {
		deps.PullRequestList = actual
	} else if actual, ok := loader.(testLegacyPullRequestListQueries); ok {
		deps.PullRequestList = testLegacyPullRequestListQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(NotificationQueries); ok {
		deps.NotificationQueries = actual
	} else if actual, ok := loader.(testLegacyNotificationQueries); ok {
		deps.NotificationQueries = testLegacyNotificationQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(DetailQueries); ok {
		deps.DetailQueries = actual
	} else if actual, ok := loader.(testLegacyDetailQueries); ok {
		deps.DetailQueries = testLegacyDetailQueriesAdapter{legacy: actual}
	}
	if actual, ok := loader.(PullRequestMutations); ok {
		deps.PullRequestMutations = actual
	} else if actual, ok := loader.(testLegacyPullRequestMutations); ok {
		deps.PullRequestMutations = testLegacyPullRequestMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(ReviewMutations); ok {
		deps.ReviewMutations = actual
	} else if actual, ok := loader.(testLegacyReviewMutations); ok {
		deps.ReviewMutations = testLegacyReviewMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(NotificationMutations); ok {
		deps.NotificationMutations = actual
	} else if actual, ok := loader.(testLegacyNotificationMutations); ok {
		deps.NotificationMutations = testLegacyNotificationMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(ReactionMutations); ok {
		deps.ReactionMutations = actual
	} else if actual, ok := loader.(testLegacyReactionMutations); ok {
		deps.ReactionMutations = testLegacyReactionMutationsAdapter{legacy: actual}
	}
	if actual, ok := loader.(BuildQueries); ok {
		deps.BuildQueries = actual
	} else if actual, ok := loader.(testLegacyBuildQueries); ok {
		deps.BuildQueries = testLegacyBuildQueriesAdapter{legacy: actual}
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

func (adapter testLegacySessionQueriesAdapter) GetConnectedUser() (githubdomain.ConnectedUser, error) {
	user, err := adapter.legacy.GetConnectedUser()
	if err != nil {
		return githubdomain.ConnectedUser{}, err
	}
	return githubcli.ToDomainConnectedUser(user), nil
}

func (adapter testLegacyPullRequestListQueriesAdapter) ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error) {
	pullRequests, err := adapter.legacy.ListPullRequests(commandArguments)
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequests(pullRequests), nil
}

func (adapter testLegacyNotificationQueriesAdapter) ListNotifications() ([]githubdomain.Notification, error) {
	notifications, err := adapter.legacy.ListNotifications()
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainNotifications(notifications), nil
}

func (adapter testLegacyNotificationQueriesAdapter) GetIssueDetail(repository string, number int) (githubdomain.IssueDetail, error) {
	detail, err := adapter.legacy.GetIssueDetail(repository, number)
	if err != nil {
		return githubdomain.IssueDetail{}, err
	}
	return githubcli.ToDomainIssueDetail(detail), nil
}

func (adapter testLegacyNotificationQueriesAdapter) GetReleaseDetail(repository string, id int) (githubdomain.ReleaseDetail, error) {
	detail, err := adapter.legacy.GetReleaseDetail(repository, id)
	if err != nil {
		return githubdomain.ReleaseDetail{}, err
	}
	return githubcli.ToDomainReleaseDetail(detail), nil
}

func (adapter testLegacyDetailQueriesAdapter) GetPullRequestDetail(repository string, number int) (githubdomain.PullRequestDetail, error) {
	detail, err := adapter.legacy.GetPullRequestDetail(repository, number)
	if err != nil {
		return githubdomain.PullRequestDetail{}, err
	}
	return githubcli.ToDomainPullRequestDetail(detail), nil
}

func (adapter testLegacyDetailQueriesAdapter) GetPullRequestDiff(repository string, number int) (githubdomain.PullRequestDiff, error) {
	diff, err := adapter.legacy.GetPullRequestDiff(repository, number)
	if err != nil {
		return githubdomain.PullRequestDiff{}, err
	}
	return githubcli.ToDomainPullRequestDiff(diff), nil
}

func (adapter testLegacyDetailQueriesAdapter) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	return adapter.legacy.GetPullRequestFileTeamOwners(repository, number, filePaths)
}

func (adapter testLegacyPullRequestMutationsAdapter) CommentOnPullRequest(repository string, number int, body string) error {
	return adapter.legacy.CommentOnPullRequest(repository, number, body)
}

func (adapter testLegacyPullRequestMutationsAdapter) UpdatePullRequestComment(commentID string, body string) error {
	return adapter.legacy.UpdatePullRequestComment(commentID, body)
}

func (adapter testLegacyPullRequestMutationsAdapter) DeletePullRequestComment(commentID string) error {
	return adapter.legacy.DeletePullRequestComment(commentID)
}

func (adapter testLegacyPullRequestMutationsAdapter) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	return adapter.legacy.RequestPullRequestReviewer(repository, number, reviewerLogin)
}

func (adapter testLegacyPullRequestMutationsAdapter) OpenPullRequestInBrowser(repository string, number int) error {
	return adapter.legacy.OpenPullRequestInBrowser(repository, number)
}

func (adapter testLegacyPullRequestMutationsAdapter) ListAssignableUsers(repository string) ([]githubdomain.PullRequestAuthor, error) {
	authors, err := adapter.legacy.ListAssignableUsers(repository)
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequestAuthors(authors), nil
}

func (adapter testLegacyPullRequestMutationsAdapter) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	return adapter.legacy.UpdatePullRequestAssignees(repository, number, addLogins, removeLogins)
}

func (adapter testLegacyPullRequestMutationsAdapter) EditPullRequestTitle(repository string, number int, title string) error {
	return adapter.legacy.EditPullRequestTitle(repository, number, title)
}

func (adapter testLegacyPullRequestMutationsAdapter) EditPullRequestDescription(repository string, number int, body string) error {
	return adapter.legacy.EditPullRequestDescription(repository, number, body)
}

func (adapter testLegacyPullRequestMutationsAdapter) MarkPullRequestReadyForReview(repository string, number int) error {
	return adapter.legacy.MarkPullRequestReadyForReview(repository, number)
}

func (adapter testLegacyPullRequestMutationsAdapter) ConvertPullRequestToDraft(repository string, number int) error {
	return adapter.legacy.ConvertPullRequestToDraft(repository, number)
}

func (adapter testLegacyPullRequestMutationsAdapter) ClosePullRequest(repository string, number int) error {
	return adapter.legacy.ClosePullRequest(repository, number)
}

func (adapter testLegacyPullRequestMutationsAdapter) ReopenPullRequest(repository string, number int) error {
	return adapter.legacy.ReopenPullRequest(repository, number)
}

func (adapter testLegacyPullRequestMutationsAdapter) SquashMergePullRequest(repository string, number int) error {
	return adapter.legacy.SquashMergePullRequest(repository, number)
}

func (adapter testLegacyReviewMutationsAdapter) ApprovePullRequest(repository string, number int) error {
	return adapter.legacy.ApprovePullRequest(repository, number)
}

func (adapter testLegacyReviewMutationsAdapter) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return adapter.legacy.ReviewPullRequestWithComment(repository, number, body)
}

func (adapter testLegacyReviewMutationsAdapter) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return adapter.legacy.RequestChangesOnPullRequest(repository, number, body)
}

func (adapter testLegacyReviewMutationsAdapter) SubmitPullRequestReview(pullRequestReviewID string, event githubdomain.PullRequestReviewEvent, body string) error {
	return adapter.legacy.SubmitPullRequestReview(pullRequestReviewID, githubcli.PullRequestReviewEventFromDomain(event), body)
}

func (adapter testLegacyReviewMutationsAdapter) AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubdomain.PullRequestReviewThreadTarget) error {
	return adapter.legacy.AddPullRequestReviewThread(pullRequestReviewID, body, githubcli.PullRequestReviewThreadTargetFromDomain(target))
}

func (adapter testLegacyReviewMutationsAdapter) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	return adapter.legacy.AddPullRequestReviewThreadReply(pullRequestReviewID, pullRequestReviewThreadID, body)
}

func (adapter testLegacyReviewMutationsAdapter) UpdatePullRequestReviewComment(commentID string, body string) error {
	return adapter.legacy.UpdatePullRequestReviewComment(commentID, body)
}

func (adapter testLegacyReviewMutationsAdapter) DeletePullRequestReviewComment(commentID string) error {
	return adapter.legacy.DeletePullRequestReviewComment(commentID)
}

func (adapter testLegacyReviewMutationsAdapter) ResolvePullRequestReviewThread(threadID string) error {
	return adapter.legacy.ResolvePullRequestReviewThread(threadID)
}

func (adapter testLegacyReviewMutationsAdapter) UnresolvePullRequestReviewThread(threadID string) error {
	return adapter.legacy.UnresolvePullRequestReviewThread(threadID)
}

func (adapter testLegacyReviewMutationsAdapter) StartPendingPullRequestReview(repository string, number int) (string, error) {
	return adapter.legacy.StartPendingPullRequestReview(repository, number)
}

func (adapter testLegacyReviewMutationsAdapter) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	return adapter.legacy.GetPendingPullRequestReviewID(repository, number)
}

func (adapter testLegacyReviewMutationsAdapter) DeletePullRequestReview(pullRequestReviewID string) error {
	return adapter.legacy.DeletePullRequestReview(pullRequestReviewID)
}

func (adapter testLegacyNotificationMutationsAdapter) MarkNotificationRead(threadID string) error {
	return adapter.legacy.MarkNotificationRead(threadID)
}

func (adapter testLegacyNotificationMutationsAdapter) MarkNotificationDone(threadID string) error {
	return adapter.legacy.MarkNotificationDone(threadID)
}

func (adapter testLegacyNotificationMutationsAdapter) MarkAllNotificationsRead() (githubdomain.NotificationBulkReadResult, error) {
	result, err := adapter.legacy.MarkAllNotificationsRead()
	if err != nil {
		return githubdomain.NotificationBulkReadResult{}, err
	}
	return githubcli.ToDomainNotificationBulkReadResult(result), nil
}

func (adapter testLegacyNotificationMutationsAdapter) MarkAllNotificationsDone(notifications []githubdomain.Notification) (int, error) {
	return adapter.legacy.MarkAllNotificationsDone(githubcli.NotificationsFromDomain(notifications))
}

func (adapter testLegacyReactionMutationsAdapter) AddReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.legacy.AddReaction(subjectID, githubcli.ReactionContentFromDomain(content))
}

func (adapter testLegacyReactionMutationsAdapter) RemoveReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.legacy.RemoveReaction(subjectID, githubcli.ReactionContentFromDomain(content))
}

func (adapter testLegacyBuildQueriesAdapter) GetPullRequestBuildRun(repository string, check githubdomain.PullRequestStatusCheck) (string, error) {
	return adapter.legacy.GetPullRequestBuildRun(repository, githubcli.PullRequestStatusCheckFromDomain(check))
}

func (adapter testLegacyBuildQueriesAdapter) GetPullRequestBuildRunJobs(repository string, check githubdomain.PullRequestStatusCheck) ([]githubdomain.PullRequestBuildRunJob, error) {
	jobs, err := adapter.legacy.GetPullRequestBuildRunJobs(repository, githubcli.PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequestBuildRunJobs(jobs), nil
}

func (adapter testLegacyBuildQueriesAdapter) GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error) {
	return adapter.legacy.GetPullRequestBuildRunJobLog(repository, jobDatabaseID)
}

func (adapter testLegacyBuildQueriesAdapter) GetPullRequestBuildRunJobLogForCheck(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.PullRequestBuildRunJob, string, error) {
	job, output, err := adapter.legacy.GetPullRequestBuildRunJobLogForCheck(repository, githubcli.PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return githubdomain.PullRequestBuildRunJob{}, "", err
	}
	return githubcli.ToDomainPullRequestBuildRunJob(job), output, nil
}
