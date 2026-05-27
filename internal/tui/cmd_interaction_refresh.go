package tui

import "github.com/jesseduffield/gocui"

type notificationRefreshCommandRuntime struct {
	executeWorkflowPlan func(workflowPlan)
	notificationReload  func() workflowPlan
}

func newNotificationRefreshCommandRuntime(program *Program, gui *gocui.Gui) notificationRefreshCommandRuntime {
	if program == nil {
		return notificationRefreshCommandRuntime{}
	}

	capturedGUI := program.captureGUI(gui)
	return notificationRefreshCommandRuntime{
		executeWorkflowPlan: func(plan workflowPlan) {
			program.executeWorkflowPlan(capturedGUI, plan)
		},
		notificationReload: program.notificationReloadPlan,
	}
}

type refreshNotificationsCmd struct{}

func (refreshNotificationsCmd) execute(program *Program, gui *gocui.Gui) {
	executeRefreshNotificationsCommand(newNotificationRefreshCommandRuntime(program, gui))
}

func executeRefreshNotificationsCommand(runtime notificationRefreshCommandRuntime) {
	if runtime.executeWorkflowPlan == nil || runtime.notificationReload == nil {
		return
	}
	runtime.executeWorkflowPlan(runtime.notificationReload())
}
