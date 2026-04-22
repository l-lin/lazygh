package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	reviewModeMetadataTitle = "[1]-Metadata"
	reviewModeFilesTitle    = "[2]-Files"
	reviewModeDiffTitle     = "[0]-Diff"
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
}

func (program *Program) executeStartReviewAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	pendingReviewID, err := program.githubLoader.StartPendingPullRequestReview(target.repository, target.number)
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.startReviewSession(summary, pendingReviewID)
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
	}
	program.model.paneLayoutSize = program.reviewModePaneLayoutSize()
	program.model.FocusPullRequestsView()
}

func (program *Program) exitReviewMode(gui *gocui.Gui, _ *gocui.View) error {
	if !program.reviewSession.active {
		return nil
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

	if gui == nil {
		return nil
	}

	return program.layout(gui)
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
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	selectableRows := reviewDiffSelectableRowIndexes(result.data.FileTree)
	if len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}
	if indexOfInt(selectableRows, program.reviewSession.selectedFileTreeRow) >= 0 {
		return
	}
	program.reviewSession.selectedFileTreeRow = selectableRows[0]
}

func (program *Program) adjustReviewSessionSelection(change int) {
	result, ok := program.reviewSessionDiffResult()
	if !ok || result.err != nil {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	selectableRows := reviewDiffSelectableRowIndexes(result.data.FileTree)
	if len(selectableRows) == 0 {
		program.reviewSession.selectedFileTreeRow = 0
		return
	}

	program.clampReviewSessionSelection()
	selectedIndex := indexOfInt(selectableRows, program.reviewSession.selectedFileTreeRow)
	if selectedIndex < 0 {
		selectedIndex = 0
	}
	selectedIndex = clampIndex(selectedIndex+change, len(selectableRows))
	program.reviewSession.selectedFileTreeRow = selectableRows[selectedIndex]
}

func (program *Program) reviewSessionMetadataContent() string {
	summary := program.reviewSession.summary
	detail := githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}
	bodyLines := []string{fmt.Sprintf("Pending review: %s", valueOrDash(program.reviewSession.pendingReviewID))}

	if result, ok := program.pullRequestDetailForSummary(summary); ok {
		if result.err != nil {
			bodyLines = append(bodyLines, "", "Could not load richer pull request metadata.", strings.TrimSpace(result.err.Error()))
		} else {
			detail = result.detail
		}
	} else {
		bodyLines = append(bodyLines, "", "Loading richer pull request metadata...")
	}

	if result, ok := program.reviewSessionDiffResult(); ok {
		if result.err != nil {
			bodyLines = append(bodyLines, "", "Could not load changed files.", strings.TrimSpace(result.err.Error()))
		} else {
			bodyLines = append(bodyLines, renderReviewSessionMetadataStats(result.data.Stats))
		}
	} else {
		bodyLines = append(bodyLines, "", "Loading changed files...")
	}

	return renderPullRequestDetailContent(renderPullRequestDetailHeader(summary, detail), strings.Join(bodyLines, "\n"))
}

func renderReviewSessionMetadataStats(stats reviewDiffStats) string {
	return strings.Join([]string{
		fmt.Sprintf("Changed files: %d", stats.ChangedFiles),
		styleText(fmt.Sprintf("+%d", stats.Additions), foregroundColorEscape(theme.DiffAdditionForegroundHex)),
		styleText(fmt.Sprintf("-%d", stats.Deletions), foregroundColorEscape(theme.DiffDeletionForegroundHex)),
	}, "  ")
}

func (program *Program) reviewSessionDetailContent() string {
	if !program.reviewSession.active {
		return ""
	}

	result, ok := program.reviewSessionDiffResult()
	if !ok {
		return program.reviewSessionLoadingDetail()
	}
	if result.err != nil {
		return program.reviewSessionDiffErrorDetail(result.err)
	}
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return program.reviewSessionNoDiffDetail()
	}
	return renderReviewDiffFile(selectedFile, program.markdownRenderer, program.detailWrapWidth)
}

func (program *Program) reviewSessionLoadingDetail() string {
	summary := program.reviewSession.summary
	repository := pullRequestRepositoryName(summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(program.reviewSession.pendingReviewID), repository, summary.Number),
		"",
		"Loading pull request diff...",
		fmt.Sprintf("Running `gh api repos/%s/pulls/%d -H 'Accept: application/vnd.github.v3.diff'`.", repository, summary.Number),
	}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionDiffErrorDetail(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub misplaced the diff again."
	}

	lines := []string{program.reviewSessionLoadingDetail(), "", message}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionNoDiffDetail() string {
	summary := program.reviewSession.summary
	repository := pullRequestRepositoryName(summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(program.reviewSession.pendingReviewID), repository, summary.Number),
		"",
		"No changed files are available for this review.",
	}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionChangedFileCount() int {
	if !program.reviewSession.active {
		return 0
	}
	if result, ok := program.reviewSessionDiffResult(); ok && result.err == nil {
		return result.data.Stats.ChangedFiles
	}
	if result, ok := program.pullRequestDetailForSummary(program.reviewSession.summary); ok && result.err == nil {
		return result.detail.ChangedFiles
	}
	return 0
}

func (program *Program) reviewSessionDetailIdentity() string {
	if !program.reviewSession.active {
		return ""
	}
	selectedFilePath := fmt.Sprintf("row:%d", program.reviewSession.selectedFileTreeRow)
	if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
		selectedFilePath = selectedFile.Path
	}
	return fmt.Sprintf(
		"review:%s:%d:%s:file:%s",
		pullRequestRepositoryName(program.reviewSession.summary.Repository),
		program.reviewSession.summary.Number,
		program.reviewSession.pendingReviewID,
		selectedFilePath,
	)
}

func (program *Program) reviewSessionDiffResult() (pullRequestDiffResult, bool) {
	if !program.reviewSession.active {
		return pullRequestDiffResult{}, false
	}
	return program.pullRequestDiffForSummary(program.reviewSession.summary)
}

func renderReadOnlyTextView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}
