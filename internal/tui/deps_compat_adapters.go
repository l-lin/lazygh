package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func (adapter legacySessionQueriesAdapter) GetConnectedUser() (githubdomain.ConnectedUser, error) {
	user, err := adapter.legacy.GetConnectedUser()
	if err != nil {
		return githubdomain.ConnectedUser{}, err
	}
	return githubcli.ToDomainConnectedUser(user), nil
}

func (adapter legacyPullRequestListQueriesAdapter) ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error) {
	pullRequests, err := adapter.legacy.ListPullRequests(commandArguments)
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequests(pullRequests), nil
}

func (adapter legacyNotificationQueriesAdapter) ListNotifications() ([]githubdomain.Notification, error) {
	notifications, err := adapter.legacy.ListNotifications()
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainNotifications(notifications), nil
}

func (adapter legacyNotificationQueriesAdapter) GetIssueDetail(repository string, number int) (githubdomain.IssueDetail, error) {
	detail, err := adapter.legacy.GetIssueDetail(repository, number)
	if err != nil {
		return githubdomain.IssueDetail{}, err
	}
	return githubcli.ToDomainIssueDetail(detail), nil
}

func (adapter legacyNotificationQueriesAdapter) GetReleaseDetail(repository string, id int) (githubdomain.ReleaseDetail, error) {
	detail, err := adapter.legacy.GetReleaseDetail(repository, id)
	if err != nil {
		return githubdomain.ReleaseDetail{}, err
	}
	return githubcli.ToDomainReleaseDetail(detail), nil
}

func (adapter legacyDetailQueriesAdapter) GetPullRequestDetail(repository string, number int) (githubdomain.PullRequestDetail, error) {
	detail, err := adapter.legacy.GetPullRequestDetail(repository, number)
	if err != nil {
		return githubdomain.PullRequestDetail{}, err
	}
	return githubcli.ToDomainPullRequestDetail(detail), nil
}

func (adapter legacyDetailQueriesAdapter) GetPullRequestDiff(repository string, number int) (githubdomain.PullRequestDiff, error) {
	diff, err := adapter.legacy.GetPullRequestDiff(repository, number)
	if err != nil {
		return githubdomain.PullRequestDiff{}, err
	}
	return githubcli.ToDomainPullRequestDiff(diff), nil
}

func (adapter legacyDetailQueriesAdapter) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	return adapter.legacy.GetPullRequestFileTeamOwners(repository, number, filePaths)
}

func (adapter legacyPullRequestMutationsAdapter) CommentOnPullRequest(repository string, number int, body string) error {
	return adapter.legacy.CommentOnPullRequest(repository, number, body)
}

func (adapter legacyPullRequestMutationsAdapter) UpdatePullRequestComment(commentID string, body string) error {
	return adapter.legacy.UpdatePullRequestComment(commentID, body)
}

func (adapter legacyPullRequestMutationsAdapter) DeletePullRequestComment(commentID string) error {
	return adapter.legacy.DeletePullRequestComment(commentID)
}

func (adapter legacyPullRequestMutationsAdapter) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	return adapter.legacy.RequestPullRequestReviewer(repository, number, reviewerLogin)
}

func (adapter legacyPullRequestMutationsAdapter) OpenPullRequestInBrowser(repository string, number int) error {
	return adapter.legacy.OpenPullRequestInBrowser(repository, number)
}

func (adapter legacyPullRequestMutationsAdapter) ListAssignableUsers(repository string) ([]githubdomain.PullRequestAuthor, error) {
	authors, err := adapter.legacy.ListAssignableUsers(repository)
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequestAuthors(authors), nil
}

func (adapter legacyPullRequestMutationsAdapter) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	return adapter.legacy.UpdatePullRequestAssignees(repository, number, addLogins, removeLogins)
}

func (adapter legacyPullRequestMutationsAdapter) EditPullRequestTitle(repository string, number int, title string) error {
	return adapter.legacy.EditPullRequestTitle(repository, number, title)
}

func (adapter legacyPullRequestMutationsAdapter) EditPullRequestDescription(repository string, number int, body string) error {
	return adapter.legacy.EditPullRequestDescription(repository, number, body)
}

func (adapter legacyPullRequestMutationsAdapter) MarkPullRequestReadyForReview(repository string, number int) error {
	return adapter.legacy.MarkPullRequestReadyForReview(repository, number)
}

func (adapter legacyPullRequestMutationsAdapter) ConvertPullRequestToDraft(repository string, number int) error {
	return adapter.legacy.ConvertPullRequestToDraft(repository, number)
}

func (adapter legacyPullRequestMutationsAdapter) ClosePullRequest(repository string, number int) error {
	return adapter.legacy.ClosePullRequest(repository, number)
}

func (adapter legacyPullRequestMutationsAdapter) ReopenPullRequest(repository string, number int) error {
	return adapter.legacy.ReopenPullRequest(repository, number)
}

func (adapter legacyPullRequestMutationsAdapter) SquashMergePullRequest(repository string, number int) error {
	return adapter.legacy.SquashMergePullRequest(repository, number)
}

func (adapter legacyReviewMutationsAdapter) ApprovePullRequest(repository string, number int) error {
	return adapter.legacy.ApprovePullRequest(repository, number)
}

func (adapter legacyReviewMutationsAdapter) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return adapter.legacy.ReviewPullRequestWithComment(repository, number, body)
}

func (adapter legacyReviewMutationsAdapter) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return adapter.legacy.RequestChangesOnPullRequest(repository, number, body)
}

func (adapter legacyReviewMutationsAdapter) SubmitPullRequestReview(pullRequestReviewID string, event githubdomain.PullRequestReviewEvent, body string) error {
	return adapter.legacy.SubmitPullRequestReview(pullRequestReviewID, githubcli.PullRequestReviewEventFromDomain(event), body)
}

func (adapter legacyReviewMutationsAdapter) AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubdomain.PullRequestReviewThreadTarget) error {
	return adapter.legacy.AddPullRequestReviewThread(pullRequestReviewID, body, githubcli.PullRequestReviewThreadTargetFromDomain(target))
}

func (adapter legacyReviewMutationsAdapter) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	return adapter.legacy.AddPullRequestReviewThreadReply(pullRequestReviewID, pullRequestReviewThreadID, body)
}

func (adapter legacyReviewMutationsAdapter) UpdatePullRequestReviewComment(commentID string, body string) error {
	return adapter.legacy.UpdatePullRequestReviewComment(commentID, body)
}

func (adapter legacyReviewMutationsAdapter) DeletePullRequestReviewComment(commentID string) error {
	return adapter.legacy.DeletePullRequestReviewComment(commentID)
}

func (adapter legacyReviewMutationsAdapter) ResolvePullRequestReviewThread(threadID string) error {
	return adapter.legacy.ResolvePullRequestReviewThread(threadID)
}

func (adapter legacyReviewMutationsAdapter) UnresolvePullRequestReviewThread(threadID string) error {
	return adapter.legacy.UnresolvePullRequestReviewThread(threadID)
}

func (adapter legacyReviewMutationsAdapter) StartPendingPullRequestReview(repository string, number int) (string, error) {
	return adapter.legacy.StartPendingPullRequestReview(repository, number)
}

func (adapter legacyReviewMutationsAdapter) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	return adapter.legacy.GetPendingPullRequestReviewID(repository, number)
}

func (adapter legacyReviewMutationsAdapter) DeletePullRequestReview(pullRequestReviewID string) error {
	return adapter.legacy.DeletePullRequestReview(pullRequestReviewID)
}

func (adapter legacyNotificationMutationsAdapter) MarkNotificationRead(threadID string) error {
	return adapter.legacy.MarkNotificationRead(threadID)
}

func (adapter legacyNotificationMutationsAdapter) MarkNotificationDone(threadID string) error {
	return adapter.legacy.MarkNotificationDone(threadID)
}

func (adapter legacyNotificationMutationsAdapter) MarkAllNotificationsRead() (githubdomain.NotificationBulkReadResult, error) {
	result, err := adapter.legacy.MarkAllNotificationsRead()
	if err != nil {
		return githubdomain.NotificationBulkReadResult{}, err
	}
	return githubcli.ToDomainNotificationBulkReadResult(result), nil
}

func (adapter legacyNotificationMutationsAdapter) MarkAllNotificationsDone(notifications []githubdomain.Notification) (int, error) {
	return adapter.legacy.MarkAllNotificationsDone(githubcli.NotificationsFromDomain(notifications))
}

func (adapter legacyReactionMutationsAdapter) AddReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.legacy.AddReaction(subjectID, githubcli.ReactionContentFromDomain(content))
}

func (adapter legacyReactionMutationsAdapter) RemoveReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.legacy.RemoveReaction(subjectID, githubcli.ReactionContentFromDomain(content))
}

func (adapter legacyBuildQueriesAdapter) GetPullRequestBuildRun(repository string, check githubdomain.PullRequestStatusCheck) (string, error) {
	return adapter.legacy.GetPullRequestBuildRun(repository, githubcli.PullRequestStatusCheckFromDomain(check))
}

func (adapter legacyBuildQueriesAdapter) GetPullRequestBuildRunJobs(repository string, check githubdomain.PullRequestStatusCheck) ([]githubdomain.PullRequestBuildRunJob, error) {
	jobs, err := adapter.legacy.GetPullRequestBuildRunJobs(repository, githubcli.PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return nil, err
	}
	return githubcli.ToDomainPullRequestBuildRunJobs(jobs), nil
}

func (adapter legacyBuildQueriesAdapter) GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error) {
	return adapter.legacy.GetPullRequestBuildRunJobLog(repository, jobDatabaseID)
}

func (adapter legacyBuildQueriesAdapter) GetPullRequestBuildRunJobLogForCheck(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.PullRequestBuildRunJob, string, error) {
	job, output, err := adapter.legacy.GetPullRequestBuildRunJobLogForCheck(repository, githubcli.PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return githubdomain.PullRequestBuildRunJob{}, "", err
	}
	return githubcli.ToDomainPullRequestBuildRunJob(job), output, nil
}
