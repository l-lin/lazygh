package tui

import "strings"

func (program *Program) footerPresenter() footerPresenter {
	if program == nil {
		return footerPresenter{}
	}

	submitAction := modalEditorSubmitAction
	submitFallback := "Alt+Enter"
	if program.overlayState.modalEditor != nil {
		submitAction = program.overlayState.modalEditor.submitAction()
		submitFallback = program.overlayState.modalEditor.submitHintFallback()
	}

	_, notificationSelectionVisible := program.selectedNotificationActionTarget()
	return footerPresenter{
		model:                        program.model,
		screenState:                  program.screenState(),
		keyResolver:                  program.keybindingLabelResolver(),
		helpVisible:                  program.overlayState.helpVisible,
		modalEditorVisible:           program.modalEditorVisible(),
		searchPromptVisible:          program.searchPromptVisible(),
		pullRequestBuildPopupVisible: program.pullRequestBuildRunPopupVisible(),
		assigneePickerVisible:        program.assigneePickerVisible(),
		notificationSelectionVisible: notificationSelectionVisible,
		actionsPopupActionCount:      len(program.currentActionsPopupActions()),
		modalEditorSubmitAction:      submitAction,
		modalEditorSubmitFallback:    submitFallback,
		paneSearchSummaries:          program.paneSearchSummaries(),
	}
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
