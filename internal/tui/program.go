package tui

import (
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

type Program struct {
	model                  *Model
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
	startupState       startupStateModel
	detailState        detailStateModel
	clipboardReader    ClipboardReader
	clipboardWriter    ClipboardWriter
	overlayState       overlayStateModel
	searchWidget       searchWidgetState
	actionsPopupWidget actionsPopupWidgetState
	navigationState    navigationStateModel
	externalEditor     ExternalEditor
	linkOpener         LinkOpener
	markdownRenderer   MarkdownRenderer
	storyGenerator     reviewStoryGenerator
	asyncRunner        asyncRunner
	uiUpdater          uiUpdater
	gui                *gocui.Gui
	runtimeConfig      runtimeConfigState
	themePresetStore   ThemePresetStore
	timingState        timingStateModel
	manualRefreshState manualRefreshStateModel
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
		model:                         model,
		sessionQueries:                resolvedDeps.SessionQueries,
		pullRequestListQueries:        resolvedDeps.PullRequestList,
		notificationQueries:           resolvedDeps.NotificationQueries,
		detailQueries:                 resolvedDeps.DetailQueries,
		pullRequestMutations:          resolvedDeps.PullRequestMutations,
		reviewMutations:               resolvedDeps.ReviewMutations,
		notificationMutations:         resolvedDeps.NotificationMutations,
		reactionMutations:             resolvedDeps.ReactionMutations,
		buildQueries:                  resolvedDeps.BuildQueries,
		markdownHTMLRenderer:          resolvedDeps.MarkdownHTMLRenderer,
		authTokenProvider:             resolvedDeps.AuthTokenProvider,
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
		externalEditor:                resolvedDeps.ExternalEditor,
		linkOpener:                    resolvedDeps.LinkOpener,
		markdownRenderer:              glamourMarkdownRenderer{imageStore: imageStore, imageProtocol: imageProtocol, terminalCellSize: screenTerminalCellSize{}},
		storyGenerator:                commandReviewStoryGenerator{generator: story.NewGenerator(nil)},
		themePresetStore:              resolvedDeps.ThemePresetStore,
		asyncRunner:                   goroutineAsyncRunner{},
		uiUpdater:                     queuedUIUpdater{},
		clipboardReader:               resolvedDeps.ClipboardReader,
		clipboardWriter:               resolvedDeps.ClipboardWriter,
		startupState:                  startupStateModel{},
		detailState:                   detailStateModel{viewState: newDetailViewState(), wrapWidth: defaultDetailWrapWidth},
		actionsPopupWidget:            actionsPopupWidgetState{assigneePickerSearchDebounceDelay: defaultAssigneePickerSearchDebounceDelay},
		navigationState:               navigationStateModel{pendingListViewportPlacements: map[string]viewportPlacement{}},
		runtimeConfig:                 runtimeConfigState{pullRequestSearches: appconfig.DefaultPullRequestSearches()},
		timingState:                   timingStateModel{now: time.Now, after: time.After, yankHighlightDuration: defaultYankHighlightDuration, transientErrorPopupDuration: defaultTransientErrorPopupDuration},
		manualRefreshState:            manualRefreshStateModel{pullRequestListPending: map[PullRequestTab]bool{}, pullRequestDetailPending: map[string]bool{}, pullRequestDiffPending: map[string]bool{}},
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
	program.gui = gui
	if !program.startupState.appStarted {
		_ = program.dispatch(gui, MsgAppStarted{})
	}
}
