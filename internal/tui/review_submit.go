package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pendingPullRequestReviewTarget struct {
	repository      string
	number          int
	pendingReviewID string
	sourceFocus     Focus
}

func (program *Program) submitPendingReviewCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "submit-pending-review-comment",
		title:   pullRequestReviewCommentComposerTitle,
		icon:    actionsPopupReviewCommentIcon,
		execute: program.executeSubmitPendingReviewCommentAction,
	}
}

func (program *Program) submitPendingReviewApprovalAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "submit-pending-review-approval",
		title:   pullRequestReviewApprovalTitle,
		icon:    actionsPopupReviewApproveIcon,
		execute: program.executeSubmitPendingReviewApprovalAction,
	}
}

func (program *Program) submitPendingReviewRequestChangesAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "submit-pending-review-request-changes",
		title:   pullRequestRequestChangesComposerTitle,
		icon:    actionsPopupReviewRequestChangesIcon,
		execute: program.executeSubmitPendingReviewRequestChangesAction,
	}
}

func (program *Program) executeSubmitPendingReviewCommentAction(gui *gocui.Gui) error {
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewCommentComposerTitle, githubdomain.PullRequestReviewEventComment)
}

func (program *Program) executeSubmitPendingReviewApprovalAction(gui *gocui.Gui) error {
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewApprovalTitle, githubdomain.PullRequestReviewEventApprove)
}

func (program *Program) executeSubmitPendingReviewRequestChangesAction(gui *gocui.Gui) error {
	return program.openPendingReviewSubmitComposer(gui, pullRequestRequestChangesComposerTitle, githubdomain.PullRequestReviewEventRequestChanges)
}

func (program *Program) openPendingReviewSubmitComposer(gui *gocui.Gui, title string, event githubdomain.PullRequestReviewEvent) error {
	target, ok := program.selectedPendingPullRequestReviewTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openModalEditorWithSubmitRequested(gui, title, "", func(body string) Msg {
			return MsgPendingPullRequestReviewSubmitRequested{Target: target, Event: event, Body: body, FeedbackTarget: feedbackTarget}
		})
	})
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
