package tui

import "unicode/utf8"

func (program *Program) actionsPopupPresenter() actionsPopupPresenter {
	if program == nil {
		return actionsPopupPresenter{}
	}

	query := ""
	filteredActionCount := 0
	searchText := ""
	searchCursor := 0
	if program.model != nil {
		query = program.model.ActionsPopupSearchQuery()
		filteredActionCount = len(program.model.ActionsPopupFilteredActionIndexes())
		searchText = query
		searchCursor = utf8.RuneCountInString(query)
	}
	if program.actionsPopupWidget.searchEditor != nil {
		searchText = program.actionsPopupWidget.searchEditor.Text()
		searchCursor = program.actionsPopupWidget.searchEditor.Cursor()
	}

	return actionsPopupPresenter{
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
}

func (program *Program) actionsPopupFrame(maxX int, maxY int) paneFrame {
	return program.actionsPopupPresenter().frame(maxX, program.layoutContentHeight(maxY))
}

func (program *Program) actionsPopupSearchFrame(maxX int, maxY int) paneFrame {
	return program.actionsPopupPresenter().searchFrame(maxX, program.layoutContentHeight(maxY))
}

func (program *Program) actionsPopupListFrame(maxX int, maxY int) paneFrame {
	return program.actionsPopupPresenter().listFrame(maxX, program.layoutContentHeight(maxY))
}

func (program *Program) actionsPopupHeight(contentMaxY int) int {
	return program.actionsPopupPresenter().height(contentMaxY)
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
