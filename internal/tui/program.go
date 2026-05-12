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

type Program struct {
	model                                   *Model
	sessionQueries                          SessionQueries
	pullRequestListQueries                  PullRequestListQueries
	notificationQueries                     NotificationQueries
	detailQueries                           DetailQueries
	pullRequestMutations                    PullRequestMutations
	reviewMutations                         ReviewMutations
	notificationMutations                   NotificationMutations
	reactionMutations                       ReactionMutations
	buildQueries                            BuildQueries
	markdownHTMLRenderer                    MarkdownHTMLRenderer
	authTokenProvider                       AuthTokenProvider
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
	pullRequestChangesRenderedRowsCache     map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow
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
	clipboardWriter                         ClipboardWriter
	feedbackMessage                         string
	optimisticMutationSequence              int
	helpVisible                             bool
	searchEditor                            *lineEditor
	actionsPopupSearchEditor                *lineEditor
	actionsPopupErrorMessage                string
	actionsPopupPendingConfirmationActionID string
	pendingPullRequestReviewCache           map[string]pendingPullRequestReviewState
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
	externalEditor                          ExternalEditor
	linkOpener                              LinkOpener
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
	themePresetStore                        ThemePresetStore
	openedPullRequestSummary                *githubcli.PullRequest
	openedPullRequestTab                    PullRequestTab
	pendingSelectionKeySequence             keySequenceState
	pendingListViewportPlacements           map[string]viewportPlacement
}

func NewProgram(depsOrLoaders ...any) *Program {
	var loader any
	if len(depsOrLoaders) > 0 {
		loader = depsOrLoaders[0]
	}

	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	return NewProgramWithModelAndLoader(model, loader)
}

func NewProgramWithModel(model *Model) *Program {
	return NewProgramWithModelAndDeps(model, AppDeps{})
}

func NewProgramWithModelAndLoader(model *Model, githubLoader any) *Program {
	return NewProgramWithModelAndDeps(model, appDepsFromCompatibilityLoader(githubLoader))
}

func NewProgramWithModelAndDeps(model *Model, deps AppDeps) *Program {
	if model == nil {
		model = NewModel(DefaultSeedData())
	}

	resolvedDeps := resolveAppDeps(deps)
	imageStore := newMemoryDetailImageStore()
	imageProtocol := kittyImageProtocol{}

	return &Program{
		model:                                model,
		sessionQueries:                       resolvedDeps.SessionQueries,
		pullRequestListQueries:               resolvedDeps.PullRequestList,
		notificationQueries:                  resolvedDeps.NotificationQueries,
		detailQueries:                        resolvedDeps.DetailQueries,
		pullRequestMutations:                 resolvedDeps.PullRequestMutations,
		reviewMutations:                      resolvedDeps.ReviewMutations,
		notificationMutations:                resolvedDeps.NotificationMutations,
		reactionMutations:                    resolvedDeps.ReactionMutations,
		buildQueries:                         resolvedDeps.BuildQueries,
		markdownHTMLRenderer:                 resolvedDeps.MarkdownHTMLRenderer,
		authTokenProvider:                    resolvedDeps.AuthTokenProvider,
		pullRequestDetailCache:               map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight:        map[string]bool{},
		pullRequestDetailDocumentCache:       map[pullRequestDetailDocumentCacheKey]detailDocument{},
		pullRequestConversationDocumentCache: map[pullRequestDetailDocumentCacheKey]browserConversationDocument{},
		pullRequestChangesRenderedRowsCache:  map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow{},
		pullRequestDiffCache:                 map[string]pullRequestDiffResult{},
		pullRequestDiffLoadInFlight:          map[string]bool{},
		issueDetailCache:                     map[string]issueDetailResult{},
		issueDetailLoadInFlight:              map[string]bool{},
		releaseDetailCache:                   map[string]releaseDetailResult{},
		releaseDetailLoadInFlight:            map[string]bool{},
		reviewDiffRenderCache:                map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{},
		pendingPullRequestReviewCache:        map[string]pendingPullRequestReviewState{},
		assignableUsersCache:                 map[string][]githubcli.PullRequestAuthor{},
		browserCollapsedSectionStates:        map[string]bool{},
		additionalPullRequestsLoadStarted:    map[PullRequestTab]bool{},
		additionalPullRequestsLoading:        map[PullRequestTab]bool{},
		additionalPullRequestsCounts:         map[PullRequestTab]pullRequestCountState{},
		notificationDoneStore:                noopNotificationDoneStore{},
		externalEditor:                       resolvedDeps.ExternalEditor,
		linkOpener:                           resolvedDeps.LinkOpener,
		detailImageStore:                     imageStore,
		detailImageManager:                   &protocolDetailImageManager{imageStore: imageStore, imageProtocol: imageProtocol, terminal: screenTerminalGraphicsTerminal{}},
		detailImageHTMLLoadInFlight:          map[string]bool{},
		detailImageHTMLLoadFailed:            map[string]bool{},
		detailImageLoadInFlight:              map[string]bool{},
		detailImageLoadFailed:                map[string]bool{},
		imageHTTPClient:                      http.DefaultClient,
		markdownRenderer:                     glamourMarkdownRenderer{imageStore: imageStore, imageProtocol: imageProtocol, terminalCellSize: screenTerminalCellSize{}},
		storyGenerator:                       commandReviewStoryGenerator{generator: story.NewGenerator(nil)},
		themePresetStore:                     resolvedDeps.ThemePresetStore,
		asyncRunner:                          goroutineAsyncRunner{},
		uiUpdater:                            queuedUIUpdater{},
		clipboardWriter:                      resolvedDeps.ClipboardWriter,
		detailViewState:                      newDetailViewState(),
		detailWrapWidth:                      defaultDetailWrapWidth,
		pullRequestSearches:                  appconfig.DefaultPullRequestSearches(),
		pendingListViewportPlacements:        map[string]viewportPlacement{},
	}
}

func resolveAppDeps(deps AppDeps) AppDeps {
	if deps.ExternalEditor == nil {
		deps.ExternalEditor = systemExternalEditor{}
	}
	if deps.LinkOpener == nil {
		deps.LinkOpener = newSystemLinkOpener(appconfig.ResolveLinksConfig(appconfig.LinksConfig{}).OpenCommand)
	}
	if deps.ThemePresetStore == nil {
		deps.ThemePresetStore = &defaultThemePresetStore{save: appconfig.SaveThemePresetDefault}
	}
	if deps.ClipboardWriter == nil {
		deps.ClipboardWriter = clip.NewSystemWriter()
	}
	return deps
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
