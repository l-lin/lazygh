package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type interactionCommandRuntime struct {
	linkOpener                          LinkOpener
	clipboardWriter                     ClipboardWriter
	clipboardReader                     ClipboardReader
	externalEditor                      ExternalEditor
	buildQueries                        BuildQueries
	submitDeps                          modalEditorSubmitCommandDeps
	runAsync                            func(func())
	dispatch                            func(*gocui.Gui, Msg) error
	dispatchAsync                       func(*gocui.Gui, Msg)
	resolveView                         func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument               func(*gocui.View) detailDocument
	syncDetailViewState                 func(detailDocument, int)
	currentDetailCursor                 func() detailPosition
	currentDetailCursorLink             func(*gocui.View) (string, bool)
	currentPullRequestBuildRunPopupLink func(*gocui.View) (string, bool)
	followSubmittedDetailSearch         func(*gocui.Gui) error
	followReverseDetailSearch           func(*gocui.Gui) error
	configureGUI                        func(*gocui.Gui)
	hasPullRequestListQueries           func() bool
	hasNotificationQueries              func() bool
	hasDetailQueries                    func() bool
	reloadNotifications                 func(*gocui.Gui)
	markManualNotificationRefresh       func() bool
	markManualPullRequestListRefresh    func(PullRequestTab) bool
	markManualPullRequestDetailRefresh  func(githubdomain.PullRequest) bool
	markManualPullRequestDiffRefresh    func(githubdomain.PullRequest) bool
	beginManualRefresh                  func(string, int)
	reviewModeActive                    func() bool
	activePullRequestTab                func() PullRequestTab
	handleDetailLineNavigation          func(*gocui.Gui, *gocui.View, int)
	handlePageNavigation                func(*gocui.Gui, *gocui.View, pageNavigationKind)
	focusDetailRenderedLine             func(*gocui.Gui, *gocui.View, int)
	prepareSelectedDetailClipboardWrite func(*gocui.Gui, *gocui.View, Focus) (writeClipboardCmd, bool)
	prepareBuildPopupClipboardWrite     func(*gocui.Gui, *gocui.View, Focus) (writeClipboardCmd, bool)
}

type openBrowserURLCmd struct {
	URL            string
	SuccessMessage string
	FailureMessage string
	Target         Focus
}

func newInteractionCommandRuntime(program *Program) interactionCommandRuntime {
	if program == nil {
		return interactionCommandRuntime{}
	}
	return interactionCommandRuntime{
		linkOpener:            program.linkOpener,
		clipboardWriter:       program.clipboardWriter,
		clipboardReader:       program.clipboardReader,
		externalEditor:        program.externalEditor,
		buildQueries:          program.buildQueries,
		submitDeps:            newModalEditorSubmitCommandDeps(program),
		runAsync:              program.runAsync,
		dispatch:              program.dispatch,
		dispatchAsync:         program.dispatchAsync,
		resolveView:           program.resolveView,
		currentDetailDocument: program.currentDetailDocument,
		syncDetailViewState:   program.syncDetailViewState,
		currentDetailCursor: func() detailPosition {
			return program.detailState.viewState.cursor
		},
		currentDetailCursorLink:             program.currentDetailCursorLink,
		currentPullRequestBuildRunPopupLink: program.currentPullRequestBuildRunPopupLink,
		followSubmittedDetailSearch:         program.followSubmittedDetailSearch,
		followReverseDetailSearch:           program.followReverseDetailSearch,
		configureGUI:                        program.configureGUI,
		hasPullRequestListQueries:           program.hasPullRequestListQueries,
		hasNotificationQueries:              program.hasNotificationQueries,
		hasDetailQueries:                    program.hasDetailQueries,
		reloadNotifications:                 program.reloadNotifications,
		markManualNotificationRefresh:       program.markManualNotificationRefresh,
		markManualPullRequestListRefresh:    program.markManualPullRequestListRefresh,
		markManualPullRequestDetailRefresh:  program.markManualPullRequestDetailRefresh,
		markManualPullRequestDiffRefresh:    program.markManualPullRequestDiffRefresh,
		beginManualRefresh:                  program.beginManualRefresh,
		reviewModeActive:                    program.reviewModeActive,
		activePullRequestTab:                program.model.ActivePullRequestTab,
		handleDetailLineNavigation: func(gui *gocui.Gui, view *gocui.View, delta int) {
			actualView := program.resolveView(gui, view, viewDetailName)
			steps := delta
			if steps < 0 {
				steps = -steps
			}
			_ = program.mutateDetailViewStateForYankMotion(gui, actualView, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
				for range steps {
					if delta > 0 {
						program.detailState.viewState.moveDown(document, viewportHeight)
						continue
					}
					program.detailState.viewState.moveUp(document, viewportHeight)
				}
			})
		},
		handlePageNavigation: func(gui *gocui.Gui, view *gocui.View, kind pageNavigationKind) {
			actualView := program.resolveView(gui, view, program.currentViewName())
			pageSize := viewPageSize(actualView)
			delta := pageNavigationDelta(kind, pageSize)
			switch kind {
			case pageNavigationKindHalfDown:
				_ = program.handlePageChange(gui, actualView, delta, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.pageDown(document, viewportHeight)
				})
			case pageNavigationKindHalfUp:
				_ = program.handlePageChange(gui, actualView, delta, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.pageUp(document, viewportHeight)
				})
			case pageNavigationKindFullDown:
				_ = program.handlePageChange(gui, actualView, delta, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.fullPageDown(document, viewportHeight)
				})
			case pageNavigationKindFullUp:
				_ = program.handlePageChange(gui, actualView, delta, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.fullPageUp(document, viewportHeight)
				})
			}
		},
		focusDetailRenderedLine: func(gui *gocui.Gui, view *gocui.View, line int) {
			_ = program.mutateDetailViewStateWithoutRefresh(gui, view, func(document detailDocument, viewportHeight int) {
				program.focusDetailLine(document, viewportHeight, line)
			})
		},
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

func (command openBrowserURLCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenBrowserURLCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeOpenBrowserURLCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command openBrowserURLCmd) {
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
	executeWriteClipboardCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeWriteClipboardCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command writeClipboardCmd) {
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

type reportErrorCmd struct {
	Message string
}

func (command reportErrorCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	program.reportError(gui, command.Message)
}

type configureGUICmd struct{}

func (configureGUICmd) execute(program *Program, gui *gocui.Gui) {
	executeConfigureGUICommand(newInteractionCommandRuntime(program), gui)
}

func executeConfigureGUICommand(runtime interactionCommandRuntime, gui *gocui.Gui) {
	if gui == nil || runtime.configureGUI == nil {
		return
	}
	runtime.configureGUI(gui)
}

type beginManualPullRequestListRefreshCmd struct {
	Tab            PullRequestTab
	SuccessMessage string
}

func (command beginManualPullRequestListRefreshCmd) execute(program *Program, gui *gocui.Gui) {
	executeBeginManualPullRequestListRefreshCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeBeginManualPullRequestListRefreshCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command beginManualPullRequestListRefreshCmd) {
	pendingOperations := 0
	if gui != nil && runtime.hasPullRequestListQueries != nil && runtime.hasPullRequestListQueries() && runtime.markManualPullRequestListRefresh != nil && runtime.markManualPullRequestListRefresh(command.Tab) {
		pendingOperations++
	}
	if runtime.beginManualRefresh != nil {
		runtime.beginManualRefresh(command.SuccessMessage, pendingOperations)
	}
}

type beginManualPullRequestRefreshCmd struct {
	Summary        githubdomain.PullRequest
	SuccessMessage string
}

func (command beginManualPullRequestRefreshCmd) execute(program *Program, gui *gocui.Gui) {
	executeBeginManualPullRequestRefreshCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeBeginManualPullRequestRefreshCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command beginManualPullRequestRefreshCmd) {
	pendingOperations := 0
	if runtime.hasDetailQueries != nil && runtime.hasDetailQueries() {
		if runtime.markManualPullRequestDetailRefresh != nil && runtime.markManualPullRequestDetailRefresh(command.Summary) {
			pendingOperations++
		}
	}
	if runtime.reviewModeActive != nil && runtime.reviewModeActive() {
		if runtime.hasDetailQueries != nil && runtime.hasDetailQueries() {
			if runtime.markManualPullRequestDiffRefresh != nil && runtime.markManualPullRequestDiffRefresh(command.Summary) {
				pendingOperations++
			}
		}
	} else if gui != nil && runtime.hasPullRequestListQueries != nil && runtime.hasPullRequestListQueries() && runtime.markManualPullRequestListRefresh != nil && runtime.activePullRequestTab != nil {
		if runtime.markManualPullRequestListRefresh(runtime.activePullRequestTab()) {
			pendingOperations++
		}
	}
	if runtime.beginManualRefresh != nil {
		runtime.beginManualRefresh(command.SuccessMessage, pendingOperations)
	}
}

type refreshNotificationsCmd struct{}

func (refreshNotificationsCmd) execute(program *Program, gui *gocui.Gui) {
	executeRefreshNotificationsCommand(newInteractionCommandRuntime(program), gui)
}

func executeRefreshNotificationsCommand(runtime interactionCommandRuntime, gui *gocui.Gui) {
	pendingOperations := 0
	if gui != nil && runtime.reviewModeActive != nil && !runtime.reviewModeActive() && runtime.hasNotificationQueries != nil && runtime.hasNotificationQueries() && runtime.markManualNotificationRefresh != nil && runtime.markManualNotificationRefresh() {
		pendingOperations++
	}
	if runtime.beginManualRefresh != nil {
		runtime.beginManualRefresh(notificationsRefreshSuccessMessage, pendingOperations)
	}
	if runtime.reloadNotifications != nil {
		runtime.reloadNotifications(gui)
	}
}

type focusReviewCommentCmd struct {
	RenderedLine int
}

func (command focusReviewCommentCmd) execute(program *Program, gui *gocui.Gui) {
	executeFocusReviewCommentCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeFocusReviewCommentCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command focusReviewCommentCmd) {
	if runtime.resolveView == nil || runtime.focusDetailRenderedLine == nil {
		return
	}
	actualView := runtime.resolveView(gui, nil, viewDetailName)
	runtime.focusDetailRenderedLine(gui, actualView, command.RenderedLine)
}

type detailLineNavigationCmd struct {
	View  *gocui.View
	Delta int
}

func (command detailLineNavigationCmd) execute(program *Program, gui *gocui.Gui) {
	executeDetailLineNavigationCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeDetailLineNavigationCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command detailLineNavigationCmd) {
	if runtime.handleDetailLineNavigation == nil {
		return
	}
	runtime.handleDetailLineNavigation(gui, command.View, command.Delta)
}

type pageNavigationCmd struct {
	View *gocui.View
	Kind pageNavigationKind
}

func (command pageNavigationCmd) execute(program *Program, gui *gocui.Gui) {
	executePageNavigationCommand(newInteractionCommandRuntime(program), gui, command)
}

func executePageNavigationCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command pageNavigationCmd) {
	if runtime.handlePageNavigation == nil {
		return
	}
	runtime.handlePageNavigation(gui, command.View, command.Kind)
}

func pageNavigationDelta(kind pageNavigationKind, pageSize int) int {
	switch kind {
	case pageNavigationKindHalfDown:
		return pageDelta(pageSize)
	case pageNavigationKindHalfUp:
		return -pageDelta(pageSize)
	case pageNavigationKindFullDown:
		return fullPageDelta(pageSize)
	case pageNavigationKindFullUp:
		return -fullPageDelta(pageSize)
	default:
		return 0
	}
}

type resolveDetailSearchWordCmd struct {
	View    *gocui.View
	Reverse bool
}

func (command resolveDetailSearchWordCmd) execute(program *Program, gui *gocui.Gui) {
	executeResolveDetailSearchWordCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeResolveDetailSearchWordCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command resolveDetailSearchWordCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentDetailDocument == nil || runtime.syncDetailViewState == nil || runtime.currentDetailCursor == nil {
		return
	}

	actualView := runtime.resolveView(gui, command.View, viewDetailName)
	document := runtime.currentDetailDocument(actualView)
	runtime.syncDetailViewState(document, viewPageSize(actualView))
	query, ok := document.wordAt(runtime.currentDetailCursor())
	if !ok {
		return
	}
	_ = runtime.dispatch(gui, MsgDetailSearchWordResolved{Query: query, Reverse: command.Reverse})
}

type followDetailSearchCmd struct {
	Reverse bool
}

func (command followDetailSearchCmd) execute(program *Program, gui *gocui.Gui) {
	executeFollowDetailSearchCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeFollowDetailSearchCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command followDetailSearchCmd) {
	if command.Reverse {
		if runtime.followReverseDetailSearch != nil {
			_ = runtime.followReverseDetailSearch(gui)
		}
		return
	}
	if runtime.followSubmittedDetailSearch != nil {
		_ = runtime.followSubmittedDetailSearch(gui)
	}
}

type openLinkUnderCursorCmd struct {
	View   *gocui.View
	Target Focus
}

func (command openLinkUnderCursorCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenLinkUnderCursorCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeOpenLinkUnderCursorCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command openLinkUnderCursorCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentDetailCursorLink == nil {
		return
	}

	actualView := runtime.resolveView(gui, command.View, viewDetailName)
	url, ok := runtime.currentDetailCursorLink(actualView)
	switch {
	case !ok:
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkUnavailableMessage})
		return
	case runtime.linkOpener == nil:
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkOpenerUnavailableMessage})
		return
	default:
		executeOpenBrowserURLCommand(runtime, gui, openBrowserURLCmd{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: command.Target})
	}
}

type openPullRequestBuildRunPopupLinkCmd struct {
	View   *gocui.View
	Target Focus
}

func (command openPullRequestBuildRunPopupLinkCmd) execute(program *Program, gui *gocui.Gui) {
	executeOpenPullRequestBuildRunPopupLinkCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeOpenPullRequestBuildRunPopupLinkCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command openPullRequestBuildRunPopupLinkCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentPullRequestBuildRunPopupLink == nil {
		return
	}
	if runtime.linkOpener == nil {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkOpenerUnavailableMessage})
		return
	}

	actualView := runtime.resolveView(gui, command.View, viewPullRequestBuildInfoName)
	url, ok := runtime.currentPullRequestBuildRunPopupLink(actualView)
	if !ok {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: openLinkUnavailableMessage})
		return
	}

	executeOpenBrowserURLCommand(runtime, gui, openBrowserURLCmd{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: command.Target})
}

type prepareSelectedDetailClipboardWriteCmd struct {
	View   *gocui.View
	Target Focus
}

func (command prepareSelectedDetailClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePrepareSelectedDetailClipboardWriteCommand(newInteractionCommandRuntime(program), gui, command)
}

func executePrepareSelectedDetailClipboardWriteCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command prepareSelectedDetailClipboardWriteCmd) {
	if runtime.prepareSelectedDetailClipboardWrite == nil {
		return
	}
	clipboardCommand, ok := runtime.prepareSelectedDetailClipboardWrite(gui, command.View, command.Target)
	if !ok {
		return
	}
	executeWriteClipboardCommand(runtime, gui, clipboardCommand)
}

type preparePullRequestBuildRunPopupClipboardWriteCmd struct {
	View   *gocui.View
	Target Focus
}

func (command preparePullRequestBuildRunPopupClipboardWriteCmd) execute(program *Program, gui *gocui.Gui) {
	executePreparePullRequestBuildRunPopupClipboardWriteCommand(newInteractionCommandRuntime(program), gui, command)
}

func executePreparePullRequestBuildRunPopupClipboardWriteCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command preparePullRequestBuildRunPopupClipboardWriteCmd) {
	if runtime.dispatch == nil || runtime.prepareBuildPopupClipboardWrite == nil {
		return
	}
	clipboardCommand, ok := runtime.prepareBuildPopupClipboardWrite(gui, command.View, command.Target)
	if !ok {
		_ = runtime.dispatch(gui, MsgFeedbackSet{Target: command.Target, Message: yankUnavailableMessage})
		return
	}
	executeWriteClipboardCommand(runtime, gui, clipboardCommand)
}

type readPullRequestURLFromClipboardCmd struct{}

func (readPullRequestURLFromClipboardCmd) execute(program *Program, gui *gocui.Gui) {
	executeReadPullRequestURLFromClipboardCommand(newInteractionCommandRuntime(program), gui)
}

func executeReadPullRequestURLFromClipboardCommand(runtime interactionCommandRuntime, gui *gocui.Gui) {
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

type modalEditorExternalEditCmd struct {
	Text string
}

func (command modalEditorExternalEditCmd) execute(program *Program, gui *gocui.Gui) {
	executeModalEditorExternalEditCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeModalEditorExternalEditCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command modalEditorExternalEditCmd) {
	if runtime.dispatch == nil {
		return
	}
	if runtime.externalEditor == nil {
		_ = runtime.dispatch(gui, MsgModalEditorExternalEditFinished{Err: errors.New("external editor is unavailable")})
		return
	}

	editedText, err := runtime.externalEditor.Edit(gui, command.Text)
	_ = runtime.dispatch(gui, MsgModalEditorExternalEditFinished{Text: editedText, Err: err})
}

type modalEditorSubmitCmd struct {
	request modalEditorSubmitRequest
}

func (command modalEditorSubmitCmd) execute(program *Program, gui *gocui.Gui) {
	executeModalEditorSubmitCommand(newInteractionCommandRuntime(program), gui, command)
}

func executeModalEditorSubmitCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command modalEditorSubmitCmd) {
	if command.request == nil || runtime.dispatch == nil {
		return
	}

	success, err := command.request.run(runtime.submitDeps)
	_ = runtime.dispatch(gui, MsgModalEditorSubmitFinished{Err: err, Success: success})
}

type pullRequestBuildRunLoadCmd struct {
	Repository string
	Target     pullRequestBuildRunTarget
}

func (command pullRequestBuildRunLoadCmd) execute(program *Program, gui *gocui.Gui) {
	executePullRequestBuildRunLoadCommand(newInteractionCommandRuntime(program), gui, command)
}

func executePullRequestBuildRunLoadCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command pullRequestBuildRunLoadCmd) {
	if runtime.buildQueries == nil || runtime.dispatchAsync == nil {
		return
	}

	run := func() {
		rawRunOutput, err := runtime.buildQueries.GetPullRequestBuildRun(command.Repository, command.Target.check)
		jobs := []githubdomain.PullRequestBuildRunJob(nil)
		jobsErr := error(nil)
		if err == nil {
			jobs, jobsErr = runtime.buildQueries.GetPullRequestBuildRunJobs(command.Repository, command.Target.check)
		}
		runtime.dispatchAsync(gui, MsgPullRequestBuildRunLoaded{Target: command.Target, RawRunOutput: rawRunOutput, Jobs: jobs, JobsErr: jobsErr, Err: err})
	}
	if runtime.runAsync != nil {
		runtime.runAsync(run)
		return
	}
	run()
}

type pullRequestBuildRunJobLogLoadCmd struct {
	Repository string
	Check      githubdomain.PullRequestStatusCheck
}

func (command pullRequestBuildRunJobLogLoadCmd) execute(program *Program, gui *gocui.Gui) {
	executePullRequestBuildRunJobLogLoadCommand(newInteractionCommandRuntime(program), gui, command)
}

func executePullRequestBuildRunJobLogLoadCommand(runtime interactionCommandRuntime, gui *gocui.Gui, command pullRequestBuildRunJobLogLoadCmd) {
	if runtime.buildQueries == nil || runtime.dispatchAsync == nil {
		return
	}

	run := func() {
		job, rawLogOutput, err := runtime.buildQueries.GetPullRequestBuildRunJobLogForCheck(command.Repository, command.Check)
		runtime.dispatchAsync(gui, MsgPullRequestBuildRunJobLogLoaded{Repository: command.Repository, Job: job, RawLogOutput: rawLogOutput, Err: err})
	}
	if runtime.runAsync != nil {
		runtime.runAsync(run)
		return
	}
	run()
}
