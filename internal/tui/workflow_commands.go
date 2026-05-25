package tui

import (
	"strings"

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

type hydratePullRequestsFromCacheCmd struct {
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

type hydrateNotificationsFromCacheCmd struct{}

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

func (command hydratePullRequestsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	pullRequests, ok := program.pullRequestsFromCache(command.tab)
	if !ok {
		return
	}
	program.executeCmds(gui, Update(program, MsgPullRequestsCacheHydrated{Tab: command.tab, PullRequests: pullRequests}))
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
	if program == nil || !program.hasDetailQueries() {
		return
	}

	summary := command.summary
	program.runAsync(func() {
		program.dispatchAsync(gui, loadPullRequestDetailResult(program, summary))
	})
}

func loadPullRequestDetailResult(program *Program, summary githubdomain.PullRequest) MsgPullRequestDetailLoaded {
	repository := pullRequestRepositoryName(summary.Repository)
	detail, err := program.detailQueries.GetPullRequestDetail(repository, summary.Number)
	pendingReviewState := pendingPullRequestReviewState{}
	pendingReviewStateKnown := false
	if program.hasReviewMutations() {
		if pendingReviewID, found, pendingReviewErr := program.reviewMutations.GetPendingPullRequestReviewID(repository, summary.Number); pendingReviewErr == nil {
			pendingReviewStateKnown = true
			if found {
				pendingReviewState.id = strings.TrimSpace(pendingReviewID)
			}
		}
	}
	return MsgPullRequestDetailLoaded{
		Summary:                 summary,
		Detail:                  detail,
		Err:                     err,
		PendingReviewState:      pendingReviewState,
		PendingReviewStateKnown: pendingReviewStateKnown,
	}
}

func (command loadPullRequestDiffCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || !program.hasDetailQueries() {
		return
	}

	summary := command.summary
	program.runAsync(func() {
		program.dispatchAsync(gui, loadPullRequestDiffResult(program, summary))
	})
}

func loadPullRequestDiffResult(program *Program, summary githubdomain.PullRequest) MsgPullRequestDiffLoaded {
	repository := pullRequestRepositoryName(summary.Repository)
	rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
	if err == nil {
		rawDiff = loadPullRequestDiffFileTeamOwners(program, repository, summary.Number, rawDiff)
	}
	return MsgPullRequestDiffLoaded{Summary: summary, RawDiff: rawDiff, Err: err}
}

func loadPullRequestDiffFileTeamOwners(program *Program, repository string, number int, rawDiff githubdomain.PullRequestDiff) githubdomain.PullRequestDiff {
	if program == nil || rawDiff.FileTeamOwnersAttempted || !program.hasDetailQueries() || !program.shouldLoadPullRequestDiffTeamOwners() {
		return rawDiff
	}

	rawDiff.FileTeamOwnersAttempted = true
	filePaths := pullRequestDiffFilePaths(rawDiff.Files)
	if len(filePaths) == 0 {
		return rawDiff
	}

	teamOwnersByPath, err := program.detailQueries.GetPullRequestFileTeamOwners(repository, number, filePaths)
	if err != nil || len(teamOwnersByPath) == 0 {
		return rawDiff
	}

	rawDiff.Files = pullRequestDiffFilesWithTeamOwners(rawDiff.Files, teamOwnersByPath)
	return rawDiff
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

func (hydrateNotificationsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	notifications, ok := program.notificationsFromCache()
	if !ok {
		return
	}
	program.executeCmds(gui, Update(program, MsgNotificationsCacheHydrated{Notifications: notifications}))
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
