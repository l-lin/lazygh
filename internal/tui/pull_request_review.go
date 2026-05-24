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
		execute: program.executeApprovePullRequestAction,
	}
}

func (program *Program) executeApprovePullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !program.hasReviewMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	return program.startActionsPopupAsyncGHCommand(gui, approvePullRequestCommand(target.repository, target.number), func() error {
		return program.reviewMutations.ApprovePullRequest(target.repository, target.number)
	}, actionsPopupAsyncInvalidatePullRequestSuccess{
		Repository:     target.repository,
		Number:         target.number,
		InvalidateDiff: true,
		Message:        pullRequestReviewSuccessMessage,
	})
}

func approvePullRequestCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "review", fmt.Sprintf("%d", number), "-R", repository, "--approve")
}

func (program *Program) reviewCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "review-comment",
		title:   pullRequestReviewCommentComposerTitle,
		icon:    actionsPopupReviewCommentIcon,
		execute: program.executeReviewCommentAction,
	}
}

func (program *Program) executeReviewCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openModalEditor(gui, pullRequestReviewCommentComposerTitle, "", func(body string) error {
			return program.submitPullRequestReviewComment(target, body)
		})
	})
}

func (program *Program) reviewRequestChangesAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "review-request-changes",
		title:   pullRequestRequestChangesComposerTitle,
		icon:    actionsPopupReviewRequestChangesIcon,
		execute: program.executeRequestChangesAction,
	}
}

func (program *Program) executeRequestChangesAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openModalEditor(gui, pullRequestRequestChangesComposerTitle, "", func(body string) error {
			return program.submitPullRequestRequestChanges(target, body)
		})
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
	program.setFeedback(program.model.Focus(), pullRequestReviewSuccessMessage)
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
	program.setFeedback(program.model.Focus(), pullRequestReviewSuccessMessage)
	return nil
}
