package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
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
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewCommentComposerTitle, githubcli.PullRequestReviewEventComment)
}

func (program *Program) executeSubmitPendingReviewApprovalAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openPendingReviewSubmitComposer(gui, pullRequestReviewApprovalTitle, githubcli.PullRequestReviewEventApprove)
}

func (program *Program) executeSubmitPendingReviewRequestChangesAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openPendingReviewSubmitComposer(gui, pullRequestRequestChangesComposerTitle, githubcli.PullRequestReviewEventRequestChanges)
}

func (program *Program) openPendingReviewSubmitComposer(gui *gocui.Gui, title string, event githubcli.PullRequestReviewEvent) actionsPopupActionResult {
	target, ok := program.selectedPendingPullRequestReviewTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openModalEditor(gui, title, "", func(body string) error {
		return program.submitPendingPullRequestReview(target, event, body)
	})
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if program.modalEditor != nil {
		program.modalEditor.afterSubmit = func(gui *gocui.Gui) {
			program.finishSubmittedPendingPullRequestReview(gui, target)
		}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) selectedPendingPullRequestReviewTarget() (pendingPullRequestReviewTarget, bool) {
	if !program.reviewSession.active {
		return pendingPullRequestReviewTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(program.reviewSession.summary.Repository))
	pendingReviewID := strings.TrimSpace(program.reviewSession.pendingReviewID)
	if repository == "" || repository == "-" || program.reviewSession.summary.Number <= 0 || pendingReviewID == "" {
		return pendingPullRequestReviewTarget{}, false
	}

	return pendingPullRequestReviewTarget{
		repository:      repository,
		number:          program.reviewSession.summary.Number,
		pendingReviewID: pendingReviewID,
		sourceFocus:     program.reviewSession.sourceFocus,
	}, true
}

func (program *Program) submitPendingPullRequestReview(target pendingPullRequestReviewTarget, event githubcli.PullRequestReviewEvent, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 || strings.TrimSpace(target.pendingReviewID) == "" {
		return errors.New("missing pull request review context")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}

	return program.reviewMutations.SubmitPullRequestReview(target.pendingReviewID, event, body)
}

func (program *Program) finishSubmittedPendingPullRequestReview(gui *gocui.Gui, target pendingPullRequestReviewTarget) {
	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, "")
	program.restorePullRequestBrowserFromReviewMode()
	program.setFeedback(target.sourceFocus, pullRequestReviewSuccessMessage)
	if gui != nil {
		_ = program.layout(gui)
	}
}
