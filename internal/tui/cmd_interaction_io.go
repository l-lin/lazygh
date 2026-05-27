package tui

import "github.com/jesseduffield/gocui"

type browserClipboardCommandRuntime struct {
	linkOpener      LinkOpener
	clipboardWriter ClipboardWriter
	clipboardReader ClipboardReader
	dispatch        func(*gocui.Gui, Msg) error
}

func newBrowserClipboardCommandRuntime(program *Program) browserClipboardCommandRuntime {
	if program == nil {
		return browserClipboardCommandRuntime{}
	}
	return browserClipboardCommandRuntime{
		linkOpener:      program.linkOpener,
		clipboardWriter: program.clipboardWriter,
		clipboardReader: program.clipboardReader,
		dispatch:        program.dispatch,
	}
}

type openBrowserURLCmd struct {
	URL            string
	SuccessMessage string
	FailureMessage string
	Target         Focus
}

func (command openBrowserURLCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenBrowserURLCommand(newBrowserClipboardCommandRuntime(program), gui, command)
}

func executeOpenBrowserURLCommand(runtime browserClipboardCommandRuntime, gui *gocui.Gui, command openBrowserURLCmd) {
	err := ErrLinkOpenerUnavailable
	if runtime.linkOpener != nil {
		err = runtime.linkOpener.Open(command.URL)
	}
	if runtime.dispatch != nil {
		_ = runtime.dispatch(gui, MsgOpenBrowserURLFinished{SuccessMessage: command.SuccessMessage, FailureMessage: command.FailureMessage, Target: command.Target, Err: err})
	}
}

type writeClipboardCmd struct {
	Text            string
	SuccessMessage  string
	FailureMessage  string
	Target          Focus
	Selection       detailSelectionRange
	SelectionTarget clipboardWriteSelectionTarget
}

func (command writeClipboardCmd) execute(program *Program, gui *gocui.Gui) {
	executeWriteClipboardCommand(newBrowserClipboardCommandRuntime(program), gui, command)
}

func executeWriteClipboardCommand(runtime browserClipboardCommandRuntime, gui *gocui.Gui, command writeClipboardCmd) {
	err := ErrClipboardUnavailable
	if runtime.clipboardWriter != nil {
		err = runtime.clipboardWriter.WriteText(command.Text)
	}
	if runtime.dispatch != nil {
		_ = runtime.dispatch(gui, MsgClipboardWriteFinished{
			SuccessMessage:  command.SuccessMessage,
			FailureMessage:  command.FailureMessage,
			Target:          command.Target,
			Err:             err,
			Selection:       command.Selection,
			SelectionTarget: command.SelectionTarget,
		})
	}
}

type readPullRequestURLFromClipboardCmd struct{}

func (readPullRequestURLFromClipboardCmd) execute(program *Program, gui *gocui.Gui) {
	executeReadPullRequestURLFromClipboardCommand(newBrowserClipboardCommandRuntime(program), gui)
}

func executeReadPullRequestURLFromClipboardCommand(runtime browserClipboardCommandRuntime, gui *gocui.Gui) {
	if runtime.dispatch == nil {
		return
	}
	if runtime.clipboardReader == nil {
		_ = runtime.dispatch(gui, MsgPullRequestURLReadFromClipboard{Err: ErrClipboardUnavailable})
		return
	}

	url, err := runtime.clipboardReader.ReadText()
	_ = runtime.dispatch(gui, MsgPullRequestURLReadFromClipboard{URL: url, Err: err})
}
