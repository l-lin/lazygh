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
	updatePullRequestBranchActionTitle              = "Update branch"
	squashMergePullRequestConfirmationPromptMessage = "Press Enter again to squash-merge PR"
	pullRequestMarkedReadyForReviewSuccessMessage   = "PR marked ready for review"
	pullRequestConvertedToDraftSuccessMessage       = "PR converted to draft"
	pullRequestClosedSuccessMessage                 = "PR closed"
	pullRequestReopenedSuccessMessage               = "PR reopened"
	pullRequestSquashMergedSuccessMessage           = "PR squash-merged"
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
		actions := []actionsPopupAction{program.convertPullRequestToDraftAction(), program.squashMergePullRequestAction()}
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
			return actionsPopupActionResult{err: err}
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
				program.setFeedback(program.model.Focus(), strings.TrimSpace(err.Error()))
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

func updatePullRequestBranchCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "update-branch", fmt.Sprintf("%d", number), "-R", repository)
}

func (program *Program) applyVisiblePullRequestLifecycleMutation(summary githubdomain.PullRequest, state string, isDraft bool) {
	if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return
	}

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
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

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{
			Title:          strings.TrimSpace(summary.Title),
			Number:         summary.Number,
			URL:            strings.TrimSpace(summary.URL),
			Body:           strings.TrimSpace(summary.Body),
			ReviewDecision: strings.TrimSpace(summary.ReviewDecision),
			State:          state,
			IsDraft:        isDraft,
		}}
	} else {
		result.detail.Title = firstNonEmpty(result.detail.Title, strings.TrimSpace(summary.Title))
		result.detail.URL = firstNonEmpty(result.detail.URL, strings.TrimSpace(summary.URL))
		result.detail.Body = firstNonEmpty(result.detail.Body, strings.TrimSpace(summary.Body))
		result.detail.ReviewDecision = firstNonEmpty(result.detail.ReviewDecision, strings.TrimSpace(summary.ReviewDecision))
		result.detail.State = state
		result.detail.IsDraft = isDraft
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
