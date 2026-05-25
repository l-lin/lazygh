package tui

import (
	"errors"
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

func (program *Program) executeSubmitPendingReviewCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewCommentComposerTitle, githubdomain.PullRequestReviewEventComment)
}

func (program *Program) executeSubmitPendingReviewApprovalAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewApprovalTitle, githubdomain.PullRequestReviewEventApprove)
}

func (program *Program) executeSubmitPendingReviewRequestChangesAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openPendingReviewSubmitComposer(gui, pullRequestRequestChangesComposerTitle, githubdomain.PullRequestReviewEventRequestChanges)
}

func (program *Program) openPendingReviewSubmitComposer(gui *gocui.Gui, title string, event githubdomain.PullRequestReviewEvent) actionsPopupActionResult {
	target, ok := program.selectedPendingPullRequestReviewTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		if err := program.openModalEditor(gui, title, "", func(body string) error {
			submitErr := program.submitPendingPullRequestReview(target, event, body)
			if submitErr != nil && event == githubdomain.PullRequestReviewEventRequestChanges {
				return newModalEditorStatusLineError(program.model.Focus(), submitErr)
			}
			return submitErr
		}); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				program.finishSubmittedPendingPullRequestReview(gui, target)
			}
		}
		return nil
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

func (program *Program) submitPendingPullRequestReview(target pendingPullRequestReviewTarget, event githubdomain.PullRequestReviewEvent, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 || strings.TrimSpace(target.pendingReviewID) == "" {
		return errors.New("missing pull request review context")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}

	if err := program.reviewMutations.SubmitPullRequestReview(target.pendingReviewID, event, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}
	return nil
}

func (program *Program) finishSubmittedPendingPullRequestReview(gui *gocui.Gui, target pendingPullRequestReviewTarget) {
	_ = program.dispatch(gui, MsgPendingPullRequestReviewSubmitted{Target: target})
}
