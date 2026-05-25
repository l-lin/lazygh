package tui

func (program *Program) reviewSessionDiffData() (reviewDiffData, bool) {
	state := program.navigationState.reviewSession
	if !state.active {
		return reviewDiffData{}, false
	}

	result, ok := program.pullRequestDiffForSummary(state.summary)
	if !ok || result.err != nil {
		return reviewDiffData{}, false
	}
	return result.data, true
}

func (program *Program) reviewSessionMainContentKind() MainContentKind {
	state := program.navigationState.reviewSession
	if !state.active {
		return MainContentKindReviewDiff
	}
	if program.model.currentSideFocus() == FocusUserView {
		return MainContentKindReviewDescription
	}
	if state.mode != reviewSessionModeStory {
		return MainContentKindReviewDiff
	}
	data, ok := program.reviewSessionDiffData()
	if !ok {
		return MainContentKindReviewDiff
	}
	if _, ok := reviewSessionSelectedStoryChapter(state, data); ok {
		return MainContentKindStoryChapter
	}
	return MainContentKindReviewDiff
}

func reviewSessionRawTree(state reviewSessionState, diffData reviewDiffData) (reviewDiffTree, []reviewDiffFile) {
	if state.mode == reviewSessionModeStory && len(state.story.Tree.Rows) > 0 {
		return state.story.Tree, diffData.Files
	}
	return diffData.FileTree, diffData.Files
}

func reviewSessionSelectedTreeRow(state reviewSessionState, diffData reviewDiffData) (reviewDiffTreeRow, reviewDiffTree, []reviewDiffFile, bool) {
	tree, files := reviewSessionRawTree(state, diffData)
	row, ok := reviewDiffVisibleRowAt(tree, state.collapsedTreeRowIDs, state.selectedFileTreeRow)
	if !ok {
		return reviewDiffTreeRow{}, reviewDiffTree{}, nil, false
	}
	return row, tree, files, true
}

func reviewSessionSelectedStoryChapter(state reviewSessionState, diffData reviewDiffData) (reviewStoryChapter, bool) {
	if state.mode != reviewSessionModeStory {
		return reviewStoryChapter{}, false
	}
	row, _, _, ok := reviewSessionSelectedTreeRow(state, diffData)
	if !ok || row.Kind != reviewDiffTreeRowKindChapter {
		return reviewStoryChapter{}, false
	}
	if row.ChapterIndex < 0 || row.ChapterIndex >= len(state.story.Chapters) {
		return reviewStoryChapter{}, false
	}
	return state.story.Chapters[row.ChapterIndex], true
}

func reviewSessionSelectedDiffFile(state reviewSessionState, diffData reviewDiffData) (reviewDiffFile, bool) {
	row, rawTree, files, ok := reviewSessionSelectedTreeRow(state, diffData)
	if !ok {
		return reviewDiffFile{}, false
	}

	fileIndex := row.FileIndex
	if fileIndex < 0 {
		fileIndex, ok = reviewDiffTreeFirstDescendantFileIndex(rawTree, row.ID)
		if !ok {
			return reviewDiffFile{}, false
		}
	}
	if fileIndex < 0 || fileIndex >= len(files) {
		return reviewDiffFile{}, false
	}
	return files[fileIndex], true
}

func reviewDiffVisibleRowAt(tree reviewDiffTree, collapsedRowIDs map[string]bool, selectedVisibleRow int) (reviewDiffTreeRow, bool) {
	if len(tree.Rows) == 0 {
		return reviewDiffTreeRow{}, false
	}

	targetVisibleRow := maxInt(selectedVisibleRow, 0)
	var (
		lastVisibleRow reviewDiffTreeRow
		haveVisibleRow bool
	)
	collapsedDepthBacking := [8]int{}
	collapsedDepths := collapsedDepthBacking[:0]
	visibleRowIndex := 0

	for _, rawRow := range tree.Rows {
		for len(collapsedDepths) > 0 && rawRow.Depth <= collapsedDepths[len(collapsedDepths)-1] {
			collapsedDepths = collapsedDepths[:len(collapsedDepths)-1]
		}
		if len(collapsedDepths) > 0 {
			continue
		}

		visibleRow := rawRow
		visibleRow.VisibleRowIndex = visibleRowIndex
		visibleRow.Collapsed = visibleRow.Foldable && reviewDiffTreeRowCollapsed(visibleRow, collapsedRowIDs)
		lastVisibleRow = visibleRow
		haveVisibleRow = true
		if visibleRowIndex == targetVisibleRow {
			return visibleRow, true
		}
		visibleRowIndex++
		if visibleRow.Collapsed {
			collapsedDepths = append(collapsedDepths, visibleRow.Depth)
		}
	}

	if !haveVisibleRow {
		return reviewDiffTreeRow{}, false
	}
	return lastVisibleRow, true
}
