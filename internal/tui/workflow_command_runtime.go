package tui

import "github.com/jesseduffield/gocui"

type workflowShellRuntime struct {
	runAsync                  func(func())
	dispatchAsyncMessage      func(Msg)
	executeUpdate             func(Msg)
	executeWorkflowPlan       func(workflowPlan)
	pullRequestListReloadPlan func(PullRequestTab) workflowPlan
}

func newWorkflowShellRuntime(program *Program, gui *gocui.Gui) workflowShellRuntime {
	if program == nil {
		return workflowShellRuntime{}
	}

	capturedGUI := program.captureGUI(gui)
	return workflowShellRuntime{
		runAsync:             program.runAsync,
		dispatchAsyncMessage: program.dispatchAsyncMessage,
		executeUpdate: func(msg Msg) {
			_ = program.executeRuntimeMessage(capturedGUI, msg)
		},
		executeWorkflowPlan: func(plan workflowPlan) {
			program.executeWorkflowPlan(capturedGUI, plan)
		},
		pullRequestListReloadPlan: program.pullRequestListReloadPlan,
	}
}

func runWorkflowCommandAsync(runAsync func(func()), run func()) {
	if run == nil {
		return
	}
	if runAsync != nil {
		runAsync(run)
		return
	}
	run()
}
