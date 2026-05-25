package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type loadConnectedUserCmd struct{}

type loadPullRequestsCmd struct {
	tab PullRequestTab
}

type reloadPullRequestsTabCmd struct {
	tab PullRequestTab
}

type loadNotificationsCmd struct{}

type loadPullRequestDetailCmd struct {
	summary githubdomain.PullRequest
}

type loadPullRequestDiffCmd struct {
	summary githubdomain.PullRequest
}

type loadIssueDetailCmd struct {
	repository string
	number     int
}

type loadReleaseDetailCmd struct {
	repository string
	id         int
}

type loadCurrentDetailImageHTMLCmd struct {
	source detailImageHTMLSource
}

type loadCurrentDetailImageCmd struct {
	imageURL string
}

func (loadConnectedUserCmd) execute(program *Program, gui *gocui.Gui) {
	program.runAsync(func() {
		program.loadConnectedUser(gui)
	})
}

func (command loadPullRequestsCmd) execute(program *Program, gui *gocui.Gui) {
	program.runAsync(func() {
		program.loadPullRequests(gui, command.tab)
	})
}

func (command reloadPullRequestsTabCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	program.executeCmds(gui, program.pullRequestListStore.planReload(program, gui, command.tab))
}

func (loadNotificationsCmd) execute(program *Program, gui *gocui.Gui) {
	program.runAsync(func() {
		program.loadNotifications(gui)
	})
}

func (command loadPullRequestDetailCmd) execute(program *Program, gui *gocui.Gui) {
	summary := command.summary
	program.runAsync(func() {
		program.loadPullRequestDetail(gui, summary)
	})
}

func (command loadPullRequestDiffCmd) execute(program *Program, gui *gocui.Gui) {
	summary := command.summary
	program.runAsync(func() {
		program.loadPullRequestDiff(gui, summary)
	})
}

func (command loadIssueDetailCmd) execute(program *Program, gui *gocui.Gui) {
	repository := command.repository
	number := command.number
	program.runAsync(func() {
		program.loadIssueDetail(gui, repository, number)
	})
}

func (command loadReleaseDetailCmd) execute(program *Program, gui *gocui.Gui) {
	repository := command.repository
	id := command.id
	program.runAsync(func() {
		program.loadReleaseDetail(gui, repository, id)
	})
}

func (command loadCurrentDetailImageHTMLCmd) execute(program *Program, gui *gocui.Gui) {
	source := command.source
	program.runAsync(func() {
		program.loadCurrentDetailImageHTML(gui, source)
	})
}

func (command loadCurrentDetailImageCmd) execute(program *Program, gui *gocui.Gui) {
	imageURL := command.imageURL
	program.runAsync(func() {
		program.loadCurrentDetailImage(gui, imageURL)
	})
}

func (program *Program) runAsync(run func()) {
	if program == nil || run == nil {
		return
	}
	if program.asyncRunner == nil {
		run()
		return
	}
	program.asyncRunner.Go(run)
}
