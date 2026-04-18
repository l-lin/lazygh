package tui

import (
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName       = "detail"
	viewUserName         = "user"
	viewPullRequestsName = "pull-requests"
)

type ConnectedUserLoader interface {
	GetConnectedUser() (githubcli.ConnectedUser, error)
}

type Program struct {
	model                    *Model
	connectedUserLoader      ConnectedUserLoader
	connectedUserLoadStarted bool
}

type keybindingSpec struct {
	viewName string
	key      any
	handler  func(*gocui.Gui, *gocui.View) error
}

func NewProgram(connectedUserLoader ConnectedUserLoader) *Program {
	return NewProgramWithModelAndLoader(NewModel(DefaultSeedData()), connectedUserLoader)
}

func NewProgramWithModel(model *Model) *Program {
	return NewProgramWithModelAndLoader(model, nil)
}

func NewProgramWithModelAndLoader(model *Model, connectedUserLoader ConnectedUserLoader) *Program {
	if model == nil {
		model = NewModel(DefaultSeedData())
	}

	return &Program{
		model:               model,
		connectedUserLoader: connectedUserLoader,
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
		{viewName: "", key: 'l', handler: program.nextSideView},
		{viewName: "", key: 'h', handler: program.previousSideView},
		{viewName: "", key: '0', handler: program.focusDetailView},
		{viewName: "", key: '1', handler: program.focusUserView},
		{viewName: "", key: '2', handler: program.focusPullRequestsView},
		{viewName: viewUserName, key: 'j', handler: program.moveSelectionDown},
		{viewName: viewUserName, key: 'k', handler: program.moveSelectionUp},
		{viewName: viewUserName, key: gocui.KeyEnter, handler: program.openDetail},
		{viewName: viewPullRequestsName, key: 'j', handler: program.moveSelectionDown},
		{viewName: viewPullRequestsName, key: 'k', handler: program.moveSelectionUp},
		{viewName: viewPullRequestsName, key: '[', handler: program.previousPullRequestTab},
		{viewName: viewPullRequestsName, key: ']', handler: program.nextPullRequestTab},
		{viewName: viewPullRequestsName, key: gocui.KeyEnter, handler: program.openDetail},
		{viewName: viewDetailName, key: gocui.KeyEsc, handler: program.closeDetail},
		{viewName: viewDetailName, key: gocui.KeyCtrlLsqBracket, handler: program.closeDetail},
		{viewName: viewDetailName, key: gocui.KeyCtrl3, handler: program.closeDetail},
		// Some terminals collapse `ctrl+[` into `[` instead of exposing a dedicated control key.
		{viewName: viewDetailName, key: '[', handler: program.closeDetail},
	}
}

func (program *Program) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (program *Program) nextSideView(gui *gocui.Gui, _ *gocui.View) error {
	program.model.NextSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) previousSideView(gui *gocui.Gui, _ *gocui.View) error {
	program.model.PreviousSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) moveSelectionDown(_ *gocui.Gui, _ *gocui.View) error {
	program.model.MoveSelectionDown()
	return nil
}

func (program *Program) moveSelectionUp(_ *gocui.Gui, _ *gocui.View) error {
	program.model.MoveSelectionUp()
	return nil
}

func (program *Program) nextPullRequestTab(_ *gocui.Gui, _ *gocui.View) error {
	program.model.NextPullRequestTab()
	return nil
}

func (program *Program) previousPullRequestTab(_ *gocui.Gui, _ *gocui.View) error {
	program.model.PreviousPullRequestTab()
	return nil
}

func (program *Program) focusDetailView(gui *gocui.Gui, _ *gocui.View) error {
	program.model.FocusDetailView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusUserView(gui *gocui.Gui, _ *gocui.View) error {
	program.model.FocusUserView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	program.model.FocusPullRequestsView()
	return program.syncCurrentView(gui)
}

func (program *Program) openDetail(gui *gocui.Gui, _ *gocui.View) error {
	program.model.OpenDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) closeDetail(gui *gocui.Gui, _ *gocui.View) error {
	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) syncCurrentView(gui *gocui.Gui) error {
	_, err := gui.SetCurrentView(program.currentViewName())
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) maybeLoadConnectedUser(gui *gocui.Gui) {
	if program.connectedUserLoader == nil || program.connectedUserLoadStarted {
		return
	}

	program.connectedUserLoadStarted = true
	go program.loadConnectedUser(gui)
}

func (program *Program) loadConnectedUser(gui *gocui.Gui) {
	user, err := program.connectedUserLoader.GetConnectedUser()

	gui.Update(func(gui *gocui.Gui) error {
		program.model.SetUsers([]Item{connectedUserStateItem(user, err)})
		return program.refreshViews(gui)
	})
}

func (program *Program) refreshViews(gui *gocui.Gui) error {
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

	return program.syncCurrentView(gui)
}

func (program *Program) currentViewName() string {
	switch program.model.Focus() {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}
