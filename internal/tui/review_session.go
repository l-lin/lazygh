package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
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
	selectedFileIdx              int
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

	changedFiles := program.reviewSessionChangedFileCount()
	title := "Diff preview unavailable"
	if changedFiles > 0 {
		title = fmt.Sprintf("%d %s pending diff load", changedFiles, pluralize(changedFiles, "file", "files"))
	}

	return []Item{{Title: title, Detail: program.reviewSessionPlaceholderDetail()}}
}

func (program *Program) selectedReviewSessionFile() (Item, bool) {
	return itemAt(program.reviewSessionFiles(), program.reviewSession.selectedFileIdx)
}

func (program *Program) adjustReviewSessionSelection(change int) {
	files := program.reviewSessionFiles()
	if len(files) == 0 {
		program.reviewSession.selectedFileIdx = 0
		return
	}

	program.reviewSession.selectedFileIdx = clampIndex(program.reviewSession.selectedFileIdx+change, len(files))
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
			bodyLines = append(bodyLines, fmt.Sprintf("Changed files: %d", detail.ChangedFiles))
		}
	} else {
		bodyLines = append(bodyLines, "", "Loading richer pull request metadata...")
	}
	bodyLines = append(bodyLines, "", "Review mode is active. The real file tree lands in TODO 19.")

	return renderPullRequestDetailContent(renderPullRequestDetailHeader(summary, detail), strings.Join(bodyLines, "\n"))
}

func (program *Program) reviewSessionDetailContent() string {
	selectedFile, ok := program.selectedReviewSessionFile()
	if !ok {
		return program.reviewSessionPlaceholderDetail()
	}

	return selectedFile.Detail
}

func (program *Program) reviewSessionPlaceholderDetail() string {
	summary := program.reviewSession.summary
	repository := pullRequestRepositoryName(summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(program.reviewSession.pendingReviewID), repository, summary.Number),
		"",
		"TODO 19 will replace this placeholder with the selected file diff and collapsed file tree.",
	}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionChangedFileCount() int {
	if !program.reviewSession.active {
		return 0
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
	return fmt.Sprintf(
		"review:%s:%d:%s:file:%d",
		pullRequestRepositoryName(program.reviewSession.summary.Repository),
		program.reviewSession.summary.Number,
		program.reviewSession.pendingReviewID,
		program.reviewSession.selectedFileIdx,
	)
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
