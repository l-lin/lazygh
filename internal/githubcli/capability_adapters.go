package githubcli

import githubdomain "github.com/l-lin/lazygh/internal/github"

type SessionAdapter struct {
	service *SessionService
}

type PullRequestListAdapter struct {
	service *PullRequestListService
}

type NotificationAdapter struct {
	service *NotificationService
}

type PullRequestDetailAdapter struct {
	service *PullRequestDetailService
}

type PullRequestMutationAdapter struct {
	service *PullRequestMutationService
}

type ReviewAdapter struct {
	service *ReviewService
}

type ReactionAdapter struct {
	service *ReactionService
}

type BuildAdapter struct {
	service *BuildService
}

func NewSessionAdapter() *SessionAdapter {
	return NewSessionAdapterWithRunner(nil)
}

func NewSessionAdapterWithRunner(runner Runner) *SessionAdapter {
	return &SessionAdapter{service: newSessionService(newSharedTransport(runner))}
}

func NewPullRequestListAdapter() *PullRequestListAdapter {
	return NewPullRequestListAdapterWithRunner(nil)
}

func NewPullRequestListAdapterWithRunner(runner Runner) *PullRequestListAdapter {
	return &PullRequestListAdapter{service: newPullRequestListService(newSharedTransport(runner))}
}

func NewNotificationAdapter() *NotificationAdapter {
	return NewNotificationAdapterWithRunner(nil)
}

func NewNotificationAdapterWithRunner(runner Runner) *NotificationAdapter {
	return &NotificationAdapter{service: newNotificationService(newSharedTransport(runner))}
}

func NewPullRequestDetailAdapter() *PullRequestDetailAdapter {
	return NewPullRequestDetailAdapterWithRunner(nil)
}

func NewPullRequestDetailAdapterWithRunner(runner Runner) *PullRequestDetailAdapter {
	return &PullRequestDetailAdapter{service: newPullRequestDetailService(newSharedTransport(runner))}
}

func NewPullRequestMutationAdapter() *PullRequestMutationAdapter {
	return NewPullRequestMutationAdapterWithRunner(nil)
}

func NewPullRequestMutationAdapterWithRunner(runner Runner) *PullRequestMutationAdapter {
	return &PullRequestMutationAdapter{service: newPullRequestMutationService(newSharedTransport(runner))}
}

func NewReviewAdapter() *ReviewAdapter {
	return NewReviewAdapterWithRunner(nil)
}

func NewReviewAdapterWithRunner(runner Runner) *ReviewAdapter {
	return &ReviewAdapter{service: newReviewService(newSharedTransport(runner))}
}

func NewReactionAdapter() *ReactionAdapter {
	return NewReactionAdapterWithRunner(nil)
}

func NewReactionAdapterWithRunner(runner Runner) *ReactionAdapter {
	return &ReactionAdapter{service: newReactionService(newSharedTransport(runner))}
}

func NewBuildAdapter() *BuildAdapter {
	return NewBuildAdapterWithRunner(nil)
}

func NewBuildAdapterWithRunner(runner Runner) *BuildAdapter {
	return &BuildAdapter{service: newBuildService(newSharedTransport(runner))}
}

func (adapter *SessionAdapter) GetConnectedUser() (githubdomain.ConnectedUser, error) {
	user, err := adapter.service.GetConnectedUser()
	if err != nil {
		return githubdomain.ConnectedUser{}, err
	}
	return ToDomainConnectedUser(user), nil
}

func (adapter *PullRequestListAdapter) ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error) {
	pullRequests, err := adapter.service.ListPullRequests(commandArguments)
	if err != nil {
		return nil, err
	}
	return ToDomainPullRequests(pullRequests), nil
}

func (adapter *NotificationAdapter) ListNotifications() ([]githubdomain.Notification, error) {
	notifications, err := adapter.service.ListNotifications()
	if err != nil {
		return nil, err
	}
	return ToDomainNotifications(notifications), nil
}

func (adapter *NotificationAdapter) GetIssueDetail(repository string, number int) (githubdomain.IssueDetail, error) {
	detail, err := adapter.service.GetIssueDetail(repository, number)
	if err != nil {
		return githubdomain.IssueDetail{}, err
	}
	return ToDomainIssueDetail(detail), nil
}

func (adapter *NotificationAdapter) GetReleaseDetail(repository string, id int) (githubdomain.ReleaseDetail, error) {
	detail, err := adapter.service.GetReleaseDetail(repository, id)
	if err != nil {
		return githubdomain.ReleaseDetail{}, err
	}
	return ToDomainReleaseDetail(detail), nil
}

func (adapter *NotificationAdapter) MarkNotificationRead(threadID string) error {
	return adapter.service.MarkNotificationRead(threadID)
}

func (adapter *NotificationAdapter) MarkNotificationDone(threadID string) error {
	return adapter.service.MarkNotificationDone(threadID)
}

func (adapter *NotificationAdapter) MarkAllNotificationsRead() (githubdomain.NotificationBulkReadResult, error) {
	result, err := adapter.service.MarkAllNotificationsRead()
	if err != nil {
		return githubdomain.NotificationBulkReadResult{}, err
	}
	return ToDomainNotificationBulkReadResult(result), nil
}

func (adapter *NotificationAdapter) MarkAllNotificationsDone(notifications []githubdomain.Notification) (int, error) {
	return adapter.service.MarkAllNotificationsDone(NotificationsFromDomain(notifications))
}

func (adapter *PullRequestDetailAdapter) GetPullRequestDetail(repository string, number int) (githubdomain.PullRequestDetail, error) {
	detail, err := adapter.service.GetPullRequestDetail(repository, number)
	if err != nil {
		return githubdomain.PullRequestDetail{}, err
	}
	return ToDomainPullRequestDetail(detail), nil
}

func (adapter *PullRequestDetailAdapter) GetPullRequestDiff(repository string, number int) (githubdomain.PullRequestDiff, error) {
	diff, err := adapter.service.GetPullRequestDiff(repository, number)
	if err != nil {
		return githubdomain.PullRequestDiff{}, err
	}
	return ToDomainPullRequestDiff(diff), nil
}

func (adapter *PullRequestDetailAdapter) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	return adapter.service.GetPullRequestFileTeamOwners(repository, number, filePaths)
}

func (adapter *PullRequestMutationAdapter) CommentOnPullRequest(repository string, number int, body string) error {
	return adapter.service.CommentOnPullRequest(repository, number, body)
}

func (adapter *PullRequestMutationAdapter) UpdatePullRequestComment(commentID string, body string) error {
	return adapter.service.UpdatePullRequestComment(commentID, body)
}

func (adapter *PullRequestMutationAdapter) DeletePullRequestComment(commentID string) error {
	return adapter.service.DeletePullRequestComment(commentID)
}

func (adapter *PullRequestMutationAdapter) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	return adapter.service.RequestPullRequestReviewer(repository, number, reviewerLogin)
}

func (adapter *PullRequestMutationAdapter) OpenPullRequestInBrowser(repository string, number int) error {
	return adapter.service.OpenPullRequestInBrowser(repository, number)
}

func (adapter *PullRequestMutationAdapter) ListAssignableUsers(repository string) ([]githubdomain.PullRequestAuthor, error) {
	authors, err := adapter.service.ListAssignableUsers(repository)
	if err != nil {
		return nil, err
	}
	return ToDomainPullRequestAuthors(authors), nil
}

func (adapter *PullRequestMutationAdapter) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	return adapter.service.UpdatePullRequestAssignees(repository, number, addLogins, removeLogins)
}

func (adapter *PullRequestMutationAdapter) EditPullRequestTitle(repository string, number int, title string) error {
	return adapter.service.EditPullRequestTitle(repository, number, title)
}

func (adapter *PullRequestMutationAdapter) EditPullRequestDescription(repository string, number int, body string) error {
	return adapter.service.EditPullRequestDescription(repository, number, body)
}

func (adapter *PullRequestMutationAdapter) MarkPullRequestReadyForReview(repository string, number int) error {
	return adapter.service.MarkPullRequestReadyForReview(repository, number)
}

func (adapter *PullRequestMutationAdapter) ConvertPullRequestToDraft(repository string, number int) error {
	return adapter.service.ConvertPullRequestToDraft(repository, number)
}

func (adapter *PullRequestMutationAdapter) ClosePullRequest(repository string, number int) error {
	return adapter.service.ClosePullRequest(repository, number)
}

func (adapter *PullRequestMutationAdapter) ReopenPullRequest(repository string, number int) error {
	return adapter.service.ReopenPullRequest(repository, number)
}

func (adapter *PullRequestMutationAdapter) SquashMergePullRequest(repository string, number int) error {
	return adapter.service.SquashMergePullRequest(repository, number)
}

func (adapter *PullRequestMutationAdapter) UpdatePullRequestBranch(repository string, number int) error {
	return adapter.service.UpdatePullRequestBranch(repository, number)
}

func (adapter *ReviewAdapter) ApprovePullRequest(repository string, number int) error {
	return adapter.service.ApprovePullRequest(repository, number)
}

func (adapter *ReviewAdapter) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return adapter.service.ReviewPullRequestWithComment(repository, number, body)
}

func (adapter *ReviewAdapter) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return adapter.service.RequestChangesOnPullRequest(repository, number, body)
}

func (adapter *ReviewAdapter) SubmitPullRequestReview(pullRequestReviewID string, event githubdomain.PullRequestReviewEvent, body string) error {
	return adapter.service.SubmitPullRequestReview(pullRequestReviewID, PullRequestReviewEventFromDomain(event), body)
}

func (adapter *ReviewAdapter) AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubdomain.PullRequestReviewThreadTarget) error {
	return adapter.service.AddPullRequestReviewThread(pullRequestReviewID, body, PullRequestReviewThreadTargetFromDomain(target))
}

func (adapter *ReviewAdapter) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	return adapter.service.AddPullRequestReviewThreadReply(pullRequestReviewID, pullRequestReviewThreadID, body)
}

func (adapter *ReviewAdapter) UpdatePullRequestReviewComment(commentID string, body string) error {
	return adapter.service.UpdatePullRequestReviewComment(commentID, body)
}

func (adapter *ReviewAdapter) DeletePullRequestReviewComment(commentID string) error {
	return adapter.service.DeletePullRequestReviewComment(commentID)
}

func (adapter *ReviewAdapter) ResolvePullRequestReviewThread(threadID string) error {
	return adapter.service.ResolvePullRequestReviewThread(threadID)
}

func (adapter *ReviewAdapter) UnresolvePullRequestReviewThread(threadID string) error {
	return adapter.service.UnresolvePullRequestReviewThread(threadID)
}

func (adapter *ReviewAdapter) StartPendingPullRequestReview(repository string, number int) (string, error) {
	return adapter.service.StartPendingPullRequestReview(repository, number)
}

func (adapter *ReviewAdapter) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	return adapter.service.GetPendingPullRequestReviewID(repository, number)
}

func (adapter *ReviewAdapter) DeletePullRequestReview(pullRequestReviewID string) error {
	return adapter.service.DeletePullRequestReview(pullRequestReviewID)
}

func (adapter *ReactionAdapter) AddReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.service.AddReaction(subjectID, ReactionContentFromDomain(content))
}

func (adapter *ReactionAdapter) RemoveReaction(subjectID string, content githubdomain.ReactionContent) error {
	return adapter.service.RemoveReaction(subjectID, ReactionContentFromDomain(content))
}

func (adapter *BuildAdapter) GetPullRequestBuildRun(repository string, check githubdomain.PullRequestStatusCheck) (string, error) {
	return adapter.service.GetPullRequestBuildRun(repository, PullRequestStatusCheckFromDomain(check))
}

func (adapter *BuildAdapter) GetPullRequestBuildRunJobs(repository string, check githubdomain.PullRequestStatusCheck) ([]githubdomain.PullRequestBuildRunJob, error) {
	jobs, err := adapter.service.GetPullRequestBuildRunJobs(repository, PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return nil, err
	}
	return ToDomainPullRequestBuildRunJobs(jobs), nil
}

func (adapter *BuildAdapter) GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error) {
	return adapter.service.GetPullRequestBuildRunJobLog(repository, jobDatabaseID)
}

func (adapter *BuildAdapter) GetPullRequestBuildRunJobLogForCheck(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.PullRequestBuildRunJob, string, error) {
	job, output, err := adapter.service.GetPullRequestBuildRunJobLogForCheck(repository, PullRequestStatusCheckFromDomain(check))
	if err != nil {
		return githubdomain.PullRequestBuildRunJob{}, "", err
	}
	return ToDomainPullRequestBuildRunJob(job), output, nil
}
