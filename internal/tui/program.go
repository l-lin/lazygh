package tui

import (
	"sync/atomic"
	"time"

	"github.com/jesseduffield/gocui"

	clip "github.com/l-lin/lazygh/internal/clipboard"
	appconfig "github.com/l-lin/lazygh/internal/config"
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
	viewActionsPopupChromeName   = "actions-popup-chrome"
	viewActionsPopupName         = "actions-popup"
	viewActionsPopupSearchName   = "actions-popup-search"
	viewTransientErrorPopupName  = "transient-error-popup"
	viewPullRequestBuildInfoName = "pull-request-build-info"

	defaultYankHighlightDuration = 240 * time.Millisecond
)

type programDeps struct {
	sessionQueries         SessionQueries
	pullRequestListQueries PullRequestListQueries
	notificationQueries    NotificationQueries
	detailQueries          DetailQueries
	pullRequestMutations   PullRequestMutations
	reviewMutations        ReviewMutations
	notificationMutations  NotificationMutations
	reactionMutations      ReactionMutations
	buildQueries           BuildQueries
	markdownHTMLRenderer   MarkdownHTMLRenderer
	authTokenProvider      AuthTokenProvider
	clipboardReader        ClipboardReader
	clipboardWriter        ClipboardWriter
	externalEditor         ExternalEditor
	linkOpener             LinkOpener
	markdownRenderer       MarkdownRenderer
	storyGenerator         reviewStoryGenerator
	themePresetStore       ThemePresetStore
}

type programStores struct {
	*sessionStore
	*persistentCacheStore
	*pullRequestListStore
	*notificationStore
	*detailStore
	*reviewStore
	*buildStore
	*statusStore
	*optimisticMutationCoordinator
	*imageLoadCoordinator
}

type listViewportRuntimeState struct {
	pendingPlacements map[string]viewportPlacement
}

type keybindingRuntimeState struct {
	registeredFingerprint string
}

type programViewRuntime struct {
	startupState        startupStateModel
	detailState         detailStateModel
	overlayState        overlayStateModel
	searchWidget        searchWidgetState
	actionsPopupWidget  actionsPopupWidgetState
	navigationState     navigationStateModel
	pastedPullRequests  pastedPullRequestTabState
	listViewportRuntime listViewportRuntimeState
	runtimeConfig       runtimeConfigState
}

type programShellRuntime struct {
	asyncRunner             asyncRunner
	uiUpdater               uiUpdater
	loadingSpinnerAnimating atomic.Bool
	gui                     *gocui.Gui
	keybindingRuntime       keybindingRuntimeState
	persistentCacheRuntime  persistentCacheRuntimeState
	timingState             timingStateModel
	manualRefreshState      manualRefreshStateModel
	refreshReadCache        refreshReadCacheState
}

type Program struct {
	model *Model
	programDeps
	programStores
	programViewRuntime
	programShellRuntime
}

func NewProgram() *Program {
	return NewProgramWithModel(defaultProgramModel())
}

func defaultProgramModel() *Model {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	return model
}

func NewProgramWithModel(model *Model) *Program {
	return NewProgramWithModelAndDeps(model, AppDeps{})
}

func NewProgramWithModelAndDeps(model *Model, deps AppDeps) *Program {
	if model == nil {
		model = defaultProgramModel()
	}

	resolvedDeps := resolveAppDeps(deps)
	imageStore := newMemoryDetailImageStore()
	imageProtocol := kittyImageProtocol{}
	persistence := &persistentCacheStore{}
	sessionState := newSessionStore()
	detailState := newDetailStore(persistence)
	reviewState := newReviewStore(persistence)
	imageCoordinator := newImageLoadCoordinator(imageStore, &protocolDetailImageManager{imageStore: imageStore, imageProtocol: imageProtocol, terminal: screenTerminalGraphicsTerminal{}})

	return &Program{
		model: model,
		programDeps: programDeps{
			sessionQueries:         resolvedDeps.SessionQueries,
			pullRequestListQueries: resolvedDeps.PullRequestList,
			notificationQueries:    resolvedDeps.NotificationQueries,
			detailQueries:          resolvedDeps.DetailQueries,
			pullRequestMutations:   resolvedDeps.PullRequestMutations,
			reviewMutations:        resolvedDeps.ReviewMutations,
			notificationMutations:  resolvedDeps.NotificationMutations,
			reactionMutations:      resolvedDeps.ReactionMutations,
			buildQueries:           resolvedDeps.BuildQueries,
			markdownHTMLRenderer:   resolvedDeps.MarkdownHTMLRenderer,
			authTokenProvider:      resolvedDeps.AuthTokenProvider,
			clipboardReader:        resolvedDeps.ClipboardReader,
			clipboardWriter:        resolvedDeps.ClipboardWriter,
			externalEditor:         resolvedDeps.ExternalEditor,
			linkOpener:             resolvedDeps.LinkOpener,
			markdownRenderer:       glamourMarkdownRenderer{imageStore: imageStore, imageProtocol: imageProtocol, terminalCellSize: screenTerminalCellSize{}},
			storyGenerator:         commandReviewStoryGenerator{generator: story.NewGenerator(nil)},
			themePresetStore:       resolvedDeps.ThemePresetStore,
		},
		programStores: programStores{
			sessionStore:                  sessionState,
			persistentCacheStore:          persistence,
			pullRequestListStore:          newPullRequestListStore(persistence),
			notificationStore:             newNotificationStore(persistence),
			detailStore:                   detailState,
			reviewStore:                   reviewState,
			buildStore:                    newBuildStore(),
			statusStore:                   newStatusStore(),
			optimisticMutationCoordinator: newOptimisticMutationCoordinator(),
			imageLoadCoordinator:          imageCoordinator,
		},
		programViewRuntime: programViewRuntime{
			startupState:        startupStateModel{},
			detailState:         detailStateModel{viewState: newDetailViewState(), wrapWidth: defaultDetailWrapWidth},
			actionsPopupWidget:  newActionsPopupWidgetState(),
			navigationState:     navigationStateModel{},
			listViewportRuntime: newListViewportRuntimeState(),
			runtimeConfig: runtimeConfigState{
				pullRequestSearches: appconfig.DefaultPullRequestSearches(),
				displayConfig:       appconfig.ResolveDisplayConfig(appconfig.DisplayConfig{}),
			},
		},
		programShellRuntime: programShellRuntime{
			asyncRunner:        goroutineAsyncRunner{},
			uiUpdater:          queuedUIUpdater{},
			timingState:        timingStateModel{now: time.Now, after: time.After, yankHighlightDuration: defaultYankHighlightDuration, transientErrorPopupDuration: defaultTransientErrorPopupDuration},
			manualRefreshState: manualRefreshStateModel{pullRequestListPending: map[PullRequestTab]bool{}, pullRequestDetailPending: map[string]bool{}, pullRequestDiffPending: map[string]bool{}},
		},
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
	if deps.ClipboardReader == nil {
		deps.ClipboardReader = clip.NewSystemReader()
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

	stopLoadingSpinner := program.startLoadingSpinner(gui)
	defer stopLoadingSpinner()
	if err := program.start(gui); err != nil {
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
	gui = program.captureGUI(gui)
	if !program.startupState.appStarted {
		_ = program.dispatch(gui, MsgAppStarted{})
	}
}
