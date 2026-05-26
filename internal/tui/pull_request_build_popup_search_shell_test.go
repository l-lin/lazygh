package tui

import "testing"

func TestUpdate_GivenMsgSubmitSearchForBuildPopup_WhenApplying_ThenItReturnsATypedBuildPopupSearchCommand(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Target"}))

	Update(subject, MsgOpenSearch{Query: "Target"})
	actual := Update(subject, MsgSubmitSearch{})

	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the build popup to stay visible")
	}
	if actualQuery := subject.pullRequestBuildRunPopup.searchQuery; actualQuery != "Target" {
		t.Fatalf("expected popup search query %q, actual %q", "Target", actualQuery)
	}
	if subject.pullRequestBuildRunPopup.searchActive {
		t.Fatal("expected the build popup search prompt to be inactive after submit")
	}
	if subject.searchWidget.hasEditor() {
		t.Fatal("expected the search editor to be cleared after popup search submit")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one build-popup search command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if command.Target != detailMotionTargetBuildPopup {
		t.Fatalf("expected build-popup motion target %v, actual %v", detailMotionTargetBuildPopup, command.Target)
	}
	if command.Operation != detailMotionOperationFollowSubmittedSearch {
		t.Fatalf("expected submitted-search operation %v, actual %v", detailMotionOperationFollowSubmittedSearch, command.Operation)
	}
}
