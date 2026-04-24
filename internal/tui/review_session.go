package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	reviewModeMetadataTitle                 = "[1]-Metadata"
	reviewModeFilesTitle                    = "[2]-Files"
	reviewModeDiffTitle                     = "[0]-Diff"
	pendingPullRequestReviewKeptOpenMessage = "Pending review kept open; start review to resume"
)

type reviewSessionState struct {
	active                       bool
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
	program.detailViewState.clearPendingPrefix()
	program.reviewSession = reviewSessionState{
		active:                       true,
		sourceFocus:                  program.model.Focus(),
		sourceDetailTab:              program.activeDetailTab,
		sourcePaneLayoutSize:         program.model.paneLayoutSize,
		sourceFullscreenPane:         program.model.fullscreenPane,
		sourceDetailFullscreenReturn: program.model.detailFullscreenReturnSize,
		summary:                      summary,
		pendingReviewID:              strings.TrimSpace(pendingReviewID),
		collapsedThreadIDs:           map[string]bool{},
	}
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

	result, ok := program.reviewSessionDiffResult()
	if !ok {
		return []Item{{Title: "Loading file tree...", Detail: program.reviewSessionLoadingDetail()}}
	}
	if result.err != nil {
		return []Item{{Title: "Could not load file tree", Detail: program.reviewSessionDiffErrorDetail(result.err)}}
	}
	if len(result.data.FileTree.Rows) == 0 {
		return []Item{{Title: "No changed files", Detail: program.reviewSessionNoDiffDetail()}}
	}

	program.clampReviewSessionSelection()
	return reviewDiffTreeItems(result.data.FileTree)
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
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		return reviewDiffFile{}, false
	}
	program.clampReviewSessionSelection()
	fileIndex, ok := reviewDiffFileIndexAtRow(result.data.FileTree, program.reviewSession.selectedFileTreeRow)
	if !ok || fileIndex < 0 || fileIndex >= len(result.data.Files) {
		return reviewDiffFile{}, false
	}
	return result.data.Files[fileIndex], true
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

func (program *Program) reviewSessionSelectableRows() ([]int, bool) {
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		return nil, false
	}

	return reviewDiffSelectableRowIndexes(result.data.FileTree), true
}
