package tui

import "github.com/jesseduffield/gocui"

type linkClipboardCommandRuntime struct {
	linkOpener                              LinkOpener
	clipboardWriter                         ClipboardWriter
	dispatch                                func(*gocui.Gui, Msg) error
	resolveView                             func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailCursorLink                 func(*gocui.View) (string, bool)
	currentPullRequestBuildRunPopupLink     func(*gocui.View) (string, bool)
	currentDetailDocument                   func(*gocui.View) detailDocument
	currentPullRequestBuildRunPopupDocument func(*gocui.View) detailDocument
}

func newLinkClipboardCommandRuntime(program *Program) linkClipboardCommandRuntime {
	if program == nil {
		return linkClipboardCommandRuntime{}
	}
	return linkClipboardCommandRuntime{
		linkOpener:                              program.linkOpener,
		clipboardWriter:                         program.clipboardWriter,
		dispatch:                                program.dispatch,
		resolveView:                             program.resolveView,
		currentDetailCursorLink:                 program.currentDetailCursorLink,
		currentPullRequestBuildRunPopupLink:     program.currentPullRequestBuildRunPopupLink,
		currentDetailDocument:                   program.currentDetailDocument,
		currentPullRequestBuildRunPopupDocument: program.currentPullRequestBuildRunPopupDocument,
	}
}

func (runtime linkClipboardCommandRuntime) browserClipboardRuntime() browserClipboardCommandRuntime {
	return browserClipboardCommandRuntime{
		linkOpener:      runtime.linkOpener,
		clipboardWriter: runtime.clipboardWriter,
		dispatch:        runtime.dispatch,
	}
}

type openLinkUnderCursorCmd struct {
	Target Focus
}

func (command openLinkUnderCursorCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenLinkUnderCursorCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executeOpenLinkUnderCursorCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command openLinkUnderCursorCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentDetailCursorLink == nil {
		return
	}

	actualView := runtime.resolveView(gui, nil, viewDetailName)
	url, ok := runtime.currentDetailCursorLink(actualView)
	_ = runtime.dispatch(gui, MsgOpenLinkUnderCursorResolved{Target: command.Target, URL: url, LinkAvailable: ok, OpenerAvailable: runtime.linkOpener != nil})
}

type openPullRequestBuildRunPopupLinkCmd struct {
	Target Focus
}

func (command openPullRequestBuildRunPopupLinkCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenPullRequestBuildRunPopupLinkCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executeOpenPullRequestBuildRunPopupLinkCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command openPullRequestBuildRunPopupLinkCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentPullRequestBuildRunPopupLink == nil {
		return
	}

	actualView := runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
	url, ok := runtime.currentPullRequestBuildRunPopupLink(actualView)
	_ = runtime.dispatch(gui, MsgOpenPullRequestBuildRunPopupLinkResolved{Target: command.Target, URL: url, LinkAvailable: ok, OpenerAvailable: runtime.linkOpener != nil})
}

type prepareSelectedDetailClipboardWriteCmd struct {
	Target Focus
}

func (command prepareSelectedDetailClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePrepareSelectedDetailClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePrepareSelectedDetailClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command prepareSelectedDetailClipboardWriteCmd) {
	if runtime.dispatch == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.dispatch(gui, MsgSelectedDetailClipboardPrepared{Target: command.Target, Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}

type preparePullRequestBuildRunPopupClipboardWriteCmd struct {
	Target Focus
}

func (command preparePullRequestBuildRunPopupClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePreparePullRequestBuildRunPopupClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePreparePullRequestBuildRunPopupClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command preparePullRequestBuildRunPopupClipboardWriteCmd) {
	if runtime.dispatch == nil || runtime.currentPullRequestBuildRunPopupDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
	}
	_ = runtime.dispatch(gui, MsgPullRequestBuildRunPopupClipboardPrepared{Target: command.Target, Document: runtime.currentPullRequestBuildRunPopupDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}
