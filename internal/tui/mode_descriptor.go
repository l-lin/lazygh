package tui

type ModeDescriptor interface {
	Mode() ScreenMode
	SidebarSchema(*Program) SidebarSchema
	ScreenState(*Program) ScreenState
	MainViewResolver(*Program, ScreenState) MainViewResolver
}

type SidebarSchema struct {
	Views []ViewState
}

func (schema SidebarSchema) ViewByNumber(number int) (ViewState, bool) {
	for _, view := range schema.Views {
		if view.Number == number {
			return view, true
		}
	}
	return ViewState{}, false
}

func (schema SidebarSchema) ViewNumbers() []int {
	numbers := make([]int, 0, len(schema.Views))
	for _, view := range schema.Views {
		numbers = append(numbers, view.Number)
	}
	return numbers
}

type BrowserMode struct{}

type ReviewMode struct{}

type StoryReviewMode struct{}

func (BrowserMode) Mode() ScreenMode {
	return ScreenModeBrowser
}

func (BrowserMode) SidebarSchema(program *Program) SidebarSchema {
	return SidebarSchema{Views: copyViewStates(program.model.ScreenState().SidePanel.Views)}
}

func (BrowserMode) ScreenState(program *Program) ScreenState {
	state := program.model.ScreenState()
	if program.browserShowsPullRequestDetailTabs() {
		state = state.WithViewTabs(mainPanelViewNumber, int(program.detailState.activeTab), program.detailScreenTabs())
	}
	return state
}

func (BrowserMode) MainViewResolver(program *Program, state ScreenState) MainViewResolver {
	resolver := state.MainViewResolver()
	if resolver.SourceView.Focus != FocusNotificationsView {
		return resolver
	}
	notification, ok := program.model.SelectedNotification()
	if !ok {
		return resolver
	}
	if _, ok = notification.PullRequestSummary(); ok {
		resolver.ContentKind = MainContentKindPullRequestDetail
	}
	return resolver
}

func (ReviewMode) Mode() ScreenMode {
	return ScreenModeReview
}

func (ReviewMode) SidebarSchema(program *Program) SidebarSchema {
	state := newReviewScreenStateWithSideFocus(ScreenModeReview, program.model.Focus(), program.model.currentSideFocus())
	return SidebarSchema{Views: copyViewStates(state.SidePanel.Views)}
}

func (ReviewMode) ScreenState(program *Program) ScreenState {
	return newReviewScreenStateWithSideFocus(ScreenModeReview, program.model.Focus(), program.model.currentSideFocus())
}

func (ReviewMode) MainViewResolver(_ *Program, state ScreenState) MainViewResolver {
	return state.MainViewResolver()
}

func (StoryReviewMode) Mode() ScreenMode {
	return ScreenModeStoryReview
}

func (StoryReviewMode) SidebarSchema(program *Program) SidebarSchema {
	state := newReviewScreenStateWithSideFocus(ScreenModeStoryReview, program.model.Focus(), program.model.currentSideFocus())
	return SidebarSchema{Views: copyViewStates(state.SidePanel.Views)}
}

func (StoryReviewMode) ScreenState(program *Program) ScreenState {
	return newReviewScreenStateWithSideFocus(ScreenModeStoryReview, program.model.Focus(), program.model.currentSideFocus())
}

func (StoryReviewMode) MainViewResolver(program *Program, state ScreenState) MainViewResolver {
	resolver := state.MainViewResolver()
	if resolver.SourceView.Focus != FocusPullRequestsView {
		return resolver
	}
	if _, ok := program.selectedReviewSessionStoryChapter(); ok {
		resolver.ContentKind = MainContentKindStoryChapter
	}
	return resolver
}

func (program *Program) modeDescriptor() ModeDescriptor {
	if program.navigationState.reviewSession.active {
		if program.navigationState.reviewSession.mode == reviewSessionModeStory {
			return StoryReviewMode{}
		}
		return ReviewMode{}
	}
	return BrowserMode{}
}

func (program *Program) mainViewResolver() MainViewResolver {
	state := program.screenState()
	return program.modeDescriptor().MainViewResolver(program, state)
}

type ActionContext struct {
	Mode            ScreenMode
	ActiveView      ViewState
	MainView        MainViewResolver
	ActiveDetailTab DetailTab
}

func (context ActionContext) IsReviewContext() bool {
	return context.Mode == ScreenModeReview || context.Mode == ScreenModeStoryReview
}

func (context ActionContext) IsNotificationContext() bool {
	if context.Mode != ScreenModeBrowser {
		return false
	}
	if context.ActiveView.Focus == FocusNotificationsView {
		return true
	}
	if context.ActiveView.Focus != FocusDetailView {
		return false
	}
	return context.MainView.ContentKind == MainContentKindNotificationDetail
}

func (context ActionContext) IsPullRequestContext() bool {
	if context.IsReviewContext() {
		return true
	}
	switch context.ActiveView.Focus {
	case FocusPullRequestsView:
		return true
	case FocusDetailView:
		return context.MainView.ContentKind == MainContentKindPullRequestDetail
	default:
		return false
	}
}

func (context ActionContext) ShowsPullRequestDescription() bool {
	if context.MainView.ContentKind == MainContentKindReviewDescription {
		return true
	}
	return context.MainView.ContentKind == MainContentKindPullRequestDetail && context.ActiveDetailTab == DescriptionDetailTab
}

func (program *Program) actionContext() ActionContext {
	state := program.screenState()
	return ActionContext{
		Mode:            state.Mode,
		ActiveView:      state.ActiveView(),
		MainView:        program.mainViewResolver(),
		ActiveDetailTab: program.detailState.activeTab,
	}
}

type DetailInputMode int

const (
	DetailInputModeNone DetailInputMode = iota
	DetailInputModePullRequestComment
	DetailInputModeBrowserChangesInlineComment
	DetailInputModeReviewInlineComment
)

type InputContext struct {
	ActionContext
	SearchUsesReviewTree bool
	DetailInputMode      DetailInputMode
}

func (program *Program) inputContext() InputContext {
	actionContext := program.actionContext()
	context := InputContext{ActionContext: actionContext}
	context.SearchUsesReviewTree = actionContext.IsReviewContext() && actionContext.ActiveView.Focus == FocusPullRequestsView
	switch {
	case actionContext.IsReviewContext() && actionContext.ActiveView.Focus == FocusDetailView:
		context.DetailInputMode = DetailInputModeReviewInlineComment
	case actionContext.MainView.ContentKind == MainContentKindPullRequestDetail && actionContext.ActiveView.Focus == FocusDetailView && program.browserChangesInlineCommentShortcutActive():
		context.DetailInputMode = DetailInputModeBrowserChangesInlineComment
	case actionContext.MainView.ContentKind == MainContentKindPullRequestDetail && actionContext.ActiveView.Focus == FocusDetailView:
		context.DetailInputMode = DetailInputModePullRequestComment
	default:
		context.DetailInputMode = DetailInputModeNone
	}
	return context
}
