package tui

import "unicode/utf8"

func (program *Program) actionsPopupPresenter() actionsPopupPresenter {
	if program == nil {
		return actionsPopupPresenter{}
	}
	if program.refreshReadCache.enabled && program.refreshReadCache.actionsPopupPresenterSet {
		return program.refreshReadCache.actionsPopupPresenter
	}

	query := ""
	filteredActionCount := 0
	searchText := ""
	searchCursor := 0
	if program.model != nil {
		query = program.model.ActionsPopupSearchQuery()
		filteredActionCount = len(program.currentActionsPopupFilteredIndexes())
		searchText = query
		searchCursor = utf8.RuneCountInString(query)
	}
	if program.actionsPopupWidget.hasSearchEditor() {
		searchText = program.actionsPopupWidget.searchEditor.Text()
		searchCursor = program.actionsPopupWidget.searchEditor.Cursor()
	}

	presenter := actionsPopupPresenter{
		assigneePickerVisible: program.assigneePickerVisible(),
		themePickerVisible:    program.themePickerVisible(),
		reactionPickerVisible: program.reactionPickerVisible(),
		errorMessage:          program.actionsPopupWidget.errorMessage,
		confirmationMessage:   program.actionsPopupConfirmationMessage(),
		searchQuery:           query,
		searchText:            searchText,
		searchCursor:          searchCursor,
		filteredActionCount:   filteredActionCount,
		totalActionCount:      len(program.currentActionsPopupActions()),
		renderedLineCount:     program.currentActionsPopupRenderedLineCount(),
	}
	if program.refreshReadCache.enabled {
		program.refreshReadCache.actionsPopupPresenter = presenter
		program.refreshReadCache.actionsPopupPresenterSet = true
	}
	return presenter
}

func (program *Program) currentActionsPopupSearchText() string {
	return program.actionsPopupPresenter().promptText()
}

func (program *Program) currentActionsPopupSearchCursor() int {
	return program.actionsPopupPresenter().promptCursor()
}

func (program *Program) actionsPopupTitle() string {
	return program.actionsPopupPresenter().title()
}

func (program *Program) actionsPopupFooter() string {
	return program.actionsPopupPresenter().footer()
}

func (program *Program) emptyActionsPopupMessage() string {
	return program.actionsPopupPresenter().emptyMessage()
}
