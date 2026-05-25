package tui

import (
	"errors"
	"fmt"
	"strings"

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
		if err := program.openModalEditor(gui, pullRequestReviewCommentComposerTitle, "", func(body string) error {
			return program.submitPullRequestReviewComment(target, body)
		}); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				_ = program.dispatch(gui, MsgFeedbackSet{Target: feedbackTarget, Message: pullRequestReviewSuccessMessage})
			}
		}
		return nil
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
		if err := program.openModalEditor(gui, pullRequestRequestChangesComposerTitle, "", func(body string) error {
			return program.submitPullRequestRequestChanges(target, body)
		}); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				_ = program.dispatch(gui, MsgFeedbackSet{Target: feedbackTarget, Message: pullRequestReviewSuccessMessage})
			}
		}
		return nil
	})
}

func (program *Program) submitPullRequestReviewComment(target pullRequestActionTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.ReviewPullRequestWithComment(target.repository, target.number, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	return nil
}

func (program *Program) submitPullRequestRequestChanges(target pullRequestActionTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.RequestChangesOnPullRequest(target.repository, target.number, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	return nil
}
