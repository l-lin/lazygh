package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgClearCacheRequested struct{}

type MsgStartPullRequestReviewRequested struct {
	Summary githubdomain.PullRequest
}

type MsgOpenPullRequestInBrowserRequested struct {
	Target pullRequestActionTarget
}

type MsgApprovePullRequestRequested struct {
	Target pullRequestActionTarget
}

type MsgReRequestPullRequestReviewRequested struct {
	Target pullRequestReviewerRequestTarget
}

type pullRequestLifecycleMutationKind int

const (
	pullRequestLifecycleMutationReadyForReview pullRequestLifecycleMutationKind = iota
	pullRequestLifecycleMutationConvertToDraft
	pullRequestLifecycleMutationClose
	pullRequestLifecycleMutationReopen
)

type MsgPullRequestLifecycleMutationRequested struct {
	Kind           pullRequestLifecycleMutationKind
	Target         pullRequestActionTarget
	Summary        githubdomain.PullRequest
	State          string
	IsDraft        bool
	SuccessMessage string
}

type pullRequestAutoMergeMutationKind int

const (
	pullRequestAutoMergeMutationEnable pullRequestAutoMergeMutationKind = iota
	pullRequestAutoMergeMutationDisable
)

type MsgPullRequestAutoMergeMutationRequested struct {
	Kind           pullRequestAutoMergeMutationKind
	Target         pullRequestActionTarget
	Summary        githubdomain.PullRequest
	Enabled        bool
	SuccessMessage string
}

type MsgPullRequestBranchUpdateRequested struct {
	Target  pullRequestActionTarget
	Summary githubdomain.PullRequest
}

type MsgPullRequestCustomSearchSubmitted struct {
	Criteria string
}

type MsgOpenAssigneePickerRequested struct {
	Target pullRequestAssigneePickerTarget
}

type MsgToggleAssigneePickerSelectionRequested struct {
	Candidate githubdomain.PullRequestAuthor
}

type MsgSubmitAssigneePickerRequested struct {
	Repository   string
	Number       int
	AddLogins    []string
	RemoveLogins []string
}

type MsgOpenReactionPickerRequested struct {
	Target pullRequestReactionActionTarget
}

type MsgAddReactionRequested struct {
	Target  pullRequestReactionActionTarget
	Content githubdomain.ReactionContent
}

type MsgOpenThemePickerRequested struct{}

type MsgThemePresetSelected struct {
	NormalizedName string
	Label          string
}

type MsgThemePresetSaved struct {
	NormalizedName string
	Label          string
	Err            error
}

type MsgRefreshPullRequestListRequested struct{}

type MsgRefreshPullRequestRequested struct {
	Target  pullRequestActionTarget
	Summary githubdomain.PullRequest
}

type MsgPullRequestTitleEditApplied struct {
	Target         pullRequestActionTarget
	Title          string
	FeedbackTarget Focus
}

type MsgPullRequestDescriptionEditApplied struct {
	Target         pullRequestActionTarget
	Body           string
	FeedbackTarget Focus
}

type MsgCancelPendingPullRequestReviewRequested struct {
	Target pendingPullRequestReviewActionTarget
}

func (MsgClearCacheRequested) isMsg()                     {}
func (MsgStartPullRequestReviewRequested) isMsg()         {}
func (MsgOpenPullRequestInBrowserRequested) isMsg()       {}
func (MsgApprovePullRequestRequested) isMsg()             {}
func (MsgReRequestPullRequestReviewRequested) isMsg()     {}
func (MsgPullRequestLifecycleMutationRequested) isMsg()   {}
func (MsgPullRequestAutoMergeMutationRequested) isMsg()   {}
func (MsgPullRequestBranchUpdateRequested) isMsg()        {}
func (MsgPullRequestCustomSearchSubmitted) isMsg()        {}
func (MsgOpenAssigneePickerRequested) isMsg()             {}
func (MsgToggleAssigneePickerSelectionRequested) isMsg()  {}
func (MsgSubmitAssigneePickerRequested) isMsg()           {}
func (MsgOpenReactionPickerRequested) isMsg()             {}
func (MsgAddReactionRequested) isMsg()                    {}
func (MsgOpenThemePickerRequested) isMsg()                {}
func (MsgThemePresetSelected) isMsg()                     {}
func (MsgThemePresetSaved) isMsg()                        {}
func (MsgRefreshPullRequestListRequested) isMsg()         {}
func (MsgRefreshPullRequestRequested) isMsg()             {}
func (MsgPullRequestTitleEditApplied) isMsg()             {}
func (MsgPullRequestDescriptionEditApplied) isMsg()       {}
func (MsgCancelPendingPullRequestReviewRequested) isMsg() {}
