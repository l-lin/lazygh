package githubcli

type sessionProvider interface {
	GetConnectedUser() (ConnectedUser, error)
}

type pullRequestListProvider interface {
	ListPullRequests([]string) ([]PullRequest, error)
}

type pullRequestDetailProvider interface {
	GetPullRequestDetail(string, int) (PullRequestDetail, error)
	GetPullRequestDiff(string, int) (PullRequestDiff, error)
	GetPullRequestFileTeamOwners(string, int, []string) (map[string][]string, error)
}

type pullRequestMutationProvider interface {
	CommentOnPullRequest(string, int, string) error
	UpdatePullRequestComment(string, string) error
	DeletePullRequestComment(string) error
	OpenPullRequestInBrowser(string, int) error
	ListAssignableUsers(string) ([]PullRequestAuthor, error)
	UpdatePullRequestAssignees(string, int, []string, []string) error
	EditPullRequestTitle(string, int, string) error
	EditPullRequestDescription(string, int, string) error
	MarkPullRequestReadyForReview(string, int) error
	ConvertPullRequestToDraft(string, int) error
	ClosePullRequest(string, int) error
	ReopenPullRequest(string, int) error
	SquashMergePullRequest(string, int) error
	RequestPullRequestReviewer(string, int, string) error
}

type reviewProvider interface {
	ApprovePullRequest(string, int) error
	ReviewPullRequestWithComment(string, int, string) error
	RequestChangesOnPullRequest(string, int, string) error
	SubmitPullRequestReview(string, PullRequestReviewEvent, string) error
	AddPullRequestReviewThread(string, string, PullRequestReviewThreadTarget) error
	AddPullRequestReviewThreadReply(string, string, string) error
	UpdatePullRequestReviewComment(string, string) error
	DeletePullRequestReviewComment(string) error
	ResolvePullRequestReviewThread(string) error
	UnresolvePullRequestReviewThread(string) error
	StartPendingPullRequestReview(string, int) (string, error)
	GetPendingPullRequestReviewID(string, int) (string, bool, error)
	DeletePullRequestReview(string) error
}

type notificationProvider interface {
	ListNotifications() ([]Notification, error)
	GetIssueDetail(string, int) (IssueDetail, error)
	GetReleaseDetail(string, int) (ReleaseDetail, error)
	MarkNotificationRead(string) error
	MarkNotificationDone(string) error
	MarkAllNotificationsRead() (NotificationBulkReadResult, error)
	MarkAllNotificationsDone([]Notification) (int, error)
}

type reactionProvider interface {
	AddReaction(string, ReactionContent) error
	RemoveReaction(string, ReactionContent) error
}

type buildProvider interface {
	GetPullRequestBuildInfo(string, int, PullRequestStatusCheck) (PullRequestBuildInfo, error)
	GetPullRequestBuildRun(string, PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(string, PullRequestStatusCheck) ([]PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(string, int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(string, PullRequestStatusCheck) (PullRequestBuildRunJob, string, error)
}

type markdownProvider interface {
	RenderMarkdownHTML(string, string) (string, error)
}

type authProvider interface {
	GetAuthToken() (string, error)
}

// ProviderFacade is a transport-typed compatibility facade kept for legacy
// callers and tests while the app composes focused services or adapters.
type ProviderFacade struct {
	session              sessionProvider
	pullRequestLists     pullRequestListProvider
	pullRequestDetails   pullRequestDetailProvider
	pullRequestMutations pullRequestMutationProvider
	review               reviewProvider
	notifications        notificationProvider
	reactions            reactionProvider
	builds               buildProvider
	markdown             markdownProvider
	auth                 authProvider
}

func NewProviderFacade() *ProviderFacade {
	return NewProviderFacadeWithRunner(execRunner{})
}

func NewProviderFacadeWithRunner(runner Runner) *ProviderFacade {
	if runner == nil {
		runner = execRunner{}
	}
	return newProviderFacade(newSharedTransport(runner))
}

func newProviderFacade(transport sharedTransport) *ProviderFacade {
	return &ProviderFacade{
		session:              newSessionService(transport),
		pullRequestLists:     newPullRequestListService(transport),
		pullRequestDetails:   newPullRequestDetailService(transport),
		pullRequestMutations: newPullRequestMutationService(transport),
		review:               newReviewService(transport),
		notifications:        newNotificationService(transport),
		reactions:            newReactionService(transport),
		builds:               newBuildService(transport),
		markdown:             newMarkdownService(transport),
		auth:                 newAuthService(transport),
	}
}

func (provider *ProviderFacade) GetConnectedUser() (ConnectedUser, error) {
	return provider.session.GetConnectedUser()
}

func (provider *ProviderFacade) ListPullRequests(commandArguments []string) ([]PullRequest, error) {
	return provider.pullRequestLists.ListPullRequests(commandArguments)
}

func (provider *ProviderFacade) ListNotifications() ([]Notification, error) {
	return provider.notifications.ListNotifications()
}

func (provider *ProviderFacade) MarkNotificationRead(threadID string) error {
	return provider.notifications.MarkNotificationRead(threadID)
}

func (provider *ProviderFacade) MarkNotificationDone(threadID string) error {
	return provider.notifications.MarkNotificationDone(threadID)
}

func (provider *ProviderFacade) MarkAllNotificationsRead() (NotificationBulkReadResult, error) {
	return provider.notifications.MarkAllNotificationsRead()
}

func (provider *ProviderFacade) MarkAllNotificationsDone(notifications []Notification) (int, error) {
	return provider.notifications.MarkAllNotificationsDone(notifications)
}

func (provider *ProviderFacade) GetPullRequestDetail(repository string, number int) (PullRequestDetail, error) {
	return provider.pullRequestDetails.GetPullRequestDetail(repository, number)
}

func (provider *ProviderFacade) GetIssueDetail(repository string, number int) (IssueDetail, error) {
	return provider.notifications.GetIssueDetail(repository, number)
}

func (provider *ProviderFacade) GetReleaseDetail(repository string, id int) (ReleaseDetail, error) {
	return provider.notifications.GetReleaseDetail(repository, id)
}

func (provider *ProviderFacade) GetPullRequestDiff(repository string, number int) (PullRequestDiff, error) {
	return provider.pullRequestDetails.GetPullRequestDiff(repository, number)
}

func (provider *ProviderFacade) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	return provider.pullRequestDetails.GetPullRequestFileTeamOwners(repository, number, filePaths)
}

func (provider *ProviderFacade) CommentOnPullRequest(repository string, number int, body string) error {
	return provider.pullRequestMutations.CommentOnPullRequest(repository, number, body)
}

func (provider *ProviderFacade) UpdatePullRequestComment(commentID string, body string) error {
	return provider.pullRequestMutations.UpdatePullRequestComment(commentID, body)
}

func (provider *ProviderFacade) DeletePullRequestComment(commentID string) error {
	return provider.pullRequestMutations.DeletePullRequestComment(commentID)
}

func (provider *ProviderFacade) ApprovePullRequest(repository string, number int) error {
	return provider.review.ApprovePullRequest(repository, number)
}

func (provider *ProviderFacade) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return provider.review.ReviewPullRequestWithComment(repository, number, body)
}

func (provider *ProviderFacade) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return provider.review.RequestChangesOnPullRequest(repository, number, body)
}

func (provider *ProviderFacade) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	return provider.pullRequestMutations.RequestPullRequestReviewer(repository, number, reviewerLogin)
}

func (provider *ProviderFacade) SubmitPullRequestReview(pullRequestReviewID string, event PullRequestReviewEvent, body string) error {
	return provider.review.SubmitPullRequestReview(pullRequestReviewID, event, body)
}

func (provider *ProviderFacade) AddPullRequestReviewThread(pullRequestReviewID string, body string, target PullRequestReviewThreadTarget) error {
	return provider.review.AddPullRequestReviewThread(pullRequestReviewID, body, target)
}

func (provider *ProviderFacade) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	return provider.review.AddPullRequestReviewThreadReply(pullRequestReviewID, pullRequestReviewThreadID, body)
}

func (provider *ProviderFacade) UpdatePullRequestReviewComment(commentID string, body string) error {
	return provider.review.UpdatePullRequestReviewComment(commentID, body)
}

func (provider *ProviderFacade) DeletePullRequestReviewComment(commentID string) error {
	return provider.review.DeletePullRequestReviewComment(commentID)
}

func (provider *ProviderFacade) ResolvePullRequestReviewThread(threadID string) error {
	return provider.review.ResolvePullRequestReviewThread(threadID)
}

func (provider *ProviderFacade) UnresolvePullRequestReviewThread(threadID string) error {
	return provider.review.UnresolvePullRequestReviewThread(threadID)
}

func (provider *ProviderFacade) AddReaction(subjectID string, content ReactionContent) error {
	return provider.reactions.AddReaction(subjectID, content)
}

func (provider *ProviderFacade) RemoveReaction(subjectID string, content ReactionContent) error {
	return provider.reactions.RemoveReaction(subjectID, content)
}

func (provider *ProviderFacade) OpenPullRequestInBrowser(repository string, number int) error {
	return provider.pullRequestMutations.OpenPullRequestInBrowser(repository, number)
}

func (provider *ProviderFacade) ListAssignableUsers(repository string) ([]PullRequestAuthor, error) {
	return provider.pullRequestMutations.ListAssignableUsers(repository)
}

func (provider *ProviderFacade) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	return provider.pullRequestMutations.UpdatePullRequestAssignees(repository, number, addLogins, removeLogins)
}

func (provider *ProviderFacade) EditPullRequestTitle(repository string, number int, title string) error {
	return provider.pullRequestMutations.EditPullRequestTitle(repository, number, title)
}

func (provider *ProviderFacade) EditPullRequestDescription(repository string, number int, body string) error {
	return provider.pullRequestMutations.EditPullRequestDescription(repository, number, body)
}

func (provider *ProviderFacade) MarkPullRequestReadyForReview(repository string, number int) error {
	return provider.pullRequestMutations.MarkPullRequestReadyForReview(repository, number)
}

func (provider *ProviderFacade) ConvertPullRequestToDraft(repository string, number int) error {
	return provider.pullRequestMutations.ConvertPullRequestToDraft(repository, number)
}

func (provider *ProviderFacade) ClosePullRequest(repository string, number int) error {
	return provider.pullRequestMutations.ClosePullRequest(repository, number)
}

func (provider *ProviderFacade) ReopenPullRequest(repository string, number int) error {
	return provider.pullRequestMutations.ReopenPullRequest(repository, number)
}

func (provider *ProviderFacade) SquashMergePullRequest(repository string, number int) error {
	return provider.pullRequestMutations.SquashMergePullRequest(repository, number)
}

func (provider *ProviderFacade) StartPendingPullRequestReview(repository string, number int) (string, error) {
	return provider.review.StartPendingPullRequestReview(repository, number)
}

func (provider *ProviderFacade) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	return provider.review.GetPendingPullRequestReviewID(repository, number)
}

func (provider *ProviderFacade) DeletePullRequestReview(pullRequestReviewID string) error {
	return provider.review.DeletePullRequestReview(pullRequestReviewID)
}

func (provider *ProviderFacade) GetPullRequestBuildInfo(repository string, number int, check PullRequestStatusCheck) (PullRequestBuildInfo, error) {
	return provider.builds.GetPullRequestBuildInfo(repository, number, check)
}

func (provider *ProviderFacade) GetPullRequestBuildRun(repository string, check PullRequestStatusCheck) (string, error) {
	return provider.builds.GetPullRequestBuildRun(repository, check)
}

func (provider *ProviderFacade) GetPullRequestBuildRunJobs(repository string, check PullRequestStatusCheck) ([]PullRequestBuildRunJob, error) {
	return provider.builds.GetPullRequestBuildRunJobs(repository, check)
}

func (provider *ProviderFacade) GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error) {
	return provider.builds.GetPullRequestBuildRunJobLog(repository, jobDatabaseID)
}

func (provider *ProviderFacade) GetPullRequestBuildRunJobLogForCheck(repository string, check PullRequestStatusCheck) (PullRequestBuildRunJob, string, error) {
	return provider.builds.GetPullRequestBuildRunJobLogForCheck(repository, check)
}

func (provider *ProviderFacade) RenderMarkdownHTML(repository string, markdown string) (string, error) {
	return provider.markdown.RenderMarkdownHTML(repository, markdown)
}

func (provider *ProviderFacade) GetAuthToken() (string, error) {
	return provider.auth.GetAuthToken()
}
