package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestBrowserMode_GivenTheCursorOnANonPendingBuild_WhenPressingEnter_ThenItOpensTheBuildInfoPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		buildInfos: map[string]githubcli.PullRequestBuildInfo{
			"https://github.com/acme/widgets/actions/runs/42": {
				Bucket:      "fail",
				CompletedAt: "2026-04-18T13:04:00Z",
				Description: "widget smoke test timed out",
				Event:       "pull_request",
				Link:        "https://github.com/acme/widgets/actions/runs/42",
				Name:        "test",
				StartedAt:   "2026-04-18T13:00:00Z",
				State:       "FAILURE",
				Workflow:    "CI",
			},
		},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-popup",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = enterHandler(gui, detailView)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Build info · CI / test") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Build info · CI / test", popupView.Title)
	}
	for _, expected := range []string{"Status: Failed", "Workflow: CI", "Name: test", "Event: pull_request", "Started: 2026-04-18 13:00 UTC", "Completed: 2026-04-18 13:04 UTC", "widget smoke test timed out", "Link: https://github.com/acme/widgets/actions/runs/42"} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if !reflect.DeepEqual(loader.buildInfoCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected build info calls %v, actual %v", []string{"acme/widgets#42"}, loader.buildInfoCalls)
	}
	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)
}

func TestBrowserMode_GivenTheCursorOnAPendingBuild_WhenPressingEnter_ThenItDoesNotOpenTheBuildInfoPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-popup",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{
					{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE", Link: "https://github.com/acme/widgets/actions/runs/42"},
					{Name: "deploy", Status: "IN_PROGRESS"},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "deploy (Pending)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = enterHandler(gui, detailView)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewPullRequestBuildInfoName)
	if len(loader.buildInfoCalls) != 0 {
		t.Fatalf("expected no build info calls, actual %v", loader.buildInfoCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
}

func TestPullRequestBuildInfoPopup_GivenVisible_WhenPressingEscape_ThenItClosesAndReturnsToDetailView(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		buildInfos: map[string]githubcli.PullRequestBuildInfo{
			"https://github.com/acme/widgets/actions/runs/42": {
				Bucket:   "fail",
				Link:     "https://github.com/acme/widgets/actions/runs/42",
				Name:     "test",
				State:    "FAILURE",
				Workflow: "CI",
			},
		},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-popup",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = enterHandler(gui, detailView)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	closeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, gocui.KeyEsc)
	actualErr = closeHandler(gui, popupView)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewPullRequestBuildInfoName)
	then_currentViewNameIs(t, gui, viewDetailName)
}
