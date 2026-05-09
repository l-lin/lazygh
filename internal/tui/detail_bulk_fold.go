package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) closeAllDetailFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.setAllDetailFolds(gui, view, true)
}

func (program *Program) openAllDetailFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.setAllDetailFolds(gui, view, false)
}

func (program *Program) setAllDetailFolds(gui *gocui.Gui, view *gocui.View, collapsed bool) error {
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		program.detailViewState.clearPendingPrefix()
		return nil
	}

	if program.reviewSession.active {
		if program.reviewSessionShowsDescription() {
			return program.setAllReviewDescriptionFolds(gui, view, collapsed)
		}
		return program.setAllReviewInlineConversationFolds(gui, view, collapsed)
	}
	return program.setAllBrowserDetailFolds(gui, view, collapsed)
}

func (program *Program) setAllReviewDescriptionFolds(gui *gocui.Gui, view *gocui.View, collapsed bool) error {
	summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
	if !ok {
		return nil
	}

	actualView := program.resolveView(gui, view, viewDetailName)
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	return program.setAllBrowserOverviewFolds(gui, summary, detail, detailDocument, viewportHeight, collapsed)
}

func (program *Program) setAllReviewInlineConversationFolds(gui *gocui.Gui, view *gocui.View, collapsed bool) error {
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return nil
	}

	actualView := program.resolveView(gui, view, viewDetailName)
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	threadAtCursor, cursorOnThread := reviewDiffThreadAtCursor(renderedRows, detailDocument, program.detailViewState)

	if !program.setAllReviewThreadsCollapsed(selectedFile.Threads, collapsed) {
		return nil
	}

	updatedDocument := program.currentReviewDiffDocument(selectedFile, detailDocument.width)
	if cursorOnThread {
		headerLineIndex := reviewDiffThreadHeaderLineIndex(program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width), threadAtCursor.ID)
		if headerLineIndex >= 0 {
			program.detailViewState.cursor = detailPosition{line: headerLineIndex, column: 0}
			program.detailViewState.preferredColumn = 0
		}
	}
	program.syncDetailViewState(updatedDocument, viewportHeight)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) setAllReviewThreadsCollapsed(threads []reviewDiffThread, collapsed bool) bool {
	if len(threads) == 0 {
		return false
	}
	if program.reviewSession.collapsedThreadIDs == nil {
		program.reviewSession.collapsedThreadIDs = map[string]bool{}
	}

	changed := false
	for _, thread := range threads {
		trimmedThreadID := strings.TrimSpace(thread.ID)
		if trimmedThreadID == "" {
			continue
		}
		if reviewDiffThreadCollapsed(thread, program.reviewSession.collapsedThreadIDs) != collapsed {
			changed = true
		}
		program.reviewSession.collapsedThreadIDs[trimmedThreadID] = collapsed
	}
	if changed {
		program.invalidateReviewDiffRenderCache()
	}
	return changed
}

func (program *Program) setAllBrowserDetailFolds(gui *gocui.Gui, view *gocui.View, collapsed bool) error {
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

	actualView := program.resolveView(gui, view, viewDetailName)
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)

	switch program.activeDetailTab {
	case ChangesDetailTab:
		return program.setAllBrowserChangesThreadFolds(gui, summary, detailDocument, viewportHeight, collapsed)
	case CommentsDetailTab:
		return program.setAllBrowserConversationFolds(gui, summary, result.detail, detailDocument, viewportHeight, collapsed)
	default:
		return program.setAllBrowserOverviewFolds(gui, summary, result.detail, detailDocument, viewportHeight, collapsed)
	}
}

func (program *Program) setAllBrowserOverviewFolds(gui *gocui.Gui, summary githubcli.PullRequest, detail githubcli.PullRequestDetail, detailDocument detailDocument, viewportHeight int, collapsed bool) error {
	sections := program.currentPullRequestOverviewSections(summary, detail, detailDocument.width)
	sectionAtCursor, cursorOnSection := program.browserOverviewSectionAtCursor(summary, detail, detailDocument.width, program.detailViewState.cursor.line)
	if !program.setBrowserDetailSectionsCollapsed(browserDetailSectionIDs(sections), collapsed) {
		return nil
	}

	updatedDocument := program.currentDetailDocument(program.resolveView(gui, nil, viewDetailName))
	if cursorOnSection {
		if headerFocusLine, ok := browserDetailSectionHeaderFocusLine(program.currentPullRequestOverviewSections(summary, detail, detailDocument.width), sectionAtCursor.section.id, false); ok {
			program.detailViewState.cursor = detailPosition{line: browserDescriptionOverviewStartLine(summary, detail) + headerFocusLine, column: 0}
			program.detailViewState.preferredColumn = 0
		}
	}
	program.syncDetailViewState(updatedDocument, viewportHeight)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) setAllBrowserConversationFolds(gui *gocui.Gui, summary githubcli.PullRequest, detail githubcli.PullRequestDetail, detailDocument detailDocument, viewportHeight int, collapsed bool) error {
	sections := program.currentPullRequestConversationSections(summary, detail, detailDocument.width)
	sectionAtCursor, cursorOnSection := program.browserConversationSectionAtCursor(summary, detail, detailDocument.width, program.detailViewState.cursor.line)
	if !program.setBrowserDetailSectionsCollapsed(browserDetailSectionIDs(sections), collapsed) {
		return nil
	}

	updatedDocument := program.currentDetailDocument(program.resolveView(gui, nil, viewDetailName))
	if cursorOnSection {
		if headerFocusLine, ok := browserDetailSectionHeaderFocusLine(program.currentPullRequestConversationSections(summary, detail, detailDocument.width), sectionAtCursor.section.id, false); ok {
			program.detailViewState.cursor = detailPosition{line: headerFocusLine, column: 0}
			program.detailViewState.preferredColumn = 0
		}
	}
	program.syncDetailViewState(updatedDocument, viewportHeight)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) setAllBrowserChangesThreadFolds(gui *gocui.Gui, summary githubcli.PullRequest, detailDocument detailDocument, viewportHeight int, collapsed bool) error {
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return nil
	}

	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	filePathAtCursor, cursorOnFile := reviewDiffFilePathAtCursor(renderedRows, detailDocument, program.detailViewState)
	sectionIDs := append(browserChangesFileSectionIDs(summary, result.data.Files), browserChangesThreadSectionIDs(summary, result.data.Files)...)
	if !program.setBrowserDetailSectionsCollapsed(sectionIDs, collapsed) {
		return nil
	}

	updatedDocument := program.currentDetailDocument(program.resolveView(gui, nil, viewDetailName))
	if cursorOnFile {
		headerLineIndex := reviewDiffFileHeaderLineIndex(program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width), filePathAtCursor)
		if headerLineIndex >= 0 {
			program.detailViewState.cursor = detailPosition{line: headerLineIndex, column: 0}
			program.detailViewState.preferredColumn = 0
		}
	}
	program.syncDetailViewState(updatedDocument, viewportHeight)
	return program.refreshViewsIfGUI(gui)
}

func browserDetailSectionIDs(sections []browserDetailSection) []string {
	sectionIDs := make([]string, 0, len(sections))
	for _, section := range sections {
		if trimmedSectionID := strings.TrimSpace(section.id); trimmedSectionID != "" {
			sectionIDs = append(sectionIDs, trimmedSectionID)
		}
	}
	return sectionIDs
}

func browserChangesFileSectionIDs(summary githubcli.PullRequest, files []reviewDiffFile) []string {
	sectionIDs := make([]string, 0, len(files))
	for _, file := range files {
		if trimmedFilePath := strings.TrimSpace(file.Path); trimmedFilePath != "" {
			sectionIDs = append(sectionIDs, browserChangesFileSectionID(summary, trimmedFilePath))
		}
	}
	return sectionIDs
}

func browserChangesThreadSectionIDs(summary githubcli.PullRequest, files []reviewDiffFile) []string {
	sectionIDs := make([]string, 0)
	for _, file := range files {
		for _, thread := range file.Threads {
			trimmedThreadID := strings.TrimSpace(thread.ID)
			if trimmedThreadID == "" {
				continue
			}
			sectionIDs = append(sectionIDs, browserChangesThreadSectionID(summary, thread))
		}
	}
	return sectionIDs
}

func (program *Program) setBrowserDetailSectionsCollapsed(sectionIDs []string, collapsed bool) bool {
	if len(sectionIDs) == 0 {
		return false
	}
	if program.browserCollapsedSectionStates == nil {
		program.browserCollapsedSectionStates = map[string]bool{}
	}

	changed := false
	for _, sectionID := range sectionIDs {
		trimmedSectionID := strings.TrimSpace(sectionID)
		if trimmedSectionID == "" {
			continue
		}
		if actualCollapsed, ok := program.browserCollapsedSectionStates[trimmedSectionID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		program.browserCollapsedSectionStates[trimmedSectionID] = collapsed
	}
	if changed {
		program.invalidatePullRequestDetailDocumentCache()
	}
	return changed
}

func browserDetailSectionHeaderFocusLine(sections []browserDetailSection, sectionID string, includeBlankLines bool) (int, bool) {
	currentLine := 0
	trimmedSectionID := strings.TrimSpace(sectionID)
	for sectionIndex, section := range sections {
		headerLineCount := renderedTextLineCount(strings.TrimSpace(section.header))
		headerFocusLine := currentLine + minInt(section.headerFocusOffset, maxInt(headerLineCount-1, 0))
		if strings.TrimSpace(section.id) == trimmedSectionID {
			return headerFocusLine, true
		}
		currentLine += headerLineCount
		if !section.collapsed {
			bodyLineCount := 0
			if strings.TrimSpace(section.body) != "" {
				bodyLineCount = renderedTextLineCount(strings.TrimSpace(section.body))
			}
			currentLine += bodyLineCount
		}
		if includeBlankLines && sectionIndex < len(sections)-1 {
			currentLine++
		}
	}
	return 0, false
}
