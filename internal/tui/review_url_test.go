package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestOpenReviewByURL_GivenAValidGitHubPRURLAfterLayout_WhenOpening_ThenItRefreshesThroughDispatch(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_model(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.OpenReviewByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	if !subject.navigationState.reviewSession.active {
		t.Fatal("expected review mode to become active immediately")
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.Title != reviewModeMetadataTitle {
		t.Fatalf("expected metadata view title %q, actual %q", reviewModeMetadataTitle, metadataView.Title)
	}
	if actual := strings.TrimSpace(metadataView.Buffer()); actual != "acme/rocket#77" {
		t.Fatalf("expected metadata view buffer %q, actual %q", "acme/rocket#77", actual)
	}
}

func TestOpenReviewByURL_GivenAValidGitHubPRURLBeforeLayout_WhenRendering_ThenItStartsDirectlyInReviewMode(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:        "Rocket PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/review-url",
				State:        "OPEN",
				ChangedFiles: 2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_model(), loader)

	actualErr := subject.OpenReviewByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	if !subject.navigationState.reviewSession.active {
		t.Fatal("expected review mode to start before the first layout")
	}
	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected review start calls %v, actual %v", []string{"acme/rocket#77"}, loader.startReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.Title != reviewModeMetadataTitle {
		t.Fatalf("expected metadata view title %q, actual %q", reviewModeMetadataTitle, metadataView.Title)
	}
	if actual := strings.TrimSpace(metadataView.Buffer()); actual != "acme/rocket#77" {
		t.Fatalf("expected metadata view buffer %q, actual %q", "acme/rocket#77", actual)
	}
}

func TestOpenReviewByURL_GivenAValidGitHubPRURLBeforeLayout_WhenYankingFromTheActionsPopup_ThenItUsesTheReviewedPullRequestURL(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	clipboardWriter := &fakeClipboardWriter{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.clipboardWriter = clipboardWriter

	actualErr := subject.OpenReviewByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("clipboard", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clipboard"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(clipboardWriter.writes, []string{"https://github.com/acme/rocket/pull/77"}) {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/rocket/pull/77"}, clipboardWriter.writes)
	}
}

func TestOpenReviewByURL_GivenAValidGitHubPRURLBeforeLayout_WhenOpeningInBrowserFromTheActionsPopup_ThenItUsesTheReviewedPullRequestIdentity(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

	actualErr := subject.OpenReviewByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "browser"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.openBrowserCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected open browser calls %v, actual %v", []string{"acme/rocket#77"}, loader.openBrowserCalls)
	}
}
