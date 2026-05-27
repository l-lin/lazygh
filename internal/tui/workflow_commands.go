package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type workflowCommandDeps struct {
	runAsync                             func(func())
	dispatchAsyncMessage                 func(Msg)
	executeUpdate                        func(Msg)
	executeWorkflowPlan                  func(workflowPlan)
	pullRequestListReloadPlan            func(PullRequestTab) workflowPlan
	pullRequestsFromCache                func(PullRequestTab) ([]githubdomain.PullRequest, bool)
	notificationsFromCache               func() ([]githubdomain.Notification, bool)
	pullRequestDetailFromPersistentCache func(githubdomain.PullRequest) (pullRequestDetailResult, bool)
	pullRequestDiffFromPersistentCache   func(githubdomain.PullRequest) (pullRequestDiffResult, bool)
	pullRequestDetailCached              func(githubdomain.PullRequest) bool
	pullRequestDiffCached                func(githubdomain.PullRequest) bool
	getConnectedUser                     func() (githubdomain.ConnectedUser, error)
	listPullRequests                     func(PullRequestTab) ([]githubdomain.PullRequest, error)
	listNotifications                    func() ([]githubdomain.Notification, error)
	getPullRequestDetail                 func(githubdomain.PullRequest) (githubdomain.PullRequestDetail, error)
	getPendingPullRequestReviewState     func(githubdomain.PullRequest) (pendingPullRequestReviewState, bool)
	getPullRequestDiff                   func(githubdomain.PullRequest) (githubdomain.PullRequestDiff, error)
	shouldLoadPullRequestDiffTeamOwners  bool
	getPullRequestFileTeamOwners         func(string, int, []string) (map[string][]string, error)
	getIssueDetail                       func(string, int) (githubdomain.IssueDetail, error)
	getReleaseDetail                     func(string, int) (githubdomain.ReleaseDetail, error)
	renderMarkdownHTML                   func(string, string) (string, error)
	loadDetailImage                      func(string) (loadedDetailImage, error)
}

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

type hydratePullRequestDetailFromCacheCmd struct {
	summary githubdomain.PullRequest
}

type hydratePullRequestDiffFromCacheCmd struct {
	summary githubdomain.PullRequest
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

func newWorkflowCommandDeps(program *Program, gui *gocui.Gui) workflowCommandDeps {
	if program == nil {
		return workflowCommandDeps{}
	}

	capturedGUI := program.captureGUI(gui)
	deps := workflowCommandDeps{
		runAsync:             program.runAsync,
		dispatchAsyncMessage: program.dispatchAsyncMessage,
		executeUpdate: func(msg Msg) {
			program.executeCmds(capturedGUI, Update(program, msg))
		},
		executeWorkflowPlan: func(plan workflowPlan) {
			program.executeWorkflowPlan(capturedGUI, plan)
		},
		pullRequestListReloadPlan:            program.pullRequestListReloadPlan,
		pullRequestsFromCache:                program.pullRequestsFromCache,
		notificationsFromCache:               program.notificationsFromCache,
		pullRequestDetailFromPersistentCache: program.pullRequestDetailFromPersistentCache,
		pullRequestDiffFromPersistentCache:   program.pullRequestDiffFromPersistentCache,
		pullRequestDetailCached: func(summary githubdomain.PullRequest) bool {
			key := pullRequestDetailKey(summary.Repository, summary.Number)
			if key == "" {
				return false
			}
			_, ok := program.pullRequestDetailCache[key]
			return ok
		},
		pullRequestDiffCached: func(summary githubdomain.PullRequest) bool {
			key := pullRequestDetailKey(summary.Repository, summary.Number)
			if key == "" {
				return false
			}
			_, ok := program.pullRequestDiffCache[key]
			return ok
		},
	}
	if program.sessionQueries != nil {
		deps.getConnectedUser = program.sessionQueries.GetConnectedUser
	}
	if program.pullRequestListQueries != nil {
		deps.listPullRequests = program.listPullRequests
	}
	if program.notificationQueries != nil {
		deps.listNotifications = program.notificationQueries.ListNotifications
		deps.getIssueDetail = program.notificationQueries.GetIssueDetail
		deps.getReleaseDetail = program.notificationQueries.GetReleaseDetail
	}
	if program.detailQueries != nil {
		deps.getPullRequestDetail = func(summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, error) {
			repository := pullRequestRepositoryName(summary.Repository)
			return program.detailQueries.GetPullRequestDetail(repository, summary.Number)
		}
		deps.shouldLoadPullRequestDiffTeamOwners = program.shouldLoadPullRequestDiffTeamOwners()
		deps.getPullRequestFileTeamOwners = program.detailQueries.GetPullRequestFileTeamOwners
		deps.getPullRequestDiff = func(summary githubdomain.PullRequest) (githubdomain.PullRequestDiff, error) {
			repository := pullRequestRepositoryName(summary.Repository)
			rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
			if err == nil {
				rawDiff = loadPullRequestDiffFileTeamOwners(deps, repository, summary.Number, rawDiff)
			}
			return rawDiff, err
		}
	}
	if program.reviewMutations != nil {
		deps.getPendingPullRequestReviewState = func(summary githubdomain.PullRequest) (pendingPullRequestReviewState, bool) {
			repository := pullRequestRepositoryName(summary.Repository)
			pendingReviewID, found, err := program.reviewMutations.GetPendingPullRequestReviewID(repository, summary.Number)
			if err != nil {
				return pendingPullRequestReviewState{}, false
			}
			state := pendingPullRequestReviewState{}
			if found {
				state.id = strings.TrimSpace(pendingReviewID)
			}
			return state, true
		}
	}
	if program.markdownHTMLRenderer != nil {
		deps.renderMarkdownHTML = program.markdownHTMLRenderer.RenderMarkdownHTML
	}
	deps.loadDetailImage = func(imageURL string) (loadedDetailImage, error) {
		githubToken := ""
		if isGitHubImageSource(imageURL) {
			githubToken = program.detailImageAuthToken()
		}
		return loadDetailImage(imageURL, program.imageHTTPClient, githubToken)
	}
	return deps
}

func (loadConnectedUserCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.getConnectedUser == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	runWorkflowCommandAsync(deps, func() {
		user, err := deps.getConnectedUser()
		deps.dispatchAsyncMessage(MsgConnectedUserLoaded{User: user, Err: err})
	})
}

func (command loadPullRequestsCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.listPullRequests == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	runWorkflowCommandAsync(deps, func() {
		pullRequests, err := deps.listPullRequests(command.tab)
		deps.dispatchAsyncMessage(MsgPullRequestsLoaded{Tab: command.tab, PullRequests: pullRequests, Err: err})
	})
}

func (command hydratePullRequestsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.pullRequestsFromCache == nil || deps.executeUpdate == nil {
		return
	}
	pullRequests, ok := deps.pullRequestsFromCache(command.tab)
	if !ok {
		return
	}
	deps.executeUpdate(MsgPullRequestsCacheHydrated{Tab: command.tab, PullRequests: pullRequests})
}

func (command reloadPullRequestsTabCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.executeWorkflowPlan == nil || deps.pullRequestListReloadPlan == nil {
		return
	}
	deps.executeWorkflowPlan(deps.pullRequestListReloadPlan(command.tab))
}

func (loadNotificationsCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.listNotifications == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	runWorkflowCommandAsync(deps, func() {
		notifications, err := deps.listNotifications()
		deps.dispatchAsyncMessage(MsgNotificationsLoaded{Notifications: notifications, Err: err})
	})
}

func (command hydratePullRequestDetailFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.pullRequestDetailCached == nil || deps.pullRequestDetailFromPersistentCache == nil || deps.executeUpdate == nil {
		return
	}
	if deps.pullRequestDetailCached(command.summary) {
		return
	}
	result, ok := deps.pullRequestDetailFromPersistentCache(command.summary)
	if !ok {
		return
	}
	deps.executeUpdate(MsgPullRequestDetailCacheHydrated{Summary: command.summary, Result: result})
}

func (command loadPullRequestDetailCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.getPullRequestDetail == nil || deps.dispatchAsyncMessage == nil {
		return
	}

	summary := command.summary
	runWorkflowCommandAsync(deps, func() {
		deps.dispatchAsyncMessage(loadPullRequestDetailResult(deps, summary))
	})
}

func loadPullRequestDetailResult(deps workflowCommandDeps, summary githubdomain.PullRequest) MsgPullRequestDetailLoaded {
	detail, err := deps.getPullRequestDetail(summary)
	pendingReviewState := pendingPullRequestReviewState{}
	pendingReviewStateKnown := false
	if deps.getPendingPullRequestReviewState != nil {
		pendingReviewState, pendingReviewStateKnown = deps.getPendingPullRequestReviewState(summary)
	}
	return MsgPullRequestDetailLoaded{
		Summary:                 summary,
		Detail:                  detail,
		Err:                     err,
		PendingReviewState:      pendingReviewState,
		PendingReviewStateKnown: pendingReviewStateKnown,
	}
}

func (command hydratePullRequestDiffFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.pullRequestDiffCached == nil || deps.pullRequestDiffFromPersistentCache == nil || deps.executeUpdate == nil {
		return
	}
	if deps.pullRequestDiffCached(command.summary) {
		return
	}
	result, ok := deps.pullRequestDiffFromPersistentCache(command.summary)
	if !ok {
		return
	}
	deps.executeUpdate(MsgPullRequestDiffCacheHydrated{Summary: command.summary, Result: result})
}

func (command loadPullRequestDiffCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.getPullRequestDiff == nil || deps.dispatchAsyncMessage == nil {
		return
	}

	summary := command.summary
	runWorkflowCommandAsync(deps, func() {
		deps.dispatchAsyncMessage(loadPullRequestDiffResult(deps, summary))
	})
}

func loadPullRequestDiffResult(deps workflowCommandDeps, summary githubdomain.PullRequest) MsgPullRequestDiffLoaded {
	rawDiff, err := deps.getPullRequestDiff(summary)
	return MsgPullRequestDiffLoaded{Summary: summary, RawDiff: rawDiff, Err: err}
}

func loadPullRequestDiffFileTeamOwners(deps workflowCommandDeps, repository string, number int, rawDiff githubdomain.PullRequestDiff) githubdomain.PullRequestDiff {
	if rawDiff.FileTeamOwnersAttempted || !deps.shouldLoadPullRequestDiffTeamOwners || deps.getPullRequestFileTeamOwners == nil {
		return rawDiff
	}

	rawDiff.FileTeamOwnersAttempted = true
	filePaths := pullRequestDiffFilePaths(rawDiff.Files)
	if len(filePaths) == 0 {
		return rawDiff
	}

	teamOwnersByPath, err := deps.getPullRequestFileTeamOwners(repository, number, filePaths)
	if err != nil || len(teamOwnersByPath) == 0 {
		return rawDiff
	}

	rawDiff.Files = pullRequestDiffFilesWithTeamOwners(rawDiff.Files, teamOwnersByPath)
	return rawDiff
}

func (command loadIssueDetailCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.getIssueDetail == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	repository := command.repository
	number := command.number
	runWorkflowCommandAsync(deps, func() {
		detail, err := deps.getIssueDetail(repository, number)
		deps.dispatchAsyncMessage(MsgIssueDetailLoaded{Repository: repository, Number: number, Detail: detail, Err: err})
	})
}

func (command loadReleaseDetailCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.getReleaseDetail == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	repository := command.repository
	id := command.id
	runWorkflowCommandAsync(deps, func() {
		detail, err := deps.getReleaseDetail(repository, id)
		deps.dispatchAsyncMessage(MsgReleaseDetailLoaded{Repository: repository, ID: id, Detail: detail, Err: err})
	})
}

func (command loadCurrentDetailImageHTMLCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.renderMarkdownHTML == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	source := command.source
	runWorkflowCommandAsync(deps, func() {
		renderedHTML, err := deps.renderMarkdownHTML(source.repository, source.markdown)
		deps.dispatchAsyncMessage(MsgCurrentDetailImageHTMLLoaded{Source: source, RenderedHTML: renderedHTML, Err: err})
	})
}

func (command loadCurrentDetailImageCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.loadDetailImage == nil || deps.dispatchAsyncMessage == nil {
		return
	}
	imageURL := command.imageURL
	runWorkflowCommandAsync(deps, func() {
		loadedImage, err := deps.loadDetailImage(imageURL)
		deps.dispatchAsyncMessage(MsgCurrentDetailImageLoaded{ImageURL: imageURL, Image: loadedImage, Err: err})
	})
}

func (hydrateNotificationsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	deps := newWorkflowCommandDeps(program, gui)
	if deps.notificationsFromCache == nil || deps.executeUpdate == nil {
		return
	}
	notifications, ok := deps.notificationsFromCache()
	if !ok {
		return
	}
	deps.executeUpdate(MsgNotificationsCacheHydrated{Notifications: notifications})
}

func runWorkflowCommandAsync(deps workflowCommandDeps, run func()) {
	if run == nil {
		return
	}
	if deps.runAsync != nil {
		deps.runAsync(run)
		return
	}
	run()
}
