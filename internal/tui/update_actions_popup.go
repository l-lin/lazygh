package tui

import (
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

func (program *Program) updateActionsPopupSearch(query string) {
	program.model.UpdateActionsPopupSearch(query, program.currentActionsPopupMatchingIndexes(query))
}

func (program *Program) syncVisibleActionsPopupSearchSelection() {
	if program == nil || program.model == nil || !program.model.ActionsPopupVisible() {
		return
	}
	program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
}

func (program *Program) applyActionsPopupRequestedMessage(requested Msg) []Cmd {
	switch actual := requested.(type) {
	case nil:
		return nil
	case MsgActionsPopupActionErrorHandled:
		return program.applyActionsPopupActionErrorHandled(actual)
	case MsgOpenPullRequestInBrowserRequested:
		return program.applyOpenPullRequestInBrowserRequested(actual)
	case MsgModalEditorOpened:
		program.applyModalEditorOpened(actual)
		return nil
	case MsgCopyPullRequestURLRequested:
		return program.applyCopyPullRequestURLRequested(actual)
	case MsgOpenLinkUnderCursorRequested:
		return program.applyOpenLinkUnderCursorRequested(actual)
	case MsgOpenAssigneePickerRequested:
		return program.applyOpenAssigneePickerRequested(actual)
	case MsgToggleAssigneePickerSelectionRequested:
		program.applyToggleAssigneePickerSelectionRequested(actual)
		return nil
	case MsgPullRequestBuildRunLoadRequested:
		return program.applyPullRequestBuildRunLoadRequested(actual)
	case MsgPullRequestBuildRunJobLogLoadRequested:
		return program.applyPullRequestBuildRunJobLogLoadRequested(actual)
	case MsgPullRequestBuildRunPopupOpened:
		program.applyPullRequestBuildRunPopupOpened(actual)
		return nil
	case MsgPullRequestCommentDeleteRequested:
		return program.applyPullRequestCommentDeleteRequested(actual)
	case MsgInlineCommentDeleteRequested:
		return program.applyInlineCommentDeleteRequested(actual)
	case MsgInlineCommentResolutionRequested:
		return program.applyInlineCommentResolutionRequested(actual)
	case MsgOpenThemePickerRequested:
		program.applyOpenThemePickerRequested()
		return nil
	case MsgThemePresetSelected:
		return program.applyThemePresetSelected(actual)
	case MsgAddReactionRequested:
		return program.applyAddReactionRequested(actual)
	case MsgReactionRemovalRequested:
		return program.applyReactionRemovalRequested(actual)
	case MsgOpenReactionPickerRequested:
		program.applyOpenReactionPickerRequested(actual)
		return nil
	case MsgNotificationReadRequested:
		return program.applyNotificationReadRequested(actual)
	case MsgNotificationDoneRequested:
		return program.applyNotificationDoneRequested(actual)
	case MsgAllNotificationsReadRequested:
		return program.applyAllNotificationsReadRequested()
	case MsgAllNotificationsDoneRequested:
		return program.applyAllNotificationsDoneRequested()
	case MsgOpenNotificationInBrowserRequested:
		return program.applyOpenNotificationInBrowserRequested()
	case MsgClearCacheRequested:
		return program.applyClearCacheRequested()
	case MsgStartPullRequestReviewRequested:
		return program.applyStartPullRequestReviewRequested(actual)
	case MsgApprovePullRequestRequested:
		return program.applyApprovePullRequestRequested(actual)
	case MsgReRequestPullRequestReviewRequested:
		return program.applyReRequestPullRequestReviewRequested(actual)
	case MsgReviewStoryRequested:
		return program.applyReviewStoryRequested(actual)
	case MsgRefreshPullRequestRequested:
		return program.applyRefreshPullRequestRequested(actual)
	case MsgRefreshPullRequestListRequested:
		return program.applyRefreshPullRequestListRequested()
	case MsgPullRequestLifecycleMutationRequested:
		return program.applyPullRequestLifecycleMutationRequested(actual)
	case MsgPullRequestAutoMergeMutationRequested:
		return program.applyPullRequestAutoMergeMutationRequested(actual)
	case MsgPullRequestBranchUpdateRequested:
		return program.applyPullRequestBranchUpdateRequested(actual)
	case MsgPullRequestSquashMergeRequested:
		return program.applyPullRequestSquashMergeRequested(actual)
	case MsgCancelPendingPullRequestReviewRequested:
		return program.applyCancelPendingPullRequestReviewRequested(actual)
	default:
		return program.applyActionsPopupActionErrorHandled(MsgActionsPopupActionErrorHandled{Err: errActionsPopupActionUnavailable})
	}
}

func (program *Program) applyActionsPopupActionRequested(message MsgActionsPopupActionRequested) []Cmd {
	if program == nil || program.model == nil || !program.model.ActionsPopupVisible() {
		return nil
	}
	if message.Action.requested == nil {
		return program.applyActionsPopupActionErrorHandled(MsgActionsPopupActionErrorHandled{Err: errActionsPopupActionUnavailable})
	}
	return program.applyActionsPopupRequestedMessage(message.Action.requested)
}

func (program *Program) applyClearCacheRequested() []Cmd {
	if program.pullRequestCache == nil {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) != clearCacheActionTitle {
		program.setActionsPopupPendingConfirmation(clearCacheActionTitle)
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	return []Cmd{clearPersistentCacheCmd{}}
}

func (program *Program) applyPersistentCacheCleared(message MsgPersistentCacheCleared) []Cmd {
	if message.Err != nil {
		program.setActionsPopupErrorMessage(message.Err.Error())
		return program.applyErrorReportedMessage(program.actionsPopupWidget.errorMessage)
	}
	if err := program.clearCachedData(); err != nil {
		program.setActionsPopupErrorMessage(err.Error())
		return program.applyErrorReportedMessage(program.actionsPopupWidget.errorMessage)
	}
	program.closeActionsPopupState()
	program.setFeedback(program.model.Focus(), clearCacheSuccessMessage)
	return nil
}

func (program *Program) clearCachedData() error {
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withWorkflowStateReset()
	})
	program.updateReviewStore(func(store reviewStore) reviewStore {
		store = store.withDiffWorkflowStateReset()
		return store.withPendingReviewCacheReset()
	})
	program.resetNotificationLoadState()
	program.clearGHCommandLoading()
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.resetPullRequestListLoadState()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
	program.syncPastedPullRequestTab()
	program.model.SetNotifications([]Item{notificationsLoadingItem()})
	return nil
}

func (program *Program) applyStartPullRequestReviewRequested(message MsgStartPullRequestReviewRequested) []Cmd {
	summary := message.Summary
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(startPullRequestReviewPopupRequest{summary: summary})
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
	program.clearDetailPendingPrefix()
	trimmedPendingReviewID := strings.TrimSpace(pendingReviewID)
	program.startReviewSessionState(reviewSessionStartDescriptor{
		mode:                         mode,
		sourceFocus:                  program.model.Focus(),
		sourceDetailTab:              program.detailState.activeTab,
		sourcePaneLayoutSize:         program.model.paneLayoutSize,
		sourceFullscreenPane:         program.model.fullscreenPane,
		sourceDetailFullscreenReturn: program.model.detailFullscreenReturnSize,
		summary:                      summary,
		pendingReviewID:              trimmedPendingReviewID,
		story:                        story,
	})
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
	program.clearReviewSession()
	program.invalidateReviewDiffRenderCache()
	program.setDetailActiveTab(sourceDetailTab)
	program.clearDetailPendingPrefix()
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

func (program *Program) applyPullRequestCustomSearchSubmitRequested(message MsgPullRequestCustomSearchSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestCustomSearchSubmitRequest{criteria: message.Criteria}}}
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

	program.setRuntimePullRequestSearches(searches)
	searches = append([]appconfig.PullRequestSearch(nil), program.runtimeConfig.pullRequestSearches...)
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(searches))
	for tab, rows := range preservedRows {
		program.model.SetPullRequestRows(tab, rows)
	}
	program.model.SetPullRequestRows(customTab, []PullRequestRow{{Item: program.pullRequestLoadingItem(customTab)}})
	program.syncPastedPullRequestTab()
	program.model.ClearPullRequestSearchQuery(customTab)
	program.setPullRequestsLoadStarted(customTab, false)
	program.setPullRequestsLoading(customTab, false)
	program.setPullRequestsCount(customTab, 0, false)

	state := program.model.ScreenState().WithViewTabs(sidePanelPullRequestsViewNumber, int(customTab), program.model.pullRequestScreenTabs()).FocusViewNumber(sidePanelPullRequestsViewNumber)
	program.model.applyBrowserScreenState(state)
	return customTab
}

func (program *Program) applyOpenAssigneePickerRequested(message MsgOpenAssigneePickerRequested) []Cmd {
	program.openAssigneePicker(message.Target)
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
	if strings.TrimSpace(message.Candidate.Login) == "" {
		return
	}
	program.toggleAssigneePickerSelectionState(message.Candidate)
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
	return program.queueActionsPopupAsyncRequest(updatePullRequestAssigneesPopupRequest{repository: repository, number: number, addLogins: addLogins, removeLogins: removeLogins, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyOpenReactionPickerRequested(message MsgOpenReactionPickerRequested) {
	program.openReactionPicker(message.Target)
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
}

func (program *Program) applyAddReactionRequested(message MsgAddReactionRequested) []Cmd {
	if reactionGroupViewerHasReacted(message.Target.reactionGroups, message.Content) {
		program.closeActionsPopupState()
		program.setFeedback(program.model.Focus(), pullRequestReactionAlreadyAddedMessage)
		return nil
	}

	return program.queueActionsPopupAsyncRequest(addReactionPopupRequest{target: message.Target, content: message.Content, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyOpenThemePickerRequested() {
	program.openThemePicker()
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

func (program *Program) applyThemePresetSaved(message MsgThemePresetSaved) []Cmd {
	if message.Err != nil {
		program.setActionsPopupErrorMessage(message.Err.Error())
		return program.applyErrorReportedMessage(program.actionsPopupWidget.errorMessage)
	}

	theme.ApplyPalette(theme.ResolvePaletteWithPreset(strings.TrimSpace(message.NormalizedName), theme.Palette{}))
	program.restylePullRequestRows()
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.clearActionsPopupErrorMessage()
	program.setFeedback(program.model.Focus(), "Theme changed to "+strings.TrimSpace(message.Label))
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return []Cmd{configureGUICmd{}}
}

func (program *Program) applyRefreshPullRequestListRequested() []Cmd {
	tab := program.model.ActivePullRequestTab()
	program.beginManualPullRequestListRefresh(tab, pullRequestListRefreshSuccessMessage)
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return []Cmd{reloadPullRequestsTabCmd{tab: tab}}
}

func (program *Program) applyRefreshPullRequestRequested(message MsgRefreshPullRequestRequested) []Cmd {
	target := message.Target
	summary := message.Summary
	if program.hasDetailQueries() {
		program.markPullRequestDetailNeedsRefresh(summary)
		if program.reviewModeActive() {
			program.markPullRequestDiffNeedsRefresh(summary)
		}
	}
	program.invalidatePersistentPullRequest(target.repository, target.number)
	program.beginManualPullRequestRefresh(summary, program.model.ActivePullRequestTab())
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
		updated := program.navigationState.reviewSession.summary
		mutate(&updated)
		program.setReviewSessionSummary(updated)
	}
}

func (program *Program) applyCancelPendingPullRequestReviewRequested(message MsgCancelPendingPullRequestReviewRequested) []Cmd {
	return program.queueActionsPopupAsyncRequest(cancelPendingPullRequestReviewPopupRequest{target: message.Target})
}
