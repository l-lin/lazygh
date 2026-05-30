package tui

import "strings"

func (program *Program) footerPresenter() footerPresenter {
	if program == nil {
		return footerPresenter{}
	}
	if program.refreshReadCache.enabled && program.refreshReadCache.footerPresenterSet {
		return program.refreshReadCache.footerPresenter
	}

	submitAction := modalEditorSubmitAction
	submitFallback := "Alt+Enter"
	if program.modalEditorVisible() {
		submitAction = program.overlayState.modalEditor.submitAction()
		submitFallback = program.overlayState.modalEditor.submitHintFallback()
	}

	_, notificationSelectionVisible := program.selectedNotificationActionTarget()
	presenter := footerPresenter{
		model:                            program.model,
		screenState:                      program.screenState(),
		keyResolver:                      program.keybindingLabelResolver(),
		helpVisible:                      program.overlayState.helpVisible,
		modalEditorVisible:               program.modalEditorVisible(),
		searchPromptVisible:              program.searchPromptVisible(),
		pullRequestBuildPopupVisible:     program.pullRequestBuildRunPopupVisible(),
		assigneePickerVisible:            program.assigneePickerVisible(),
		notificationSelectionVisible:     notificationSelectionVisible,
		commentShortcutAvailable:         program.paneFooterCommentShortcutAvailable(),
		inlineCommentResolutionAvailable: program.inlineCommentResolutionShortcutAvailable(),
		inlineCommentResolutionHintLabel: program.inlineCommentResolutionShortcutHintLabel(),
		pullRequestBrowserAvailable:      program.pullRequestBrowserShortcutAvailable(),
		actionsPopupAvailable:            program.paneFooterActionsAvailable(),
		modalEditorSubmitAction:          submitAction,
		modalEditorSubmitFallback:        submitFallback,
		paneSearchSummaries:              program.paneSearchSummaries(),
	}
	program.cacheFooterPresenter(presenter)
	return presenter
}

func (program *Program) paneFooterCommentShortcutAvailable() bool {
	if program == nil || program.model == nil {
		return false
	}

	focus := program.screenState().ActiveView().Focus
	switch focus {
	case FocusPullRequestsView:
		if program.actionContext().IsReviewContext() {
			return false
		}
		_, ok := program.selectedPullRequestCommentTarget()
		return ok
	case FocusDetailView:
		switch program.inputContext().DetailInputMode {
		case DetailInputModePullRequestComment:
			_, ok := program.selectedPullRequestCommentTarget()
			return ok
		case DetailInputModeBrowserChangesInlineComment:
			_, err := program.selectedBrowserChangesInlineCommentSelection()
			return err == nil
		case DetailInputModeReviewInlineComment:
			_, err := program.selectedReviewInlineCommentSelection()
			return err == nil
		default:
			return false
		}
	default:
		return false
	}
}

func (program *Program) paneFooterActionsAvailable() bool {
	if program == nil || program.model == nil {
		return false
	}
	_, ok := paneFooterActionsActionID(program.screenState().ActiveView().Focus)
	return ok
}

func (program *Program) paneSearchSummaries() map[Focus]string {
	if program == nil || program.model == nil {
		return nil
	}

	detailQuery := program.model.appliedSearchQuery(SearchTargetDetail, MyPullRequestsTab)
	detailSearchSummary := ""
	if strings.TrimSpace(detailQuery) != "" {
		detailSearchSummary = searchSummaryText(detailQuery, len(program.currentDetailDocument(nil).searchMatches(detailQuery)))
	}
	if program.actionContext().IsReviewContext() {
		reviewTreeQuery := program.reviewFileTreeSearchQuery()
		return map[Focus]string{
			FocusPullRequestsView: searchSummaryText(reviewTreeQuery, program.reviewFileTreeSearchMatchCount(reviewTreeQuery)),
			FocusDetailView:       detailSearchSummary,
		}
	}

	activePullRequestTab := program.model.ActivePullRequestTab()
	return map[Focus]string{
		FocusUserView:          searchSummaryText(program.model.appliedSearchQuery(SearchTargetUser, MyPullRequestsTab), len(program.model.VisibleUsers())),
		FocusPullRequestsView:  searchSummaryText(program.model.appliedSearchQuery(SearchTargetPullRequests, activePullRequestTab), len(program.model.VisiblePullRequests())),
		FocusNotificationsView: searchSummaryText(program.model.appliedSearchQuery(SearchTargetNotifications, MyPullRequestsTab), len(program.model.VisibleNotifications())),
		FocusDetailView:        detailSearchSummary,
	}
}
