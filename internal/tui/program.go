package tui

import (
	"github.com/jesseduffield/gocui"

	clip "codeberg.org/l-lin/lazygh/internal/clipboard"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName       = "detail"
	viewUserName         = "user"
	viewPullRequestsName = "pull-requests"
	viewSearchName       = "search"
)

type GitHubLoader interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
	ListMyPullRequests() ([]githubcli.PullRequest, error)
	ListRequestedPullRequests() ([]githubcli.PullRequest, error)
	GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error)
	CommentOnPullRequest(repository string, number int, body string) error
}

type Program struct {
	model                            *Model
	githubLoader                     GitHubLoader
	connectedUserLoadStarted         bool
	myPullRequestsLoadStarted        bool
	requestedPullRequestsLoadStarted bool
	myPullRequestsCount              int
	myPullRequestsCountKnown         bool
	requestedPullRequestsCount       int
	requestedPullRequestsCountKnown  bool
	pullRequestDetailCache           map[string]pullRequestDetailResult
	pullRequestDetailLoadInFlight    map[string]bool
	detailWrapWidth                  int
	activeDetailTab                  DetailTab
	lastDetailIdentity               string
	clipboardWriter                  clip.Writer
	feedbackFocus                    Focus
	feedbackMessage                  string
	helpVisible                      bool
	searchEditor                     *lineEditor
	modalEditor                      *modalEditorState
	markdownRenderer                 MarkdownRenderer
	asyncRunner                      asyncRunner
	uiUpdater                        uiUpdater
}

type keybindingSpec struct {
	viewName string
	key      any
	handler  func(*gocui.Gui, *gocui.View) error
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
		markdownRenderer:              glamourMarkdownRenderer{},
		asyncRunner:                   goroutineAsyncRunner{},
		uiUpdater:                     queuedUIUpdater{},
		clipboardWriter:               clip.NewSystemWriter(),
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

func (program *Program) setKeybindings(gui *gocui.Gui) error {
	for _, binding := range program.keybindingSpecs() {
		if err := gui.SetKeybinding(binding.viewName, binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}

	return nil
}

func (program *Program) keybindingSpecs() []keybindingSpec {
	return []keybindingSpec{
		{viewName: "", key: gocui.KeyCtrlC, handler: program.quit},
		{viewName: "", key: gocui.KeyTab, handler: program.nextSideView},
		{viewName: "", key: gocui.KeyBacktab, handler: program.previousSideView},
		{viewName: viewUserName, key: '?', handler: program.toggleHelp},
		{viewName: viewPullRequestsName, key: '?', handler: program.toggleHelp},
		{viewName: viewDetailName, key: '?', handler: program.toggleHelp},
		{viewName: viewUserName, key: 'l', handler: program.nextSideView},
		{viewName: viewPullRequestsName, key: 'l', handler: program.nextSideView},
		{viewName: viewDetailName, key: 'l', handler: program.nextSideView},
		{viewName: viewUserName, key: 'h', handler: program.previousSideView},
		{viewName: viewPullRequestsName, key: 'h', handler: program.previousSideView},
		{viewName: viewDetailName, key: 'h', handler: program.previousSideView},
		{viewName: viewUserName, key: '0', handler: program.focusDetailView},
		{viewName: viewPullRequestsName, key: '0', handler: program.focusDetailView},
		{viewName: viewDetailName, key: '0', handler: program.focusDetailView},
		{viewName: viewUserName, key: '1', handler: program.focusUserView},
		{viewName: viewPullRequestsName, key: '1', handler: program.focusUserView},
		{viewName: viewDetailName, key: '1', handler: program.focusUserView},
		{viewName: viewUserName, key: '2', handler: program.focusPullRequestsView},
		{viewName: viewPullRequestsName, key: '2', handler: program.focusPullRequestsView},
		{viewName: viewDetailName, key: '2', handler: program.focusPullRequestsView},
		{viewName: viewUserName, key: '/', handler: program.openSearch},
		{viewName: viewUserName, key: 'j', handler: program.moveSelectionDown},
		{viewName: viewUserName, key: 'k', handler: program.moveSelectionUp},
		{viewName: viewUserName, key: gocui.KeyCtrlD, handler: program.pageDown},
		{viewName: viewUserName, key: gocui.KeyCtrlU, handler: program.pageUp},
		{viewName: viewUserName, key: gocui.KeyEnter, handler: program.openDetail},
		{viewName: viewUserName, key: 'y', handler: program.copyPullRequestURL},
		{viewName: viewPullRequestsName, key: '/', handler: program.openSearch},
		{viewName: viewPullRequestsName, key: 'j', handler: program.moveSelectionDown},
		{viewName: viewPullRequestsName, key: 'k', handler: program.moveSelectionUp},
		{viewName: viewPullRequestsName, key: gocui.KeyCtrlD, handler: program.pageDown},
		{viewName: viewPullRequestsName, key: gocui.KeyCtrlU, handler: program.pageUp},
		{viewName: viewPullRequestsName, key: '[', handler: program.previousPullRequestTab},
		{viewName: viewPullRequestsName, key: ']', handler: program.nextPullRequestTab},
		{viewName: viewPullRequestsName, key: gocui.KeyEnter, handler: program.openDetail},
		{viewName: viewPullRequestsName, key: 'y', handler: program.copyPullRequestURL},
		{viewName: viewPullRequestsName, key: 'c', handler: program.openPullRequestCommentComposer},
		{viewName: viewDetailName, key: '/', handler: program.openSearch},
		{viewName: viewDetailName, key: 'j', handler: program.moveSelectionDown},
		{viewName: viewDetailName, key: 'k', handler: program.moveSelectionUp},
		{viewName: viewDetailName, key: gocui.KeyCtrlD, handler: program.pageDown},
		{viewName: viewDetailName, key: gocui.KeyCtrlU, handler: program.pageUp},
		{viewName: viewDetailName, key: '[', handler: program.previousDetailTab},
		{viewName: viewDetailName, key: ']', handler: program.nextDetailTab},
		{viewName: viewDetailName, key: 'y', handler: program.copyPullRequestURL},
		{viewName: viewDetailName, key: 'c', handler: program.openPullRequestCommentComposer},
		{viewName: viewDetailName, key: gocui.KeyEsc, handler: program.closeDetail},
		{viewName: viewDetailName, key: gocui.KeyCtrlLsqBracket, handler: program.closeDetail},
		{viewName: viewSearchName, key: gocui.KeyEnter, handler: program.submitSearch},
		{viewName: viewSearchName, key: gocui.KeyCtrlJ, handler: program.submitSearch},
		{viewName: viewSearchName, key: gocui.KeyEsc, handler: program.cancelSearch},
		{viewName: viewSearchName, key: gocui.KeyCtrlLsqBracket, handler: program.cancelSearch},
		{viewName: viewModalEditorName, key: gocui.KeyAltEnter, handler: program.submitModalEditor},
		{viewName: viewModalEditorName, key: gocui.KeyEsc, handler: program.closeModalEditor},
		{viewName: viewModalEditorName, key: gocui.KeyCtrlLsqBracket, handler: program.closeModalEditor},
		{viewName: viewHelpName, key: gocui.KeyEsc, handler: program.closeHelp},
		{viewName: viewHelpName, key: gocui.KeyCtrlLsqBracket, handler: program.closeHelp},
	}
}

func (program *Program) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (program *Program) nextSideView(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}

	program.model.NextSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) previousSideView(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}

	program.model.PreviousSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) moveSelectionDown(_ *gocui.Gui, view *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailDownLine(view)
		return nil
	}

	program.model.MoveSelectionDown()
	return nil
}

func (program *Program) moveSelectionUp(_ *gocui.Gui, view *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailUpLine(view)
		return nil
	}

	program.model.MoveSelectionUp()
	return nil
}

func (program *Program) pageDown(_ *gocui.Gui, view *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailDown(view)
		return nil
	}

	program.model.PageDown(viewPageSize(view))
	return nil
}

func (program *Program) pageUp(_ *gocui.Gui, view *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailUp(view)
		return nil
	}

	program.model.PageUp(viewPageSize(view))
	return nil
}

func (program *Program) nextPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}

	program.model.NextPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) previousPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}

	program.model.PreviousPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) focusDetailView(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}

	program.model.FocusDetailView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusUserView(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}

	program.model.FocusUserView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}

	program.model.FocusPullRequestsView()
	return program.syncCurrentView(gui)
}

func (program *Program) openDetail(gui *gocui.Gui, _ *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}

	program.model.OpenDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) closeDetail(gui *gocui.Gui, _ *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}

	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}

	program.model.StartSearch()
	program.searchEditor = newLineEditor(program.model.SearchDraft())
	return program.layout(gui)
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.model.SubmitSearch()
	return program.closeSearch(gui)
}

func (program *Program) cancelSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.model.CancelSearch()
	return program.closeSearch(gui)
}

func (program *Program) closeSearch(gui *gocui.Gui) error {
	program.searchEditor = nil

	actualErr := gui.DeleteView(viewSearchName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}

	return program.refreshViews(gui)
}

func (program *Program) syncCurrentView(gui *gocui.Gui) error {
	gui.Cursor = program.model.SearchActive() || program.modalEditorVisible()
	if program.helpVisible {
		gui.Cursor = false
		_, err := gui.SetCurrentView(viewHelpName)
		if isUnknownViewError(err) {
			return nil
		}

		return err
	}

	_, err := gui.SetCurrentView(program.currentViewName())
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || program.connectedUserLoadStarted {
		return
	}

	program.connectedUserLoadStarted = true
	program.asyncRunner.Go(func() {
		program.loadConnectedUser(gui)
	})
}

func (program *Program) maybeLoadMyPullRequests(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || program.myPullRequestsLoadStarted || program.model.ActivePullRequestTab() != MyPullRequestsTab {
		return
	}

	program.myPullRequestsLoadStarted = true
	program.asyncRunner.Go(func() {
		program.loadMyPullRequests(gui)
	})
}

func (program *Program) maybeLoadRequestedPullRequests(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || program.requestedPullRequestsLoadStarted || program.model.ActivePullRequestTab() != RequestedPullRequestsTab {
		return
	}

	program.requestedPullRequestsLoadStarted = true
	program.asyncRunner.Go(func() {
		program.loadRequestedPullRequests(gui)
	})
}

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil {
		return
	}

	switch program.model.ActivePullRequestTab() {
	case RequestedPullRequestsTab:
		program.requestedPullRequestsLoadStarted = true
		program.asyncRunner.Go(func() {
			program.loadRequestedPullRequests(gui)
		})
	default:
		program.myPullRequestsLoadStarted = true
		program.asyncRunner.Go(func() {
			program.loadMyPullRequests(gui)
		})
	}
}

func (program *Program) loadConnectedUser(gui *gocui.Gui) {
	user, err := program.githubLoader.GetConnectedUser()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.model.SetUsers([]Item{connectedUserStateItem(user, err)})
		return program.refreshViews(gui)
	})
}

func (program *Program) loadMyPullRequests(gui *gocui.Gui) {
	pullRequests, err := program.githubLoader.ListMyPullRequests()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.myPullRequestsCount = len(pullRequests)
		program.myPullRequestsCountKnown = err == nil
		program.model.SetPullRequestRows(MyPullRequestsTab, myPullRequestsStateRows(pullRequests, err))
		return program.refreshViews(gui)
	})
}

func (program *Program) loadRequestedPullRequests(gui *gocui.Gui) {
	pullRequests, err := program.githubLoader.ListRequestedPullRequests()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.requestedPullRequestsCount = len(pullRequests)
		program.requestedPullRequestsCountKnown = err == nil
		program.model.SetPullRequestRows(RequestedPullRequestsTab, requestedPullRequestsStateRows(pullRequests, err))
		return program.refreshViews(gui)
	})
}

func (program *Program) refreshViews(gui *gocui.Gui) error {
	program.maybeLoadSelectedPullRequestDetail(gui)

	userView, err := gui.View(viewUserName)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	if err == nil {
		program.configureUserView(userView)
		program.renderUserView(userView)
	}

	pullRequestsView, err := gui.View(viewPullRequestsName)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	if err == nil {
		program.configurePullRequestsView(pullRequestsView)
		program.renderPullRequestsView(pullRequestsView)
	}

	detailView, err := gui.View(viewDetailName)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	if err == nil {
		program.configureDetailView(detailView)
		program.renderDetailView(detailView)
	}

	if program.helpVisible {
		helpView, err := gui.View(viewHelpName)
		if err != nil && !isUnknownViewError(err) {
			return err
		}
		if err == nil {
			program.configureHelpView(helpView)
			program.renderHelpView(helpView)
			_, err = gui.SetViewOnTop(viewHelpName)
			if err != nil && !isUnknownViewError(err) {
				return err
			}
		}
	}

	if program.model.SearchActive() {
		searchView, err := gui.View(viewSearchName)
		if err != nil && !isUnknownViewError(err) {
			return err
		}
		if err == nil {
			program.configureSearchView(searchView)
			program.renderSearchView(searchView)
			_, err = gui.SetViewOnTop(viewSearchName)
			if err != nil && !isUnknownViewError(err) {
				return err
			}
		}
	}

	if program.modalEditorVisible() {
		modalView, err := gui.View(viewModalEditorName)
		if err != nil && !isUnknownViewError(err) {
			return err
		}
		if err == nil {
			program.configureModalEditorView(modalView)
			program.renderModalEditorView(modalView)
			_, err = gui.SetViewOnTop(viewModalEditorName)
			if err != nil && !isUnknownViewError(err) {
				return err
			}
		}
	}

	return program.syncCurrentView(gui)
}

func (program *Program) currentViewName() string {
	if program.modalEditorVisible() {
		return viewModalEditorName
	}
	if program.model.SearchActive() {
		return viewSearchName
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}

func (program *Program) toggleHelp(gui *gocui.Gui, _ *gocui.View) error {
	if program.model.SearchActive() {
		return nil
	}

	program.helpVisible = !program.helpVisible
	if !program.helpVisible {
		return program.closeHelp(gui, nil)
	}

	return program.layout(gui)
}

func (program *Program) closeHelp(gui *gocui.Gui, _ *gocui.View) error {
	program.helpVisible = false
	actualErr := gui.DeleteView(viewHelpName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}

	return program.syncCurrentView(gui)
}

func (program *Program) scrollDetailDown(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollDown(pageDelta(view.InnerHeight()))
}

func (program *Program) scrollDetailUp(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollUp(pageDelta(view.InnerHeight()))
}

func (program *Program) scrollDetailDownLine(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollDown(1)
}

func (program *Program) scrollDetailUpLine(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollUp(1)
}

func viewPageSize(view *gocui.View) int {
	if view == nil {
		return 1
	}

	return view.InnerHeight()
}
