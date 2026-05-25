package tui

import "testing"

func TestDetailMotionCommand_GivenPendingYank_WhenExecutingTheLinewiseFinish_ThenItCopiesTheCurrentLineWithoutMovingTheCursor(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha\nBeta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = &fakeClipboardWriter{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_detailCursorOnSegment(t, gui, subject, "Alpha")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expectedCursor := subject.detailState.viewState.cursor
	subject.detailState.viewState.armPendingYank()

	detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationFinishPendingYank, View: detailView, SelectionKind: detailYankMotionLinewise}.execute(subject, gui)

	if actual := subject.clipboardWriter.(*fakeClipboardWriter).writes; len(actual) != 1 || actual[0] != "Alpha" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"Alpha"}, actual)
	}
	if subject.detailState.viewState.cursor != expectedCursor {
		t.Fatalf("expected cursor %+v after the linewise yank, actual %+v", expectedCursor, subject.detailState.viewState.cursor)
	}
}

func TestDetailMotionCommand_GivenPendingCharacterMotion_WhenExecutingTheTarget_ThenItMovesTheDetailCursor(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha.Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_detailCursorOnSegment(t, gui, subject, "Alpha.Beta")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expected := given_detailPositionOfSegmentOccurrence(t, gui, subject, ".", 0)
	subject.detailState.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

	detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationConsumePendingCharacter, View: detailView, SelectionKind: detailYankMotionCharacterInclusive, Rune: '.'}.execute(subject, gui)

	then_detailCursorIs(t, subject.detailState.viewState, expected)
	if subject.detailState.viewState.hasPendingCharacterMotion() {
		t.Fatal("expected the pending character motion to be consumed")
	}
}

func TestDetailMotionCommand_GivenBuildPopupCharacterMotion_WhenExecutingTheTarget_ThenItMovesThePopupCursor(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Build.zeta"}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	subject.pullRequestBuildRunPopup.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

	detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationConsumePendingCharacter, View: popupView, SelectionKind: detailYankMotionCharacterInclusive, Rune: '.'}.execute(subject, gui)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 5}) {
		t.Fatalf("expected popup cursor %+v, actual %+v", detailPosition{line: 0, column: 5}, actual)
	}
	if subject.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		t.Fatal("expected the popup pending character motion to be consumed")
	}
}
