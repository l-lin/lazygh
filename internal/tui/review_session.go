package tui

import (
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
	collapsedTreeRowIDs          map[string]bool
	collapsedThreadIDs           map[string]bool
	story                        reviewStoryData
}

func (program *Program) startReviewAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "start-review",
		title:   "Start review",
		icon:    actionsPopupStartReviewIcon,
		execute: actionsPopupExecuteErr(program.executeStartReviewAction),
	}
}

func (program *Program) executeStartReviewAction(gui *gocui.Gui) error {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if !program.hasReviewMutations() {
		return errActionsPopupActionUnavailable
	}
	return program.dispatch(gui, MsgStartPullRequestReviewRequested{Summary: summary})
}

func (program *Program) exitReviewMode(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgExitReviewMode{})
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
	if !program.navigationState.reviewSession.active {
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
	return program.navigationState.reviewSession.selectedFileTreeRow
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
	if !ok {
		program.navigationState.reviewSession.clampSelection(nil, nil)
		return
	}

	fileRows, fileRowsOK := program.reviewSessionFileRows()
	if !fileRowsOK {
		fileRows = nil
	}
	program.navigationState.reviewSession.clampSelection(selectableRows, fileRows)
}

func (program *Program) adjustReviewSessionSelection(change int) {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok {
		program.navigationState.reviewSession.adjustSelection(nil, change)
		return
	}

	program.navigationState.reviewSession.adjustSelection(selectableRows, change)
}

func (program *Program) moveReviewSessionSelectionToTop() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok {
		program.navigationState.reviewSession.moveSelectionToTop(nil)
		return
	}

	program.navigationState.reviewSession.moveSelectionToTop(selectableRows)
}

func (program *Program) moveReviewSessionSelectionToBottom() {
	selectableRows, ok := program.reviewSessionSelectableRows()
	if !ok {
		program.navigationState.reviewSession.moveSelectionToBottom(nil)
		return
	}

	program.navigationState.reviewSession.moveSelectionToBottom(selectableRows)
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
	if program.navigationState.reviewSession.mode == reviewSessionModeStory && len(program.navigationState.reviewSession.story.Tree.Rows) > 0 {
		return program.navigationState.reviewSession.story.Tree, result.data.Files, true
	}
	return result.data.FileTree, result.data.Files, true
}

func (program *Program) reviewSessionCurrentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	tree, files, ok := program.reviewSessionRawTree()
	if !ok {
		return reviewDiffTree{}, nil, false
	}
	return reviewDiffTreeVisibleRows(tree, program.navigationState.reviewSession.collapsedTreeRowIDs), files, true
}

func (program *Program) selectedReviewSessionTreeRow() (reviewDiffTreeRow, []reviewDiffFile, bool) {
	tree, files, ok := program.reviewSessionCurrentTree()
	if !ok || len(tree.Rows) == 0 {
		return reviewDiffTreeRow{}, nil, false
	}
	program.clampReviewSessionSelection()
	rowIndex := clampIndex(program.navigationState.reviewSession.selectedFileTreeRow, len(tree.Rows))
	return tree.Rows[rowIndex], files, true
}

func (program *Program) selectedReviewSessionStoryChapter() (reviewStoryChapter, bool) {
	row, _, ok := program.selectedReviewSessionTreeRow()
	if !ok || row.Kind != reviewDiffTreeRowKindChapter {
		return reviewStoryChapter{}, false
	}
	if row.ChapterIndex < 0 || row.ChapterIndex >= len(program.navigationState.reviewSession.story.Chapters) {
		return reviewStoryChapter{}, false
	}
	return program.navigationState.reviewSession.story.Chapters[row.ChapterIndex], true
}

func (program *Program) reviewSessionFileRows() ([]int, bool) {
	tree, _, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil, false
	}
	return reviewDiffSelectableRowIndexes(tree), true
}
