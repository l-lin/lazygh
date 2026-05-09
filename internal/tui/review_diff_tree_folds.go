package tui

import (
	"strconv"
	"strings"
)

func reviewDiffTreeRowIDForDirectory(path string) string {
	return "dir:" + strings.TrimSpace(path)
}

func reviewDiffTreeRowIDForFile(path string) string {
	return "file:" + strings.TrimSpace(path)
}

func reviewDiffTreeRowIDForChapter(chapterID string, index int) string {
	trimmedChapterID := strings.TrimSpace(chapterID)
	if trimmedChapterID == "" {
		trimmedChapterID = strconv.Itoa(index)
	}
	return "chapter:" + trimmedChapterID
}

func reviewDiffTreeVisibleRows(tree reviewDiffTree, collapsedRowIDs map[string]bool) reviewDiffTree {
	if len(tree.Rows) == 0 {
		return tree
	}

	visibleRows := make([]reviewDiffTreeRow, 0, len(tree.Rows))
	collapsedDepths := make([]int, 0)
	for _, rawRow := range tree.Rows {
		for len(collapsedDepths) > 0 && rawRow.Depth <= collapsedDepths[len(collapsedDepths)-1] {
			collapsedDepths = collapsedDepths[:len(collapsedDepths)-1]
		}
		if len(collapsedDepths) > 0 {
			continue
		}

		visibleRow := rawRow
		visibleRow.VisibleRowIndex = len(visibleRows)
		visibleRow.Collapsed = visibleRow.Foldable && reviewDiffTreeRowCollapsed(visibleRow, collapsedRowIDs)
		visibleRows = append(visibleRows, visibleRow)
		if visibleRow.Collapsed {
			collapsedDepths = append(collapsedDepths, visibleRow.Depth)
		}
	}
	return reviewDiffTree{Rows: visibleRows}
}

func reviewDiffTreeRowCollapsed(row reviewDiffTreeRow, collapsedRowIDs map[string]bool) bool {
	if !row.Foldable || collapsedRowIDs == nil {
		return false
	}
	return collapsedRowIDs[strings.TrimSpace(row.ID)]
}

func reviewDiffTreeVisibleRowIndexByID(tree reviewDiffTree, rowID string) (int, bool) {
	trimmedRowID := strings.TrimSpace(rowID)
	for _, row := range tree.Rows {
		if strings.TrimSpace(row.ID) != trimmedRowID {
			continue
		}
		return row.VisibleRowIndex, true
	}
	return 0, false
}

func reviewDiffTreeRowByID(tree reviewDiffTree, rowID string) (reviewDiffTreeRow, int, bool) {
	trimmedRowID := strings.TrimSpace(rowID)
	for index, row := range tree.Rows {
		if strings.TrimSpace(row.ID) != trimmedRowID {
			continue
		}
		return row, index, true
	}
	return reviewDiffTreeRow{}, 0, false
}

func reviewDiffTreeFirstDescendantFileIndex(tree reviewDiffTree, rowID string) (int, bool) {
	row, index, ok := reviewDiffTreeRowByID(tree, rowID)
	if !ok {
		return 0, false
	}
	if row.FileIndex >= 0 {
		return row.FileIndex, true
	}

	for nextIndex := index + 1; nextIndex < len(tree.Rows); nextIndex++ {
		candidate := tree.Rows[nextIndex]
		if candidate.Depth <= row.Depth {
			break
		}
		if candidate.FileIndex >= 0 {
			return candidate.FileIndex, true
		}
	}
	return 0, false
}

func reviewDiffTreePreferredVisibleRowIndex(rawTree reviewDiffTree, visibleTree reviewDiffTree, rowID string) int {
	if visibleRowIndex, ok := reviewDiffTreeVisibleRowIndexByID(visibleTree, rowID); ok {
		return visibleRowIndex
	}

	row, rawIndex, ok := reviewDiffTreeRowByID(rawTree, rowID)
	if ok {
		currentDepth := row.Depth
		for index := rawIndex - 1; index >= 0; index-- {
			candidate := rawTree.Rows[index]
			if candidate.Depth >= currentDepth {
				continue
			}
			currentDepth = candidate.Depth
			if visibleRowIndex, visible := reviewDiffTreeVisibleRowIndexByID(visibleTree, candidate.ID); visible {
				return visibleRowIndex
			}
		}
	}

	selectableRows := reviewDiffSelectableTreeRowIndexes(visibleTree)
	if len(selectableRows) == 0 {
		return 0
	}
	return selectableRows[0]
}
