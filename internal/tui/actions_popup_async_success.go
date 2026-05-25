package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

type actionsPopupAsyncSuccess interface {
	apply(*Program) []Cmd
}

func applyActionsPopupFeedback(program *Program, target Focus, message string) {
	if program == nil {
		return
	}
	program.applyFeedbackSet(MsgFeedbackSet{Target: target, Message: message})
}

type actionsPopupAsyncFeedbackSuccess struct {
	Message string
}

func (success actionsPopupAsyncFeedbackSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncPullRequestLifecycleSuccess struct {
	Summary githubdomain.PullRequest
	State   string
	IsDraft bool
	Message string
}

func (success actionsPopupAsyncPullRequestLifecycleSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.applyVisiblePullRequestLifecycleMutation(success.Summary, success.State, success.IsDraft)
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncPullRequestAutoMergeSuccess struct {
	Summary githubdomain.PullRequest
	Enabled bool
	Message string
}

func (success actionsPopupAsyncPullRequestAutoMergeSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.applyVisiblePullRequestAutoMergeMutation(success.Summary, success.Enabled)
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncPullRequestBranchUpdateSuccess struct {
	Summary githubdomain.PullRequest
	Message string
}

func (success actionsPopupAsyncPullRequestBranchUpdateSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.applyVisiblePullRequestBranchUpdate(success.Summary)
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncInvalidatePullRequestSuccess struct {
	Repository     string
	Number         int
	InvalidateDiff bool
	Message        string
}

func (success actionsPopupAsyncInvalidatePullRequestSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.invalidatePullRequestDetail(success.Repository, success.Number)
	if success.InvalidateDiff {
		program.invalidatePullRequestDiff(success.Repository, success.Number)
	}
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncPullRequestAssigneesUpdatedSuccess struct {
	Repository   string
	Number       int
	AddLogins    []string
	RemoveLogins []string
	Message      string
}

func (success actionsPopupAsyncPullRequestAssigneesUpdatedSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyUpdatePullRequestAssignees(success.Repository, success.Number, success.AddLogins, success.RemoveLogins)
	applyActionsPopupFeedback(program, program.model.Focus(), success.Message)
	return nil
}

type actionsPopupAsyncStartReviewSuccess struct {
	Summary         githubdomain.PullRequest
	PendingReviewID string
}

func (success actionsPopupAsyncStartReviewSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.startReviewSession(success.Summary, success.PendingReviewID)
	return nil
}

type actionsPopupAsyncReactionAddedSuccess struct {
	Target  pullRequestReactionActionTarget
	Content githubdomain.ReactionContent
}

func (success actionsPopupAsyncReactionAddedSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyAddReaction(success.Target, success.Content)
	applyActionsPopupFeedback(program, program.model.Focus(), pullRequestReactionAddedSuccessMessage)
	return nil
}

type actionsPopupAsyncThemeAppliedSuccess struct {
	NormalizedName string
	Label          string
}

func (success actionsPopupAsyncThemeAppliedSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	theme.ApplyPalette(theme.ResolvePaletteWithPreset(success.NormalizedName, theme.Palette{}))
	program.restylePullRequestRows()
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.actionsPopupWidget.errorMessage = ""
	applyActionsPopupFeedback(program, program.model.Focus(), "Theme changed to "+success.Label)
	if program.gui != nil {
		program.configureGUI(program.gui)
	}
	return nil
}

type actionsPopupAsyncPendingReviewCanceledSuccess struct {
	Target pendingPullRequestReviewActionTarget
}

func (success actionsPopupAsyncPendingReviewCanceledSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.invalidatePullRequestDetail(success.Target.repository, success.Target.number)
	program.invalidatePullRequestDiff(success.Target.repository, success.Target.number)
	program.setPendingPullRequestReviewStateByIdentity(success.Target.repository, success.Target.number, "")
	applyActionsPopupFeedback(program, success.Target.sourceFocus, pendingPullRequestReviewCanceledMessage)
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
}
