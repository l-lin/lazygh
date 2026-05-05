package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	reviewModeMetadataTitle                 = "[1]-Metadata"
	reviewModeFilesTitle                    = "[2]-Files"
	reviewModeChaptersTitle                 = "[2]-Chapters"
	reviewModeDescriptionTitle              = "[0]-Description"
	reviewModeDiffTitle                     = "[0]-Diff"
	reviewModeChapterTitle                  = "[0]-Chapter"
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
	summary                      githubcli.PullRequest
	pendingReviewID              string
	selectedFileTreeRow          int
	fileTreeSearchQuery          string
	collapsedThreadIDs           map[string]bool
	story                        reviewStoryData
}

func (program *Program) startReviewAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "start-review",
		title:    "Start review",
		icon:     actionsPopupStartReviewIcon,
		keywords: []string{"start", "review", "pending", "session", "inline"},
		execute:  program.executeStartReviewAction,
	}
}

func (program *Program) executeStartReviewAction(_ *gocui.Gui) actionsPopupActionResult {
	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.openPullRequestReview(summary); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) startReviewSession(summary githubcli.PullRequest, pendingReviewID string) {
	program.startReviewSessionWithMode(summary, pendingReviewID, reviewSessionModeDiff, reviewStoryData{})
}

func (program *Program) startStoryReviewSession(summary githubcli.PullRequest, pendingReviewID string, story reviewStoryData) {
	program.startReviewSessionWithMode(summary, pendingReviewID, reviewSessionModeStory, story)
}

func (program *Program) startReviewSessionWithMode(summary githubcli.PullRequest, pendingReviewID string, mode reviewSessionMode, story reviewStoryData) {
	program.detailViewState.clearPendingPrefix()
	program.reviewSession = reviewSessionState{
		active:                       true,
		mode:                         mode,
		sourceFocus:                  program.model.Focus(),
		sourceDetailTab:              program.activeDetailTab,
		sourcePaneLayoutSize:         program.model.paneLayoutSize,
		sourceFullscreenPane:         program.model.fullscreenPane,
		sourceDetailFullscreenReturn: program.model.detailFullscreenReturnSize,
		summary:                      summary,
		pendingReviewID:              strings.TrimSpace(pendingReviewID),
		collapsedThreadIDs:           map[string]bool{},
		story:                        story,
	}
	program.invalidateReviewDiffRenderCache()
	program.model.paneLayoutSize = program.reviewModePaneLayoutSize()
	program.model.FocusPullRequestsView()
}

func (program *Program) exitReviewMode(gui *gocui.Gui, _ *gocui.View) error {
	if !program.reviewSession.active {
		return nil
	}

	sourceFocus := program.reviewSession.sourceFocus
	pendingReviewID := strings.TrimSpace(program.reviewSession.pendingReviewID)
	program.restorePullRequestBrowserFromReviewMode()
	if pendingReviewID != "" {
		program.setFeedback(sourceFocus, pendingPullRequestReviewKeptOpenMessage)
	}
	if gui == nil {
		return nil
	}

	return program.layout(gui)
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
	program.model.paneLayoutSize = sourcePaneLayoutSize
	program.model.fullscreenPane = sourceFullscreenPane
	program.model.detailFullscreenReturnSize = sourceDetailFullscreenReturn

	switch sourceFocus {
	case FocusDetailView:
		program.model.lastSideFocus = FocusPullRequestsView
		program.model.focus = FocusDetailView
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
	if !ok || row.FileIndex < 0 || row.FileIndex >= len(files) {
		return reviewDiffFile{}, false
	}
	return files[row.FileIndex], true
}

func (program *Program) clampReviewSessionSelection() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
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
	if program.reviewSession.mode == reviewSessionModeStory {
		return reviewDiffSelectableRowIndexesIncludingChapters(tree), true
	}
	return reviewDiffSelectableRowIndexes(tree), true
}

func (program *Program) reviewSessionCurrentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		return reviewDiffTree{}, nil, false
	}
	if program.reviewSession.mode == reviewSessionModeStory && len(program.reviewSession.story.Tree.Rows) > 0 {
		return program.reviewSession.story.Tree, result.data.Files, true
	}
	return result.data.FileTree, result.data.Files, true
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
