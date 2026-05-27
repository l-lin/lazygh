package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type buildCommandRuntime struct {
	buildQueries         BuildQueries
	runAsync             func(func())
	dispatchAsyncMessage func(Msg)
}

func newBuildCommandRuntime(program *Program, gui *gocui.Gui) buildCommandRuntime {
	if program == nil {
		return buildCommandRuntime{}
	}
	program.captureGUI(gui)
	return buildCommandRuntime{
		buildQueries:         program.buildQueries,
		runAsync:             program.runAsync,
		dispatchAsyncMessage: program.dispatchAsyncMessage,
	}
}

type pullRequestBuildRunLoadCmd struct {
	Repository string
	Target     pullRequestBuildRunTarget
}

func (command pullRequestBuildRunLoadCmd) execute(program *Program, gui *gocui.Gui) {
	executePullRequestBuildRunLoadCommand(newBuildCommandRuntime(program, gui), command)
}

func executePullRequestBuildRunLoadCommand(runtime buildCommandRuntime, command pullRequestBuildRunLoadCmd) {
	if runtime.buildQueries == nil || runtime.dispatchAsyncMessage == nil {
		return
	}

	run := func() {
		rawRunOutput, err := runtime.buildQueries.GetPullRequestBuildRun(command.Repository, command.Target.check)
		jobs := []githubdomain.PullRequestBuildRunJob(nil)
		jobsErr := error(nil)
		if err == nil {
			jobs, jobsErr = runtime.buildQueries.GetPullRequestBuildRunJobs(command.Repository, command.Target.check)
		}
		runtime.dispatchAsyncMessage(MsgPullRequestBuildRunLoaded{Target: command.Target, RawRunOutput: rawRunOutput, Jobs: jobs, JobsErr: jobsErr, Err: err})
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
	executePullRequestBuildRunJobLogLoadCommand(newBuildCommandRuntime(program, gui), command)
}

func executePullRequestBuildRunJobLogLoadCommand(runtime buildCommandRuntime, command pullRequestBuildRunJobLogLoadCmd) {
	if runtime.buildQueries == nil || runtime.dispatchAsyncMessage == nil {
		return
	}

	run := func() {
		job, rawLogOutput, err := runtime.buildQueries.GetPullRequestBuildRunJobLogForCheck(command.Repository, command.Check)
		runtime.dispatchAsyncMessage(MsgPullRequestBuildRunJobLogLoaded{Repository: command.Repository, Job: job, RawLogOutput: rawLogOutput, Err: err})
	}
	if runtime.runAsync != nil {
		runtime.runAsync(run)
		return
	}
	run()
}
