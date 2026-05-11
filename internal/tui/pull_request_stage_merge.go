package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	markPullRequestReadyForReviewActionTitle        = "Mark ready for review"
	convertPullRequestToDraftActionTitle            = "Convert to draft"
	closePullRequestActionTitle                     = "Close PR"
	reopenPullRequestActionTitle                    = "Reopen PR"
	squashMergePullRequestActionTitle               = "Squash and merge PR"
	squashMergePullRequestConfirmationPromptMessage = "Press Enter again to squash-merge PR"
	pullRequestMarkedReadyForReviewSuccessMessage   = "PR marked ready for review"
	pullRequestConvertedToDraftSuccessMessage       = "PR converted to draft"
	pullRequestClosedSuccessMessage                 = "PR closed"
	pullRequestReopenedSuccessMessage               = "PR reopened"
	pullRequestSquashMergedSuccessMessage           = "PR squash-merged"
)

func (program *Program) currentPullRequestStageAndMergeActions() []actionsPopupAction {
	if !program.pullRequestStateMutationVisible() {
		return nil
	}

	status, ok := program.currentPullRequestStageAndMergeStatus()
	if !ok {
		return nil
	}

	switch status {
	case "DRAFT":
		return []actionsPopupAction{program.markPullRequestReadyForReviewAction(), program.closePullRequestAction()}
	case "OPEN":
		return []actionsPopupAction{program.convertPullRequestToDraftAction(), program.squashMergePullRequestAction(), program.closePullRequestAction()}
	case "CLOSED":
		return []actionsPopupAction{program.reopenPullRequestAction()}
	default:
		return nil
	}
}

func (program *Program) pullRequestStateMutationVisible() bool {
	if !program.isPullRequestContext() || program.reviewSession.active {
		return false
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return true
	case FocusDetailView:
		return program.shouldShowPullRequestDetailTabs() && program.activeDetailTab == DescriptionDetailTab
	default:
		return false
	}
}

func (program *Program) currentPullRequestStageAndMergeStatus() (string, bool) {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return "", false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return detailStatus(result.detail, summary), true
	}
	return effectivePullRequestStatus(summary.State, summary.IsDraft), true
}

func (program *Program) markPullRequestReadyForReviewAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "mark-pull-request-ready-for-review",
		title:   markPullRequestReadyForReviewActionTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeMarkPullRequestReadyForReviewAction,
	}
}

func (program *Program) convertPullRequestToDraftAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "convert-pull-request-to-draft",
		title:   convertPullRequestToDraftActionTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeConvertPullRequestToDraftAction,
	}
}

func (program *Program) closePullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "close-pull-request",
		title:   closePullRequestActionTitle,
		icon:    actionsPopupClosePullRequestIcon,
		execute: program.executeClosePullRequestAction,
	}
}

func (program *Program) reopenPullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "reopen-pull-request",
		title:   reopenPullRequestActionTitle,
		icon:    actionsPopupReopenPullRequestIcon,
		execute: program.executeReopenPullRequestAction,
	}
}

func (program *Program) squashMergePullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "squash-merge-pull-request",
		title:   squashMergePullRequestActionTitle,
		icon:    actionsPopupReviewApproveIcon,
		execute: program.executeSquashMergePullRequestAction,
	}
}

func (program *Program) executeMarkPullRequestReadyForReviewAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		"gh pr ready",
		func(repository string, number int) error {
			return program.githubLoader.MarkPullRequestReadyForReview(repository, number)
		},
		"OPEN",
		false,
		pullRequestMarkedReadyForReviewSuccessMessage,
	)
}

func (program *Program) executeConvertPullRequestToDraftAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		"gh pr ready",
		func(repository string, number int) error {
			return program.githubLoader.ConvertPullRequestToDraft(repository, number)
		},
		"OPEN",
		true,
		pullRequestConvertedToDraftSuccessMessage,
	)
}

func (program *Program) executeClosePullRequestAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		"gh pr close",
		func(repository string, number int) error {
			return program.githubLoader.ClosePullRequest(repository, number)
		},
		"CLOSED",
		program.currentPullRequestDraftState(),
		pullRequestClosedSuccessMessage,
	)
}

func (program *Program) executeReopenPullRequestAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		"gh pr reopen",
		func(repository string, number int) error {
			return program.githubLoader.ReopenPullRequest(repository, number)
		},
		"OPEN",
		program.currentPullRequestDraftState(),
		pullRequestReopenedSuccessMessage,
	)
}

func (program *Program) executeSquashMergePullRequestAction(_ *gocui.Gui) actionsPopupActionResult {
	if strings.TrimSpace(program.actionsPopupPendingConfirmationActionID) != squashMergePullRequestActionTitle {
		program.actionsPopupPendingConfirmationActionID = squashMergePullRequestActionTitle
		program.actionsPopupErrorMessage = ""
		return actionsPopupActionResult{}
	}

	program.clearActionsPopupPendingConfirmation()
	return program.executePullRequestLifecycleMutation(
		"gh pr merge",
		func(repository string, number int) error {
			return program.githubLoader.SquashMergePullRequest(repository, number)
		},
		"MERGED",
		false,
		pullRequestSquashMergedSuccessMessage,
	)
}

func (program *Program) currentPullRequestDraftState() bool {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return result.detail.IsDraft || summary.IsDraft
	}
	return summary.IsDraft
}

func (program *Program) executePullRequestLifecycleMutation(commandName string, mutate func(string, int) error, state string, isDraft bool, successMessage string) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := mutate(target.repository, target.number); err != nil {
		return actionsPopupActionResult{err: normalizedPullRequestMutationError(err, commandName)}
	}

	program.applyVisiblePullRequestLifecycleMutation(summary, state, isDraft)
	program.setFeedback(program.model.Focus(), successMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) applyVisiblePullRequestLifecycleMutation(summary githubcli.PullRequest, state string, isDraft bool) {
	if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return
	}

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubcli.PullRequest) {
		current.State = state
		current.IsDraft = isDraft
		if strings.EqualFold(state, "MERGED") {
			current.ReviewDecision = ""
			current.ReviewRequests = nil
			current.Mergeable = ""
			current.MergeStateStatus = ""
			current.StatusCheckRollupState = ""
		}
	})
	program.mutateOrSeedPullRequestDetail(summary, state, isDraft)
	program.invalidatePullRequestMutationCaches(summary)
}

func (program *Program) mutateLoadedPullRequestSummaries(identity githubcli.PullRequest, mutate func(*githubcli.PullRequest)) {
	if program == nil || program.model == nil {
		return
	}

	for _, tab := range program.model.PullRequestTabs() {
		rows := program.model.PullRequestRows(tab)
		if len(rows) == 0 {
			continue
		}

		updatedRows := append([]PullRequestRow(nil), rows...)
		updated := false
		for index, row := range rows {
			if row.Summary == nil || !samePullRequestIdentity(*row.Summary, identity) {
				continue
			}

			summary := *row.Summary
			mutate(&summary)
			updatedRows[index] = pullRequestRow(summary)
			updated = true
		}
		if updated {
			program.model.SetPullRequestRows(tab, updatedRows)
		}
	}

	if program.openedPullRequestSummary != nil && samePullRequestIdentity(*program.openedPullRequestSummary, identity) {
		updated := *program.openedPullRequestSummary
		mutate(&updated)
		program.pinOpenedPullRequestSummary(program.openedPullRequestTab, updated)
	}
	if samePullRequestIdentity(program.reviewSession.summary, identity) {
		mutate(&program.reviewSession.summary)
	}
}

func (program *Program) mutateOrSeedPullRequestDetail(summary githubcli.PullRequest, state string, isDraft bool) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: githubcli.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   state,
			IsDraft: isDraft,
		}}
	} else {
		result.detail.Title = firstNonEmpty(result.detail.Title, strings.TrimSpace(summary.Title))
		result.detail.URL = firstNonEmpty(result.detail.URL, strings.TrimSpace(summary.URL))
		result.detail.Body = firstNonEmpty(result.detail.Body, strings.TrimSpace(summary.Body))
		result.detail.State = state
		result.detail.IsDraft = isDraft
	}
	if strings.EqualFold(state, "MERGED") {
		result.detail.Mergeable = ""
		result.detail.MergeStateStatus = ""
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	program.pullRequestDetailCache[key] = result
	delete(program.pullRequestDetailLoadInFlight, key)
}

func (program *Program) invalidatePullRequestMutationCaches(summary githubcli.PullRequest) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	delete(program.pullRequestDiffCache, key)
	delete(program.pullRequestDiffLoadInFlight, key)
	program.forgetPendingPullRequestReviewState(pullRequestRepositoryName(summary.Repository), summary.Number)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidatePersistentPullRequest(pullRequestRepositoryName(summary.Repository), summary.Number)
}

func normalizedPullRequestMutationError(err error, commandName string) error {
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(err.Error())
	prefix := "run `" + strings.TrimSpace(commandName) + "`:"
	message = strings.TrimSpace(strings.TrimPrefix(message, prefix))
	if strings.HasPrefix(message, "exit status ") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if message == "" {
		return err
	}
	return errors.New(message)
}
