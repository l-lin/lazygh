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

type reviewSessionStartDescriptor struct {
	mode                         reviewSessionMode
	sourceFocus                  Focus
	sourceDetailTab              DetailTab
	sourcePaneLayoutSize         PaneLayoutSize
	sourceFullscreenPane         Focus
	sourceDetailFullscreenReturn PaneLayoutSize
	summary                      githubdomain.PullRequest
	pendingReviewID              string
	story                        reviewStoryData
}

func (program *Program) startReviewAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	summary, ok := program.currentPullRequestSummary()
	if ok && program.hasReviewMutations() {
		requested = MsgStartPullRequestReviewRequested{Summary: summary}
	}
	return actionsPopupAction{
		id:        "start-review",
		title:     "Start review",
		icon:      actionsPopupStartReviewIcon,
		requested: requested,
	}
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
