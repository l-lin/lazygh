package tui

import (
	"errors"
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
		id:       "review-approve",
		title:    pullRequestReviewApprovalTitle,
		icon:     actionsPopupReviewApproveIcon,
		keywords: []string{"review", "approve", "lgtm", "shipit"},
		execute:  program.executeApprovePullRequestAction,
	}
}

func (program *Program) executeApprovePullRequestAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.githubLoader.ApprovePullRequest(target.repository, target.number); err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(program.model.Focus(), pullRequestReviewSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) reviewCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "review-comment",
		title:    pullRequestReviewCommentComposerTitle,
		icon:     actionsPopupReviewCommentIcon,
		keywords: []string{"review", "comment", "feedback"},
		execute:  program.executeReviewCommentAction,
	}
}

func (program *Program) executeReviewCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openModalEditor(gui, pullRequestReviewCommentComposerTitle, "", func(body string) error {
		return program.submitPullRequestReviewComment(target, body)
	})
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) reviewRequestChangesAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "review-request-changes",
		title:    pullRequestRequestChangesComposerTitle,
		icon:     actionsPopupReviewRequestChangesIcon,
		keywords: []string{"review", "request", "changes", "block"},
		execute:  program.executeRequestChangesAction,
	}
}

func (program *Program) executeRequestChangesAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openModalEditor(gui, pullRequestRequestChangesComposerTitle, "", func(body string) error {
		return program.submitPullRequestRequestChanges(target, body)
	})
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) submitPullRequestReviewComment(target pullRequestActionTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.ReviewPullRequestWithComment(target.repository, target.number, body); err != nil {
		return err
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
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.RequestChangesOnPullRequest(target.repository, target.number, body); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(program.model.Focus(), pullRequestReviewSuccessMessage)
	return nil
}
