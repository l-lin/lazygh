package tui

import (
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewDetailName       = "detail"
	viewUserName         = "user"
	viewPullRequestsName = "pull-requests"
)

type Program struct {
	model *Model
}

func NewProgram() *Program {
	return NewProgramWithModel(NewModel(DefaultSeedData()))
}

func NewProgramWithModel(model *Model) *Program {
	if model == nil {
		model = NewModel(DefaultSeedData())
	}

	return &Program{model: model}
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
	gui.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	gui.SelFrameColor = gocui.GetColor(theme.ActiveBorderHex)
	gui.SelBgColor = gocui.GetColor(theme.ActiveSelectionBackgroundHex)
	gui.SelFgColor = gocui.GetColor(theme.ActiveSelectionForegroundHex)
}

func (program *Program) setKeybindings(gui *gocui.Gui) error {
	bindings := []struct {
		key     any
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{key: gocui.KeyCtrlC, handler: program.quit},
		{key: gocui.KeyTab, handler: program.nextSideView},
		{key: gocui.KeyBacktab, handler: program.previousSideView},
		{key: 'l', handler: program.nextSideView},
		{key: 'h', handler: program.previousSideView},
		{key: 'j', handler: program.moveSelectionDown},
		{key: 'k', handler: program.moveSelectionUp},
		{key: ']', handler: program.nextPullRequestTab},
		{key: '[', handler: program.previousPullRequestTab},
		{key: gocui.KeyEnter, handler: program.openDetail},
		{key: gocui.KeyEsc, handler: program.closeDetail},
		{key: gocui.KeyCtrlLsqBracket, handler: program.closeDetail},
		{key: gocui.KeyCtrl3, handler: program.closeDetail},
	}

	for _, binding := range bindings {
		if err := gui.SetKeybinding("", binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}

	return nil
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
