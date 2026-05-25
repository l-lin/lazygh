package tui

import (
	"reflect"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestPullRequestTabNavigation_GivenRenderedProgram_WhenReturningToLoadedPullRequestTabs_ThenEachSelectionDispatchesAReloadCommand(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Reloaded My PR",
			Number:     42,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/42",
			State:      "OPEN",
		}},
		requestedPullRequests: []githubcli.PullRequest{{
			Title:      "Reloaded Requested PR",
			Number:     84,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/84",
			State:      "OPEN",
		}},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	loader.listPullRequestCommands = nil

	actualErr = subject.nextPullRequestTab(gui, nil)
	then_noError(t, actualErr)
	if !reflect.DeepEqual(loader.listPullRequestCommands, [][]string{subject.pullRequestSearch(RequestedPullRequestsTab).Command}) {
		t.Fatalf("expected requested-tab reload commands %v, actual %v", [][]string{subject.pullRequestSearch(RequestedPullRequestsTab).Command}, loader.listPullRequestCommands)
	}

	loader.listPullRequestCommands = nil
	actualErr = subject.previousPullRequestTab(gui, nil)
	then_noError(t, actualErr)
	if !reflect.DeepEqual(loader.listPullRequestCommands, [][]string{subject.pullRequestSearch(MyPullRequestsTab).Command}) {
		t.Fatalf("expected my-tab reload commands %v, actual %v", [][]string{subject.pullRequestSearch(MyPullRequestsTab).Command}, loader.listPullRequestCommands)
	}
}
