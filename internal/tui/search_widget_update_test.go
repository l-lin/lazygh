package tui

import "testing"

func TestUpdate_GivenMsgOpenSearchForDetailTarget_WhenApplying_ThenItOpensTheSearchWidgetAndSeedsTheDraft(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	Update(subject, MsgOpenSearch{Query: "Alpha"})

	if !subject.model.SearchActive() {
		t.Fatal("expected the search prompt to become active")
	}
	if actual := subject.model.SearchDraft(); actual != "Alpha" {
		t.Fatalf("expected search draft %q, actual %q", "Alpha", actual)
	}
	if !subject.searchWidget.hasEditor() {
		t.Fatal("expected the search widget editor to open")
	}
	if actual := subject.searchWidget.editor.Text(); actual != "Alpha" {
		t.Fatalf("expected search widget text %q, actual %q", "Alpha", actual)
	}
}

func TestUpdate_GivenMsgCancelSearchForDetailTarget_WhenApplying_ThenItClearsTheSearchWidgetAndPreservesTheCurrentDirection(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.searchWidget.detailReversed = true

	Update(subject, MsgOpenSearch{Query: "Alpha"})
	Update(subject, MsgCancelSearch{})

	if subject.model.SearchActive() {
		t.Fatal("expected the search prompt to be inactive after cancel")
	}
	if subject.searchWidget.hasEditor() {
		t.Fatal("expected the search widget editor to be cleared after cancel")
	}
	if !subject.searchWidget.detailReversed {
		t.Fatal("expected cancel to preserve the current detail-search direction")
	}
}

func TestUpdate_GivenMsgSubmitSearchForDetailTargetAfterAReversedSearch_WhenApplying_ThenItResetsTheStoredDirectionToForward(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.searchWidget.detailReversed = true

	Update(subject, MsgOpenSearch{Query: "Alpha"})
	actual := Update(subject, MsgSubmitSearch{})

	if subject.searchWidget.detailReversed {
		t.Fatal("expected detail search submit to reset the stored direction to forward")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if command.Reverse {
		t.Fatal("expected the submitted search follow command to stay forward")
	}
}
