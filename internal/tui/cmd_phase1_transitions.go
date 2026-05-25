package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type openBrowserURLCmd struct {
	URL            string
	SuccessMessage string
	FailureMessage string
	Target         Focus
}

func (command openBrowserURLCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	err := ErrLinkOpenerUnavailable
	if program.linkOpener != nil {
		err = program.linkOpener.Open(command.URL)
	}
	_ = program.dispatch(gui, MsgOpenBrowserURLFinished{SuccessMessage: command.SuccessMessage, FailureMessage: command.FailureMessage, Target: command.Target, Err: err})
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
	if program == nil {
		return
	}

	err := ErrClipboardUnavailable
	if program.clipboardWriter != nil {
		err = program.clipboardWriter.WriteText(command.Text)
	}
	_ = program.dispatch(gui, MsgClipboardWriteFinished{
		SuccessMessage:  command.SuccessMessage,
		FailureMessage:  command.FailureMessage,
		Target:          command.Target,
		Err:             err,
		Selection:       command.Selection,
		SelectionTarget: command.SelectionTarget,
	})
}

type readPullRequestURLFromClipboardCmd struct{}

func (readPullRequestURLFromClipboardCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	if program.clipboardReader == nil {
		_ = program.dispatch(gui, MsgPullRequestURLReadFromClipboard{Err: ErrClipboardUnavailable})
		return
	}

	url, err := program.clipboardReader.ReadText()
	_ = program.dispatch(gui, MsgPullRequestURLReadFromClipboard{URL: url, Err: err})
}

type modalEditorExternalEditCmd struct {
	Text string
}

func (command modalEditorExternalEditCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	if program.externalEditor == nil {
		_ = program.dispatch(gui, MsgModalEditorExternalEditFinished{Err: errors.New("external editor is unavailable")})
		return
	}

	editedText, err := program.externalEditor.Edit(gui, command.Text)
	_ = program.dispatch(gui, MsgModalEditorExternalEditFinished{Text: editedText, Err: err})
}

type modalEditorSubmitCmd struct {
	Text        string
	Submit      func(string) error
	AfterSubmit func(*Program)
}

func (command modalEditorSubmitCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	if command.Submit == nil {
		_ = program.dispatch(gui, MsgModalEditorSubmitFinished{AfterSubmit: command.AfterSubmit})
		return
	}

	err := command.Submit(command.Text)
	_ = program.dispatch(gui, MsgModalEditorSubmitFinished{Err: err, AfterSubmit: command.AfterSubmit})
}

type pullRequestBuildRunLoadCmd struct {
	Repository string
	Target     pullRequestBuildRunTarget
}

func (command pullRequestBuildRunLoadCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	program.runAsync(func() {
		rawRunOutput, err := program.buildQueries.GetPullRequestBuildRun(command.Repository, command.Target.check)
		jobs := []githubdomain.PullRequestBuildRunJob(nil)
		jobsErr := error(nil)
		if err == nil {
			jobs, jobsErr = program.buildQueries.GetPullRequestBuildRunJobs(command.Repository, command.Target.check)
		}
		program.dispatchAsync(gui, MsgPullRequestBuildRunLoaded{Target: command.Target, RawRunOutput: rawRunOutput, Jobs: jobs, JobsErr: jobsErr, Err: err})
	})
}

type pullRequestBuildRunJobLogLoadCmd struct {
	Repository string
	Check      githubdomain.PullRequestStatusCheck
}

func (command pullRequestBuildRunJobLogLoadCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	program.runAsync(func() {
		job, rawLogOutput, err := program.buildQueries.GetPullRequestBuildRunJobLogForCheck(command.Repository, command.Check)
		program.dispatchAsync(gui, MsgPullRequestBuildRunJobLogLoaded{Repository: command.Repository, Job: job, RawLogOutput: rawLogOutput, Err: err})
	})
}
