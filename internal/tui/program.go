package tui

import (
	"net/http"
	"sync"

	"github.com/jesseduffield/gocui"

	clip "github.com/l-lin/lazygh/internal/clipboard"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
	"github.com/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName               = "detail"
	viewUserName                 = "user"
	viewPullRequestsName         = "pull-requests"
	viewNotificationsName        = "notifications"
	viewSearchName               = "search"
	viewStatusLineName           = "status-line"
	viewStatusLineKeyHintsName   = "status-line-key-hints"
	viewActionsPopupName         = "actions-popup"
	viewActionsPopupSearchName   = "actions-popup-search"
	viewPullRequestBuildInfoName = "pull-request-build-info"
)

type GitHubLoader interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
	ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error)
	ListNotifications() ([]githubcli.Notification, error)
	MarkNotificationRead(threadID string) error
	MarkNotificationDone(threadID string) error
	MarkAllNotificationsRead() (githubcli.NotificationBulkReadResult, error)
	MarkAllNotificationsDone(notifications []githubcli.Notification) (int, error)
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	GetIssueDetail(repository string, number int) (githubcli.IssueDetail, error)
	GetReleaseDetail(repository string, id int) (githubcli.ReleaseDetail, error)
	GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error)
	CommentOnPullRequest(repository string, number int, body string) error
	ApprovePullRequest(repository string, number int) error
	ReviewPullRequestWithComment(repository string, number int, body string) error
	RequestChangesOnPullRequest(repository string, number int, body string) error
	SubmitPullRequestReview(pullRequestReviewID string, event githubcli.PullRequestReviewEvent, body string) error
	AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubcli.PullRequestReviewThreadTarget) error
	AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error
	UpdatePullRequestReviewComment(commentID string, body string) error
	DeletePullRequestReviewComment(commentID string) error
	ResolvePullRequestReviewThread(threadID string) error
	UnresolvePullRequestReviewThread(threadID string) error
	AddReaction(subjectID string, content githubcli.ReactionContent) error
	RemoveReaction(subjectID string, content githubcli.ReactionContent) error
	OpenPullRequestInBrowser(repository string, number int) error
	ListAssignableUsers(repository string) ([]githubcli.PullRequestAuthor, error)
	UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error
	EditPullRequestTitle(repository string, number int, title string) error
	EditPullRequestDescription(repository string, number int, body string) error
	MarkPullRequestReadyForReview(repository string, number int) error
	ConvertPullRequestToDraft(repository string, number int) error
	SquashMergePullRequest(repository string, number int) error
	StartPendingPullRequestReview(repository string, number int) (string, error)
	GetPullRequestBuildRun(repository string, check githubcli.PullRequestStatusCheck) (string, error)
	GetPullRequestBuildRunJobs(repository string, check githubcli.PullRequestStatusCheck) ([]githubcli.PullRequestBuildRunJob, error)
	GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error)
	GetPullRequestBuildRunJobLogForCheck(repository string, check githubcli.PullRequestStatusCheck) (githubcli.PullRequestBuildRunJob, string, error)
	RenderMarkdownHTML(repository string, markdown string) (string, error)
	GetAuthToken() (string, error)
}

type Program struct {
	model                                   *Model
	githubLoader                            GitHubLoader
	connectedUserLoadStarted                bool
	connectedUserLogin                      string
	myPullRequestsLoadStarted               bool
	requestedPullRequestsLoadStarted        bool
	notificationsLoadStarted                bool
	myPullRequestsLoading                   bool
	requestedPullRequestsLoading            bool
	notificationsLoading                    bool
	notificationsLoadingDetailMessage       string
	myPullRequestsCount                     int
	myPullRequestsCountKnown                bool
	requestedPullRequestsCount              int
	requestedPullRequestsCountKnown         bool
	additionalPullRequestsLoadStarted       map[PullRequestTab]bool
	additionalPullRequestsLoading           map[PullRequestTab]bool
	additionalPullRequestsCounts            map[PullRequestTab]pullRequestCountState
	pullRequestCache                        persistentPullRequestCache
	notificationDoneStore                   notificationDoneStore
	pullRequestDetailCache                  map[string]pullRequestDetailResult
	pullRequestDetailLoadInFlight           map[string]bool
	pullRequestDetailDocumentCache          map[pullRequestDetailDocumentCacheKey]detailDocument
	pullRequestConversationDocumentCache    map[pullRequestDetailDocumentCacheKey]browserConversationDocument
	pullRequestDiffCache                    map[string]pullRequestDiffResult
	pullRequestDiffLoadInFlight             map[string]bool
	issueDetailCache                        map[string]issueDetailResult
	issueDetailLoadInFlight                 map[string]bool
	releaseDetailCache                      map[string]releaseDetailResult
	releaseDetailLoadInFlight               map[string]bool
	reviewDiffRenderCache                   map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry
	storyReviewLoading                      bool
	loadingSpinnerFrameIndex                int
	detailWrapWidth                         int
	activeDetailTab                         DetailTab
	lastDetailIdentity                      string
	detailViewState                         detailViewState
	clipboardWriter                         clip.Writer
	feedbackMessage                         string
	optimisticMutationSequence              int
	helpVisible                             bool
	searchEditor                            *lineEditor
	actionsPopupSearchEditor                *lineEditor
	actionsPopupErrorMessage                string
	actionsPopupPendingConfirmationActionID string
	reactionPicker                          *reactionPickerState
	themePicker                             *themePickerState
	assigneePicker                          *assigneePickerState
	assigneePickerLoad                      *assigneePickerLoadState
	assignableUsersCache                    map[string][]githubcli.PullRequestAuthor
	reviewSession                           reviewSessionState
	browserCollapsedSectionStates           map[string]bool
	pullRequestBuildRunLoad                 *pullRequestBuildRunLoadState
	pullRequestBuildRunPopup                *pullRequestBuildRunPopupState
	modalEditor                             *modalEditorState
	externalEditor                          externalEditor
	linkOpener                              linkOpener
	detailImageStore                        detailImageStore
	detailImageManager                      detailImageManager
	detailImageHTMLLoadInFlight             map[string]bool
	detailImageHTMLLoadFailed               map[string]bool
	detailImageLoadInFlight                 map[string]bool
	detailImageLoadFailed                   map[string]bool
	githubAuthToken                         string
	githubAuthTokenLoaded                   bool
	detailImageAuthTokenMu                  sync.Mutex
	imageHTTPClient                         *http.Client
	markdownRenderer                        MarkdownRenderer
	storyGenerator                          reviewStoryGenerator
	asyncRunner                             asyncRunner
	uiUpdater                               uiUpdater
	gui                                     *gocui.Gui
	keymapOverrides                         appconfig.KeymapOverrides
	pullRequestSearches                     []appconfig.PullRequestSearch
	storyReviewConfig                       story.Config
	themePresetStore                        themePresetStore
	openedPullRequestSummary                *githubcli.PullRequest
	openedPullRequestTab                    PullRequestTab
	pendingSelectionKeySequence             keySequenceState
	pendingListViewportPlacements           map[string]viewportPlacement
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

	imageStore := newMemoryDetailImageStore()
	imageProtocol := kittyImageProtocol{}

	return &Program{
		model:                                model,
		githubLoader:                         githubLoader,
		pullRequestDetailCache:               map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight:        map[string]bool{},
		pullRequestDetailDocumentCache:       map[pullRequestDetailDocumentCacheKey]detailDocument{},
		pullRequestConversationDocumentCache: map[pullRequestDetailDocumentCacheKey]browserConversationDocument{},
		pullRequestDiffCache:                 map[string]pullRequestDiffResult{},
		pullRequestDiffLoadInFlight:          map[string]bool{},
		issueDetailCache:                     map[string]issueDetailResult{},
		issueDetailLoadInFlight:              map[string]bool{},
		releaseDetailCache:                   map[string]releaseDetailResult{},
		releaseDetailLoadInFlight:            map[string]bool{},
		reviewDiffRenderCache:                map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{},
		assignableUsersCache:                 map[string][]githubcli.PullRequestAuthor{},
		browserCollapsedSectionStates:        map[string]bool{},
		additionalPullRequestsLoadStarted:    map[PullRequestTab]bool{},
		additionalPullRequestsLoading:        map[PullRequestTab]bool{},
		additionalPullRequestsCounts:         map[PullRequestTab]pullRequestCountState{},
		notificationDoneStore:                noopNotificationDoneStore{},
		externalEditor:                       systemExternalEditor{},
		linkOpener:                           newSystemLinkOpener(appconfig.ResolveLinksConfig(appconfig.LinksConfig{}).OpenCommand),
		detailImageStore:                     imageStore,
		detailImageManager:                   &protocolDetailImageManager{imageStore: imageStore, imageProtocol: imageProtocol, terminal: screenTerminalGraphicsTerminal{}},
		detailImageHTMLLoadInFlight:          map[string]bool{},
		detailImageHTMLLoadFailed:            map[string]bool{},
		detailImageLoadInFlight:              map[string]bool{},
		detailImageLoadFailed:                map[string]bool{},
		imageHTTPClient:                      http.DefaultClient,
		markdownRenderer:                     glamourMarkdownRenderer{imageStore: imageStore, imageProtocol: imageProtocol, terminalCellSize: screenTerminalCellSize{}},
		storyGenerator:                       commandReviewStoryGenerator{generator: story.NewGenerator(nil)},
		themePresetStore:                     &defaultThemePresetStore{save: appconfig.SaveThemePresetDefault},
		asyncRunner:                          goroutineAsyncRunner{},
		uiUpdater:                            queuedUIUpdater{},
		clipboardWriter:                      clip.NewSystemWriter(),
		detailViewState:                      newDetailViewState(),
		detailWrapWidth:                      defaultDetailWrapWidth,
		pullRequestSearches:                  appconfig.DefaultPullRequestSearches(),
		pendingListViewportPlacements:        map[string]viewportPlacement{},
	}
}

func (program *Program) Run() error {
	if program.pullRequestCache != nil {
		defer func() {
			_ = program.pullRequestCache.Close()
		}()
	}

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
	gui.BgColor = gocuiColorOrDefault(theme.BackgroundHex)
	gui.FgColor = gocui.GetColor(theme.InactiveTextHex)
	gui.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	gui.SelBgColor = gocuiColorOrDefault(theme.BackgroundHex)
	gui.SelFgColor = gocui.GetColor(theme.ActiveTextHex)
	gui.SelFrameColor = gocui.GetColor(theme.ActiveBorderHex)
}
