package tui

import (
	"errors"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const reviewThreadTargetUnavailableMessage = "Inline comments require a diff line or valid diff-line selection"

var errReviewThreadTargetUnavailable = errors.New(reviewThreadTargetUnavailableMessage)

type reviewDiffRenderedRowAnchor struct {
	Path string
	Line reviewDiffLine
}

func reviewDiffThreadTargetForSelection(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (githubdomain.PullRequestReviewThreadTarget, error) {
	selectedRenderedRowIndexes := reviewDiffSelectedRenderedRowIndexes(document, state)
	return reviewDiffThreadTargetForRenderedRows(renderedRows, selectedRenderedRowIndexes)
}

func reviewDiffSelectedRenderedRowIndexes(document detailDocument, state detailViewState) []int {
	selectedRenderedRowIndexes := make([]int, 0)
	appendRenderedRowIndex := func(rowIndex int) {
		if rowIndex < 0 || rowIndex >= len(document.rows) {
			return
		}
		renderedRowIndex := document.rows[rowIndex].line
		if len(selectedRenderedRowIndexes) > 0 && selectedRenderedRowIndexes[len(selectedRenderedRowIndexes)-1] == renderedRowIndex {
			return
		}
		selectedRenderedRowIndexes = append(selectedRenderedRowIndexes, renderedRowIndex)
	}

	if startRow, endRow, ok := state.visualRowSelection(document); ok {
		for rowIndex := startRow; rowIndex <= endRow; rowIndex++ {
			appendRenderedRowIndex(rowIndex)
		}
		return selectedRenderedRowIndexes
	}
	if start, end, ok := state.visualSelection(document); ok {
		startRow := document.rowIndexForPosition(start)
		endRow := document.rowIndexForPosition(end)
		for rowIndex := startRow; rowIndex <= endRow; rowIndex++ {
			appendRenderedRowIndex(rowIndex)
		}
		return selectedRenderedRowIndexes
	}

	appendRenderedRowIndex(document.rowIndexForPosition(state.cursor))
	return selectedRenderedRowIndexes
}

func reviewDiffThreadTargetForRenderedRows(renderedRows []reviewDiffRenderedRow, selectedRenderedRowIndexes []int) (githubdomain.PullRequestReviewThreadTarget, error) {
	selectedLines := make([]reviewDiffLine, 0, len(selectedRenderedRowIndexes))
	path := ""
	for _, renderedRowIndex := range selectedRenderedRowIndexes {
		if renderedRowIndex < 0 || renderedRowIndex >= len(renderedRows) {
			continue
		}
		anchor := renderedRows[renderedRowIndex].Anchor
		if anchor == nil {
			continue
		}
		if path == "" {
			path = anchor.Path
		}
		selectedLines = append(selectedLines, anchor.Line)
	}

	target, ok := reviewDiffThreadTargetForLines(path, selectedLines)
	if !ok {
		return githubdomain.PullRequestReviewThreadTarget{}, errReviewThreadTargetUnavailable
	}
	return target, nil
}
