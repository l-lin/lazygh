package tui

func itemAt(items []Item, selectedIndex int) (Item, bool) {
	if len(items) == 0 {
		return Item{}, false
	}

	index := clampIndex(selectedIndex, len(items))
	return items[index], true
}

func pullRequestRowAt(rows []PullRequestRow, selectedIndex int) (PullRequestRow, bool) {
	if len(rows) == 0 {
		return PullRequestRow{}, false
	}

	index := clampIndex(selectedIndex, len(rows))
	return rows[index], true
}

func clampIndex(index int, itemCount int) int {
	if itemCount == 0 {
		return 0
	}

	if index < 0 {
		return 0
	}

	maxIndex := itemCount - 1
	if index > maxIndex {
		return maxIndex
	}

	return index
}

func pageDelta(pageSize int) int {
	if pageSize <= 1 {
		return 1
	}

	return maxInt(1, pageSize/2)
}

func copyItems(items []Item) []Item {
	copiedItems := make([]Item, len(items))
	copy(copiedItems, items)
	return copiedItems
}

func copyPullRequestRows(rows []PullRequestRow) []PullRequestRow {
	copiedRows := make([]PullRequestRow, 0, len(rows))
	for _, row := range rows {
		copiedRow := PullRequestRow{Item: row.Item}
		if row.Summary != nil {
			summaryCopy := *row.Summary
			copiedRow.Summary = &summaryCopy
		}
		copiedRows = append(copiedRows, copiedRow)
	}
	return copiedRows
}

func pullRequestRowsFromItems(items []Item) []PullRequestRow {
	rows := make([]PullRequestRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, PullRequestRow{Item: item})
	}
	return rows
}

func pullRequestItems(rows []PullRequestRow) []Item {
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Item)
	}
	return items
}

func (model *Model) pullRequestRows(tab PullRequestTab) []PullRequestRow {
	index := int(tab)
	if index < 0 || index >= len(model.pullRequestTabs) {
		return nil
	}

	return model.pullRequestTabs[index].rows
}

func indexOfInt(items []int, expected int) int {
	for index, item := range items {
		if item == expected {
			return index
		}
	}

	return -1
}
