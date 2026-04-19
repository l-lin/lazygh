package tui

func (model *Model) OpenActionsPopup(actionCount int) {
	if actionCount <= 0 {
		model.actionsPopup = actionsPopupState{}
		return
	}

	model.actionsPopup = actionsPopupState{
		visible:               true,
		filteredActionIndexes: actionIndexes(actionCount),
		selectedActionIndex:   0,
	}
}

func (model *Model) CloseActionsPopup() {
	model.actionsPopup = actionsPopupState{}
}

func (model *Model) ActionsPopupVisible() bool {
	return model.actionsPopup.visible
}

func (model *Model) ActionsPopupSearchActive() bool {
	return model.actionsPopup.visible && model.actionsPopup.searchActive
}

func (model *Model) FocusActionsPopupSearch() {
	if !model.actionsPopup.visible {
		return
	}

	model.actionsPopup.searchActive = true
}

func (model *Model) BlurActionsPopupSearch() {
	if !model.actionsPopup.visible {
		return
	}

	model.actionsPopup.searchActive = false
}

func (model *Model) ActionsPopupSearchQuery() string {
	return model.actionsPopup.searchQuery
}

func (model *Model) ActionsPopupFilteredActionIndexes() []int {
	return append([]int(nil), model.actionsPopup.filteredActionIndexes...)
}

func (model *Model) ActionsPopupSelectedActionIndex() int {
	return model.actionsPopup.selectedActionIndex
}

func (model *Model) ActionsPopupSelectedVisibleIndex() int {
	return model.selectedVisibleIndex(model.actionsPopup.selectedActionIndex, model.actionsPopup.filteredActionIndexes)
}

func (model *Model) UpdateActionsPopupSearch(query string, filteredActionIndexes []int) {
	if !model.actionsPopup.visible {
		return
	}

	model.actionsPopup.searchQuery = query
	model.actionsPopup.filteredActionIndexes = append([]int(nil), filteredActionIndexes...)
	if len(model.actionsPopup.filteredActionIndexes) == 0 {
		model.actionsPopup.selectedActionIndex = 0
		return
	}
	if indexOfInt(model.actionsPopup.filteredActionIndexes, model.actionsPopup.selectedActionIndex) >= 0 {
		return
	}

	model.actionsPopup.selectedActionIndex = model.actionsPopup.filteredActionIndexes[0]
}

func (model *Model) MoveActionsPopupSelectionDown() {
	if !model.actionsPopup.visible {
		return
	}

	model.actionsPopup.selectedActionIndex = model.adjustVisibleSelection(model.actionsPopup.selectedActionIndex, model.actionsPopup.filteredActionIndexes, 1)
}

func (model *Model) MoveActionsPopupSelectionUp() {
	if !model.actionsPopup.visible {
		return
	}

	model.actionsPopup.selectedActionIndex = model.adjustVisibleSelection(model.actionsPopup.selectedActionIndex, model.actionsPopup.filteredActionIndexes, -1)
}

type actionsPopupState struct {
	visible               bool
	searchActive          bool
	searchQuery           string
	filteredActionIndexes []int
	selectedActionIndex   int
}

func actionIndexes(count int) []int {
	indexes := make([]int, 0, count)
	for index := range count {
		indexes = append(indexes, index)
	}
	return indexes
}
