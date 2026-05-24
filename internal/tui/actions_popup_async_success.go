package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type actionsPopupAsyncSuccess interface {
	apply(*Program)
}

type actionsPopupAsyncFeedbackSuccess struct {
	Message string
}

func (success actionsPopupAsyncFeedbackSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.setFeedback(program.model.Focus(), success.Message)
}

type actionsPopupAsyncPullRequestLifecycleSuccess struct {
	Summary githubdomain.PullRequest
	State   string
	IsDraft bool
	Message string
}

func (success actionsPopupAsyncPullRequestLifecycleSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.applyVisiblePullRequestLifecycleMutation(success.Summary, success.State, success.IsDraft)
	program.setFeedback(program.model.Focus(), success.Message)
}

type actionsPopupAsyncPullRequestAutoMergeSuccess struct {
	Summary githubdomain.PullRequest
	Enabled bool
	Message string
}

func (success actionsPopupAsyncPullRequestAutoMergeSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.applyVisiblePullRequestAutoMergeMutation(success.Summary, success.Enabled)
	program.setFeedback(program.model.Focus(), success.Message)
}

type actionsPopupAsyncPullRequestBranchUpdateSuccess struct {
	Summary githubdomain.PullRequest
	Message string
}

func (success actionsPopupAsyncPullRequestBranchUpdateSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.applyVisiblePullRequestBranchUpdate(success.Summary)
	program.setFeedback(program.model.Focus(), success.Message)
}

type actionsPopupAsyncInvalidatePullRequestSuccess struct {
	Repository     string
	Number         int
	InvalidateDiff bool
	Message        string
}

func (success actionsPopupAsyncInvalidatePullRequestSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.invalidatePullRequestDetail(success.Repository, success.Number)
	if success.InvalidateDiff {
		program.invalidatePullRequestDiff(success.Repository, success.Number)
	}
	program.setFeedback(program.model.Focus(), success.Message)
}

type actionsPopupAsyncPullRequestAssigneesUpdatedSuccess struct {
	Repository   string
	Number       int
	AddLogins    []string
	RemoveLogins []string
	Message      string
}

func (success actionsPopupAsyncPullRequestAssigneesUpdatedSuccess) apply(program *Program) {
	if program == nil {
		return
	}
	program.optimisticallyUpdatePullRequestAssignees(success.Repository, success.Number, success.AddLogins, success.RemoveLogins)
	program.setFeedback(program.model.Focus(), success.Message)
}
