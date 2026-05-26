package tui

import "unicode/utf8"

func (program *Program) searchViewPresenter() searchViewPresenter {
	if program == nil {
		return searchViewPresenter{}
	}

	searchText := ""
	searchCursor := 0
	var notificationRows []NotificationRow
	if program.model != nil {
		searchText = program.model.SearchDraft()
		searchCursor = utf8.RuneCountInString(searchText)
		notificationRows = program.model.NotificationRows()
	}
	if program.searchWidget.hasEditor() {
		searchText = program.searchWidget.editor.Text()
		searchCursor = program.searchWidget.editor.Cursor()
	}

	return searchViewPresenter{
		mode:                       program.modeDescriptor().Mode(),
		mainContentKind:            program.mainViewResolver().ContentKind,
		showsPullRequestDetailTabs: program.shouldShowPullRequestDetailTabs(),
		searchText:                 searchText,
		searchCursor:               searchCursor,
		notificationRows:           notificationRows,
	}
}

func (program *Program) userViewTitle() string {
	return program.searchViewPresenter().userViewTitle()
}

func (program *Program) detailViewTitle() string {
	return program.searchViewPresenter().detailViewTitle()
}

func (program *Program) notificationsViewTitle() string {
	return program.searchViewPresenter().notificationsViewTitle()
}

func (program *Program) pullRequestsViewTitle() string {
	return program.searchViewPresenter().pullRequestsViewTitle()
}

func (program *Program) currentSearchText() string {
	return program.searchViewPresenter().promptText()
}

func (program *Program) currentSearchCursor() int {
	return program.searchViewPresenter().promptCursor()
}
