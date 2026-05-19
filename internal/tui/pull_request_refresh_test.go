package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

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
	actualErr = subject.refreshViews(gui)
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
	actualErr = subject.refreshViews(gui)
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
	actualErr = subject.refreshViews(gui)
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
