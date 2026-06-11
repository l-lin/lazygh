package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) closeAllDetailFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSetAllDetailFolds{Collapsed: true})
}

func (program *Program) openAllDetailFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSetAllDetailFolds{Collapsed: false})
}

func (program *Program) setAllDetailFolds(detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	if program.model.Focus() != FocusDetailView || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return detailViewSyncPlan{}, false
	}

	if program.reviewModeActive() {
		if program.reviewSessionShowsDescription() {
			return program.setAllReviewDescriptionFolds(detailDocument, collapsed)
		}
		return program.setAllReviewInlineConversationFolds(detailDocument, collapsed)
	}
	return program.setAllBrowserDetailFolds(detailDocument, collapsed)
}

func (program *Program) setAllReviewDescriptionFolds(detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
	if !ok {
		return detailViewSyncPlan{}, false
	}
	return program.setAllBrowserOverviewFolds(summary, detail, detailDocument, collapsed)
}

func (program *Program) setAllReviewInlineConversationFolds(detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return detailViewSyncPlan{}, false
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	threadAtCursor, cursorOnThread := reviewDiffThreadAtCursor(renderedRows, detailDocument, program.detailState.viewState)

	if !program.setAllReviewSessionThreadsCollapsed(selectedFile.Threads, collapsed) {
		return detailViewSyncPlan{}, false
	}
	program.invalidateReviewDiffRenderCache()

	plan := detailViewSyncPlan{document: program.currentReviewDiffDocument(selectedFile, detailDocument.width)}
	if cursorOnThread {
		headerLineIndex := reviewDiffThreadHeaderLineIndex(program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width), threadAtCursor.ID)
		if headerLineIndex >= 0 {
			plan.focusLine = headerLineIndex
			plan.focusLineKnown = true
		}
	}
	return plan, true
}

func (program *Program) setAllBrowserDetailFolds(detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
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

	switch program.detailState.activeTab {
	case ChangesDetailTab:
		return program.setAllBrowserChangesThreadFolds(summary, detailDocument, collapsed)
	case CommitChangesDetailTab:
		return program.setAllCommitDiffFolds(summary, detailDocument, collapsed)
	case CommentsDetailTab:
		return program.setAllBrowserConversationFolds(summary, result.detail, detailDocument, collapsed)
	default:
		return program.setAllBrowserOverviewFolds(summary, result.detail, detailDocument, collapsed)
	}
}

func (program *Program) setAllBrowserOverviewFolds(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	sections := program.currentPullRequestOverviewSections(summary, detail, detailDocument.width)
	sectionAtCursor, cursorOnSection := program.browserOverviewSectionAtCursor(summary, detail, detailDocument.width, program.detailState.viewState.cursor.line)
	if !program.setBrowserDetailSectionsCollapsed(browserDetailSectionIDs(sections), collapsed) {
		return detailViewSyncPlan{}, false
	}

	plan := detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width)}
	if cursorOnSection {
		if headerFocusLine, ok := browserDetailSectionHeaderFocusLine(program.currentPullRequestOverviewSections(summary, detail, detailDocument.width), sectionAtCursor.section.id, false); ok {
			plan.focusLine = browserDescriptionOverviewStartLine(summary, detail) + headerFocusLine
			plan.focusLineKnown = true
		}
	}
	return plan, true
}

func (program *Program) setAllBrowserConversationFolds(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	sections := program.currentPullRequestConversationSections(summary, detail, detailDocument.width)
	sectionAtCursor, cursorOnSection := program.browserConversationSectionAtCursor(summary, detail, detailDocument.width, program.detailState.viewState.cursor.line)
	if !program.setBrowserDetailSectionsCollapsed(browserDetailSectionIDs(sections), collapsed) {
		return detailViewSyncPlan{}, false
	}

	plan := detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width)}
	if cursorOnSection {
		if headerFocusLine, ok := browserDetailSectionHeaderFocusLine(program.currentPullRequestConversationSections(summary, detail, detailDocument.width), sectionAtCursor.section.id, false); ok {
			plan.focusLine = headerFocusLine
			plan.focusLineKnown = true
		}
	}
	return plan, true
}

func (program *Program) setAllBrowserChangesThreadFolds(summary githubdomain.PullRequest, detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return detailViewSyncPlan{}, false
	}

	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	filePathAtCursor, cursorOnFile := reviewDiffFilePathAtCursor(renderedRows, detailDocument, program.detailState.viewState)
	sectionIDs := append(browserChangesFileSectionIDs(summary, result.data.Files), browserChangesThreadSectionIDs(summary, result.data.Files)...)
	if !program.setBrowserDetailSectionsCollapsed(sectionIDs, collapsed) {
		return detailViewSyncPlan{}, false
	}

	plan := detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width)}
	if cursorOnFile {
		headerLineIndex := reviewDiffFileHeaderLineIndex(program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width), filePathAtCursor)
		if headerLineIndex >= 0 {
			plan.focusLine = headerLineIndex
			plan.focusLineKnown = true
		}
	}
	return plan, true
}

func (program *Program) setAllCommitDiffFolds(summary githubdomain.PullRequest, detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	if !program.detailState.commitDiffTab.visibleForPullRequestKey(pullRequestKey) {
		return detailViewSyncPlan{}, false
	}
	commitOID := program.detailState.commitDiffTab.commitOID
	result, ok := program.commitDiffResultForTarget(pullRequestKey, commitOID)
	if !ok || result.err != nil {
		return detailViewSyncPlan{}, false
	}

	renderedRows := program.currentCommitDiffRenderedRows(pullRequestKey, commitOID, result.data.Files, detailDocument.width)
	filePathAtCursor, cursorOnFile := reviewDiffFilePathAtCursor(renderedRows, detailDocument, program.detailState.viewState)
	sectionIDs := append(commitDiffFileSectionIDs(pullRequestKey, commitOID, result.data.Files), commitDiffThreadSectionIDs(pullRequestKey, commitOID, result.data.Files)...)
	if !program.setBrowserDetailSectionsCollapsed(sectionIDs, collapsed) {
		return detailViewSyncPlan{}, false
	}

	plan := detailViewSyncPlan{document: program.buildCurrentDetailDocument(detailDocument.width)}
	if cursorOnFile {
		headerLineIndex := reviewDiffFileHeaderLineIndex(program.currentCommitDiffRenderedRows(pullRequestKey, commitOID, result.data.Files, detailDocument.width), filePathAtCursor)
		if headerLineIndex >= 0 {
			plan.focusLine = headerLineIndex
			plan.focusLineKnown = true
		}
	}
	return plan, true
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

func browserChangesFileSectionIDs(summary githubdomain.PullRequest, files []reviewDiffFile) []string {
	sectionIDs := make([]string, 0, len(files))
	for _, file := range files {
		if trimmedFilePath := strings.TrimSpace(file.Path); trimmedFilePath != "" {
			sectionIDs = append(sectionIDs, browserChangesFileSectionID(summary, trimmedFilePath))
		}
	}
	return sectionIDs
}

func browserChangesThreadSectionIDs(summary githubdomain.PullRequest, files []reviewDiffFile) []string {
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

func commitDiffFileSectionIDs(pullRequestKey string, commitOID string, files []reviewDiffFile) []string {
	sectionIDs := make([]string, 0, len(files))
	for _, file := range files {
		if trimmedFilePath := strings.TrimSpace(file.Path); trimmedFilePath != "" {
			sectionIDs = append(sectionIDs, commitDiffFileSectionID(pullRequestKey, commitOID, trimmedFilePath))
		}
	}
	return sectionIDs
}

func commitDiffThreadSectionIDs(pullRequestKey string, commitOID string, files []reviewDiffFile) []string {
	sectionIDs := make([]string, 0)
	for _, file := range files {
		for _, thread := range file.Threads {
			trimmedThreadID := strings.TrimSpace(thread.ID)
			if trimmedThreadID == "" {
				continue
			}
			sectionIDs = append(sectionIDs, commitDiffThreadSectionID(pullRequestKey, commitOID, thread))
		}
	}
	return sectionIDs
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
