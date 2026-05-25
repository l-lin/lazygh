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
		return program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == DescriptionDetailTab
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
		execute: actionsPopupExecuteErr(program.executeMarkPullRequestReadyForReviewAction),
	}
}

func (program *Program) convertPullRequestToDraftAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "convert-pull-request-to-draft",
		title:   convertPullRequestToDraftActionTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeConvertPullRequestToDraftAction),
	}
}

func (program *Program) closePullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "close-pull-request",
		title:   closePullRequestActionTitle,
		icon:    actionsPopupClosePullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeClosePullRequestAction),
	}
}

func (program *Program) reopenPullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "reopen-pull-request",
		title:   reopenPullRequestActionTitle,
		icon:    actionsPopupReopenPullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeReopenPullRequestAction),
	}
}

func (program *Program) squashMergePullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "squash-merge-pull-request",
		title:   squashMergePullRequestActionTitle,
		icon:    actionsPopupReviewApproveIcon,
		execute: actionsPopupExecuteErr(program.executeSquashMergePullRequestAction),
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
		execute: actionsPopupExecuteErr(program.executeEnablePullRequestAutoMergeAction),
	}
}

func (program *Program) disablePullRequestAutoMergeAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "disable-pull-request-auto-merge",
		title:   disablePullRequestAutoMergeActionTitle,
		icon:    actionsPopupClosePullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeDisablePullRequestAutoMergeAction),
	}
}

func (program *Program) updatePullRequestBranchAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "update-pull-request-branch",
		title:   updatePullRequestBranchActionTitle,
		icon:    actionsPopupRefreshPullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeUpdatePullRequestBranchAction),
	}
}

func (program *Program) executeMarkPullRequestReadyForReviewAction(gui *gocui.Gui) error {
	return program.executePullRequestLifecycleMutation(gui, pullRequestLifecycleMutationReadyForReview, "OPEN", false, pullRequestMarkedReadyForReviewSuccessMessage)
}

func (program *Program) executeConvertPullRequestToDraftAction(gui *gocui.Gui) error {
	return program.executePullRequestLifecycleMutation(gui, pullRequestLifecycleMutationConvertToDraft, "OPEN", true, pullRequestConvertedToDraftSuccessMessage)
}

func (program *Program) executeClosePullRequestAction(gui *gocui.Gui) error {
	return program.executePullRequestLifecycleMutation(gui, pullRequestLifecycleMutationClose, "CLOSED", program.currentPullRequestDraftState(), pullRequestClosedSuccessMessage)
}

func (program *Program) executeReopenPullRequestAction(gui *gocui.Gui) error {
	return program.executePullRequestLifecycleMutation(gui, pullRequestLifecycleMutationReopen, "OPEN", program.currentPullRequestDraftState(), pullRequestReopenedSuccessMessage)
}

func (program *Program) executeSquashMergePullRequestAction(gui *gocui.Gui) error {
	if strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) != squashMergePullRequestActionTitle {
		program.actionsPopupWidget.pendingConfirmationActionID = squashMergePullRequestActionTitle
		program.actionsPopupWidget.errorMessage = ""
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	return program.startSquashMergePullRequestMutation(gui)
}

func (program *Program) executeEnablePullRequestAutoMergeAction(gui *gocui.Gui) error {
	return program.executePullRequestAutoMergeMutation(gui, pullRequestAutoMergeMutationEnable, true, pullRequestAutoMergeEnabledSuccessMessage)
}

func (program *Program) executeDisablePullRequestAutoMergeAction(gui *gocui.Gui) error {
	return program.executePullRequestAutoMergeMutation(gui, pullRequestAutoMergeMutationDisable, false, pullRequestAutoMergeDisabledSuccessMessage)
}

func (program *Program) executeUpdatePullRequestBranchAction(gui *gocui.Gui) error {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return err
	}
	return program.dispatch(gui, MsgPullRequestBranchUpdateRequested{Target: target, Summary: summary})
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

func (program *Program) executePullRequestLifecycleMutation(gui *gocui.Gui, kind pullRequestLifecycleMutationKind, state string, isDraft bool, successMessage string) error {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return err
	}
	return program.dispatch(gui, MsgPullRequestLifecycleMutationRequested{
		Kind:           kind,
		Target:         target,
		Summary:        summary,
		State:          state,
		IsDraft:        isDraft,
		SuccessMessage: successMessage,
	})
}

func (program *Program) executePullRequestAutoMergeMutation(gui *gocui.Gui, kind pullRequestAutoMergeMutationKind, enabled bool, successMessage string) error {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return err
	}
	return program.dispatch(gui, MsgPullRequestAutoMergeMutationRequested{
		Kind:           kind,
		Target:         target,
		Summary:        summary,
		Enabled:        enabled,
		SuccessMessage: successMessage,
	})
}

func (program *Program) startSquashMergePullRequestMutation(gui *gocui.Gui) error {
	target, summary, err := program.selectedPullRequestMutationContext()
	if err != nil {
		return err
	}
	return program.dispatch(gui, MsgPullRequestSquashMergeRequested{Target: target, Summary: summary})
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
		result.needsRefresh = true
		program.pullRequestDetailCache[key] = result
		delete(program.pullRequestDetailLoadInFlight, key)
	}
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidatePersistentPullRequest(pullRequestRepositoryName(summary.Repository), summary.Number)
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
