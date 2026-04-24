package tui

import (
	"github.com/jesseduffield/gocui"

	clip "codeberg.org/l-lin/lazygh/internal/clipboard"
	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName             = "detail"
	viewUserName               = "user"
	viewPullRequestsName       = "pull-requests"
	viewSearchName             = "search"
	viewStatusLineName         = "status-line"
	viewActionsPopupName       = "actions-popup"
	viewActionsPopupSearchName = "actions-popup-search"
)

type GitHubLoader interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
	ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error)
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error)
	CommentOnPullRequest(repository string, number int, body string) error
	ApprovePullRequest(repository string, number int) error
	ReviewPullRequestWithComment(repository string, number int, body string) error
	RequestChangesOnPullRequest(repository string, number int, body string) error
	SubmitPullRequestReview(pullRequestReviewID string, event githubcli.PullRequestReviewEvent, body string) error
	AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubcli.PullRequestReviewThreadTarget) error
	OpenPullRequestInBrowser(repository string, number int) error
	EditPullRequestTitle(repository string, number int, title string) error
	EditPullRequestDescription(repository string, number int, body string) error
	StartPendingPullRequestReview(repository string, number int) (string, error)
}

type Program struct {
	model                             *Model
	githubLoader                      GitHubLoader
	connectedUserLoadStarted          bool
	myPullRequestsLoadStarted         bool
	requestedPullRequestsLoadStarted  bool
	myPullRequestsLoading             bool
	requestedPullRequestsLoading      bool
	myPullRequestsCount               int
	myPullRequestsCountKnown          bool
	requestedPullRequestsCount        int
	requestedPullRequestsCountKnown   bool
	additionalPullRequestsLoadStarted map[PullRequestTab]bool
	additionalPullRequestsLoading     map[PullRequestTab]bool
	additionalPullRequestsCounts      map[PullRequestTab]pullRequestCountState
	pullRequestDetailCache            map[string]pullRequestDetailResult
	pullRequestDetailLoadInFlight     map[string]bool
	pullRequestDiffCache              map[string]pullRequestDiffResult
	pullRequestDiffLoadInFlight       map[string]bool
	loadingSpinnerFrameIndex          int
	detailWrapWidth                   int
	activeDetailTab                   DetailTab
	lastDetailIdentity                string
	detailViewState                   detailViewState
	clipboardWriter                   clip.Writer
	feedbackMessage                   string
	helpVisible                       bool
	searchEditor                      *lineEditor
	actionsPopupSearchEditor          *lineEditor
	actionsPopupErrorMessage          string
	reviewSession                     reviewSessionState
	modalEditor                       *modalEditorState
	externalEditor                    externalEditor
	markdownRenderer                  MarkdownRenderer
	asyncRunner                       asyncRunner
	uiUpdater                         uiUpdater
	gui                               *gocui.Gui
	keymapOverrides                   appconfig.KeymapOverrides
	pullRequestSearches               []appconfig.PullRequestSearch
}

func NewProgram(githubLoaders ...GitHubLoader) *Program {
	var githubLoader GitHubLoader
	if len(githubLoaders) > 0 {
		githubLoader = githubLoaders[0]
	}

	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	return NewProgramWithModelAndLoader(model, githubLoader)
}

func NewProgramWithModel(model *Model) *Program {
	return NewProgramWithModelAndLoader(model, nil)
}

func NewProgramWithModelAndLoader(model *Model, githubLoader GitHubLoader) *Program {
	if model == nil {
		model = NewModel(DefaultSeedData())
	}

	return &Program{
		model:                             model,
		githubLoader:                      githubLoader,
		pullRequestDetailCache:            map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight:     map[string]bool{},
		pullRequestDiffCache:              map[string]pullRequestDiffResult{},
		pullRequestDiffLoadInFlight:       map[string]bool{},
		additionalPullRequestsLoadStarted: map[PullRequestTab]bool{},
		additionalPullRequestsLoading:     map[PullRequestTab]bool{},
		additionalPullRequestsCounts:      map[PullRequestTab]pullRequestCountState{},
		externalEditor:                    systemExternalEditor{},
		markdownRenderer:                  glamourMarkdownRenderer{},
		asyncRunner:                       goroutineAsyncRunner{},
		uiUpdater:                         queuedUIUpdater{},
		clipboardWriter:                   clip.NewSystemWriter(),
		detailViewState:                   newDetailViewState(),
		detailWrapWidth:                   defaultDetailWrapWidth,
		pullRequestSearches:               appconfig.DefaultPullRequestSearches(),
	}
}

func (program *Program) Run() error {
	gui, err := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if err != nil {
		return err
	}
	defer gui.Close()

	program.configureGUI(gui)
	stopLoadingSpinner := program.startLoadingSpinner(gui)
	defer stopLoadingSpinner()
	gui.SetManagerFunc(program.layout)

	if err := program.setKeybindings(gui); err != nil {
		return err
	}

	if err := gui.MainLoop(); err != nil && !isQuitError(err) {
		return err
	}

	return nil
}

func (program *Program) configureGUI(gui *gocui.Gui) {
	gui.Highlight = true
	gui.InputEsc = true
	gui.Cursor = false
	gui.ShowListFooter = true
	gui.BgColor = gocui.ColorDefault
	gui.FgColor = gocui.GetColor(theme.InactiveTextHex)
	gui.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	gui.SelBgColor = gocui.ColorDefault
	gui.SelFgColor = gocui.GetColor(theme.ActiveTextHex)
	gui.SelFrameColor = gocui.GetColor(theme.ActiveBorderHex)
}
