package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

type issueDetailResult struct {
	detail githubcli.IssueDetail
	err    error
}

type releaseDetailResult struct {
	detail githubcli.ReleaseDetail
	err    error
}

func (program *Program) maybeLoadSelectedNotificationDetail(gui *gocui.Gui) {
	program.executeWorkflowCommands(gui, program.detailStore.planSelectedNotificationDetailLoad(program, gui))
}

func (program *Program) loadIssueDetail(gui *gocui.Gui, repository string, number int) {
	detail, err := program.notificationQueries.GetIssueDetail(repository, number)
	key := notificationDetailKey(repository, number)
	result := issueDetailResult{err: err}
	if err == nil {
		result.detail = githubcli.IssueDetailFromDomain(detail)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.issueDetailLoadInFlight, key)
		program.issueDetailCache[key] = result
		return program.refreshViews(gui)
	})
}

func (program *Program) loadReleaseDetail(gui *gocui.Gui, repository string, id int) {
	detail, err := program.notificationQueries.GetReleaseDetail(repository, id)
	key := notificationDetailKey(repository, id)
	result := releaseDetailResult{err: err}
	if err == nil {
		result.detail = githubcli.ReleaseDetailFromDomain(detail)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.releaseDetailLoadInFlight, key)
		program.releaseDetailCache[key] = result
		return program.refreshViews(gui)
	})
}

func (program *Program) issueDetailLoaded(key string) bool {
	_, ok := program.issueDetailCache[key]
	return ok
}

func (program *Program) releaseDetailLoaded(key string) bool {
	_, ok := program.releaseDetailCache[key]
	return ok
}

func (program *Program) issueDetailForNotification(notification githubcli.Notification) (issueDetailResult, bool) {
	repository, number, ok := notification.IssueIdentity()
	if !ok {
		return issueDetailResult{}, false
	}
	result, ok := program.issueDetailCache[notificationDetailKey(repository, number)]
	return result, ok
}

func (program *Program) releaseDetailForNotification(notification githubcli.Notification) (releaseDetailResult, bool) {
	repository, id, ok := notification.ReleaseIdentity()
	if !ok {
		return releaseDetailResult{}, false
	}
	result, ok := program.releaseDetailCache[notificationDetailKey(repository, id)]
	return result, ok
}

func (program *Program) selectedNotificationDetailLoading() bool {
	if program.reviewSession.active || program.model.currentSideFocus() != FocusNotificationsView {
		return false
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return false
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return program.issueDetailLoadInFlight[notificationDetailKey(repository, number)]
	}
	if repository, id, ok := notification.ReleaseIdentity(); ok {
		return program.releaseDetailLoadInFlight[notificationDetailKey(repository, id)]
	}
	return false
}

func (program *Program) selectedNotificationDetailLoadingStatus() string {
	if !program.selectedNotificationDetailLoading() {
		return ""
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return ""
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return fmt.Sprintf("Running `gh api repos/%s/issues/%d`.", repository, number)
	}
	if repository, id, ok := notification.ReleaseIdentity(); ok {
		return fmt.Sprintf("Running `gh api repos/%s/releases/%d`.", repository, id)
	}
	return ""
}

func notificationDetailKey(repository string, id int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || id <= 0 {
		return ""
	}
	return fmt.Sprintf("%s#%d", trimmedRepository, id)
}
