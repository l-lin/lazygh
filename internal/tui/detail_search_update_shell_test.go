package tui

import "testing"

func TestUpdate_GivenMsgSearchWordUnderCursor_WhenApplying_ThenItReturnsATypedResolveDetailSearchWordCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgSearchWordUnderCursor{})

	if len(actual) != 1 {
		t.Fatalf("expected one detail-search word command, actual %d", len(actual))
	}
	if _, ok := actual[0].(resolveDetailSearchWordCmd); !ok {
		t.Fatalf("expected a resolveDetailSearchWordCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgDetailSearchWordResolvedContext_WhenApplying_ThenItAppliesTheSearchStateAndReturnsATypedDetailMotionCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 0}

	actual := Update(subject, MsgDetailSearchWordResolved{Document: newDetailDocument("Alpha Beta", 40), ViewportHeight: 3, Reverse: true})

	if actualQuery := subject.model.DetailSearchQuery(); actualQuery != "Alpha" {
		t.Fatalf("expected detail search query %q, actual %q", "Alpha", actualQuery)
	}
	if subject.model.SearchActive() {
		t.Fatal("expected the direct word search to leave the search prompt inactive")
	}
	if subject.searchWidget.hasEditor() {
		t.Fatal("expected the search editor to stay cleared after direct word search")
	}
	if !subject.searchWidget.detailReversed {
		t.Fatal("expected the detail search direction to stay reversed")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if command.Target != detailMotionTargetDetail || command.Operation != detailMotionOperationFollowSubmittedSearch {
		t.Fatalf("expected a detail follow-search motion command, actual %+v", command)
	}
	if !command.Reverse {
		t.Fatal("expected the follow command to stay reversed")
	}
}

func TestUpdate_GivenMsgSubmitSearchForDetailTarget_WhenApplying_ThenItReturnsATypedDetailMotionCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	Update(subject, MsgOpenSearch{Query: "Alpha"})
	actual := Update(subject, MsgSubmitSearch{})

	if actualQuery := subject.model.DetailSearchQuery(); actualQuery != "Alpha" {
		t.Fatalf("expected applied detail search query %q, actual %q", "Alpha", actualQuery)
	}
	if subject.model.SearchActive() {
		t.Fatal("expected the search to be inactive after submit")
	}
	if subject.searchWidget.hasEditor() {
		t.Fatal("expected the search editor to be cleared after submit")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if command.Target != detailMotionTargetDetail || command.Operation != detailMotionOperationFollowSubmittedSearch {
		t.Fatalf("expected a detail follow-search motion command, actual %+v", command)
	}
	if command.Reverse {
		t.Fatal("expected the submitted search follow command to stay forward")
	}
}

func TestUpdate_GivenMsgRepeatDetailSearchRequested_WhenApplying_ThenItReturnsATypedDetailMotionCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	Update(subject, MsgOpenSearch{Query: "Alpha"})
	Update(subject, MsgSubmitSearch{})
	actual := Update(subject, MsgRepeatDetailSearchRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one repeat-detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if command.Target != detailMotionTargetDetail || command.Operation != detailMotionOperationRepeatSearch {
		t.Fatalf("expected a detail repeat-search motion command, actual %+v", command)
	}
}
