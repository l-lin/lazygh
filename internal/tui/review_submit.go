package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pendingPullRequestReviewTarget struct {
	repository      string
	number          int
	pendingReviewID string
	sourceFocus     Focus
}

func (program *Program) submitPendingReviewCommentAction() actionsPopupAction {
	return program.pendingReviewSubmitAction("submit-pending-review-comment", pullRequestReviewCommentComposerTitle, actionsPopupReviewCommentIcon, githubdomain.PullRequestReviewEventComment)
}

func (program *Program) submitPendingReviewApprovalAction() actionsPopupAction {
	return program.pendingReviewSubmitAction("submit-pending-review-approval", pullRequestReviewApprovalTitle, actionsPopupReviewApproveIcon, githubdomain.PullRequestReviewEventApprove)
}

func (program *Program) submitPendingReviewRequestChangesAction() actionsPopupAction {
	return program.pendingReviewSubmitAction("submit-pending-review-request-changes", pullRequestRequestChangesComposerTitle, actionsPopupReviewRequestChangesIcon, githubdomain.PullRequestReviewEventRequestChanges)
}

func (program *Program) pendingReviewSubmitAction(id string, title string, icon string, event githubdomain.PullRequestReviewEvent) actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPendingPullRequestReviewTarget()
	if ok {
		feedbackTarget := program.model.Focus()
		requested = MsgModalEditorOpened{Descriptor: newModalEditorOpenDescriptorWithSubmitDescriptor(title, "", newPendingPullRequestReviewSubmitDescriptor(target, event, feedbackTarget))}
	}
	return actionsPopupAction{id: id, title: title, icon: icon, requested: requested}
}

func (program *Program) selectedPendingPullRequestReviewTarget() (pendingPullRequestReviewTarget, bool) {
	if !program.reviewModeActive() {
		return pendingPullRequestReviewTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(program.navigationState.reviewSession.summary.Repository))
	pendingReviewID := strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID)
	if repository == "" || repository == "-" || program.navigationState.reviewSession.summary.Number <= 0 || pendingReviewID == "" {
		return pendingPullRequestReviewTarget{}, false
	}

	return pendingPullRequestReviewTarget{
		repository:      repository,
		number:          program.navigationState.reviewSession.summary.Number,
		pendingReviewID: pendingReviewID,
		sourceFocus:     program.navigationState.reviewSession.sourceFocus,
	}, true
}
