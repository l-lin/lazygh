package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) toggleInlineConversationVisibility(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgToggleInlineConversationVisibility{})
}

func (program *Program) toggleInlineConversationVisibilityState(detailDocument detailDocument) (detailViewSyncPlan, bool) {
	program.detailState.viewState.clearPendingPrefix()
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return detailViewSyncPlan{}, false
	}
	if program.reviewModeActive() {
		if program.reviewSessionShowsDescription() {
			return program.toggleReviewDescriptionSectionVisibility(detailDocument)
		}
		return program.toggleReviewInlineConversationVisibility(detailDocument)
	}
	return program.toggleBrowserDetailSectionVisibility(detailDocument)
}

func (program *Program) toggleReviewDescriptionSectionVisibility(detailDocument detailDocument) (detailViewSyncPlan, bool) {
	summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
	if !ok {
		return detailViewSyncPlan{}, false
	}
	return program.toggleOverviewSectionVisibility(summary, detail, detailDocument)
}

func (program *Program) toggleReviewInlineConversationVisibility(detailDocument detailDocument) (detailViewSyncPlan, bool) {
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return detailViewSyncPlan{}, false
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	thread, ok := reviewDiffThreadAtCursor(renderedRows, detailDocument, program.detailState.viewState)
	if !ok {
		return detailViewSyncPlan{}, false
	}

	collapsed := reviewDiffThreadCollapsed(thread, program.navigationState.reviewSession.collapsedThreadIDs)
	program.setReviewSessionThreadCollapsed(thread.ID, !collapsed)
	program.invalidateReviewDiffRenderCache()

	updatedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	plan := detailViewSyncPlan{document: program.currentReviewDiffDocument(selectedFile, detailDocument.width)}
	headerLineIndex := reviewDiffThreadHeaderLineIndex(updatedRows, thread.ID)
	if headerLineIndex >= 0 {
		plan.focusLine = headerLineIndex
		plan.focusLineKnown = true
	}
	return plan, true
}

func (program *Program) toggleBrowserDetailSectionVisibility(detailDocument detailDocument) (detailViewSyncPlan, bool) {
	if !program.shouldShowPullRequestDetailTabs() {
		return detailViewSyncPlan{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return detailViewSyncPlan{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return detailViewSyncPlan{}, false
	}

	if program.detailState.activeTab == ChangesDetailTab {
		return program.toggleBrowserChangesVisibility(summary, detailDocument)
	}

	if program.detailState.activeTab == CommentsDetailTab {
		sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, detailDocument.width, program.detailState.viewState.cursor.line)
		if !ok {
			return detailViewSyncPlan{}, false
		}
		program.setBrowserDetailSectionCollapsed(sectionAtCursor.section.id, !sectionAtCursor.section.collapsed)
		return detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width), focusLine: sectionAtCursor.headerFocusLine, focusLineKnown: true}, true
	}

	return program.toggleOverviewSectionVisibility(summary, result.detail, detailDocument)
}

func (program *Program) toggleOverviewSectionVisibility(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, detailDocument detailDocument) (detailViewSyncPlan, bool) {
	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, detailDocument.width, program.detailState.viewState.cursor.line)
	if !ok {
		return detailViewSyncPlan{}, false
	}
	program.setBrowserDetailSectionCollapsed(sectionAtCursor.section.id, !sectionAtCursor.section.collapsed)
	return detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width), focusLine: sectionAtCursor.headerFocusLine, focusLineKnown: true}, true
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
