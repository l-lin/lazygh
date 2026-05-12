package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	cancelPendingPullRequestReviewActionTitle = "Cancel pending review"
	pendingPullRequestReviewCanceledMessage   = "Pending review canceled"
)

type pendingPullRequestReviewState struct {
	id string
}

type pendingPullRequestReviewActionTarget struct {
	repository      string
	number          int
	pendingReviewID string
	sourceFocus     Focus
}

func (program *Program) pendingPullRequestReviewStateForSummary(summary githubcli.PullRequest) (pendingPullRequestReviewState, bool) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return pendingPullRequestReviewState{}, false
	}

	state, ok := program.pendingPullRequestReviewCache[key]
	return state, ok
}

func (program *Program) setPendingPullRequestReviewState(summary githubcli.PullRequest, pendingReviewID string) {
	program.setPendingPullRequestReviewStateByIdentity(pullRequestRepositoryName(summary.Repository), summary.Number, pendingReviewID)
}

func (program *Program) setPendingPullRequestReviewStateByIdentity(repository string, number int, pendingReviewID string) {
	key := pullRequestKeyFromIdentity(repository, number)
	if key == "" {
		return
	}

	program.pendingPullRequestReviewCache[key] = pendingPullRequestReviewState{id: strings.TrimSpace(pendingReviewID)}
}

func (program *Program) forgetPendingPullRequestReviewState(repository string, number int) {
	key := pullRequestKeyFromIdentity(repository, number)
	if key == "" {
		return
	}

	delete(program.pendingPullRequestReviewCache, key)
}

func pullRequestKeyFromIdentity(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" || number <= 0 {
		return ""
	}

	return fmt.Sprintf("%s#%d", trimmedRepository, number)
}

func (program *Program) currentCancelPendingPullRequestReviewAction() (actionsPopupAction, bool) {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return actionsPopupAction{}, false
	}

	pendingState, known := program.pendingPullRequestReviewStateForSummary(summary)
	if !known || strings.TrimSpace(pendingState.id) == "" {
		return actionsPopupAction{}, false
	}

	return program.cancelPendingPullRequestReviewAction(), true
}

func (program *Program) cancelPendingPullRequestReviewAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "cancel-pending-review",
		title:   cancelPendingPullRequestReviewActionTitle,
		icon:    actionsPopupCancelPendingReviewIcon,
		execute: program.executeCancelPendingPullRequestReviewAction,
	}
}

func (program *Program) selectedPendingPullRequestReviewActionTarget() (pendingPullRequestReviewActionTarget, bool) {
	if program.reviewSession.active {
		return pendingPullRequestReviewActionTarget{}, false
	}

	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pendingPullRequestReviewActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return pendingPullRequestReviewActionTarget{}, false
	}

	pendingState, known := program.pendingPullRequestReviewStateForSummary(summary)
	pendingReviewID := strings.TrimSpace(pendingState.id)
	if !known || pendingReviewID == "" {
		return pendingPullRequestReviewActionTarget{}, false
	}

	return pendingPullRequestReviewActionTarget{
		repository:      repository,
		number:          summary.Number,
		pendingReviewID: pendingReviewID,
		sourceFocus:     program.model.Focus(),
	}, true
}

func (program *Program) executeCancelPendingPullRequestReviewAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPendingPullRequestReviewActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !program.hasReviewMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.reviewMutations.DeletePullRequestReview(target.pendingReviewID); err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, "")
	program.reloadActivePullRequestsTab(gui)
	program.setFeedback(target.sourceFocus, pendingPullRequestReviewCanceledMessage)
	return actionsPopupActionResult{closePopup: true}
}
