package tui

type reviewCommentNavigationFile struct {
	fileTreeRow  int
	renderedRows []reviewDiffRenderedRow
}

type reviewCommentNavigationReadModel struct {
	active              bool
	selectedFileTreeRow int
	selection           detailCursorSelection
	files               []reviewCommentNavigationFile
}

func (model reviewCommentNavigationReadModel) currentPosition() (int, int) {
	currentFileTreeRow := model.selectedFileTreeRow
	if !model.active {
		return currentFileTreeRow, 0
	}

	rowIndex := model.selection.document.rowIndexForPosition(model.selection.state.cursor)
	if rowIndex < 0 || rowIndex >= len(model.selection.document.rows) {
		return currentFileTreeRow, 0
	}
	return currentFileTreeRow, model.selection.document.rows[rowIndex].line
}

func (model reviewCommentNavigationReadModel) commentLocations() []reviewCommentLocation {
	locations := make([]reviewCommentLocation, 0)
	for _, file := range model.files {
		for renderedLine, row := range file.renderedRows {
			if !reviewDiffRenderedRowIsThreadStatus(row) {
				continue
			}
			locations = append(locations, reviewCommentLocation{fileTreeRow: file.fileTreeRow, renderedLine: renderedLine})
		}
	}
	return locations
}

func (model reviewCommentNavigationReadModel) target(direction reviewNavigationDirection) (reviewCommentLocation, bool) {
	locations := model.commentLocations()
	if len(locations) == 0 {
		return reviewCommentLocation{}, false
	}

	currentFileTreeRow, currentRenderedLine := model.currentPosition()
	if direction > 0 {
		for _, location := range locations {
			if location.fileTreeRow > currentFileTreeRow || (location.fileTreeRow == currentFileTreeRow && location.renderedLine > currentRenderedLine) {
				return location, true
			}
		}
		return reviewCommentLocation{}, false
	}

	for index := len(locations) - 1; index >= 0; index-- {
		location := locations[index]
		if location.fileTreeRow < currentFileTreeRow || (location.fileTreeRow == currentFileTreeRow && location.renderedLine < currentRenderedLine) {
			return location, true
		}
	}
	return reviewCommentLocation{}, false
}
