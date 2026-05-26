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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestActionTarget()
	if ok {
		if !program.hasReviewMutations() {
			requested = actionsPopupErrorRequested(errors.New("github loader is unavailable"))
		} else {
			requested = MsgApprovePullRequestRequested{Target: target}
		}
	}
	return actionsPopupAction{
		id:        "review-approve",
		title:     pullRequestReviewApprovalTitle,
		icon:      actionsPopupReviewApproveIcon,
		requested: requested,
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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestActionTarget()
	if ok {
		feedbackTarget := program.model.Focus()
		requested = MsgModalEditorOpened{State: newModalEditorStateWithSubmitRequested(pullRequestReviewCommentComposerTitle, "", func(body string) Msg {
			return MsgPullRequestReviewCommentSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		})}
	}
	return actionsPopupAction{
		id:        "review-comment",
		title:     pullRequestReviewCommentComposerTitle,
		icon:      actionsPopupReviewCommentIcon,
		requested: requested,
	}
}

func (program *Program) executeReviewCommentAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorWithSubmitRequested(gui, pullRequestReviewCommentComposerTitle, "", func(body string) Msg {
		return MsgPullRequestReviewCommentSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
	})
}

func (program *Program) reviewRequestChangesAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestActionTarget()
	if ok {
		feedbackTarget := program.model.Focus()
		requested = MsgModalEditorOpened{State: newModalEditorStateWithSubmitRequested(pullRequestRequestChangesComposerTitle, "", func(body string) Msg {
			return MsgPullRequestRequestChangesSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		})}
	}
	return actionsPopupAction{
		id:        "review-request-changes",
		title:     pullRequestRequestChangesComposerTitle,
		icon:      actionsPopupReviewRequestChangesIcon,
		requested: requested,
	}
}

func (program *Program) executeRequestChangesAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorWithSubmitRequested(gui, pullRequestRequestChangesComposerTitle, "", func(body string) Msg {
		return MsgPullRequestRequestChangesSubmitRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
	})
}
