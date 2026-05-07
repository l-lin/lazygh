package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) currentReviewDiffRenderedRows(file reviewDiffFile, width int) []reviewDiffRenderedRow {
	cacheKey := program.reviewDiffRenderKey(file, width)
	if entry, ok := program.cachedReviewDiffRenderEntry(cacheKey); ok && len(entry.rows) > 0 {
		return entry.rows
	}

	rows := buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file, program.markdownRenderer, width, program.reviewSession.collapsedThreadIDs, program.currentConnectedUserLogin())
	entry, _ := program.cachedReviewDiffRenderEntry(cacheKey)
	entry.rows = rows
	program.storeReviewDiffRenderEntry(cacheKey, entry)
	return rows
}

func (program *Program) currentReviewDiffDocument(file reviewDiffFile, width int) detailDocument {
	cacheKey := program.reviewDiffRenderKey(file, width)
	if entry, ok := program.cachedReviewDiffRenderEntry(cacheKey); ok && len(entry.document.rows) > 0 {
		return entry.document
	}

	rows := program.currentReviewDiffRenderedRows(file, width)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.Text)
	}
	document := newDetailDocumentWithWrap(strings.Join(lines, "\n"), width, false)
	entry, _ := program.cachedReviewDiffRenderEntry(cacheKey)
	entry.document = document
	program.storeReviewDiffRenderEntry(cacheKey, entry)
	return document
}

func (program *Program) armInlineConversationTogglePrefix(gui *gocui.Gui, view *gocui.View) error {
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		program.detailViewState.clearPendingPrefix()
		return nil
	}

	return program.armOrHandleDetailKeySequence(detailViewportPlacementTarget(), func() error {
		return program.recenterDetailView(gui, view)
	})
}

func (program *Program) toggleInlineConversationVisibility(gui *gocui.Gui, view *gocui.View) error {
	program.detailViewState.clearPendingPrefix()
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}
	if program.reviewSession.active {
		return program.toggleReviewInlineConversationVisibility(gui, view)
	}
	return program.toggleBrowserDetailSectionVisibility(gui, view)
}

func (program *Program) toggleReviewInlineConversationVisibility(gui *gocui.Gui, view *gocui.View) error {
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return nil
	}

	actualView := view
	if actualView == nil && gui != nil {
		detailView, err := gui.View(viewDetailName)
		if err == nil {
			actualView = detailView
		}
	}
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	thread, ok := reviewDiffThreadAtCursor(renderedRows, detailDocument, program.detailViewState)
	if !ok {
		return nil
	}

	collapsed := reviewDiffThreadCollapsed(thread, program.reviewSession.collapsedThreadIDs)
	program.setReviewThreadCollapsed(thread.ID, !collapsed)

	updatedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	headerLineIndex := reviewDiffThreadHeaderLineIndex(updatedRows, thread.ID)
	if headerLineIndex >= 0 {
		program.detailViewState.cursor = detailPosition{line: headerLineIndex, column: 0}
		program.detailViewState.preferredColumn = 0
	}
	if gui == nil {
		return nil
	}
	return program.refreshViews(gui)
}

func (program *Program) toggleBrowserDetailSectionVisibility(gui *gocui.Gui, view *gocui.View) error {
	if !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return nil
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return nil
	}

	actualView := view
	if actualView == nil && gui != nil {
		detailView, err := gui.View(viewDetailName)
		if err == nil {
			actualView = detailView
		}
	}
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)

	cursorLine := program.detailViewState.cursor.line
	var sectionAtCursor browserDetailSectionCursor
	switch program.activeDetailTab {
	case CommentsDetailTab:
		sectionAtCursor, ok = program.browserConversationSectionAtCursor(summary, result.detail, detailDocument.width, cursorLine)
	default:
		sectionAtCursor, ok = program.browserOverviewSectionAtCursor(summary, result.detail, detailDocument.width, cursorLine)
	}
	if !ok {
		return nil
	}

	program.setBrowserDetailSectionCollapsed(sectionAtCursor.section.id, !sectionAtCursor.section.collapsed)
	program.detailViewState.cursor = detailPosition{line: sectionAtCursor.headerFocusLine, column: 0}
	program.detailViewState.preferredColumn = 0
	if gui == nil {
		return nil
	}
	return program.refreshViews(gui)
}

func (program *Program) setReviewThreadCollapsed(threadID string, collapsed bool) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return
	}
	if program.reviewSession.collapsedThreadIDs == nil {
		program.reviewSession.collapsedThreadIDs = map[string]bool{}
	}
	program.reviewSession.collapsedThreadIDs[trimmedThreadID] = collapsed
	program.invalidateReviewDiffRenderCache()
}

func reviewDiffThreadCollapsed(thread reviewDiffThread, collapsedThreadIDs map[string]bool) bool {
	trimmedThreadID := strings.TrimSpace(thread.ID)
	if collapsedThreadIDs != nil {
		if collapsed, ok := collapsedThreadIDs[trimmedThreadID]; ok {
			return collapsed
		}
	}
	return thread.IsResolved
}

func reviewDiffThreadHeaderLineIndex(renderedRows []reviewDiffRenderedRow, threadID string) int {
	trimmedThreadID := strings.TrimSpace(threadID)
	for index, row := range renderedRows {
		if row.Kind != reviewDiffRenderedRowKindInlineCommentHeader || row.Thread == nil || strings.TrimSpace(row.Thread.ID) != trimmedThreadID {
			continue
		}
		return index
	}
	return -1
}
