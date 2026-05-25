package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	reviewModeMetadataTitle                 = "[1]-" + reviewModeMetadataIcon + " Metadata"
	reviewModeFilesTitle                    = "[2]-" + reviewDiffDirectoryIcon + " Files"
	reviewModeChaptersTitle                 = "[2]-" + reviewModeChapterIcon + " Chapters"
	reviewModeDescriptionTitle              = "[0]-" + detailDescriptionIcon + " Description"
	reviewModeDiffTitle                     = "[0]-" + detailChangesIcon + " Diff"
	reviewModeChapterTitle                  = "[0]-" + reviewModeChapterIcon + " Chapter"
	pendingPullRequestReviewKeptOpenMessage = "Pending review kept open; start review to resume"
)

type reviewSessionMode int

const (
	reviewSessionModeDiff reviewSessionMode = iota
	reviewSessionModeStory
)

type reviewSessionState struct {
	active                       bool
	mode                         reviewSessionMode
	sourceFocus                  Focus
	sourceDetailTab              DetailTab
	sourcePaneLayoutSize         PaneLayoutSize
	sourceFullscreenPane         Focus
	sourceDetailFullscreenReturn PaneLayoutSize
	summary                      githubdomain.PullRequest
	pendingReviewID              string
	selectedFileTreeRow          int
	fileTreeSearchQuery          string
	collapsedTreeRowIDs          map[string]bool
	collapsedThreadIDs           map[string]bool
	story                        reviewStoryData
}

func (program *Program) startReviewAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "start-review",
		title:   "Start review",
		icon:    actionsPopupStartReviewIcon,
		execute: program.executeStartReviewAction,
	}
}

func (program *Program) executeStartReviewAction(_ *gocui.Gui) actionsPopupActionResult {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.openPullRequestReview(summary); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) startReviewSession(summary any, pendingReviewID string) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	program.startReviewSessionWithMode(summaryValue, pendingReviewID, reviewSessionModeDiff, reviewStoryData{})
}

func (program *Program) startStoryReviewSession(summary any, pendingReviewID string, story reviewStoryData) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	program.startReviewSessionWithMode(summaryValue, pendingReviewID, reviewSessionModeStory, story)
}

func (program *Program) startReviewSessionWithMode(summary githubdomain.PullRequest, pendingReviewID string, mode reviewSessionMode, story reviewStoryData) {
	program.detailViewState.clearPendingPrefix()
	trimmedPendingReviewID := strings.TrimSpace(pendingReviewID)
	program.reviewSession = reviewSessionState{
		active:                       true,
		mode:                         mode,
		sourceFocus:                  program.model.Focus(),
		sourceDetailTab:              program.activeDetailTab,
		sourcePaneLayoutSize:         program.model.paneLayoutSize,
		sourceFullscreenPane:         program.model.fullscreenPane,
		sourceDetailFullscreenReturn: program.model.detailFullscreenReturnSize,
		summary:                      summary,
		pendingReviewID:              trimmedPendingReviewID,
		selectedFileTreeRow:          -1,
		collapsedTreeRowIDs:          map[string]bool{},
		collapsedThreadIDs:           map[string]bool{},
		story:                        story,
	}
	if trimmedPendingReviewID != "" {
		program.setPendingPullRequestReviewState(summary, trimmedPendingReviewID)
	}
	program.invalidateReviewDiffRenderCache()
	program.model.SetPaneLayoutSize(program.reviewModePaneLayoutSize())
	program.model.FocusPullRequestsView()
}

func (program *Program) exitReviewMode(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgExitReviewMode{})
}

func (program *Program) restorePullRequestBrowserFromReviewMode() {
	if !program.reviewSession.active {
		return
	}

	sourceFocus := program.reviewSession.sourceFocus
	sourceDetailTab := program.reviewSession.sourceDetailTab
	sourcePaneLayoutSize := program.reviewSession.sourcePaneLayoutSize
	sourceFullscreenPane := program.reviewSession.sourceFullscreenPane
	sourceDetailFullscreenReturn := program.reviewSession.sourceDetailFullscreenReturn
	program.reviewSession = reviewSessionState{}
	program.invalidateReviewDiffRenderCache()
	program.activeDetailTab = sourceDetailTab
	program.detailViewState.clearPendingPrefix()
	program.model.SetPaneLayoutSize(sourcePaneLayoutSize)
	program.model.SetFullscreenPane(sourceFullscreenPane)
	program.model.SetDetailFullscreenReturnSize(sourceDetailFullscreenReturn)

	switch sourceFocus {
	case FocusDetailView:
		program.model.FocusDetailView()
	default:
		program.model.FocusPullRequestsView()
	}
}

func (program *Program) reviewModePaneLayoutSize() PaneLayoutSize {
	if program.model.paneLayoutSize != PaneLayoutFullscreen {
		return program.model.paneLayoutSize
	}
	if program.model.fullscreenPane == FocusDetailView && program.model.detailFullscreenReturnSize != PaneLayoutFullscreen {
		return program.model.detailFullscreenReturnSize
	}
	return PaneLayoutDefault
}

func (program *Program) reviewSessionFiles() []Item {
	if !program.reviewSession.active {
		return nil
	}

	tree, files, ok := program.reviewSessionCurrentTree()
	if !ok {
		if result, diffOk := program.reviewSessionDiffResult(); diffOk && result.err != nil {
			return []Item{{Title: "Could not load file tree", Detail: program.reviewSessionDiffErrorDetail(result.err)}}
		}
		return []Item{{Title: "Loading file tree...", Detail: program.reviewSessionLoadingDetail()}}
	}
	if len(tree.Rows) == 0 {
		return []Item{{Title: "No changed files", Detail: program.reviewSessionNoDiffDetail()}}
	}

	program.clampReviewSessionSelection()
	return reviewDiffTreeItems(tree, files)
}

func (program *Program) reviewSessionSelectedVisibleLine() int {
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		return 0
	}
	program.clampReviewSessionSelection()
	return program.reviewSession.selectedFileTreeRow
}

func (program *Program) selectedReviewSessionDiffFile() (reviewDiffFile, bool) {
	row, files, ok := program.selectedReviewSessionTreeRow()
	if !ok {
		return reviewDiffFile{}, false
	}
	fileIndex := row.FileIndex
	if fileIndex < 0 {
		rawTree, _, rawTreeOK := program.reviewSessionRawTree()
		if !rawTreeOK {
			return reviewDiffFile{}, false
		}
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

func (program *Program) clampReviewSessionSelection() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}
	if program.reviewSession.selectedFileTreeRow < 0 {
		if program.reviewSession.mode != reviewSessionModeStory {
			if fileRows, fileRowsOK := program.reviewSessionFileRows(); fileRowsOK && len(fileRows) > 0 {
				program.reviewSession.selectedFileTreeRow = fileRows[0]
				return
			}
		}
		program.reviewSession.selectedFileTreeRow = selectableRows[0]
		return
	}

	program.reviewSession.selectedFileTreeRow = adjustVisibleSelection(program.reviewSession.selectedFileTreeRow, selectableRows, 0)
}

func (program *Program) adjustReviewSessionSelection(change int) {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	program.reviewSession.selectedFileTreeRow = adjustVisibleSelection(program.reviewSession.selectedFileTreeRow, selectableRows, change)
}

func (program *Program) moveReviewSessionSelectionToTop() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	program.reviewSession.selectedFileTreeRow = selectableRows[0]
}

func (program *Program) moveReviewSessionSelectionToBottom() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	program.reviewSession.selectedFileTreeRow = selectableRows[len(selectableRows)-1]
}

func (program *Program) reviewSessionSelectableRows() ([]int, bool) {
	tree, _, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil, false
	}
	return reviewDiffSelectableTreeRowIndexes(tree), true
}

func (program *Program) reviewSessionRawTree() (reviewDiffTree, []reviewDiffFile, bool) {
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		return reviewDiffTree{}, nil, false
	}
	if program.reviewSession.mode == reviewSessionModeStory && len(program.reviewSession.story.Tree.Rows) > 0 {
		return program.reviewSession.story.Tree, result.data.Files, true
	}
	return result.data.FileTree, result.data.Files, true
}

func (program *Program) reviewSessionCurrentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	tree, files, ok := program.reviewSessionRawTree()
	if !ok {
		return reviewDiffTree{}, nil, false
	}
	return reviewDiffTreeVisibleRows(tree, program.reviewSession.collapsedTreeRowIDs), files, true
}

func (program *Program) selectedReviewSessionTreeRow() (reviewDiffTreeRow, []reviewDiffFile, bool) {
	tree, files, ok := program.reviewSessionCurrentTree()
	if !ok || len(tree.Rows) == 0 {
		return reviewDiffTreeRow{}, nil, false
	}
	program.clampReviewSessionSelection()
	rowIndex := clampIndex(program.reviewSession.selectedFileTreeRow, len(tree.Rows))
	return tree.Rows[rowIndex], files, true
}

func (program *Program) selectedReviewSessionStoryChapter() (reviewStoryChapter, bool) {
	row, _, ok := program.selectedReviewSessionTreeRow()
	if !ok || row.Kind != reviewDiffTreeRowKindChapter {
		return reviewStoryChapter{}, false
	}
	if row.ChapterIndex < 0 || row.ChapterIndex >= len(program.reviewSession.story.Chapters) {
		return reviewStoryChapter{}, false
	}
	return program.reviewSession.story.Chapters[row.ChapterIndex], true
}

func (program *Program) reviewSessionFileRows() ([]int, bool) {
	tree, _, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil, false
	}
	return reviewDiffSelectableRowIndexes(tree), true
}
