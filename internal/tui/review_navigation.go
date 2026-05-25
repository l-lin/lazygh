package tui

import "github.com/jesseduffield/gocui"

type reviewNavigationDirection int

const (
	reviewNavigationBackward reviewNavigationDirection = -1
	reviewNavigationForward  reviewNavigationDirection = 1
)

func (program *Program) previousReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewFile{Delta: int(reviewNavigationBackward)})
}

func (program *Program) nextReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewFile{Delta: int(reviewNavigationForward)})
}

func (program *Program) previousReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationBackward})
}

func (program *Program) nextReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationForward})
}

type reviewCommentLocation struct {
	fileTreeRow  int
	renderedLine int
}

func (program *Program) currentReviewCommentPosition() (int, int) {
	currentFileTreeRow := program.navigationState.reviewSession.selectedFileTreeRow
	if !program.reviewModeActive() {
		return currentFileTreeRow, 0
	}

	document := program.currentDetailDocument(nil)
	currentRowIndex := document.rowIndexForPosition(program.detailState.viewState.cursor)
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

	width := program.detailState.wrapWidth
	if detailView != nil && detailView.InnerWidth() > 0 {
		width = max(detailView.InnerWidth(), 1)
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
