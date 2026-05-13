package tui

type ScreenMode int

const (
	ScreenModeBrowser ScreenMode = iota
	ScreenModeReview
	ScreenModeStoryReview
)

type ActivePanel int

const (
	ActivePanelSide ActivePanel = iota
	ActivePanelMain
)

type OverlayKind int

const (
	OverlayKindHelp OverlayKind = iota
	OverlayKindSearch
	OverlayKindActionsPopup
	OverlayKindActionsPopupSearch
	OverlayKindModalEditor
	OverlayKindBuildInfo
)

type KeyHintContext int

const (
	KeyHintContextNone KeyHintContext = iota
	KeyHintContextSidePanel
	KeyHintContextMainPanel
	KeyHintContextSearch
	KeyHintContextActionsPopup
	KeyHintContextActionsPopupSearch
	KeyHintContextModalEditor
	KeyHintContextBuildInfo
)

const (
	mainPanelViewNumber              = 0
	sidePanelUserViewNumber          = 1
	sidePanelPullRequestsViewNumber  = 2
	sidePanelNotificationsViewNumber = 3
)

type ScreenState struct {
	Mode        ScreenMode
	ActivePanel ActivePanel
	MainPanel   PanelState
	SidePanel   PanelState
	Overlays    []OverlayState
}

type PanelState struct {
	Views            []ViewState
	ActiveViewNumber int
}

type ViewState struct {
	Number    int
	Focus     Focus
	Tabs      []TabState
	ActiveTab int
}

type TabState struct {
	Label string
}

type OverlayState struct {
	Kind OverlayKind
}

type MainContentKind int

const (
	MainContentKindUserDetail MainContentKind = iota
	MainContentKindPullRequestDetail
	MainContentKindNotificationDetail
	MainContentKindReviewDescription
	MainContentKindReviewDiff
	MainContentKindStoryChapter
)

type MainViewResolver struct {
	SourceView  ViewState
	ContentKind MainContentKind
}

func newBrowserScreenState(activeFocus Focus, activePullRequestTab PullRequestTab, mainTabs []TabState, pullRequestTabs []TabState) ScreenState {
	return newBrowserScreenStateWithSideFocus(activeFocus, defaultBrowserSideFocus(activeFocus), activePullRequestTab, mainTabs, pullRequestTabs)
}

func newBrowserScreenStateWithSideFocus(activeFocus Focus, activeSideFocus Focus, activePullRequestTab PullRequestTab, mainTabs []TabState, pullRequestTabs []TabState) ScreenState {
	state := ScreenState{
		Mode:        ScreenModeBrowser,
		ActivePanel: activePanelForFocus(activeFocus),
		MainPanel: PanelState{
			Views:            []ViewState{{Number: mainPanelViewNumber, Focus: FocusDetailView, Tabs: copyTabStates(mainTabs), ActiveTab: clampScreenTabIndex(0, len(mainTabs))}},
			ActiveViewNumber: mainPanelViewNumber,
		},
		SidePanel: PanelState{
			Views: []ViewState{
				{Number: sidePanelUserViewNumber, Focus: FocusUserView},
				{Number: sidePanelPullRequestsViewNumber, Focus: FocusPullRequestsView, Tabs: copyTabStates(pullRequestTabs), ActiveTab: clampScreenTabIndex(int(activePullRequestTab), len(pullRequestTabs))},
				{Number: sidePanelNotificationsViewNumber, Focus: FocusNotificationsView},
			},
			ActiveViewNumber: browserSideViewNumber(activeSideFocus),
		},
	}
	if state.ActivePanel == ActivePanelSide {
		state.SidePanel.ActiveViewNumber = browserSideViewNumber(activeFocus)
	}
	return state
}

func newReviewScreenStateWithSideFocus(mode ScreenMode, activeFocus Focus, activeSideFocus Focus) ScreenState {
	state := ScreenState{
		Mode:        mode,
		ActivePanel: activePanelForFocus(activeFocus),
		MainPanel: PanelState{
			Views:            []ViewState{{Number: mainPanelViewNumber, Focus: FocusDetailView}},
			ActiveViewNumber: mainPanelViewNumber,
		},
		SidePanel: PanelState{
			Views: []ViewState{
				{Number: sidePanelUserViewNumber, Focus: FocusUserView},
				{Number: sidePanelPullRequestsViewNumber, Focus: FocusPullRequestsView},
			},
			ActiveViewNumber: reviewSideViewNumber(activeSideFocus),
		},
	}
	if state.ActivePanel == ActivePanelSide {
		state.SidePanel.ActiveViewNumber = reviewSideViewNumber(activeFocus)
	}
	return state
}

func defaultBrowserSideFocus(activeFocus Focus) Focus {
	switch activeFocus {
	case FocusPullRequestsView, FocusNotificationsView:
		return activeFocus
	default:
		return FocusUserView
	}
}

func activePanelForFocus(activeFocus Focus) ActivePanel {
	if activeFocus == FocusDetailView {
		return ActivePanelMain
	}
	return ActivePanelSide
}

func browserSideViewNumber(focus Focus) int {
	switch focus {
	case FocusPullRequestsView:
		return sidePanelPullRequestsViewNumber
	case FocusNotificationsView:
		return sidePanelNotificationsViewNumber
	default:
		return sidePanelUserViewNumber
	}
}

func reviewSideViewNumber(focus Focus) int {
	if focus == FocusUserView {
		return sidePanelUserViewNumber
	}
	return sidePanelPullRequestsViewNumber
}

func clampScreenTabIndex(index int, count int) int {
	if count == 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= count {
		return count - 1
	}
	return index
}

func (state ScreenState) ActiveView() ViewState {
	if state.ActivePanel == ActivePanelMain {
		return state.MainPanel.activeView()
	}
	return state.SidePanel.activeView()
}

func (state ScreenState) ActiveSideView() ViewState {
	return state.SidePanel.activeView()
}

func (state ScreenState) ViewByNumber(number int) (ViewState, bool) {
	if view, ok := state.MainPanel.viewByNumber(number); ok {
		return view, true
	}
	return state.SidePanel.viewByNumber(number)
}

func (panel PanelState) activeView() ViewState {
	if view, ok := panel.viewByNumber(panel.ActiveViewNumber); ok {
		return view
	}
	if len(panel.Views) == 0 {
		return ViewState{}
	}
	return panel.Views[0]
}

func (panel PanelState) viewByNumber(number int) (ViewState, bool) {
	for _, view := range panel.Views {
		if view.Number == number {
			return view, true
		}
	}
	return ViewState{}, false
}

func (state ScreenState) FocusViewNumber(number int) ScreenState {
	if _, ok := state.ViewByNumber(number); !ok {
		return state
	}
	if number == mainPanelViewNumber {
		state.ActivePanel = ActivePanelMain
		return state
	}
	state.ActivePanel = ActivePanelSide
	state.SidePanel.ActiveViewNumber = number
	return state
}

func (state ScreenState) NextSideView() ScreenState {
	return state.cycleSideView(1)
}

func (state ScreenState) PreviousSideView() ScreenState {
	return state.cycleSideView(-1)
}

func (state ScreenState) cycleSideView(delta int) ScreenState {
	if state.ActivePanel != ActivePanelSide || len(state.SidePanel.Views) <= 1 {
		return state
	}

	currentIndex := 0
	for index, view := range state.SidePanel.Views {
		if view.Number == state.SidePanel.ActiveViewNumber {
			currentIndex = index
			break
		}
	}
	currentIndex = (currentIndex + delta + len(state.SidePanel.Views)) % len(state.SidePanel.Views)
	state.SidePanel.ActiveViewNumber = state.SidePanel.Views[currentIndex].Number
	return state
}

func (state ScreenState) NextTab() ScreenState {
	return state.cycleActiveViewTab(1)
}

func (state ScreenState) PreviousTab() ScreenState {
	return state.cycleActiveViewTab(-1)
}

func (state ScreenState) cycleActiveViewTab(delta int) ScreenState {
	activeView := state.ActiveView()
	if len(activeView.Tabs) <= 1 {
		return state
	}
	activeView.ActiveTab = (activeView.ActiveTab + delta + len(activeView.Tabs)) % len(activeView.Tabs)
	return state.WithViewTabs(activeView.Number, activeView.ActiveTab, activeView.Tabs)
}

func (state ScreenState) WithViewTabs(viewNumber int, activeTab int, tabs []TabState) ScreenState {
	updated, ok := state.MainPanel.withViewTabs(viewNumber, activeTab, tabs)
	if ok {
		state.MainPanel = updated
		return state
	}
	updated, ok = state.SidePanel.withViewTabs(viewNumber, activeTab, tabs)
	if ok {
		state.SidePanel = updated
	}
	return state
}

func (panel PanelState) withViewTabs(viewNumber int, activeTab int, tabs []TabState) (PanelState, bool) {
	for index, view := range panel.Views {
		if view.Number != viewNumber {
			continue
		}
		panel.Views = copyViewStates(panel.Views)
		panel.Views[index].Tabs = copyTabStates(tabs)
		panel.Views[index].ActiveTab = clampScreenTabIndex(activeTab, len(tabs))
		return panel, true
	}
	return panel, false
}

func (state ScreenState) WithOverlay(overlay OverlayState) ScreenState {
	filtered := make([]OverlayState, 0, len(state.Overlays)+1)
	for _, actual := range state.Overlays {
		if actual.Kind == overlay.Kind {
			continue
		}
		filtered = append(filtered, actual)
	}
	state.Overlays = append(filtered, overlay)
	return state
}

func (state ScreenState) MainViewResolver() MainViewResolver {
	sourceView := state.ActiveSideView()
	return MainViewResolver{SourceView: sourceView, ContentKind: defaultMainContentKind(state.Mode, sourceView)}
}

func defaultMainContentKind(mode ScreenMode, sourceView ViewState) MainContentKind {
	switch mode {
	case ScreenModeReview:
		if sourceView.Focus == FocusUserView {
			return MainContentKindReviewDescription
		}
		return MainContentKindReviewDiff
	case ScreenModeStoryReview:
		if sourceView.Focus == FocusUserView {
			return MainContentKindReviewDescription
		}
		return MainContentKindReviewDiff
	default:
		switch sourceView.Focus {
		case FocusPullRequestsView:
			return MainContentKindPullRequestDetail
		case FocusNotificationsView:
			return MainContentKindNotificationDetail
		default:
			return MainContentKindUserDetail
		}
	}
}

func (state ScreenState) AllowsMainCursor() bool {
	return state.ActivePanel == ActivePanelMain && state.ActiveView().Number == mainPanelViewNumber
}

func (state ScreenState) KeyHintContext() KeyHintContext {
	if overlay, ok := state.activeOverlay(); ok {
		switch overlay.Kind {
		case OverlayKindSearch:
			return KeyHintContextSearch
		case OverlayKindActionsPopup:
			return KeyHintContextActionsPopup
		case OverlayKindActionsPopupSearch:
			return KeyHintContextActionsPopupSearch
		case OverlayKindModalEditor:
			return KeyHintContextModalEditor
		case OverlayKindBuildInfo:
			return KeyHintContextBuildInfo
		default:
			return KeyHintContextNone
		}
	}
	if state.ActivePanel == ActivePanelMain {
		return KeyHintContextMainPanel
	}
	return KeyHintContextSidePanel
}

func (state ScreenState) activeOverlay() (OverlayState, bool) {
	if len(state.Overlays) == 0 {
		return OverlayState{}, false
	}
	return state.Overlays[len(state.Overlays)-1], true
}

func copyViewStates(views []ViewState) []ViewState {
	copied := make([]ViewState, 0, len(views))
	for _, view := range views {
		copied = append(copied, ViewState{Number: view.Number, Focus: view.Focus, Tabs: copyTabStates(view.Tabs), ActiveTab: view.ActiveTab})
	}
	return copied
}

func copyTabStates(tabs []TabState) []TabState {
	if len(tabs) == 0 {
		return nil
	}
	copied := make([]TabState, 0, len(tabs))
	for _, tab := range tabs {
		copied = append(copied, TabState{Label: tab.Label})
	}
	return copied
}
