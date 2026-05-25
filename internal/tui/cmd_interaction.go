package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type interactionCommandRuntime struct {
	linkOpener      LinkOpener
	clipboardWriter ClipboardWriter
	clipboardReader ClipboardReader
	externalEditor  ExternalEditor
	buildQueries    BuildQueries
	submitDeps      modalEditorSubmitCommandDeps
	runAsync        func(func())
	dispatch        func(*gocui.Gui, Msg) error
	dispatchAsync   func(*gocui.Gui, Msg)
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
		linkOpener:      program.linkOpener,
		clipboardWriter: program.clipboardWriter,
		clipboardReader: program.clipboardReader,
		externalEditor:  program.externalEditor,
		buildQueries:    program.buildQueries,
		submitDeps:      newModalEditorSubmitCommandDeps(program),
		runAsync:        program.runAsync,
		dispatch:        program.dispatch,
		dispatchAsync:   program.dispatchAsync,
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
