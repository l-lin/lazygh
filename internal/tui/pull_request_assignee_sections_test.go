package tui

import (
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestAssigneePickerCandidateSections_GivenSearchResults_WhenCollectingSections_ThenItKeepsPinnedAndSearchedCandidatesSeparate(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.actionsPopupWidget.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	actual := subject.currentAssigneePickerCandidateSections("char")
	if !reflect.DeepEqual(given_assigneePickerCandidateLogins(actual.pinned), []string{"bob", "alice"}) {
		t.Fatalf("expected pinned logins %v, actual %v", []string{"bob", "alice"}, given_assigneePickerCandidateLogins(actual.pinned))
	}
	if !reflect.DeepEqual(given_assigneePickerCandidateLogins(actual.searchResults), []string{"charlie"}) {
		t.Fatalf("expected searched logins %v, actual %v", []string{"charlie"}, given_assigneePickerCandidateLogins(actual.searchResults))
	}
}

func TestAssigneePickerCandidateSections_GivenSelectedSearchResultAndANewQuery_WhenCollectingSections_ThenItPromotesTheSelectionIntoPinned(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.searchAssignableUsers["acme/widgets|dora"] = []githubcli.PullRequestAuthor{{Login: "dora", Name: "Dora"}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.actionsPopupWidget.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	actualErr = subject.toggleAssigneePickerSelection(githubdomain.PullRequestAuthor{Login: "charlie", Name: "Charlie"})
	if actualErr != nil {
		t.Fatalf("expected no toggle error, actual %v", actualErr)
	}
	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlU, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+u to clear the assignee picker query")
	}
	for _, ch := range "dora" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	actual := subject.currentAssigneePickerCandidateSections("dora")
	if !reflect.DeepEqual(given_assigneePickerCandidateLogins(actual.pinned), []string{"bob", "alice", "charlie"}) {
		t.Fatalf("expected pinned logins %v, actual %v", []string{"bob", "alice", "charlie"}, given_assigneePickerCandidateLogins(actual.pinned))
	}
	if !reflect.DeepEqual(given_assigneePickerCandidateLogins(actual.searchResults), []string{"dora"}) {
		t.Fatalf("expected searched logins %v, actual %v", []string{"dora"}, given_assigneePickerCandidateLogins(actual.searchResults))
	}
}

func given_assigneePickerCandidateLogins(candidates []githubdomain.PullRequestAuthor) []string {
	logins := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		logins = append(logins, candidate.Login)
	}
	return logins
}
