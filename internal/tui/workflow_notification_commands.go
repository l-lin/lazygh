package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type notificationWorkflowRuntime struct {
	workflowShellRuntime
	notificationsFromCache func() ([]githubdomain.Notification, bool)
	listNotifications      func() ([]githubdomain.Notification, error)
	getIssueDetail         func(string, int) (githubdomain.IssueDetail, error)
	getReleaseDetail       func(string, int) (githubdomain.ReleaseDetail, error)
}

type loadNotificationsCmd struct {
	statusLineOperationID uint64
}

type loadIssueDetailCmd struct {
	repository string
	number     int
}

type loadReleaseDetailCmd struct {
	repository string
	id         int
}

type hydrateNotificationsFromCacheCmd struct{}

func newNotificationWorkflowRuntime(program *Program, gui *gocui.Gui) notificationWorkflowRuntime {
	runtime := notificationWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program != nil {
		runtime.notificationsFromCache = program.notificationsFromCache
		if program.notificationQueries != nil {
			runtime.listNotifications = program.notificationQueries.ListNotifications
			runtime.getIssueDetail = program.notificationQueries.GetIssueDetail
			runtime.getReleaseDetail = program.notificationQueries.GetReleaseDetail
		}
	}
	return runtime
}

func (command loadNotificationsCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newNotificationWorkflowRuntime(program, gui)
	if runtime.listNotifications == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	statusLineOperationID := command.statusLineOperationID
	if statusLineOperationID == 0 && program != nil {
		statusLineOperationID = program.notificationsStatusLineOperationID()
	}
	runWorkflowCommandAsync(runtime.runAsync, func() {
		notifications, err := runtime.listNotifications()
		runtime.dispatchAsyncMessage(MsgNotificationsLoaded{Notifications: notifications, Err: err, StatusLineOperationID: statusLineOperationID})
	})
}

func (hydrateNotificationsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newNotificationWorkflowRuntime(program, gui)
	if runtime.notificationsFromCache == nil || runtime.executeUpdate == nil {
		return
	}
	notifications, ok := runtime.notificationsFromCache()
	if !ok {
		return
	}
	runtime.executeUpdate(MsgNotificationsCacheHydrated{Notifications: notifications})
}

func (command loadIssueDetailCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newNotificationWorkflowRuntime(program, gui)
	if runtime.getIssueDetail == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	repository := command.repository
	number := command.number
	runWorkflowCommandAsync(runtime.runAsync, func() {
		detail, err := runtime.getIssueDetail(repository, number)
		runtime.dispatchAsyncMessage(MsgIssueDetailLoaded{Repository: repository, Number: number, Detail: detail, Err: err})
	})
}

func (command loadReleaseDetailCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newNotificationWorkflowRuntime(program, gui)
	if runtime.getReleaseDetail == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	repository := command.repository
	id := command.id
	runWorkflowCommandAsync(runtime.runAsync, func() {
		detail, err := runtime.getReleaseDetail(repository, id)
		runtime.dispatchAsyncMessage(MsgReleaseDetailLoaded{Repository: repository, ID: id, Detail: detail, Err: err})
	})
}
