package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) currentReviewDiffRenderedRows(file reviewDiffFile, width int) []reviewDiffRenderedRow {
	return buildReviewDiffRenderedRowsWithCollapsedThreads(file, program.markdownRenderer, width, program.reviewSession.collapsedThreadIDs)
}

func (program *Program) currentReviewDiffDocument(file reviewDiffFile, width int) detailDocument {
	return newDetailDocument(renderReviewDiffFileWithCollapsedThreads(file, program.markdownRenderer, width, program.reviewSession.collapsedThreadIDs), width)
}

func (program *Program) toggleInlineConversationVisibility(gui *gocui.Gui, view *gocui.View) error {
	if !program.reviewSession.active || program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}

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

func (program *Program) setReviewThreadCollapsed(threadID string, collapsed bool) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return
	}
	if program.reviewSession.collapsedThreadIDs == nil {
		program.reviewSession.collapsedThreadIDs = map[string]bool{}
	}
	program.reviewSession.collapsedThreadIDs[trimmedThreadID] = collapsed
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
		if row.Thread == nil || strings.TrimSpace(row.Thread.ID) != trimmedThreadID {
			continue
		}
		return index
	}
	return -1
}
