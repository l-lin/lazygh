package githubcli

type SessionService struct {
	serviceBase
}

type PullRequestListService struct {
	serviceBase
}

type PullRequestDetailService struct {
	serviceBase
}

type PullRequestMutationService struct {
	serviceBase
}

type ReviewService struct {
	serviceBase
}

type NotificationService struct {
	serviceBase
}

type ReactionService struct {
	serviceBase
}

type BuildService struct {
	serviceBase
}

type MarkdownService struct {
	serviceBase
}

type AuthService struct {
	serviceBase
}

func NewSessionService() *SessionService {
	return NewSessionServiceWithRunner(nil)
}

func NewSessionServiceWithRunner(runner Runner) *SessionService {
	return newSessionService(newSharedTransport(runner))
}

func NewPullRequestListService() *PullRequestListService {
	return NewPullRequestListServiceWithRunner(nil)
}

func NewPullRequestListServiceWithRunner(runner Runner) *PullRequestListService {
	return newPullRequestListService(newSharedTransport(runner))
}

func NewPullRequestDetailService() *PullRequestDetailService {
	return NewPullRequestDetailServiceWithRunner(nil)
}

func NewPullRequestDetailServiceWithRunner(runner Runner) *PullRequestDetailService {
	return newPullRequestDetailService(newSharedTransport(runner))
}

func NewPullRequestMutationService() *PullRequestMutationService {
	return NewPullRequestMutationServiceWithRunner(nil)
}

func NewPullRequestMutationServiceWithRunner(runner Runner) *PullRequestMutationService {
	return newPullRequestMutationService(newSharedTransport(runner))
}

func NewReviewService() *ReviewService {
	return NewReviewServiceWithRunner(nil)
}

func NewReviewServiceWithRunner(runner Runner) *ReviewService {
	return newReviewService(newSharedTransport(runner))
}

func NewNotificationService() *NotificationService {
	return NewNotificationServiceWithRunner(nil)
}

func NewNotificationServiceWithRunner(runner Runner) *NotificationService {
	return newNotificationService(newSharedTransport(runner))
}

func NewReactionService() *ReactionService {
	return NewReactionServiceWithRunner(nil)
}

func NewReactionServiceWithRunner(runner Runner) *ReactionService {
	return newReactionService(newSharedTransport(runner))
}

func NewBuildService() *BuildService {
	return NewBuildServiceWithRunner(nil)
}

func NewBuildServiceWithRunner(runner Runner) *BuildService {
	return newBuildService(newSharedTransport(runner))
}

func NewMarkdownService() *MarkdownService {
	return NewMarkdownServiceWithRunner(nil)
}

func NewMarkdownServiceWithRunner(runner Runner) *MarkdownService {
	return newMarkdownService(newSharedTransport(runner))
}

func NewAuthService() *AuthService {
	return NewAuthServiceWithRunner(nil)
}

func NewAuthServiceWithRunner(runner Runner) *AuthService {
	return newAuthService(newSharedTransport(runner))
}

func newSessionService(transport sharedTransport) *SessionService {
	return &SessionService{serviceBase: newServiceBase(transport)}
}

func newPullRequestListService(transport sharedTransport) *PullRequestListService {
	return &PullRequestListService{serviceBase: newServiceBase(transport)}
}

func newPullRequestDetailService(transport sharedTransport) *PullRequestDetailService {
	return &PullRequestDetailService{serviceBase: newServiceBase(transport)}
}

func newPullRequestMutationService(transport sharedTransport) *PullRequestMutationService {
	return &PullRequestMutationService{serviceBase: newServiceBase(transport)}
}

func newReviewService(transport sharedTransport) *ReviewService {
	return &ReviewService{serviceBase: newServiceBase(transport)}
}

func newNotificationService(transport sharedTransport) *NotificationService {
	return &NotificationService{serviceBase: newServiceBase(transport)}
}

func newReactionService(transport sharedTransport) *ReactionService {
	return &ReactionService{serviceBase: newServiceBase(transport)}
}

func newBuildService(transport sharedTransport) *BuildService {
	return &BuildService{serviceBase: newServiceBase(transport)}
}

func newMarkdownService(transport sharedTransport) *MarkdownService {
	return &MarkdownService{serviceBase: newServiceBase(transport)}
}

func newAuthService(transport sharedTransport) *AuthService {
	return &AuthService{serviceBase: newServiceBase(transport)}
}
