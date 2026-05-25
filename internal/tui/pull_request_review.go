package tui

import (
	"errors"
	"fmt"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestReviewSuccessMessage        = "Review submitted"
	pullRequestReviewApprovalTitle         = "Review: Approve PR"
	pullRequestReviewCommentComposerTitle  = "Review: Comment on PR"
	pullRequestRequestChangesComposerTitle = "Review: Request changes"
)

func (program *Program) reviewApproveAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "review-approve",
		title:   pullRequestReviewApprovalTitle,
		icon:    actionsPopupReviewApproveIcon,
		execute: actionsPopupExecuteErr(program.executeApprovePullRequestAction),
	}
}

func (program *Program) executeApprovePullRequestAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	return program.dispatch(gui, MsgApprovePullRequestRequested{Target: target})
}

func approvePullRequestCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "review", fmt.Sprintf("%d", number), "-R", repository, "--approve")
}

func (program *Program) reviewCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "review-comment",
		title:   pullRequestReviewCommentComposerTitle,
		icon:    actionsPopupReviewCommentIcon,
		execute: actionsPopupExecuteErr(program.executeReviewCommentAction),
	}
}

func (program *Program) executeReviewCommentAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openModalEditorWithSubmitRequested(gui, pullRequestReviewCommentComposerTitle, "", func(body string) Msg {
			return MsgPullRequestReviewCommentSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		})
	})
}

func (program *Program) reviewRequestChangesAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "review-request-changes",
		title:   pullRequestRequestChangesComposerTitle,
		icon:    actionsPopupReviewRequestChangesIcon,
		execute: actionsPopupExecuteErr(program.executeRequestChangesAction),
	}
}

func (program *Program) executeRequestChangesAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openModalEditorWithSubmitRequested(gui, pullRequestRequestChangesComposerTitle, "", func(body string) Msg {
			return MsgPullRequestRequestChangesSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		})
	})
}
