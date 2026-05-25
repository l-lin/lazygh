package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) toggleInlineConversationVisibility(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgToggleInlineConversationVisibility{View: view})
}

func (program *Program) toggleInlineConversationVisibilityState(view *gocui.View) error {
	program.detailViewState.clearPendingPrefix()
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}
	if program.reviewModeActive() {
		if program.reviewSessionShowsDescription() {
			return program.toggleReviewDescriptionSectionVisibility(nil, view)
		}
		return program.toggleReviewInlineConversationVisibility(nil, view)
	}
	return program.toggleBrowserDetailSectionVisibility(nil, view)
}

func (program *Program) toggleReviewDescriptionSectionVisibility(gui *gocui.Gui, view *gocui.View) error {
	summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
	if !ok {
		return nil
	}
	return program.toggleOverviewSectionVisibility(gui, view, summary, detail)
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
	return nil
}

func (program *Program) toggleBrowserDetailSectionVisibility(gui *gocui.Gui, view *gocui.View) error {
	if !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
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

	if program.activeDetailTab == ChangesDetailTab {
		return program.toggleBrowserChangesVisibility(gui, summary, detailDocument)
	}

	if program.activeDetailTab == CommentsDetailTab {
		sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, detailDocument.width, program.detailViewState.cursor.line)
		if !ok {
			return nil
		}
		program.setBrowserDetailSectionCollapsed(sectionAtCursor.section.id, !sectionAtCursor.section.collapsed)
		program.detailViewState.cursor = detailPosition{line: sectionAtCursor.headerFocusLine, column: 0}
		program.detailViewState.preferredColumn = 0
		return nil
	}

	return program.toggleOverviewSectionVisibility(gui, view, summary, result.detail)
}

func (program *Program) toggleOverviewSectionVisibility(gui *gocui.Gui, view *gocui.View, summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) error {
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

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, detailDocument.width, program.detailViewState.cursor.line)
	if !ok {
		return nil
	}
	program.setBrowserDetailSectionCollapsed(sectionAtCursor.section.id, !sectionAtCursor.section.collapsed)
	program.detailViewState.cursor = detailPosition{line: sectionAtCursor.headerFocusLine, column: 0}
	program.detailViewState.preferredColumn = 0
	return nil
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
