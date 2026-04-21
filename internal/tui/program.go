package tui

import (
	"github.com/jesseduffield/gocui"

	clip "codeberg.org/l-lin/lazygh/internal/clipboard"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName             = "detail"
	viewUserName               = "user"
	viewPullRequestsName       = "pull-requests"
	viewSearchName             = "search"
	viewActionsPopupName       = "actions-popup"
	viewActionsPopupSearchName = "actions-popup-search"
)

type GitHubLoader interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
	ListMyPullRequests() ([]githubcli.PullRequest, error)
	ListRequestedPullRequests() ([]githubcli.PullRequest, error)
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	CommentOnPullRequest(repository string, number int, body string) error
	ApprovePullRequest(repository string, number int) error
	ReviewPullRequestWithComment(repository string, number int, body string) error
	RequestChangesOnPullRequest(repository string, number int, body string) error
	OpenPullRequestInBrowser(repository string, number int) error
	EditPullRequestTitle(repository string, number int, title string) error
	EditPullRequestDescription(repository string, number int, body string) error
}

type Program struct {
	model                            *Model
	githubLoader                     GitHubLoader
	connectedUserLoadStarted         bool
	myPullRequestsLoadStarted        bool
	requestedPullRequestsLoadStarted bool
	myPullRequestsLoading            bool
	requestedPullRequestsLoading     bool
	myPullRequestsCount              int
	myPullRequestsCountKnown         bool
	requestedPullRequestsCount       int
	requestedPullRequestsCountKnown  bool
	pullRequestDetailCache           map[string]pullRequestDetailResult
	pullRequestDetailLoadInFlight    map[string]bool
	detailWrapWidth                  int
	activeDetailTab                  DetailTab
	lastDetailIdentity               string
	detailViewState                  detailViewState
	clipboardWriter                  clip.Writer
	feedbackFocus                    Focus
	feedbackMessage                  string
	helpVisible                      bool
	searchEditor                     *lineEditor
	actionsPopupSearchEditor         *lineEditor
	actionsPopupErrorMessage         string
	modalEditor                      *modalEditorState
	externalEditor                   externalEditor
	markdownRenderer                 MarkdownRenderer
	asyncRunner                      asyncRunner
	uiUpdater                        uiUpdater
	gui                              *gocui.Gui
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
		model:                         model,
		githubLoader:                  githubLoader,
		pullRequestDetailCache:        map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight: map[string]bool{},
		externalEditor:                systemExternalEditor{},
		markdownRenderer:              glamourMarkdownRenderer{},
		asyncRunner:                   goroutineAsyncRunner{},
		uiUpdater:                     queuedUIUpdater{},
		clipboardWriter:               clip.NewSystemWriter(),
		detailViewState:               newDetailViewState(),
		detailWrapWidth:               defaultDetailWrapWidth,
	}
}

func (program *Program) Run() error {
	gui, err := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if err != nil {
		return err
	}
	defer gui.Close()

	program.configureGUI(gui)
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
	gui.BgColor = gocui.ColorDefault
	gui.FgColor = gocui.GetColor(theme.InactiveTextHex)
	gui.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	gui.SelBgColor = gocui.ColorDefault
	gui.SelFgColor = gocui.GetColor(theme.ActiveTextHex)
	gui.SelFrameColor = gocui.GetColor(theme.ActiveBorderHex)
}
