package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	markPullRequestReadyForReviewActionTitle        = "Mark ready for review"
	convertPullRequestToDraftActionTitle            = "Convert to draft"
	closePullRequestActionTitle                     = "Close PR"
	reopenPullRequestActionTitle                    = "Reopen PR"
	squashMergePullRequestActionTitle               = "Squash and merge PR"
	enablePullRequestAutoMergeActionTitle           = "Enable auto-merge"
	disablePullRequestAutoMergeActionTitle          = "Disable auto-merge"
	updatePullRequestBranchActionTitle              = "Update branch"
	squashMergePullRequestConfirmationPromptMessage = "Press Enter again to squash-merge PR"
	pullRequestMarkedReadyForReviewSuccessMessage   = "PR marked ready for review"
	pullRequestConvertedToDraftSuccessMessage       = "PR converted to draft"
	pullRequestClosedSuccessMessage                 = "PR closed"
	pullRequestReopenedSuccessMessage               = "PR reopened"
	pullRequestSquashMergedSuccessMessage           = "PR squash-merged"
	pullRequestAutoMergeEnabledSuccessMessage       = "PR auto-merge enabled"
	pullRequestAutoMergeDisabledSuccessMessage      = "PR auto-merge disabled"
	pullRequestBranchUpdatedSuccessMessage          = "PR branch updated"
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
		actions := []actionsPopupAction{program.markPullRequestReadyForReviewAction()}
		if program.currentPullRequestCanUpdateBranch() {
			actions = append(actions, program.updatePullRequestBranchAction())
		}
		return append(actions, program.closePullRequestAction())
	case "OPEN":
		actions := []actionsPopupAction{program.convertPullRequestToDraftAction()}
		if program.currentPullRequestShouldOfferAutoMerge() {
			actions = append(actions, program.currentPullRequestAutoMergeAction())
		} else {
			actions = append(actions, program.squashMergePullRequestAction())
		}
		if program.currentPullRequestCanUpdateBranch() {
			actions = append(actions, program.updatePullRequestBranchAction())
		}
		return append(actions, program.closePullRequestAction())
	case "CLOSED":
		return []actionsPopupAction{program.reopenPullRequestAction()}
	default:
		return nil
	}
}

func (program *Program) pullRequestStateMutationVisible() bool {
	if !program.isPullRequestContext() || program.reviewModeActive() {
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

func (program *Program) currentPullRequestAutoMergeAction() actionsPopupAction {
	if program.currentPullRequestAutoMergeEnabled() {
		return program.disablePullRequestAutoMergeAction()
	}
	return program.enablePullRequestAutoMergeAction()
}

func (program *Program) enablePullRequestAutoMergeAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "enable-pull-request-auto-merge",
		title:   enablePullRequestAutoMergeActionTitle,
		icon:    actionsPopupReviewApproveIcon,
		execute: program.executeEnablePullRequestAutoMergeAction,
	}
}

func (program *Program) disablePullRequestAutoMergeAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "disable-pull-request-auto-merge",
		title:   disablePullRequestAutoMergeActionTitle,
		icon:    actionsPopupClosePullRequestIcon,
		execute: program.executeDisablePullRequestAutoMergeAction,
	}
}

func (program *Program) updatePullRequestBranchAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "update-pull-request-branch",
		title:   updatePullRequestBranchActionTitle,
		icon:    actionsPopupRefreshPullRequestIcon,
		execute: program.executeUpdatePullRequestBranchAction,
	}
}

func (program *Program) executeMarkPullRequestReadyForReviewAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		gui,
		"gh pr ready",
		func(target pullRequestActionTarget) string {
			return pullRequestReadyCommand(target.repository, target.number, false)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.MarkPullRequestReadyForReview(repository, number)
		},
		"OPEN",
		false,
		pullRequestMarkedReadyForReviewSuccessMessage,
	)
}

func (program *Program) executeConvertPullRequestToDraftAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		gui,
		"gh pr ready",
		func(target pullRequestActionTarget) string {
			return pullRequestReadyCommand(target.repository, target.number, true)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.ConvertPullRequestToDraft(repository, number)
		},
		"OPEN",
		true,
		pullRequestConvertedToDraftSuccessMessage,
	)
}

func (program *Program) executeClosePullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		gui,
		"gh pr close",
		func(target pullRequestActionTarget) string {
			return closePullRequestCommand(target.repository, target.number)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.ClosePullRequest(repository, number)
		},
		"CLOSED",
		program.currentPullRequestDraftState(),
		pullRequestClosedSuccessMessage,
	)
}

func (program *Program) executeReopenPullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestLifecycleMutation(
		gui,
		"gh pr reopen",
		func(target pullRequestActionTarget) string {
			return reopenPullRequestCommand(target.repository, target.number)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.ReopenPullRequest(repository, number)
		},
		"OPEN",
		program.currentPullRequestDraftState(),
		pullRequestReopenedSuccessMessage,
	)
}

func (program *Program) executeSquashMergePullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	if strings.TrimSpace(program.actionsPopupPendingConfirmationActionID) != squashMergePullRequestActionTitle {
		program.actionsPopupPendingConfirmationActionID = squashMergePullRequestActionTitle
		program.actionsPopupErrorMessage = ""
		return actionsPopupActionResult{}
	}

	program.clearActionsPopupPendingConfirmation()
	return program.startSquashMergePullRequestMutation(gui)
}

func (program *Program) executeEnablePullRequestAutoMergeAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestAutoMergeMutation(
		gui,
		func(target pullRequestActionTarget) string {
			return enablePullRequestAutoMergeCommand(target.repository, target.number)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.EnablePullRequestAutoMerge(repository, number)
		},
		true,
		pullRequestAutoMergeEnabledSuccessMessage,
	)
}

func (program *Program) executeDisablePullRequestAutoMergeAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.executePullRequestAutoMergeMutation(
		gui,
		func(target pullRequestActionTarget) string {
			return disablePullRequestAutoMergeCommand(target.repository, target.number)
		},
		func(repository string, number int) error {
			return program.pullRequestMutations.DisablePullRequestAutoMerge(repository, number)
		},
		false,
		pullRequestAutoMergeDisabledSuccessMessage,
	)
}

func (program *Program) executeUpdatePullRequestBranchAction(gui *gocui.Gui) actionsPopupActionResult {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	return program.startActionsPopupAsyncGHCommand(gui, updatePullRequestBranchCommand(target.repository, target.number), func() error {
		return normalizedPullRequestMutationError(program.pullRequestMutations.UpdatePullRequestBranch(target.repository, target.number), "gh pr update-branch")
	}, func() {
		program.applyVisiblePullRequestBranchUpdate(summary)
		program.setFeedback(program.model.Focus(), pullRequestBranchUpdatedSuccessMessage)
	})
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

func (program *Program) currentPullRequestShouldOfferAutoMerge() bool {
	if program.currentPullRequestMergeabilityStatus() == pullRequestOverviewStatusFailure {
		return false
	}
	return program.currentPullRequestHasUnmetMergeRequirements()
}

func (program *Program) currentPullRequestHasUnmetMergeRequirements() bool {
	for _, status := range []pullRequestOverviewStatus{
		program.currentPullRequestReviewDecisionStatus(),
		program.currentPullRequestStatusCheckStatus(),
		program.currentPullRequestMergeabilityStatus(),
	} {
		switch status {
		case pullRequestOverviewStatusPending, pullRequestOverviewStatusFailure:
			return true
		}
	}
	return false
}

func (program *Program) currentPullRequestHasOngoingBuilds() bool {
	return program.currentPullRequestStatusCheckStatus() == pullRequestOverviewStatusPending
}

func (program *Program) currentPullRequestReviewDecisionStatus() pullRequestOverviewStatus {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestOverviewStatusMuted
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return pullRequestOverviewStatusForReviewDecision(firstNonEmpty(result.detail.ReviewDecision, summary.ReviewDecision))
	}
	return pullRequestOverviewStatusForReviewDecision(summary.ReviewDecision)
}

func (program *Program) currentPullRequestStatusCheckStatus() pullRequestOverviewStatus {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestOverviewStatusMuted
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil && len(result.detail.StatusCheckRollup) > 0 {
		return pullRequestOverviewStatusForChecks(result.detail.StatusCheckRollup)
	}
	return pullRequestOverviewStatusForStatusCheckRollupState(summary.StatusCheckRollupState)
}

func (program *Program) currentPullRequestMergeabilityStatus() pullRequestOverviewStatus {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestOverviewStatusMuted
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return pullRequestOverviewStatusForMergeability(firstNonEmpty(result.detail.Mergeable, summary.Mergeable), firstNonEmpty(result.detail.MergeStateStatus, summary.MergeStateStatus))
	}
	return pullRequestOverviewStatusForMergeability(summary.Mergeable, summary.MergeStateStatus)
}

func (program *Program) currentPullRequestAutoMergeEnabled() bool {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return result.detail.AutoMergeRequest != nil
	}
	return summary.AutoMergeRequest != nil
}

func (program *Program) currentPullRequestCanUpdateBranch() bool {
	summary, ok := program.currentPullRequestSummary()
	if !ok || !strings.EqualFold(strings.TrimSpace(summary.State), "OPEN") {
		return false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return pullRequestOutOfDateWithBase(result.detail)
	}
	return strings.EqualFold(strings.TrimSpace(summary.MergeStateStatus), "BEHIND")
}

func (program *Program) selectedPullRequestMutationContext() (pullRequestActionTarget, githubdomain.PullRequest, error) {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return pullRequestActionTarget{}, githubdomain.PullRequest{}, errActionsPopupActionUnavailable
	}
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestActionTarget{}, githubdomain.PullRequest{}, errActionsPopupActionUnavailable
	}
	if !program.hasPullRequestMutations() {
		return pullRequestActionTarget{}, githubdomain.PullRequest{}, errors.New("github loader is unavailable")
	}
	return target, summary, nil
}

func (program *Program) executePullRequestLifecycleMutation(gui *gocui.Gui, commandName string, command func(pullRequestActionTarget) string, mutate func(string, int) error, state string, isDraft bool, successMessage string) actionsPopupActionResult {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	return program.startActionsPopupAsyncGHCommand(gui, command(target), func() error {
		return normalizedPullRequestMutationError(mutate(target.repository, target.number), commandName)
	}, func() {
		program.applyVisiblePullRequestLifecycleMutation(summary, state, isDraft)
		program.setFeedback(program.model.Focus(), successMessage)
	})
}

func (program *Program) executePullRequestAutoMergeMutation(gui *gocui.Gui, command func(pullRequestActionTarget) string, mutate func(string, int) error, enabled bool, successMessage string) actionsPopupActionResult {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	return program.startActionsPopupAsyncGHCommand(gui, command(target), func() error {
		return normalizedPullRequestMutationError(mutate(target.repository, target.number), "gh pr merge")
	}, func() {
		program.applyVisiblePullRequestAutoMergeMutation(summary, enabled)
		program.setFeedback(program.model.Focus(), successMessage)
	})
}

func (program *Program) startSquashMergePullRequestMutation(gui *gocui.Gui) actionsPopupActionResult {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	command := squashMergePullRequestCommand(target.repository, target.number)
	program.startGHCommandLoading(command)
	if gui == nil {
		if err := program.runSquashMergePullRequestMutation(target); err != nil {
			program.clearGHCommandLoading()
			return actionsPopupActionResult{err: newTransientErrorPopupActionError(err)}
		}
		program.clearGHCommandLoading()
		program.applyVisiblePullRequestLifecycleMutation(summary, "MERGED", false)
		program.setFeedback(program.model.Focus(), pullRequestSquashMergedSuccessMessage)
		return actionsPopupActionResult{closePopup: true}
	}

	program.asyncRunner.Go(func() {
		err := program.runSquashMergePullRequestMutation(target)
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.clearGHCommandLoading()
			if err != nil {
				program.reportError(gui, strings.TrimSpace(err.Error()))
				return program.refreshViews(gui)
			}
			program.applyVisiblePullRequestLifecycleMutation(summary, "MERGED", false)
			program.setFeedback(program.model.Focus(), pullRequestSquashMergedSuccessMessage)
			return program.refreshViews(gui)
		})
	})
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) runSquashMergePullRequestMutation(target pullRequestActionTarget) error {
	if err := program.pullRequestMutations.SquashMergePullRequest(target.repository, target.number); err != nil {
		return normalizedPullRequestMutationError(err, "gh pr merge")
	}
	return nil
}

func pullRequestReadyCommand(repository string, number int, undo bool) string {
	command := formatStatusLineCommand("gh", "pr", "ready", fmt.Sprintf("%d", number), "-R", repository)
	if undo {
		return command + " --undo"
	}
	return command
}

func closePullRequestCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "close", fmt.Sprintf("%d", number), "-R", repository)
}

func reopenPullRequestCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "reopen", fmt.Sprintf("%d", number), "-R", repository)
}

func squashMergePullRequestCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "merge", fmt.Sprintf("%d", number), "-R", repository, "--squash")
}

func enablePullRequestAutoMergeCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "merge", fmt.Sprintf("%d", number), "-R", repository, "--auto", "--squash")
}

func disablePullRequestAutoMergeCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "merge", fmt.Sprintf("%d", number), "-R", repository, "--disable-auto")
}

func updatePullRequestBranchCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "update-branch", fmt.Sprintf("%d", number), "-R", repository)
}

func pullRequestHasOngoingBuilds(checks []githubdomain.PullRequestStatusCheck) bool {
	for _, check := range checks {
		if !strings.EqualFold(strings.TrimSpace(check.Status), "COMPLETED") {
			return true
		}
	}
	return false
}

func pullRequestStatusCheckRollupStateIsOngoing(state string) bool {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "PENDING", "EXPECTED":
		return true
	default:
		return false
	}
}

func (program *Program) applyVisiblePullRequestLifecycleMutation(summary githubdomain.PullRequest, state string, isDraft bool) {
	if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return
	}

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
		current.State = state
		current.IsDraft = isDraft
		if !strings.EqualFold(state, "OPEN") {
			current.AutoMergeRequest = nil
		}
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

func (program *Program) mutateLoadedPullRequestSummaries(identity githubdomain.PullRequest, mutate func(*githubdomain.PullRequest)) {
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

func (program *Program) mutateOrSeedPullRequestDetail(summary githubdomain.PullRequest, state string, isDraft bool) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	autoMergeRequest := clonePullRequestAutoMergeRequest(summary.AutoMergeRequest)
	if !strings.EqualFold(state, "OPEN") {
		autoMergeRequest = nil
	}

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{
			Title:            strings.TrimSpace(summary.Title),
			Number:           summary.Number,
			URL:              strings.TrimSpace(summary.URL),
			Body:             strings.TrimSpace(summary.Body),
			ReviewDecision:   strings.TrimSpace(summary.ReviewDecision),
			State:            state,
			IsDraft:          isDraft,
			AutoMergeRequest: autoMergeRequest,
		}}
	} else {
		result.detail.Title = firstNonEmpty(result.detail.Title, strings.TrimSpace(summary.Title))
		result.detail.URL = firstNonEmpty(result.detail.URL, strings.TrimSpace(summary.URL))
		result.detail.Body = firstNonEmpty(result.detail.Body, strings.TrimSpace(summary.Body))
		result.detail.ReviewDecision = firstNonEmpty(result.detail.ReviewDecision, strings.TrimSpace(summary.ReviewDecision))
		result.detail.State = state
		result.detail.IsDraft = isDraft
		result.detail.AutoMergeRequest = autoMergeRequest
	}
	if strings.EqualFold(state, "MERGED") {
		result.detail.ReviewDecision = ""
		result.detail.ReviewRequests = nil
		result.detail.Mergeable = ""
		result.detail.MergeStateStatus = ""
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	program.pullRequestDetailCache[key] = result
	delete(program.pullRequestDetailLoadInFlight, key)
}

func (program *Program) applyVisiblePullRequestAutoMergeMutation(summary githubdomain.PullRequest, enabled bool) {
	autoMergeRequest := clonePullRequestAutoMergeRequest(summary.AutoMergeRequest)
	if enabled {
		autoMergeRequest = &githubdomain.PullRequestAutoMergeRequest{}
	} else {
		autoMergeRequest = nil
	}

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
		current.AutoMergeRequest = clonePullRequestAutoMergeRequest(autoMergeRequest)
	})

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}
	if result, ok := program.pullRequestDetailCache[key]; ok && result.err == nil {
		result.detail.AutoMergeRequest = clonePullRequestAutoMergeRequest(autoMergeRequest)
		result.sourceUpdatedAt = ""
		program.pullRequestDetailCache[key] = result
		delete(program.pullRequestDetailLoadInFlight, key)
	}
}

func (program *Program) applyVisiblePullRequestBranchUpdate(summary githubdomain.PullRequest) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key != "" {
		if result, ok := program.pullRequestDetailCache[key]; ok && result.err == nil {
			result.detail.OutOfDateWithBase = false
			if strings.EqualFold(strings.TrimSpace(result.detail.MergeStateStatus), "BEHIND") {
				result.detail.MergeStateStatus = ""
			}
			result.sourceUpdatedAt = ""
			result.needsRefresh = true
			program.pullRequestDetailCache[key] = result
			delete(program.pullRequestDetailLoadInFlight, key)
		}
	}

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
		if strings.EqualFold(strings.TrimSpace(current.MergeStateStatus), "BEHIND") {
			current.MergeStateStatus = ""
		}
	})
	program.invalidatePullRequestMutationCaches(summary)
}

func (program *Program) invalidatePullRequestMutationCaches(summary githubdomain.PullRequest) {
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

func normalizedPullRequestMutationError(err error, _ string) error {
	return normalizeGHCommandError(err)
}
