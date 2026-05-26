package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

type linkClipboardCommandRuntime struct {
	linkOpener                          LinkOpener
	clipboardWriter                     ClipboardWriter
	dispatch                            func(*gocui.Gui, Msg) error
	resolveView                         func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailCursorLink             func(*gocui.View) (string, bool)
	currentPullRequestBuildRunPopupLink func(*gocui.View) (string, bool)
	prepareSelectedDetailClipboardWrite func(*gocui.Gui, *gocui.View, Focus) (writeClipboardCmd, bool)
	prepareBuildPopupClipboardWrite     func(*gocui.Gui, *gocui.View, Focus) (writeClipboardCmd, bool)
}

func newLinkClipboardCommandRuntime(program *Program) linkClipboardCommandRuntime {
	if program == nil {
		return linkClipboardCommandRuntime{}
	}
	return linkClipboardCommandRuntime{
		linkOpener:                          program.linkOpener,
		clipboardWriter:                     program.clipboardWriter,
		dispatch:                            program.dispatch,
		resolveView:                         program.resolveView,
		currentDetailCursorLink:             program.currentDetailCursorLink,
		currentPullRequestBuildRunPopupLink: program.currentPullRequestBuildRunPopupLink,
		prepareSelectedDetailClipboardWrite: func(gui *gocui.Gui, view *gocui.View, target Focus) (writeClipboardCmd, bool) {
			actualView := program.resolveView(gui, view, viewDetailName)
			detailDocument := program.currentDetailDocument(actualView)
			program.syncDetailViewState(detailDocument, viewPageSize(actualView))
			selection, _ := detailSelectionForCurrentMode(program.detailState.viewState, detailDocument)
			text := program.detailState.viewState.selectedText(detailDocument)
			program.detailState.viewState.exitVisualMode()
			return writeClipboardCmd{Text: text, SuccessMessage: detailYankSuccessMessage, FailureMessage: detailYankFailureMessage, Target: target, Selection: selection, SelectionTarget: clipboardWriteSelectionDetail}, true
		},
		prepareBuildPopupClipboardWrite: func(gui *gocui.Gui, view *gocui.View, target Focus) (writeClipboardCmd, bool) {
			popup := program.pullRequestBuildRunPopup
			if popup == nil {
				return writeClipboardCmd{}, false
			}
			actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
			document := program.currentPullRequestBuildRunPopupDocument(actualView)
			popup.viewState.sync(document, viewPageSize(actualView))
			popup.viewState.clearPendingPrefix()
			if popup.viewState.mode.isVisual() {
				selection, _ := detailSelectionForCurrentMode(popup.viewState, document)
				text := popup.viewState.selectedText(document)
				popup.viewState.exitVisualMode()
				return writeClipboardCmd{Text: text, SuccessMessage: detailYankSuccessMessage, FailureMessage: detailYankFailureMessage, Target: target, Selection: selection, SelectionTarget: clipboardWriteSelectionBuildPopup}, true
			}
			trimmedRunURL := strings.TrimSpace(popup.runURL)
			if trimmedRunURL == "" {
				return writeClipboardCmd{}, false
			}
			return writeClipboardCmd{Text: trimmedRunURL, SuccessMessage: yankSuccessMessage, FailureMessage: yankFailureMessage, Target: target}, true
		},
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
	switch {
	case !ok:
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkUnavailableMessage})
		return
	case runtime.linkOpener == nil:
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkOpenerUnavailableMessage})
		return
	default:
		executeOpenBrowserURLCommand(runtime.browserClipboardRuntime(), gui, openBrowserURLCmd{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: command.Target})
	}
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
	if runtime.linkOpener == nil {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkOpenerUnavailableMessage})
		return
	}

	actualView := runtime.resolveView(gui, nil, viewPullRequestBuildInfoName)
	url, ok := runtime.currentPullRequestBuildRunPopupLink(actualView)
	if !ok {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkUnavailableMessage})
		return
	}

	executeOpenBrowserURLCommand(runtime.browserClipboardRuntime(), gui, openBrowserURLCmd{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: command.Target})
}

type prepareSelectedDetailClipboardWriteCmd struct {
	Target Focus
}

func (command prepareSelectedDetailClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePrepareSelectedDetailClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePrepareSelectedDetailClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command prepareSelectedDetailClipboardWriteCmd) {
	if runtime.prepareSelectedDetailClipboardWrite == nil {
		return
	}
	clipboardCommand, ok := runtime.prepareSelectedDetailClipboardWrite(gui, nil, command.Target)
	if !ok {
		return
	}
	executeWriteClipboardCommand(runtime.browserClipboardRuntime(), gui, clipboardCommand)
}

type preparePullRequestBuildRunPopupClipboardWriteCmd struct {
	Target Focus
}

func (command preparePullRequestBuildRunPopupClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePreparePullRequestBuildRunPopupClipboardWriteCommand(newLinkClipboardCommandRuntime(program), gui, command)
}

func executePreparePullRequestBuildRunPopupClipboardWriteCommand(runtime linkClipboardCommandRuntime, gui *gocui.Gui, command preparePullRequestBuildRunPopupClipboardWriteCmd) {
	if runtime.dispatch == nil || runtime.prepareBuildPopupClipboardWrite == nil {
		return
	}
	clipboardCommand, ok := runtime.prepareBuildPopupClipboardWrite(gui, nil, command.Target)
	if !ok {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: yankUnavailableMessage})
		return
	}
	executeWriteClipboardCommand(runtime.browserClipboardRuntime(), gui, clipboardCommand)
}
