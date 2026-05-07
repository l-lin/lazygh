package tui

import (
	"errors"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func (program *Program) OpenPullRequestByURL(rawURL string) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	summary, err := githubcli.ParsePullRequestURL(rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestInBrowser(summary)
}

func (program *Program) openPullRequestInBrowser(summary githubcli.PullRequest) error {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return errors.New("missing pull request identity")
	}

	summary.Repository.NameWithOwner = repository
	program.pinOpenedPullRequestSummary(MyPullRequestsTab, summary)
	program.model.activePullRequestTab = MyPullRequestsTab
	program.model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	program.model.selectedPullRequestIndexes[MyPullRequestsTab] = 0
	program.setPullRequestsLoadStarted(MyPullRequestsTab, true)
	program.setPullRequestsLoading(MyPullRequestsTab, false)
	program.setPullRequestsCount(MyPullRequestsTab, 1, true)
	program.reviewSession = reviewSessionState{}
	program.invalidateReviewDiffRenderCache()
	program.activeDetailTab = DescriptionDetailTab
	program.detailViewState.reset()
	program.detailViewState.clearPendingPrefix()
	program.clearPendingSelectionPrefix()
	program.invalidatePullRequestDetailDocumentCache()
	program.showOpenedPullRequestInDetailFullscreen()

	if program.gui != nil {
		return program.refreshViews(program.gui)
	}
	return nil
}

func (program *Program) showOpenedPullRequestInDetailFullscreen() {
	returnSize := program.model.paneLayoutSize
	if returnSize == PaneLayoutFullscreen {
		returnSize = PaneLayoutDefault
	}
	program.model.detailFullscreenReturnSize = returnSize
	program.model.paneLayoutSize = PaneLayoutFullscreen
	program.model.fullscreenPane = FocusDetailView
	program.model.lastSideFocus = FocusPullRequestsView
	program.model.focus = FocusDetailView
}

func (program *Program) pinOpenedPullRequestSummary(tab PullRequestTab, summary githubcli.PullRequest) {
	summaryCopy := summary
	program.openedPullRequestSummary = &summaryCopy
	program.openedPullRequestTab = tab
}

func (program *Program) openedPullRequestSummaryForTab(tab PullRequestTab) (githubcli.PullRequest, bool) {
	if program.openedPullRequestSummary == nil || program.openedPullRequestTab != tab {
		return githubcli.PullRequest{}, false
	}
	return *program.openedPullRequestSummary, true
}

func (program *Program) pullRequestsWithOpenedPullRequestSummary(tab PullRequestTab, pullRequests []githubcli.PullRequest) []githubcli.PullRequest {
	openedSummary, ok := program.openedPullRequestSummaryForTab(tab)
	if !ok {
		return pullRequests
	}

	updatedPullRequests := append([]githubcli.PullRequest(nil), pullRequests...)
	for index, pullRequest := range updatedPullRequests {
		if !samePullRequestIdentity(pullRequest, openedSummary) {
			continue
		}
		program.pinOpenedPullRequestSummary(tab, pullRequest)
		updatedPullRequests[index] = pullRequest
		return updatedPullRequests
	}

	return append([]githubcli.PullRequest{openedSummary}, updatedPullRequests...)
}

func (program *Program) selectOpenedPullRequestRow(tab PullRequestTab) {
	openedSummary, ok := program.openedPullRequestSummaryForTab(tab)
	if !ok {
		return
	}

	rows := program.model.PullRequestRows(tab)
	for index, row := range rows {
		if row.Summary == nil || !samePullRequestIdentity(*row.Summary, openedSummary) {
			continue
		}
		program.model.selectedPullRequestIndexes[tab] = index
		return
	}
}

func pullRequestSummaryRowCount(rows []PullRequestRow) int {
	count := 0
	for _, row := range rows {
		if row.Summary != nil {
			count++
		}
	}
	return count
}

func samePullRequestIdentity(left githubcli.PullRequest, right githubcli.PullRequest) bool {
	leftKey := pullRequestDetailKey(left.Repository, left.Number)
	rightKey := pullRequestDetailKey(right.Repository, right.Number)
	return leftKey != "" && leftKey == rightKey
}
