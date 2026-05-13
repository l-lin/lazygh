package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
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
		appendReviewDiffTreeRows(childNode, []string{childNode.name}, 0, &rows)
	}
	return reviewDiffTree{Rows: rows}
}

func appendReviewDiffTreeRows(node *reviewDiffTreeNode, pathSegments []string, depth int, rows *[]reviewDiffTreeRow) {
	if node == nil {
		return
	}
	if node.isFile() {
		filePath := strings.Join(pathSegments, "/")
		appendReviewDiffTreeRow(rows, reviewDiffTreeRowIDForFile(filePath), depth, node.name, node.fileIndex, 0, reviewDiffTreeRowKindFile, false)
		return
	}

	currentPathSegments := append([]string(nil), pathSegments...)
	labelSegments := []string{node.name}
	currentNode := node
	for currentNode.isDirectory() && len(currentNode.children) == 1 {
		onlyChild := currentNode.children[0]
		if onlyChild.isFile() {
			directoryPath := strings.Join(currentPathSegments, "/") + "/"
			directoryLabel := strings.Join(labelSegments, "/") + "/"
			appendReviewDiffTreeRow(rows, reviewDiffTreeRowIDForDirectory(directoryPath), depth, directoryLabel, -1, 0, reviewDiffTreeRowKindDirectory, true)
			appendReviewDiffTreeRow(rows, reviewDiffTreeRowIDForFile(directoryPath+onlyChild.name), depth+1, onlyChild.name, onlyChild.fileIndex, 0, reviewDiffTreeRowKindFile, false)
			return
		}
		currentPathSegments = append(currentPathSegments, onlyChild.name)
		labelSegments = append(labelSegments, onlyChild.name)
		currentNode = onlyChild
	}

	directoryPath := strings.Join(currentPathSegments, "/") + "/"
	directoryLabel := strings.Join(labelSegments, "/") + "/"
	appendReviewDiffTreeRow(rows, reviewDiffTreeRowIDForDirectory(directoryPath), depth, directoryLabel, -1, 0, reviewDiffTreeRowKindDirectory, true)
	for _, childNode := range currentNode.children {
		appendReviewDiffTreeRows(childNode, append(append([]string(nil), currentPathSegments...), childNode.name), depth+1, rows)
	}
}

func appendReviewDiffTreeRow(rows *[]reviewDiffTreeRow, id string, depth int, label string, fileIndex int, chapterIndex int, kind reviewDiffTreeRowKind, foldable bool) {
	if rows == nil {
		return
	}
	*rows = append(*rows, reviewDiffTreeRow{
		ID:              strings.TrimSpace(id),
		VisibleRowIndex: len(*rows),
		Depth:           depth,
		Label:           label,
		FileIndex:       fileIndex,
		ChapterIndex:    chapterIndex,
		Kind:            kind,
		Foldable:        foldable,
	})
}

func reviewDiffTreeItems(tree reviewDiffTree, files []reviewDiffFile) []Item {
	items := make([]Item, 0, len(tree.Rows))
	for _, row := range tree.Rows {
		items = append(items, Item{Title: strings.Repeat("  ", row.Depth) + reviewDiffTreeRowText(row, files)})
	}
	return items
}

func reviewDiffTreeRowText(row reviewDiffTreeRow, files []reviewDiffFile) string {
	return reviewDiffTreeRowPrefix(row) + reviewDiffTreeRowDisplayLabel(row, files)
}

func reviewDiffTreeRowPrefix(row reviewDiffTreeRow) string {
	parts := make([]string, 0, 2)
	if chevron := reviewDiffTreeRowChevron(row); chevron != "" {
		parts = append(parts, chevron)
	}
	parts = append(parts, reviewDiffTreeRowIcon(row))
	return strings.Join(parts, " ") + " "
}

func reviewDiffTreeRowDisplayLabel(row reviewDiffTreeRow, files []reviewDiffFile) string {
	return row.Label + reviewDiffTreeRowTeamOwnersSuffix(row, files) + reviewDiffTreeRowCommentSuffix(row, files)
}

func reviewDiffTreeRowSearchText(row reviewDiffTreeRow, files []reviewDiffFile) string {
	return row.Label + reviewDiffTreeRowTeamOwnersSuffix(row, files)
}

func reviewDiffTreeRowTeamOwnersSuffix(row reviewDiffTreeRow, files []reviewDiffFile) string {
	teamOwners := reviewDiffTreeRowTeamOwners(row, files)
	if len(teamOwners) == 0 {
		return ""
	}
	return fmt.Sprintf("  %s %s", reviewDiffTeamOwnershipIcon, strings.Join(teamOwners, ", "))
}

func reviewDiffTreeRowTeamOwners(row reviewDiffTreeRow, files []reviewDiffFile) []string {
	if row.FileIndex < 0 || row.FileIndex >= len(files) {
		return nil
	}
	return normalizeReviewDiffTeamOwners(files[row.FileIndex].TeamOwners)
}

func reviewDiffTreeRowCommentSuffix(row reviewDiffTreeRow, files []reviewDiffFile) string {
	commentCount := reviewDiffTreeRowCommentCount(row, files)
	if commentCount <= 0 {
		return ""
	}
	return fmt.Sprintf(" %s %d", reviewDiffTreeCommentCountIcon, commentCount)
}

func reviewDiffTreeRowCommentCount(row reviewDiffTreeRow, files []reviewDiffFile) int {
	if row.FileIndex < 0 || row.FileIndex >= len(files) {
		return 0
	}

	commentCount := 0
	for _, thread := range files[row.FileIndex].Threads {
		commentCount += len(thread.Comments)
	}
	return commentCount
}

func reviewDiffTreeRowStyledPrefix(row reviewDiffTreeRow, files []reviewDiffFile) string {
	return reviewDiffTreeRowIndent(row) + reviewDiffTreeRowStyledLeadingIcons(row, files)
}

func reviewDiffTreeRowStyledLeadingIcons(row reviewDiffTreeRow, files []reviewDiffFile) string {
	parts := make([]string, 0, 2)
	if chevron := reviewDiffTreeRowStyledChevron(row, files); chevron != "" {
		parts = append(parts, chevron)
	}
	parts = append(parts, reviewDiffTreeRowStyledIcon(row, files))
	return strings.Join(parts, " ") + " "
}

func reviewDiffTreeRowIndent(row reviewDiffTreeRow) string {
	return strings.Repeat("  ", row.Depth)
}

func reviewDiffTreeRowStyledChevron(row reviewDiffTreeRow, files []reviewDiffFile) string {
	chevron := reviewDiffTreeRowChevron(row)
	if chevron == "" {
		return ""
	}
	foregroundHex := reviewDiffTreeRowForegroundHex(row, files)
	if strings.TrimSpace(foregroundHex) != "" && foregroundHex != theme.ActiveTextHex {
		return styleText(chevron, foregroundColorEscape(foregroundHex))
	}
	return chevron
}

func reviewDiffTreeRowStyledIcon(row reviewDiffTreeRow, files []reviewDiffFile) string {
	icon := reviewDiffTreeRowIcon(row)
	foregroundHex := reviewDiffTreeRowForegroundHex(row, files)
	if strings.TrimSpace(foregroundHex) != "" && foregroundHex != theme.ActiveTextHex {
		return styleText(icon, foregroundColorEscape(foregroundHex))
	}
	return icon
}

func reviewDiffTreeRowForegroundHex(row reviewDiffTreeRow, files []reviewDiffFile) string {
	if row.Kind == reviewDiffTreeRowKindChapter {
		return theme.MarkdownHeadingHex
	}
	if row.FileIndex < 0 {
		return theme.DiffLineNumberHex
	}
	if row.FileIndex >= len(files) {
		return theme.ActiveTextHex
	}

	switch files[row.FileIndex].ChangeType {
	case reviewDiffChangeTypeAdded:
		return theme.DiffAdditionHex
	case reviewDiffChangeTypeRemoved:
		return theme.DiffDeletionHex
	default:
		return theme.ActiveTextHex
	}
}

func renderReviewDiffTreeRow(row reviewDiffTreeRow, files []reviewDiffFile, query string, selected bool) string {
	label := row.Label
	teamOwnersSuffix := reviewDiffTreeRowTeamOwnersSuffix(row, files)
	commentSuffix := reviewDiffTreeRowCommentSuffix(row, files)
	if !selected {
		highlightedLabel, _ := highlightSearchMatches(label, query)
		return reviewDiffTreeRowStyledPrefix(row, files) + highlightedLabel + renderReviewDiffTreeRowTeamOwnersSuffix(teamOwnersSuffix, query, false, "") + commentSuffix
	}

	selectedPrefix := ansiBold + foregroundColorEscape(theme.ActiveTextHex) + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	highlightedLabel, _ := highlightSearchMatchesWithPrefixes(label, query, selectedPrefix, ansiBold+backgroundColorEscape(theme.SearchHighlightHex))
	prefix := renderSelectedReviewDiffTreeRowPrefix(row, files, selectedPrefix)
	return prefix + highlightedLabel + renderReviewDiffTreeRowTeamOwnersSuffix(teamOwnersSuffix, query, true, selectedPrefix) + applyPrefix(commentSuffix, selectedPrefix)
}

func renderReviewDiffTreeRowTeamOwnersSuffix(teamOwnersSuffix string, query string, selected bool, selectedPrefix string) string {
	trimmedSuffix := strings.TrimSpace(teamOwnersSuffix)
	if trimmedSuffix == "" {
		return ""
	}

	teamOwnershipPrefix := foregroundColorEscape(theme.TeamOwnershipHex)
	teamOwnershipMatchPrefix := foregroundColorEscape(theme.TeamOwnershipHex) + backgroundColorEscape(theme.SearchHighlightHex)
	if selected {
		teamOwnershipPrefix = selectedPrefix + foregroundColorEscape(theme.TeamOwnershipHex)
		teamOwnershipMatchPrefix = ansiBold + foregroundColorEscape(theme.TeamOwnershipHex) + backgroundColorEscape(theme.SearchHighlightHex)
	}

	renderedSuffix, _ := highlightSearchMatchesWithPrefixes(teamOwnersSuffix, query, teamOwnershipPrefix, teamOwnershipMatchPrefix)
	return renderedSuffix
}

func renderSelectedReviewDiffTreeRowPrefix(row reviewDiffTreeRow, files []reviewDiffFile, selectedPrefix string) string {
	indent := applyPrefix(reviewDiffTreeRowIndent(row), selectedPrefix)
	foregroundHex := reviewDiffTreeRowForegroundHex(row, files)
	segments := make([]string, 0, 2)
	for _, segment := range []string{reviewDiffTreeRowChevron(row), reviewDiffTreeRowIcon(row)} {
		if segment == "" {
			continue
		}
		if strings.TrimSpace(foregroundHex) != "" && foregroundHex != theme.ActiveTextHex {
			segments = append(segments, selectedPrefix+foregroundColorEscape(foregroundHex)+segment+ansiReset)
			continue
		}
		segments = append(segments, applyPrefix(segment, selectedPrefix))
	}
	return indent + strings.Join(segments, applyPrefix(" ", selectedPrefix)) + applyPrefix(" ", selectedPrefix)
}

func (program *Program) renderReviewDiffTreeView(view *gocui.View, tree reviewDiffTree, files []reviewDiffFile, query string, selectedVisibleLine int) {
	if view == nil {
		return
	}

	view.Clear()
	showSelectedLine := program.reviewModeActive() || (program.usesManualSelectedLineRendering(query) && program.shouldHighlightSelection(FocusPullRequestsView, true))
	for _, row := range tree.Rows {
		fmt.Fprintln(view, renderReviewDiffTreeRow(row, files, query, showSelectedLine && row.VisibleRowIndex == selectedVisibleLine))
	}
	program.selectListLine(view, selectedVisibleLine, len(tree.Rows))
}

func reviewDiffTreeRowChevron(row reviewDiffTreeRow) string {
	if !row.Foldable {
		return ""
	}
	if row.Collapsed {
		return browserDetailCollapsedChevron
	}
	return browserDetailExpandedChevron
}

func reviewDiffSelectableTreeRowIndexes(tree reviewDiffTree) []int {
	indexes := make([]int, 0, len(tree.Rows))
	for _, row := range tree.Rows {
		indexes = append(indexes, row.VisibleRowIndex)
	}
	return indexes
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

func (node *reviewDiffTreeNode) isFile() bool {
	return node != nil && node.fileIndex >= 0
}

func (node *reviewDiffTreeNode) isDirectory() bool {
	return node != nil && node.fileIndex < 0
}
