package tui

import (
	"time"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

type startupStateModel struct {
	loadingSpinnerFrameIndex int
	appStarted               bool
}

type detailStateModel struct {
	wrapWidth        int
	wordWrapDisabled bool
	activeTab        DetailTab
	commitDiffTab    commitDiffTabState
	lastIdentity     string
	viewState        detailViewState
}

type overlayStateModel struct {
	helpVisible         bool
	transientErrorPopup transientErrorPopupState
	errorMessages       []string
	modalEditor         modalEditorState
}

type navigationStateModel struct {
	reviewSession               reviewSessionState
	openedPullRequestSummary    *githubdomain.PullRequest
	openedPullRequestTab        PullRequestTab
	pendingSelectionKeySequence keySequenceState
}

type runtimeConfigState struct {
	keymapOverrides     appconfig.KeymapOverrides
	pullRequestSearches []appconfig.PullRequestSearch
	displayConfig       appconfig.DisplayConfig
	storyReviewConfig   story.Config
}

type timingStateModel struct {
	now                         func() time.Time
	after                       func(time.Duration) <-chan time.Time
	yankHighlightDuration       time.Duration
	transientErrorPopupDuration time.Duration
}

type manualRefreshStateModel struct {
	pullRequestListPending   map[PullRequestTab]bool
	pullRequestDetailPending map[string]bool
	pullRequestDiffPending   map[string]bool
	notificationPending      bool
	feedback                 *manualRefreshFeedbackState
}
