package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

type reviewDiffTreeNode struct {
	name         string
	fileIndex    int
	children     []*reviewDiffTreeNode
	childrenByID map[string]*reviewDiffTreeNode
}

func buildReviewDiffFileTree(files []reviewDiffFile) reviewDiffTree {
	root := &reviewDiffTreeNode{fileIndex: -1, childrenByID: map[string]*reviewDiffTreeNode{}}
	for fileIndex, file := range files {
		path := strings.TrimSpace(file.Path)
		if path == "" {
			continue
		}

		segments := strings.Split(path, "/")
		currentNode := root
		for segmentIndex, segment := range segments {
			isLeaf := segmentIndex == len(segments)-1
			childNode, ok := currentNode.childrenByID[segment]
			if !ok {
				childNode = &reviewDiffTreeNode{name: segment, fileIndex: -1, childrenByID: map[string]*reviewDiffTreeNode{}}
				currentNode.childrenByID[segment] = childNode
				currentNode.children = append(currentNode.children, childNode)
			}
			if isLeaf {
				childNode.fileIndex = fileIndex
			}
			currentNode = childNode
		}
	}

	rows := make([]reviewDiffTreeRow, 0)
	for _, childNode := range root.children {
		appendReviewDiffTreeRows(childNode, 0, &rows)
	}
	return reviewDiffTree{Rows: rows}
}

func appendReviewDiffTreeRows(node *reviewDiffTreeNode, depth int, rows *[]reviewDiffTreeRow) {
	if node == nil {
		return
	}
	if node.isFile() {
		appendReviewDiffTreeRow(rows, depth, node.name, node.fileIndex)
		return
	}

	pathSegments := []string{node.name}
	currentNode := node
	for currentNode.isDirectory() && len(currentNode.children) == 1 {
		onlyChild := currentNode.children[0]
		if onlyChild.isFile() {
			appendReviewDiffTreeRow(rows, depth, strings.Join(pathSegments, "/")+"/", -1)
			appendReviewDiffTreeRow(rows, depth+1, onlyChild.name, onlyChild.fileIndex)
			return
		}
		pathSegments = append(pathSegments, onlyChild.name)
		currentNode = onlyChild
	}

	appendReviewDiffTreeRow(rows, depth, strings.Join(pathSegments, "/")+"/", -1)
	for _, childNode := range currentNode.children {
		appendReviewDiffTreeRows(childNode, depth+1, rows)
	}
}

func appendReviewDiffTreeRow(rows *[]reviewDiffTreeRow, depth int, label string, fileIndex int) {
	if rows == nil {
		return
	}
	*rows = append(*rows, reviewDiffTreeRow{
		VisibleRowIndex: len(*rows),
		Depth:           depth,
		Label:           label,
		FileIndex:       fileIndex,
	})
}

func reviewDiffTreeItems(tree reviewDiffTree) []Item {
	items := make([]Item, 0, len(tree.Rows))
	for _, row := range tree.Rows {
		items = append(items, Item{Title: strings.Repeat("  ", row.Depth) + reviewDiffTreeRowText(row)})
	}
	return items
}

func reviewDiffTreeRowText(row reviewDiffTreeRow) string {
	return reviewDiffTreeRowIcon(row) + " " + row.Label
}

func reviewDiffTreeRowStyledText(row reviewDiffTreeRow, files []reviewDiffFile) string {
	icon := reviewDiffTreeRowIcon(row)
	foregroundHex := reviewDiffTreeRowForegroundHex(row, files)
	if strings.TrimSpace(foregroundHex) != "" && foregroundHex != theme.ActiveTextHex {
		icon = styleText(icon, foregroundColorEscape(foregroundHex))
	}
	return strings.Repeat("  ", row.Depth) + icon + " " + row.Label
}

func reviewDiffTreeRowForegroundHex(row reviewDiffTreeRow, files []reviewDiffFile) string {
	if row.FileIndex < 0 {
		return theme.DiffLineNumberHex
	}
	if row.FileIndex >= len(files) {
		return theme.ActiveTextHex
	}

	switch files[row.FileIndex].ChangeType {
	case reviewDiffChangeTypeAdded:
		return theme.DiffAdditionForegroundHex
	case reviewDiffChangeTypeRemoved:
		return theme.DiffDeletionForegroundHex
	default:
		return theme.ActiveTextHex
	}
}

func (program *Program) renderReviewDiffTreeView(view *gocui.View, tree reviewDiffTree, files []reviewDiffFile, selectedVisibleLine int) {
	if view == nil {
		return
	}

	view.Clear()
	for _, row := range tree.Rows {
		fmt.Fprintln(view, reviewDiffTreeRowStyledText(row, files))
	}
	program.selectListLine(view, selectedVisibleLine)
}

func reviewDiffSelectableRowIndexes(tree reviewDiffTree) []int {
	indexes := make([]int, 0, len(tree.Rows))
	for _, row := range tree.Rows {
		if row.FileIndex >= 0 {
			indexes = append(indexes, row.VisibleRowIndex)
		}
	}
	return indexes
}

func reviewDiffFileIndexAtRow(tree reviewDiffTree, rowIndex int) (int, bool) {
	if len(tree.Rows) == 0 {
		return 0, false
	}
	clampedRowIndex := clampIndex(rowIndex, len(tree.Rows))
	if tree.Rows[clampedRowIndex].FileIndex >= 0 {
		return tree.Rows[clampedRowIndex].FileIndex, true
	}
	return 0, false
}

func (node *reviewDiffTreeNode) isFile() bool {
	return node != nil && node.fileIndex >= 0
}

func (node *reviewDiffTreeNode) isDirectory() bool {
	return node != nil && node.fileIndex < 0
}
