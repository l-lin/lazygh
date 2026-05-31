package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type browserChangesReadModel struct {
	summary            githubdomain.PullRequest
	files              []reviewDiffFile
	selection          detailCursorSelection
	renderedRows       []reviewDiffRenderedRow
	renderer           MarkdownRenderer
	wordWrapEnabled    bool
	connectedUserLogin string
	collapsedFileIDs   map[string]bool
	collapsedThreadIDs map[string]bool
}

func (model browserChangesReadModel) filePathAtCursor() (string, bool) {
	row, ok := reviewDiffRenderedRowAtCursor(model.renderedRows, model.selection.document, model.selection.state)
	if !ok || strings.TrimSpace(row.FilePath) == "" {
		return "", false
	}
	return strings.TrimSpace(row.FilePath), true
}

func (model browserChangesReadModel) threadAtCursor() (reviewDiffThread, bool) {
	return reviewDiffThreadAtCursor(model.renderedRows, model.selection.document, model.selection.state)
}

func (model browserChangesReadModel) fileVisibilityPlan(filePath string) (detailViewSyncPlan, bool) {
	trimmedFilePath := strings.TrimSpace(filePath)
	if trimmedFilePath == "" {
		return detailViewSyncPlan{}, false
	}
	return model.withFileCollapsedToggled(trimmedFilePath).syncPlanForFile(trimmedFilePath), true
}

func (model browserChangesReadModel) threadVisibilityPlan(thread reviewDiffThread) (detailViewSyncPlan, bool) {
	trimmedThreadID := strings.TrimSpace(thread.ID)
	if trimmedThreadID == "" {
		return detailViewSyncPlan{}, false
	}
	return model.withThreadCollapsedToggled(thread).syncPlanForThread(trimmedThreadID), true
}

func (model browserChangesReadModel) withFileCollapsedToggled(filePath string) browserChangesReadModel {
	trimmedFilePath := strings.TrimSpace(filePath)
	if trimmedFilePath == "" {
		return model
	}

	collapsedFileIDs := copyWorkflowStringBoolMap(model.collapsedFileIDs)
	collapsedFileIDs[trimmedFilePath] = !collapsedFileIDs[trimmedFilePath]
	model.collapsedFileIDs = collapsedFileIDs
	model.renderedRows = model.buildRenderedRows()
	return model
}

func (model browserChangesReadModel) withThreadCollapsedToggled(thread reviewDiffThread) browserChangesReadModel {
	trimmedThreadID := strings.TrimSpace(thread.ID)
	if trimmedThreadID == "" {
		return model
	}

	collapsedThreadIDs := copyWorkflowStringBoolMap(model.collapsedThreadIDs)
	collapsedThreadIDs[trimmedThreadID] = !reviewDiffThreadCollapsed(thread, model.collapsedThreadIDs)
	model.collapsedThreadIDs = collapsedThreadIDs
	model.renderedRows = model.buildRenderedRows()
	return model
}

func (model browserChangesReadModel) syncPlanForFile(filePath string) detailViewSyncPlan {
	plan := detailViewSyncPlan{document: model.detailDocument()}
	headerLineIndex := reviewDiffFileHeaderLineIndex(model.renderedRows, filePath)
	if headerLineIndex >= 0 {
		plan.focusLine = headerLineIndex
		plan.focusLineKnown = true
	}
	return plan
}

func (model browserChangesReadModel) syncPlanForThread(threadID string) detailViewSyncPlan {
	plan := detailViewSyncPlan{document: model.detailDocument()}
	headerLineIndex := reviewDiffThreadHeaderLineIndex(model.renderedRows, threadID)
	if headerLineIndex >= 0 {
		plan.focusLine = headerLineIndex
		plan.focusLineKnown = true
	}
	return plan
}

func (model browserChangesReadModel) detailDocument() detailDocument {
	return newReviewDiffDetailDocumentWithWordWrap(model.renderedRows, maxInt(model.selection.document.width, 1), model.wordWrapEnabled)
}

func (model browserChangesReadModel) buildRenderedRows() []reviewDiffRenderedRow {
	return buildPullRequestChangesRenderedRowsForViewerWithWordWrap(
		model.files,
		model.renderer,
		maxInt(model.selection.document.width, 1),
		model.wordWrapEnabled,
		model.collapsedThreadIDs,
		model.collapsedFileIDs,
		model.connectedUserLogin,
	)
}
