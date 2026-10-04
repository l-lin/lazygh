package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestScheduledRefresh_GivenThePullRequestWasCheckedAtTheRemoteVersion_WhenTheListFinishes_ThenItDoesNotPrefetchDetailOrDiff(t *testing.T) {
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Title:      "Checked PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		UpdatedAt:  "2026-05-05T13:00:00Z",
	})
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", summary.UpdatedAt)
	})

	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{sourceUpdatedAt: summary.UpdatedAt, fileTeamOwnersAttempted: true}

	actual := subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Source:       pullRequestLoadSourceScheduled,
		PullRequests: []githubdomain.PullRequest{summary},
	})

	if len(actual) != 1 {
		t.Fatalf("expected one scheduled refresh command, actual %d", len(actual))
	}
	fanout, ok := actual[0].(scheduledPullRequestRefreshCmd)
	if !ok {
		t.Fatalf("expected a scheduled refresh command, actual %T", actual[0])
	}
	if len(fanout.detailSummaries) != 0 || len(fanout.diffSummaries) != 0 {
		t.Fatalf("expected no detail or diff prefetch targets, actual detail=%v diff=%v", fanout.detailSummaries, fanout.diffSummaries)
	}
}

func TestScheduledRefresh_GivenAListPullRequestWasCheckedBeforeItsLatestUpdate_WhenTheListFinishes_ThenItPrefetchesDescriptionAndDiff(t *testing.T) {
	initial := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Title:      "Previously checked PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		UpdatedAt:  "2026-05-05T12:00:00Z",
	})
	latest := initial
	latest.UpdatedAt = "2026-05-05T13:00:00Z"
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", initial.UpdatedAt)
	})

	actual := subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab:          MyPullRequestsTab,
		Source:       pullRequestLoadSourceScheduled,
		PullRequests: []githubdomain.PullRequest{latest},
	})

	if len(actual) != 1 {
		t.Fatalf("expected one scheduled refresh command, actual %d", len(actual))
	}
	fanout, ok := actual[0].(scheduledPullRequestRefreshCmd)
	if !ok {
		t.Fatalf("expected a scheduled refresh command, actual %T", actual[0])
	}
	if len(fanout.detailSummaries) != 1 || len(fanout.diffSummaries) != 1 {
		t.Fatalf("expected one detail and diff prefetch target, actual detail=%v diff=%v", fanout.detailSummaries, fanout.diffSummaries)
	}
	if fanout.detailSummaries[0].UpdatedAt != latest.UpdatedAt || fanout.diffSummaries[0].UpdatedAt != latest.UpdatedAt {
		t.Fatalf("expected the latest summary in both prefetch targets, actual detail=%v diff=%v", fanout.detailSummaries, fanout.diffSummaries)
	}
}

func TestScheduledRefresh_GivenASeenPastedPullRequest_WhenTheScheduleIsDue_ThenItProbesRemoteMetadata(t *testing.T) {
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Title:      "Checked pasted PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		UpdatedAt:  "2026-05-05T13:00:00Z",
	})
	subject := NewProgram()
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestsLoaded([]githubdomain.PullRequest{summary})
	})
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", summary.UpdatedAt)
	})

	subject.runtimeConfig.pullRequestConfig.Refresh = time.Minute
	actual := subject.applyScheduledPullRequestRefreshDue(MsgScheduledPullRequestRefreshDue{RefreshPasted: true})
	expected := 2
	if len(actual) != expected {
		t.Fatalf("expected %d commands, actual %v", expected, actual)
	}
	probe, ok := actual[0].(scheduledPullRequestProbeCmd)
	if !ok || len(probe.summaries) != 1 || probe.summaries[0].UpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected remote probe of seen pasted PR, actual %+v", actual)
	}
}

func TestActionsPopup_GivenRefreshCurrentPullRequestInformationActionOutsideReviewMode_WhenExecuting_ThenItReloadsTheSelectedPullRequestSummaryAndDetail(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Refreshed PR",
			Number:     42,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/42",
			Body:       "Refreshed summary body",
			State:      "OPEN",
		}},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Refreshed PR",
				Number:      42,
				Body:        "Refreshed body",
				BaseRefName: "main",
				HeadRefName: "refresh-branch",
				State:       "OPEN",
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:       "Old PR",
		Number:      42,
		Body:        "Old body",
		BaseRefName: "main",
		HeadRefName: "old-branch",
		State:       "OPEN",
	})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Old body") {
		t.Fatalf("expected detail buffer to contain %q before refreshing, actual %q", "Old body", detailView.Buffer())
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Refresh current PR information") {
		t.Fatalf("expected actions popup to contain %q, actual %q", "Refresh current PR information", popupView.Buffer())
	}
	subject.model.UpdateActionsPopupSearch("refresh", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "refresh"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Refreshed body") {
		t.Fatalf("expected detail buffer to contain %q after refreshing, actual %q", "Refreshed body", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Old body") {
		t.Fatalf("expected detail buffer to drop %q after refreshing, actual %q", "Old body", detailView.Buffer())
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Refreshed PR") {
		t.Fatalf("expected pull requests buffer to contain %q after refreshing, actual %q", "Refreshed PR", pullRequestsView.Buffer())
	}
	then_statusLineContains(t, gui, "Pull request refreshed")
}

func TestActionsPopup_GivenRefreshPullRequestListAction_WhenExecuting_ThenItReloadsTheActivePullRequestList(t *testing.T) {
	loader := &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{
		Title:      "Refreshed list PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		State:      "OPEN",
	}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "First PR") {
		t.Fatalf("expected pull requests buffer to contain %q before refreshing, actual %q", "First PR", pullRequestsView.Buffer())
	}
	loader.detailCalls = nil

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestListRefreshActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestListRefreshActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	if len(loader.listPullRequestCommands) != 1 {
		t.Fatalf("expected one pull request list refresh call, actual %d", len(loader.listPullRequestCommands))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string(nil)) {
		t.Fatalf("expected no detail refresh calls, actual %v", loader.detailCalls)
	}

	pullRequestsView, actualErr = gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Refreshed list PR") {
		t.Fatalf("expected pull requests buffer to contain %q after refreshing, actual %q", "Refreshed list PR", pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), "First PR") {
		t.Fatalf("expected pull requests buffer to drop %q after refreshing, actual %q", "First PR", pullRequestsView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestListRefreshSuccessMessage)
}

func TestActionsPopup_GivenRefreshCurrentPullRequestInformationActionInReviewMode_WhenExecuting_ThenItReloadsTheActivePullRequestDetailAndDiff(t *testing.T) {
	staleDiff := given_reviewSessionPullRequestDiff()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Stale PR",
				Number:      42,
				Body:        "Stale body",
				BaseRefName: "main",
				HeadRefName: "feature-old",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": staleDiff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "new line") {
		t.Fatalf("expected detail buffer to contain %q before refreshing, actual %q", "new line", detailView.Buffer())
	}

	loader.details["acme/widgets#42"] = githubcli.PullRequestDetail{
		Title:       "Refreshed PR",
		Number:      42,
		Body:        "Refreshed body",
		BaseRefName: "main",
		HeadRefName: "feature-refresh",
		State:       "OPEN",
	}
	refreshedDiff := staleDiff
	refreshedDiff.UnifiedDiff = strings.ReplaceAll(refreshedDiff.UnifiedDiff, "+new line", "+fresh line")
	loader.diffs["acme/widgets#42"] = refreshedDiff

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Refresh current PR information") {
		t.Fatalf("expected actions popup to contain %q, actual %q", "Refresh current PR information", popupView.Buffer())
	}
	subject.model.UpdateActionsPopupSearch("refresh", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "refresh"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "fresh line") {
		t.Fatalf("expected detail buffer to contain %q after refreshing, actual %q", "fresh line", detailView.Buffer())
	}
	then_statusLineContains(t, gui, "Pull request refreshed")
}

func TestActionsPopup_GivenRefreshCurrentPullRequestInformationActionOutsideReviewMode_WhenTheDetailReloadFails_ThenItShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Refreshed PR",
			Number:     42,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/42",
			State:      "OPEN",
		}},
		detailErrors: map[string]error{"acme/widgets#42": errors.New("detail refresh refused")},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.timingState.transientErrorPopupDuration = 0
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:       "Old PR",
		Number:      42,
		Body:        "Old body",
		BaseRefName: "main",
		HeadRefName: "old-branch",
		State:       "OPEN",
	})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("refresh current", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "refresh current"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)
	then_transientErrorPopupContains(t, gui, "detail refresh refused")
}

func TestActionsPopup_GivenRefreshPullRequestListAction_WhenTheReloadFails_ThenItShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{listPullRequestsErr: errors.New("list refresh refused")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.timingState.transientErrorPopupDuration = 0
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestListRefreshActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestListRefreshActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_statusLineDoesNotContain(t, gui, pullRequestListRefreshSuccessMessage)
	then_transientErrorPopupContains(t, gui, "list refresh refused")
}

func TestActionsPopup_GivenRefreshCurrentPullRequestInformationActionInReviewMode_WhenTheDiffReloadFails_ThenItShowsATransientErrorPopup(t *testing.T) {
	staleDiff := given_reviewSessionPullRequestDiff()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Stale PR",
				Number:      42,
				Body:        "Stale body",
				BaseRefName: "main",
				HeadRefName: "feature-old",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": staleDiff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.timingState.transientErrorPopupDuration = 0
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	loader.diffErrors = map[string]error{"acme/widgets#42": errors.New("diff refresh refused")}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("refresh current", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "refresh current"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)
	then_transientErrorPopupContains(t, gui, "diff refresh refused")
}
