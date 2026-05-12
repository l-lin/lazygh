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
