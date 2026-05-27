package tui

import (
	"strings"
	"testing"
)

func TestUpdate_GivenMsgDetailMotionResolvedWithPendingLinewiseYank_WhenApplying_ThenItRestoresTheCursorAndReturnsAClipboardWriteCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha\nBeta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	document := newDetailDocument("Alpha\nBeta", 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 0}
	subject.detailState.viewState.preferredColumn = 0
	subject.detailState.viewState.armPendingYank()

	actual := Update(subject, MsgDetailMotionResolved{Target: detailMotionTargetDetail, Operation: detailMotionOperationFinishPendingYank, SelectionKind: detailYankMotionLinewise, Document: document, ViewportHeight: 2})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard command, actual %d", len(actual))
	}
	command, ok := actual[0].(writeClipboardCmd)
	if !ok {
		t.Fatalf("expected a writeClipboardCmd, actual %T", actual[0])
	}
	if actual := command.Text; actual != "Alpha" {
		t.Fatalf("expected clipboard text %q, actual %q", "Alpha", actual)
	}
	if actual := command.SelectionTarget; actual != clipboardWriteSelectionDetail {
		t.Fatalf("expected clipboard selection target %v, actual %v", clipboardWriteSelectionDetail, actual)
	}
	if subject.detailState.viewState.cursor != (detailPosition{line: 0, column: 0}) {
		t.Fatalf("expected cursor %+v after linewise yank, actual %+v", detailPosition{line: 0, column: 0}, subject.detailState.viewState.cursor)
	}
	if subject.detailState.viewState.hasPendingYank() {
		t.Fatal("expected the pending yank to clear after returning the clipboard command")
	}
}

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
	_, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expectedCursor := subject.detailState.viewState.cursor
	subject.detailState.viewState.armPendingYank()

	detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationFinishPendingYank, SelectionKind: detailYankMotionLinewise}.execute(subject, gui)

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
	_, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expected := given_detailPositionOfSegmentOccurrence(t, gui, subject, ".", 0)
	subject.detailState.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

	detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationConsumePendingCharacter, SelectionKind: detailYankMotionCharacterInclusive, Rune: '.'}.execute(subject, gui)

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
	_, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	subject.pullRequestBuildRunPopup.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

	detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationConsumePendingCharacter, SelectionKind: detailYankMotionCharacterInclusive, Rune: '.'}.execute(subject, gui)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 5}) {
		t.Fatalf("expected popup cursor %+v, actual %+v", detailPosition{line: 0, column: 5}, actual)
	}
	if subject.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		t.Fatal("expected the popup pending character motion to be consumed")
	}
}

func TestDetailMotionCommand_GivenBuildPopupMoveDownOperation_WhenExecuting_ThenItMovesThePopupCursor(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha\nBeta"}))
	_, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)

	detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationMoveDown}.execute(subject, gui)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != 1 {
		t.Fatalf("expected popup cursor line %d, actual %d", 1, actual)
	}
}

func TestDetailMotionCommand_GivenBuildPopupSubmittedSearchOperation_WhenExecuting_ThenItMovesThePopupCursorToTheFirstMatch(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		body:       strings.Join([]string{"Alpha", "Target", "Omega"}, "\n"),
	}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	expectedLineIndex, _ := given_detailDocumentLineContaining(t, document, "Target")
	subject.pullRequestBuildRunPopup.searchQuery = "Target"

	detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFollowSubmittedSearch}.execute(subject, gui)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != expectedLineIndex {
		t.Fatalf("expected popup cursor line %d after submitted search, actual %d", expectedLineIndex, actual)
	}
}
