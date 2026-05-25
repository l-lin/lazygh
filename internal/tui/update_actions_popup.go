package tui

import (
	"errors"
	"fmt"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

func (program *Program) clearActionsPopupPendingConfirmation() {
	program.actionsPopupWidget.pendingConfirmationActionID = ""
}

func (program *Program) updateActionsPopupSearch(query string) {
	program.model.UpdateActionsPopupSearch(query, program.currentActionsPopupMatchingIndexes(query))
}

func (program *Program) applyClearCacheRequested() {
	if program.pullRequestCache == nil {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return
	}
	if strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) != clearCacheActionTitle {
		program.actionsPopupWidget.pendingConfirmationActionID = clearCacheActionTitle
		program.actionsPopupWidget.errorMessage = ""
		return
	}

	program.clearActionsPopupPendingConfirmation()
	if err := program.clearCachedData(); err != nil {
		program.actionsPopupWidget.errorMessage = strings.TrimSpace(err.Error())
		program.reportError(program.gui, program.actionsPopupWidget.errorMessage)
		return
	}
	program.closeActionsPopupState()
	program.setFeedback(program.model.Focus(), clearCacheSuccessMessage)
}

func (program *Program) clearCachedData() error {
	if program.pullRequestCache == nil {
		return fmt.Errorf("persistent cache is unavailable")
	}
	if err := program.pullRequestCache.Clear(); err != nil {
		return err
	}

	program.pullRequestDetailCache = map[string]pullRequestDetailResult{}
	program.pullRequestDetailLoadInFlight = map[string]bool{}
	program.pullRequestDiffCache = map[string]pullRequestDiffResult{}
	program.pullRequestDiffLoadInFlight = map[string]bool{}
	program.pendingPullRequestReviewCache = map[string]pendingPullRequestReviewState{}
	program.issueDetailCache = map[string]issueDetailResult{}
	program.issueDetailLoadInFlight = map[string]bool{}
	program.releaseDetailCache = map[string]releaseDetailResult{}
	program.releaseDetailLoadInFlight = map[string]bool{}
	program.notificationsLoadStarted = false
	program.notificationsLoading = false
	program.notificationsLoadingDetailMessage = ""
	program.ghCommandLoadingMessage = ""
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.resetPullRequestSearchState()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
	program.model.SetNotifications([]Item{notificationsLoadingItem()})
	return nil
}

func (program *Program) applyStartPullRequestReviewRequested(message MsgStartPullRequestReviewRequested) []Cmd {
	summary := message.Summary
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	command := formatStatusLineCommand("start", "review", repository, fmt.Sprintf("#%d", summary.Number))
	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		pendingReviewID, err := program.startPendingPullRequestReview(summary)
		if err != nil {
			return nil, err
		}
		return actionsPopupAsyncStartReviewSuccess{Summary: summary, PendingReviewID: pendingReviewID}, nil
	}}}
}

func (program *Program) startPendingPullRequestReview(summary githubdomain.PullRequest) (string, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return "", errors.New("missing pull request identity")
	}

	pendingReviewID, err := program.reviewMutations.StartPendingPullRequestReview(repository, summary.Number)
	if err != nil {
		return "", newTransientErrorPopupActionError(err)
	}
	return pendingReviewID, nil
}

func (program *Program) startReviewSession(summary any, pendingReviewID string) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	program.startReviewSessionWithMode(summaryValue, pendingReviewID, reviewSessionModeDiff, reviewStoryData{})
}

func (program *Program) startStoryReviewSession(summary any, pendingReviewID string, story reviewStoryData) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	program.startReviewSessionWithMode(summaryValue, pendingReviewID, reviewSessionModeStory, story)
}

func (program *Program) startReviewSessionWithMode(summary githubdomain.PullRequest, pendingReviewID string, mode reviewSessionMode, story reviewStoryData) {
	program.detailState.viewState.clearPendingPrefix()
	trimmedPendingReviewID := strings.TrimSpace(pendingReviewID)
	program.navigationState.reviewSession = reviewSessionState{
		active:                       true,
		mode:                         mode,
		sourceFocus:                  program.model.Focus(),
		sourceDetailTab:              program.detailState.activeTab,
		sourcePaneLayoutSize:         program.model.paneLayoutSize,
		sourceFullscreenPane:         program.model.fullscreenPane,
		sourceDetailFullscreenReturn: program.model.detailFullscreenReturnSize,
		summary:                      summary,
		pendingReviewID:              trimmedPendingReviewID,
		selectedFileTreeRow:          -1,
		collapsedTreeRowIDs:          map[string]bool{},
		collapsedThreadIDs:           map[string]bool{},
		story:                        story,
	}
	if trimmedPendingReviewID != "" {
		program.setPendingPullRequestReviewState(summary, trimmedPendingReviewID)
	}
	program.invalidateReviewDiffRenderCache()
	program.model.SetPaneLayoutSize(program.reviewModePaneLayoutSize())
	program.model.FocusPullRequestsView()
}

func (program *Program) restorePullRequestBrowserFromReviewMode() {
	if !program.reviewModeActive() {
		return
	}

	sourceFocus := program.navigationState.reviewSession.sourceFocus
	sourceDetailTab := program.navigationState.reviewSession.sourceDetailTab
	sourcePaneLayoutSize := program.navigationState.reviewSession.sourcePaneLayoutSize
	sourceFullscreenPane := program.navigationState.reviewSession.sourceFullscreenPane
	sourceDetailFullscreenReturn := program.navigationState.reviewSession.sourceDetailFullscreenReturn
	program.navigationState.reviewSession = reviewSessionState{}
	program.invalidateReviewDiffRenderCache()
	program.detailState.activeTab = sourceDetailTab
	program.detailState.viewState.clearPendingPrefix()
	program.model.SetPaneLayoutSize(sourcePaneLayoutSize)
	program.model.SetFullscreenPane(sourceFullscreenPane)
	program.model.SetDetailFullscreenReturnSize(sourceDetailFullscreenReturn)

	switch sourceFocus {
	case FocusDetailView:
		program.model.FocusDetailView()
	default:
		program.model.FocusPullRequestsView()
	}
}

func (program *Program) applyPullRequestCustomSearchSubmitted(message MsgPullRequestCustomSearchSubmitted) []Cmd {
	command := pullRequestCustomSearchCommand(message.Criteria)
	if len(command) == 0 {
		return nil
	}

	customTab := program.upsertPullRequestCustomSearch(appconfig.PullRequestSearch{Label: pullRequestCustomSearchLabel, Command: command})
	return []Cmd{reloadPullRequestsTabCmd{tab: customTab}}
}

func (program *Program) upsertPullRequestCustomSearch(search appconfig.PullRequestSearch) PullRequestTab {
	searches := append([]appconfig.PullRequestSearch(nil), appconfig.ResolvePullRequestSearches(program.runtimeConfig.pullRequestSearches)...)
	customTab, customTabExists := pullRequestCustomSearchTab(searches)
	preservedRows := make(map[PullRequestTab][]PullRequestRow, len(program.model.PullRequestTabs()))
	for _, tab := range program.model.PullRequestTabs() {
		if customTabExists && tab == customTab {
			continue
		}
		preservedRows[tab] = program.model.PullRequestRows(tab)
	}

	if customTabExists {
		searches[customTab] = search
	} else {
		customTab = PullRequestTab(len(searches))
		searches = append(searches, search)
	}

	program.runtimeConfig.pullRequestSearches = searches
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(searches))
	for tab, rows := range preservedRows {
		program.model.SetPullRequestRows(tab, rows)
	}
	program.model.SetPullRequestRows(customTab, []PullRequestRow{{Item: program.pullRequestLoadingItem(customTab)}})
	program.model.ClearPullRequestSearchQuery(customTab)
	program.setPullRequestsLoadStarted(customTab, false)
	program.setPullRequestsLoading(customTab, false)
	program.setPullRequestsCount(customTab, 0, false)

	state := program.model.ScreenState().WithViewTabs(sidePanelPullRequestsViewNumber, int(customTab), program.model.pullRequestScreenTabs()).FocusViewNumber(sidePanelPullRequestsViewNumber)
	program.model.applyBrowserScreenState(state)
	return customTab
}

func (program *Program) applyOpenAssigneePickerRequested(message MsgOpenAssigneePickerRequested) []Cmd {
	program.actionsPopupWidget.assigneePicker = newAssigneePickerState(message.Target, program.currentConnectedUserLogin(), program.currentConnectedUserName())
	program.actionsPopupWidget.searchEditor = nil
	program.actionsPopupWidget.errorMessage = ""
	program.model.OpenActionsPopup(program.currentAssigneePickerActionCount())
	program.updateActionsPopupSearch("")

	requestID := program.resetAssigneePickerSearch("")
	program.markAssigneePickerSearchLoading("")
	if requestID <= 0 || !program.hasPullRequestMutations() {
		return nil
	}
	return []Cmd{assigneePickerSearchCmd{RequestID: requestID, Query: "", DispatchLoading: false}}
}

func (program *Program) applyToggleAssigneePickerSelectionRequested(message MsgToggleAssigneePickerSelectionRequested) {
	if !program.assigneePickerVisible() {
		return
	}

	trimmedLogin := strings.TrimSpace(message.Candidate.Login)
	if trimmedLogin == "" {
		return
	}
	if program.actionsPopupWidget.assigneePicker.selectedLogins[trimmedLogin] {
		delete(program.actionsPopupWidget.assigneePicker.selectedLogins, trimmedLogin)
	} else {
		program.actionsPopupWidget.assigneePicker.selectedLogins[trimmedLogin] = true
	}
	program.actionsPopupWidget.assigneePicker.rememberCandidates([]githubdomain.PullRequestAuthor{message.Candidate})
}

func (program *Program) toggleAssigneePickerSelection(candidate githubdomain.PullRequestAuthor) error {
	if !program.assigneePickerVisible() {
		return errActionsPopupActionUnavailable
	}
	if strings.TrimSpace(candidate.Login) == "" {
		return errActionsPopupActionUnavailable
	}
	program.applyToggleAssigneePickerSelectionRequested(MsgToggleAssigneePickerSelectionRequested{Candidate: candidate})
	return nil
}

func (program *Program) applySubmitAssigneePickerRequested(message MsgSubmitAssigneePickerRequested) []Cmd {
	if len(message.AddLogins) == 0 && len(message.RemoveLogins) == 0 {
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
		return nil
	}

	repository := message.Repository
	number := message.Number
	addLogins := append([]string(nil), message.AddLogins...)
	removeLogins := append([]string(nil), message.RemoveLogins...)
	command := updatePullRequestAssigneesCommand(repository, number, addLogins, removeLogins)
	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		err := normalizedAssigneePickerError(program.pullRequestMutations.UpdatePullRequestAssignees(repository, number, addLogins, removeLogins))
		if err != nil {
			return nil, err
		}
		return actionsPopupAsyncPullRequestAssigneesUpdatedSuccess{
			Repository:   repository,
			Number:       number,
			AddLogins:    addLogins,
			RemoveLogins: removeLogins,
			Message:      pullRequestAssigneesUpdatedSuccessMessage,
		}, nil
	}}}
}

func (program *Program) applyOpenReactionPickerRequested(message MsgOpenReactionPickerRequested) {
	program.actionsPopupWidget.reactionPicker = &reactionPickerState{target: message.Target}
	program.actionsPopupWidget.searchEditor = nil
	program.actionsPopupWidget.errorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
}

func (program *Program) applyAddReactionRequested(message MsgAddReactionRequested) []Cmd {
	if reactionGroupViewerHasReacted(message.Target.reactionGroups, message.Content) {
		program.closeActionsPopupState()
		program.setFeedback(program.model.Focus(), pullRequestReactionAlreadyAddedMessage)
		return nil
	}

	content := message.Content
	target := message.Target
	command := formatStatusLineCommand("add", "reaction", strings.TrimSpace(string(content)))
	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.reactionMutations.AddReaction(target.subjectID, content); err != nil {
			return nil, newTransientErrorPopupActionError(err)
		}
		return actionsPopupAsyncReactionAddedSuccess{Target: target, Content: content}, nil
	}}}
}

func (program *Program) applyOpenThemePickerRequested() {
	program.actionsPopupWidget.themePicker = &themePickerState{}
	program.actionsPopupWidget.searchEditor = nil
	program.actionsPopupWidget.errorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
}

func (program *Program) applyThemePresetSelected(message MsgThemePresetSelected) []Cmd {
	normalizedName := strings.TrimSpace(message.NormalizedName)
	if normalizedName == "" {
		return nil
	}
	return []Cmd{saveThemePresetCmd{NormalizedName: normalizedName, Label: strings.TrimSpace(message.Label)}}
}

func (program *Program) restylePullRequestRows() {
	if program == nil || program.model == nil {
		return
	}

	for _, tab := range program.model.PullRequestTabs() {
		rows := program.model.PullRequestRows(tab)
		if len(rows) == 0 {
			continue
		}
		program.model.SetPullRequestRows(tab, restyledPullRequestRows(rows))
	}
}

func (program *Program) applyThemePresetSaved(message MsgThemePresetSaved) {
	if message.Err != nil {
		program.actionsPopupWidget.errorMessage = strings.TrimSpace(message.Err.Error())
		program.reportError(program.gui, program.actionsPopupWidget.errorMessage)
		return
	}

	theme.ApplyPalette(theme.ResolvePaletteWithPreset(strings.TrimSpace(message.NormalizedName), theme.Palette{}))
	program.restylePullRequestRows()
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.actionsPopupWidget.errorMessage = ""
	program.setFeedback(program.model.Focus(), "Theme changed to "+strings.TrimSpace(message.Label))
	if program.gui != nil {
		program.configureGUI(program.gui)
	}
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
}

func (program *Program) applyRefreshPullRequestListRequested() []Cmd {
	tab := program.model.ActivePullRequestTab()
	pendingOperations := 0
	if program.gui != nil && program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(tab) {
		pendingOperations++
	}
	program.beginManualRefresh(pullRequestListRefreshSuccessMessage, pendingOperations)
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return []Cmd{reloadPullRequestsTabCmd{tab: tab}}
}

func (program *Program) applyRefreshPullRequestRequested(message MsgRefreshPullRequestRequested) []Cmd {
	target := message.Target
	summary := message.Summary
	pendingOperations := 0
	if program.hasDetailQueries() {
		if program.markManualPullRequestDetailRefresh(summary) {
			pendingOperations++
		}
		program.markPullRequestDetailNeedsRefresh(summary)
	}
	if program.reviewModeActive() {
		if program.hasDetailQueries() {
			if program.markManualPullRequestDiffRefresh(summary) {
				pendingOperations++
			}
			program.markPullRequestDiffNeedsRefresh(summary)
		}
	} else if program.gui != nil && program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(program.model.ActivePullRequestTab()) {
		pendingOperations++
	}
	program.beginManualRefresh(pullRequestRefreshSuccessMessage, pendingOperations)
	program.invalidatePersistentPullRequest(target.repository, target.number)
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	if program.reviewModeActive() {
		return nil
	}
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
}

func (program *Program) applyPullRequestTitleEditApplied(message MsgPullRequestTitleEditApplied) []Cmd {
	program.optimisticallyUpdatePullRequestTitle(message.Target.repository, message.Target.number, message.Title)
	program.setFeedback(message.FeedbackTarget, pullRequestTitleEditSuccessMessage)
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
}

func (program *Program) applyPullRequestDescriptionEditApplied(message MsgPullRequestDescriptionEditApplied) []Cmd {
	program.optimisticallyUpdatePullRequestDescription(message.Target.repository, message.Target.number, message.Body)
	program.setFeedback(message.FeedbackTarget, pullRequestDescriptionEditSuccessMessage)
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
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

	if program.navigationState.openedPullRequestSummary != nil && samePullRequestIdentity(*program.navigationState.openedPullRequestSummary, identity) {
		updated := *program.navigationState.openedPullRequestSummary
		mutate(&updated)
		program.pinOpenedPullRequestSummary(program.navigationState.openedPullRequestTab, updated)
	}
	if samePullRequestIdentity(program.navigationState.reviewSession.summary, identity) {
		mutate(&program.navigationState.reviewSession.summary)
	}
}

func (program *Program) applyCancelPendingPullRequestReviewRequested(message MsgCancelPendingPullRequestReviewRequested) []Cmd {
	target := message.Target
	command := formatStatusLineCommand("cancel", "pending", "review", target.repository, fmt.Sprintf("#%d", target.number))
	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.reviewMutations.DeletePullRequestReview(target.pendingReviewID); err != nil {
			return nil, newTransientErrorPopupActionError(err)
		}
		return actionsPopupAsyncPendingReviewCanceledSuccess{Target: target}, nil
	}}}
}
