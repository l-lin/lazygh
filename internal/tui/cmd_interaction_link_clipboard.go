package tui

import "github.com/jesseduffield/gocui"

type linkClipboardCommandRuntime struct {
	linkOpener                              LinkOpener
	clipboardWriter                         ClipboardWriter
	executeMessage                          func(*gocui.Gui, Msg) error
	resolveView                             func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailCursorLink                 func() (string, bool)
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
		executeMessage:                          program.executeRuntimeMessage,
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
		executeMessage:  runtime.executeMessage,
	}
}

type openLinkUnderCursorCmd struct {
	Target Focus
}

func (command openLinkUnderCursorCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenLinkUnderCursorCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executeOpenLinkUnderCursorCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command openLinkUnderCursorCmd) {
	if runtime.executeMessage == nil || runtime.currentDetailCursorLink == nil {
		return
	}

	url, ok := runtime.currentDetailCursorLink()
	_ = runtime.executeMessage(gui, MsgOpenLinkUnderCursorResolved{Target: command.Target, URL: url, LinkAvailable: ok, OpenerAvailable: runtime.linkOpener != nil})
}

type openPullRequestBuildRunPopupLinkCmd struct {
	Target Focus
}

func (command openPullRequestBuildRunPopupLinkCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenPullRequestBuildRunPopupLinkCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executeOpenPullRequestBuildRunPopupLinkCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command openPullRequestBuildRunPopupLinkCmd) {
	if runtime.executeMessage == nil || runtime.resolveView == nil || runtime.currentPullRequestBuildRunPopupLink == nil {
		return
	}

	actualView := runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
	url, ok := runtime.currentPullRequestBuildRunPopupLink(actualView)
	_ = runtime.executeMessage(gui, MsgOpenPullRequestBuildRunPopupLinkResolved{Target: command.Target, URL: url, LinkAvailable: ok, OpenerAvailable: runtime.linkOpener != nil})
}

type prepareSelectedDetailClipboardWriteCmd struct {
	Target Focus
}

func (command prepareSelectedDetailClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePrepareSelectedDetailClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePrepareSelectedDetailClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command prepareSelectedDetailClipboardWriteCmd) {
	if runtime.executeMessage == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.executeMessage(gui, MsgSelectedDetailClipboardPrepared{Target: command.Target, Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}

type preparePullRequestBuildRunPopupClipboardWriteCmd struct {
	Target Focus
}

func (command preparePullRequestBuildRunPopupClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePreparePullRequestBuildRunPopupClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePreparePullRequestBuildRunPopupClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command preparePullRequestBuildRunPopupClipboardWriteCmd) {
	if runtime.executeMessage == nil || runtime.currentPullRequestBuildRunPopupDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
	}
	_ = runtime.executeMessage(gui, MsgPullRequestBuildRunPopupClipboardPrepared{Target: command.Target, Document: runtime.currentPullRequestBuildRunPopupDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}
