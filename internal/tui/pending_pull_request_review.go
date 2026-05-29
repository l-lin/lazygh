package tui

import (
	"errors"
	"fmt"
	"strings"
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

func (program *Program) pendingPullRequestReviewStateForSummary(summary any) (pendingPullRequestReviewState, bool) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return pendingPullRequestReviewState{}, false
	}
	key := pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)
	if key == "" {
		return pendingPullRequestReviewState{}, false
	}

	state, ok := program.pendingPullRequestReviewCache[key]
	return state, ok
}

func (program *Program) setPendingPullRequestReviewState(summary any, pendingReviewID string) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	program.setPendingPullRequestReviewStateByIdentity(pullRequestRepositoryName(summaryValue.Repository), summaryValue.Number, pendingReviewID)
}

func (program *Program) setPendingPullRequestReviewStateByIdentity(repository string, number int, pendingReviewID string) {
	key := pullRequestKeyFromIdentity(repository, number)
	if key == "" {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withPendingPullRequestReviewCached(key, pendingPullRequestReviewState{id: strings.TrimSpace(pendingReviewID)})
	})
}

func (program *Program) forgetPendingPullRequestReviewState(repository string, number int) {
	key := pullRequestKeyFromIdentity(repository, number)
	if key == "" {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withoutPendingPullRequestReview(key)
	})
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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPendingPullRequestReviewActionTarget()
	if ok {
		if !program.hasReviewMutations() {
			requested = actionsPopupErrorRequested(errors.New("github loader is unavailable"))
		} else {
			requested = MsgCancelPendingPullRequestReviewRequested{Target: target}
		}
	}
	return actionsPopupAction{
		id:        "cancel-pending-review",
		title:     cancelPendingPullRequestReviewActionTitle,
		icon:      actionsPopupCancelPendingReviewIcon,
		requested: requested,
	}
}

func (program *Program) selectedPendingPullRequestReviewActionTarget() (pendingPullRequestReviewActionTarget, bool) {
	if program.reviewModeActive() {
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
