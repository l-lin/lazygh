package tui

import (
	"github.com/jesseduffield/gocui"
)

type reviewNavigationDirection int

const (
	reviewNavigationBackward reviewNavigationDirection = -1
	reviewNavigationForward  reviewNavigationDirection = 1
)

func (program *Program) previousReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.moveReviewSessionFile(gui, int(reviewNavigationBackward))
}

func (program *Program) nextReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.moveReviewSessionFile(gui, int(reviewNavigationForward))
}

func (program *Program) previousReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.moveReviewSessionComment(gui, reviewNavigationBackward)
}

func (program *Program) nextReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.moveReviewSessionComment(gui, reviewNavigationForward)
}

func (program *Program) moveReviewSessionFile(gui *gocui.Gui, change int) error {
	if !program.reviewSession.active {
		return nil
	}

	selectableRows, ok := program.reviewSessionFileRows()
	if !ok || len(selectableRows) == 0 {
		return nil
	}

	originalRow := program.reviewSession.selectedFileTreeRow
	program.reviewSession.selectedFileTreeRow = adjustVisibleSelection(program.reviewSession.selectedFileTreeRow, selectableRows, change)
	if program.reviewSession.selectedFileTreeRow == originalRow {
		return nil
	}

	return program.refreshViewsIfGUI(gui)
}

type reviewCommentLocation struct {
	fileTreeRow  int
	renderedLine int
}

func (program *Program) moveReviewSessionComment(gui *gocui.Gui, direction reviewNavigationDirection) error {
	if !program.reviewSession.active {
		return nil
	}

	detailView := program.resolveView(gui, nil, viewDetailName)
	currentFileTreeRow, currentRenderedLine := program.currentReviewCommentPosition(detailView)
	target, ok := program.reviewSessionCommentTarget(detailView, currentFileTreeRow, currentRenderedLine, direction)
	if !ok {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.reviewSession.selectedFileTreeRow = target.fileTreeRow
	if actualErr := program.mutateDetailViewStateWithoutRefresh(gui, detailView, func(document detailDocument, viewportHeight int) {
		program.detailViewState.cursor = document.clampPosition(detailPosition{line: target.renderedLine, column: 0})
		program.detailViewState.preferredColumn = 0
		program.detailViewState.sync(document, viewportHeight)
	}); actualErr != nil {
		return actualErr
	}

	return program.refreshViewsIfGUI(gui)
}

func (program *Program) currentReviewCommentPosition(detailView *gocui.View) (int, int) {
	currentFileTreeRow := program.reviewSession.selectedFileTreeRow
	if !program.reviewSession.active {
		return currentFileTreeRow, 0
	}

	document := program.currentDetailDocument(detailView)
	program.syncDetailViewState(document, viewPageSize(detailView))
	currentRowIndex := document.rowIndexForPosition(program.detailViewState.cursor)
	if currentRowIndex < 0 || currentRowIndex >= len(document.rows) {
		return currentFileTreeRow, 0
	}
	return currentFileTreeRow, document.rows[currentRowIndex].line
}

func (program *Program) reviewSessionCommentTarget(detailView *gocui.View, currentFileTreeRow int, currentRenderedLine int, direction reviewNavigationDirection) (reviewCommentLocation, bool) {
	locations := program.reviewSessionCommentLocations(detailView)
	if len(locations) == 0 {
		return reviewCommentLocation{}, false
	}

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

func (program *Program) reviewSessionCommentLocations(detailView *gocui.View) []reviewCommentLocation {
	tree, files, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil
	}

	fileTreeRows := map[int]int{}
	for _, row := range tree.Rows {
		if row.FileIndex < 0 {
			continue
		}
		fileTreeRows[row.FileIndex] = row.VisibleRowIndex
	}

	width := program.detailWrapWidth
	if detailView != nil && detailView.InnerWidth() > 0 {
		width = detailView.InnerWidth()
		if width < 1 {
			width = 1
		}
	}

	locations := make([]reviewCommentLocation, 0)
	for fileIndex, file := range files {
		fileTreeRow, ok := fileTreeRows[fileIndex]
		if !ok {
			continue
		}

		for renderedLine, row := range program.currentReviewDiffRenderedRows(file, width) {
			if !reviewDiffRenderedRowIsThreadStatus(row) {
				continue
			}
			locations = append(locations, reviewCommentLocation{fileTreeRow: fileTreeRow, renderedLine: renderedLine})
		}
	}
	return locations
}

func reviewDiffRenderedRowIsThreadStatus(row reviewDiffRenderedRow) bool {
	return row.Thread != nil && row.Kind == reviewDiffRenderedRowKindInlineCommentHeader
}
