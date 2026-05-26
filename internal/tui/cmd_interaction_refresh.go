package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestListRefreshCommandRuntime struct {
	hasPullRequestListQueries        func() bool
	markManualPullRequestListRefresh func(PullRequestTab) bool
	beginManualRefresh               func(string, int)
}

func newPullRequestListRefreshCommandRuntime(program *Program) pullRequestListRefreshCommandRuntime {
	if program == nil {
		return pullRequestListRefreshCommandRuntime{}
	}
	return pullRequestListRefreshCommandRuntime{
		hasPullRequestListQueries:        program.hasPullRequestListQueries,
		markManualPullRequestListRefresh: program.markManualPullRequestListRefresh,
		beginManualRefresh:               program.beginManualRefresh,
	}
}

type pullRequestRefreshCommandRuntime struct {
	hasDetailQueries                   func() bool
	reviewModeActive                   func() bool
	hasPullRequestListQueries          func() bool
	activePullRequestTab               func() PullRequestTab
	markManualPullRequestDetailRefresh func(githubdomain.PullRequest) bool
	markManualPullRequestDiffRefresh   func(githubdomain.PullRequest) bool
	markManualPullRequestListRefresh   func(PullRequestTab) bool
	beginManualRefresh                 func(string, int)
}

func newPullRequestRefreshCommandRuntime(program *Program) pullRequestRefreshCommandRuntime {
	if program == nil {
		return pullRequestRefreshCommandRuntime{}
	}
	return pullRequestRefreshCommandRuntime{
		hasDetailQueries:                   program.hasDetailQueries,
		reviewModeActive:                   program.reviewModeActive,
		hasPullRequestListQueries:          program.hasPullRequestListQueries,
		activePullRequestTab:               program.model.ActivePullRequestTab,
		markManualPullRequestDetailRefresh: program.markManualPullRequestDetailRefresh,
		markManualPullRequestDiffRefresh:   program.markManualPullRequestDiffRefresh,
		markManualPullRequestListRefresh:   program.markManualPullRequestListRefresh,
		beginManualRefresh:                 program.beginManualRefresh,
	}
}

type notificationRefreshCommandRuntime struct {
	reviewModeActive              func() bool
	hasNotificationQueries        func() bool
	markManualNotificationRefresh func() bool
	beginManualRefresh            func(string, int)
	reloadNotifications           func(*gocui.Gui)
}

func newNotificationRefreshCommandRuntime(program *Program) notificationRefreshCommandRuntime {
	if program == nil {
		return notificationRefreshCommandRuntime{}
	}
	return notificationRefreshCommandRuntime{
		reviewModeActive:              program.reviewModeActive,
		hasNotificationQueries:        program.hasNotificationQueries,
		markManualNotificationRefresh: program.markManualNotificationRefresh,
		beginManualRefresh:            program.beginManualRefresh,
		reloadNotifications:           program.reloadNotifications,
	}
}

type beginManualPullRequestListRefreshCmd struct {
	Tab            PullRequestTab
	SuccessMessage string
}

func (command beginManualPullRequestListRefreshCmd) execute(program *Program, gui *gocui.Gui) {
	executeBeginManualPullRequestListRefreshCommand(newPullRequestListRefreshCommandRuntime(program), gui, command)
}

func executeBeginManualPullRequestListRefreshCommand(runtime pullRequestListRefreshCommandRuntime, gui *gocui.Gui, command beginManualPullRequestListRefreshCmd) {
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
	executeBeginManualPullRequestRefreshCommand(newPullRequestRefreshCommandRuntime(program), gui, command)
}

func executeBeginManualPullRequestRefreshCommand(runtime pullRequestRefreshCommandRuntime, gui *gocui.Gui, command beginManualPullRequestRefreshCmd) {
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
	executeRefreshNotificationsCommand(newNotificationRefreshCommandRuntime(program), gui)
}

func executeRefreshNotificationsCommand(runtime notificationRefreshCommandRuntime, gui *gocui.Gui) {
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
