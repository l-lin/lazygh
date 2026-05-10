package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestRenderPullRequestOverviewSection_GivenEligibleReviewedUsers_WhenFormatting_ThenItShowsTheReRequestReviewIndicatorOnlyForThoseUsers(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
		Reviews: []githubcli.PullRequestReview{
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-blocked"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-04-21T11:00:00Z"},
		},
	}

	actualDocument := newDetailDocument(renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80), 80)
	actual := string(actualDocument.text)

	for _, expected := range []string{
		"@reviewer-approved " + pullRequestOverviewReRequestReviewIcon,
		"@reviewer-blocked " + pullRequestOverviewReRequestReviewIcon,
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected overview section to contain %q, actual %q", expected, actual)
		}
	}
	for _, unexpected := range []string{
		"@reviewer-requested " + pullRequestOverviewReRequestReviewIcon,
		"Requested team " + pullRequestOverviewReRequestReviewIcon,
		"Requested teams " + pullRequestOverviewReRequestReviewIcon,
	} {
		if strings.Contains(actual, unexpected) {
			t.Fatalf("expected overview section to omit %q, actual %q", unexpected, actual)
		}
	}
}

func TestActionsPopup_GivenDetailCursorOnAReRequestableReviewer_WhenOpening_ThenItShowsTheReRequestReviewAction(t *testing.T) {
	subject := given_pullRequestReviewerProgram()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_openPullRequestReviewerOverview(t, gui, subject, "@reviewer-approved")
	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	expected := actionsPopupLabel(actionsPopupReRequestReviewIcon, reRequestPullRequestReviewActionTitle("reviewer-approved"))
	if !strings.Contains(popupView.Buffer(), expected) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
	}
}

func TestActionsPopup_GivenDetailCursorOutsideAReRequestableReviewer_WhenOpening_ThenItHidesTheReRequestReviewAction(t *testing.T) {
	testCases := []struct {
		name    string
		segment string
	}{
		{name: "summary line", segment: "1 reviewer has requested changes."},
		{name: "already requested reviewer", segment: "@reviewer-requested"},
		{name: "requested team detail", segment: "Platform"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_pullRequestReviewerProgram()
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)

			given_openPullRequestReviewerOverview(t, gui, subject, testCase.segment)
			actualErr := subject.openActionsPopup(gui, nil)
			then_noError(t, actualErr)

			popupView, actualErr := gui.View(viewActionsPopupName)
			then_noError(t, actualErr)
			if strings.Contains(popupView.Buffer(), reRequestPullRequestReviewActionTitle("reviewer-approved")) {
				t.Fatalf("expected popup buffer to hide %q, actual %q", reRequestPullRequestReviewActionTitle("reviewer-approved"), popupView.Buffer())
			}
		})
	}
}

func TestReRequestPullRequestReview_GivenSelectedReviewer_WhenExecuting_ThenItRequestsTheReviewerAgainAndRefreshesTheVisibleDetail(t *testing.T) {
	loader := given_pullRequestReviewerLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_openPullRequestReviewerOverview(t, gui, subject, "@reviewer-approved")
	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("re-request review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "re-request review"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.requestReviewerCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected reviewer request calls %v, actual %v", []string{"acme/widgets#42"}, loader.requestReviewerCalls)
	}
	if !reflect.DeepEqual(loader.requestReviewerLogins, []string{"reviewer-approved"}) {
		t.Fatalf("expected reviewer request logins %v, actual %v", []string{"reviewer-approved"}, loader.requestReviewerLogins)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "@reviewer-approved "+pullRequestOverviewReRequestReviewIcon) {
		t.Fatalf("expected the refreshed detail to drop the re-request indicator, actual %q", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestReviewReRequestedSuccessMessage)
}

func TestReviewMode_GivenDescriptionCursorOnAReRequestableReviewer_WhenOpeningActionsPopup_ThenItShowsTheReRequestReviewAction(t *testing.T) {
	loader := given_pullRequestReviewerLoader()
	loader.startReviewID = "PRR_pending"
	loader.diffs = map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "@reviewer-approved")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	expected := actionsPopupLabel(actionsPopupReRequestReviewIcon, reRequestPullRequestReviewActionTitle("reviewer-approved"))
	if !strings.Contains(popupView.Buffer(), expected) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
	}
}

func given_openPullRequestReviewerOverview(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, segment)
}

func given_pullRequestReviewerProgram() *Program {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), given_pullRequestReviewerLoader())
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}

func given_pullRequestReviewerLoader() *fakePullRequestDetailLoader {
	return &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/re-review",
				State:       "OPEN",
				ReviewRequests: []githubcli.PullRequestReviewRequest{
					{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}},
					{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
				},
				Reviews: []githubcli.PullRequestReview{
					{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"},
					{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-blocked"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-04-21T11:00:00Z"},
				},
			},
		},
	}
}
