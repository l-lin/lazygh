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

func TestUpdate_GivenMsgDetailSearchWordResolved_WhenApplying_ThenItAppliesTheSearchStateAndReturnsATypedFollowCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgDetailSearchWordResolved{Query: "Alpha", Reverse: true})

	if actualQuery := subject.model.DetailSearchQuery(); actualQuery != "Alpha" {
		t.Fatalf("expected detail search query %q, actual %q", "Alpha", actualQuery)
	}
	if subject.model.SearchActive() {
		t.Fatal("expected the direct word search to leave the search prompt inactive")
	}
	if subject.searchWidget.editor != nil {
		t.Fatal("expected the search editor to stay cleared after direct word search")
	}
	if !subject.searchWidget.detailReversed {
		t.Fatal("expected the detail search direction to stay reversed")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-detail-search command, actual %d", len(actual))
	}
	command, ok := actual[0].(followDetailSearchCmd)
	if !ok {
		t.Fatalf("expected a followDetailSearchCmd, actual %T", actual[0])
	}
	if !command.Reverse {
		t.Fatal("expected the follow command to stay reversed")
	}
}

func TestUpdate_GivenMsgSubmitSearchForDetailTarget_WhenApplying_ThenItReturnsATypedFollowDetailSearchCommand(t *testing.T) {
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
	if subject.searchWidget.editor != nil {
		t.Fatal("expected the search editor to be cleared after submit")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-detail-search command, actual %d", len(actual))
	}
	command, ok := actual[0].(followDetailSearchCmd)
	if !ok {
		t.Fatalf("expected a followDetailSearchCmd, actual %T", actual[0])
	}
	if command.Reverse {
		t.Fatal("expected the submitted search follow command to stay forward")
	}
}

func TestUpdate_GivenMsgRepeatDetailSearchRequested_WhenApplying_ThenItReturnsATypedRepeatDetailSearchCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	Update(subject, MsgOpenSearch{Query: "Alpha"})
	Update(subject, MsgSubmitSearch{})
	actual := Update(subject, MsgRepeatDetailSearchRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one repeat-detail-search command, actual %d", len(actual))
	}
	if _, ok := actual[0].(repeatDetailSearchCmd); !ok {
		t.Fatalf("expected a repeatDetailSearchCmd, actual %T", actual[0])
	}
}
