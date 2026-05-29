package tui

import (
	"strings"
	"testing"
)

func TestUpdate_GivenMsgDetailMotionRequestedWithDetailTarget_WhenApplying_ThenItReturnsATypedDetailMotionCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveLeft})

	if len(actual) != 1 {
		t.Fatalf("expected one detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if actual := command.Target; actual != detailMotionTargetDetail {
		t.Fatalf("expected target %v, actual %v", detailMotionTargetDetail, actual)
	}
	if actual := command.Operation; actual != detailMotionOperationMoveLeft {
		t.Fatalf("expected operation %v, actual %v", detailMotionOperationMoveLeft, actual)
	}
}

func TestUpdate_GivenMsgDetailMotionRequestedForBuildPopupRepeatSearchWithoutQuery_WhenApplying_ThenItReturnsNoCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestBuildRunPopup = &pullRequestBuildRunPopupState{}

	actual := Update(subject, MsgDetailMotionRequested{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatSearch})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
}

func TestUpdate_GivenMsgDetailMotionRequestedWithoutLastCharacterMotion_WhenApplyingRepeatCharacter_ThenItReturnsNoCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationRepeatCharacter, SelectionKind: detailYankMotionCharacterInclusive})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
}

func TestUpdate_GivenMsgDetailViewportResolvedWithActiveSearch_WhenApplying_ThenItRefreshesTheSearchMatchAndPreservesViewportSync(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: strings.Join([]string{"Alpha", "Target", "Omega"}, "\n")}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	Update(subject, MsgOpenSearch{Query: "Target"})
	Update(subject, MsgSubmitSearch{})

	document := newDetailDocument(strings.Join([]string{"Alpha", "Target", "Omega"}, "\n"), 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.cursor = detailPosition{line: 1, column: 0}
	subject.detailState.viewState.currentSearchMatch = -1

	actual := Update(subject, MsgDetailViewportResolved{Operation: detailViewportOperationPlaceTop, Document: document, ViewportHeight: 2})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.detailState.viewState.currentSearchMatch; actual != 0 {
		t.Fatalf("expected refreshed current search match %d, actual %d", 0, actual)
	}
	expectedOrigin := placedViewportOrigin(1, 2, document.rowCount(), viewportPlacementTop)
	if actual := subject.detailState.viewState.originRow; actual != expectedOrigin {
		t.Fatalf("expected placed origin row %d, actual %d", expectedOrigin, actual)
	}
	if actual := subject.detailState.viewState.preserveViewportSyncCount; actual != 3 {
		t.Fatalf("expected preserve-sync count %d, actual %d", 3, actual)
	}
}

func TestUpdate_GivenMsgFocusDetailRenderedLineResolvedWithActiveSearch_WhenApplying_ThenItRefreshesTheSearchMatchForTheFocusedLine(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: strings.Join([]string{"Alpha", "Target one", "Omega", "Target two"}, "\n")}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	Update(subject, MsgOpenSearch{Query: "Target"})
	Update(subject, MsgSubmitSearch{})

	document := newDetailDocument(strings.Join([]string{"Alpha", "Target one", "Omega", "Target two"}, "\n"), 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.currentSearchMatch = -1

	actual := Update(subject, MsgFocusDetailRenderedLineResolved{RenderedLine: 3, Document: document, ViewportHeight: 2})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.detailState.viewState.cursor.line; actual != 3 {
		t.Fatalf("expected focused cursor line %d, actual %d", 3, actual)
	}
	if actual := subject.detailState.viewState.currentSearchMatch; actual != 1 {
		t.Fatalf("expected refreshed current search match %d, actual %d", 1, actual)
	}
}

func TestUpdate_GivenMsgDetailViewSyncPlanResolvedWithFocusedLineAndActiveSearch_WhenApplying_ThenItRefreshesTheSearchMatchForTheAppliedLine(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: strings.Join([]string{"Alpha", "Target one", "Omega", "Target two"}, "\n")}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	Update(subject, MsgOpenSearch{Query: "Target"})
	Update(subject, MsgSubmitSearch{})

	document := newDetailDocument(strings.Join([]string{"Alpha", "Target one", "Omega", "Target two"}, "\n"), 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.currentSearchMatch = -1

	actual := Update(subject, MsgDetailViewSyncPlanResolved{Plan: detailViewSyncPlan{document: document, focusLine: 3, focusLineKnown: true}, ViewportHeight: 2})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.detailState.viewState.cursor.line; actual != 3 {
		t.Fatalf("expected focused cursor line %d, actual %d", 3, actual)
	}
	if actual := subject.detailState.viewState.currentSearchMatch; actual != 1 {
		t.Fatalf("expected refreshed current search match %d, actual %d", 1, actual)
	}
}

func TestUpdate_GivenMsgDetailYankRequestedWithVisualDetailSelection_WhenApplying_ThenItReturnsATypedClipboardPrepareCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.detailState.viewState.enterVisualMode()

	actual := Update(subject, MsgDetailYankRequested{Target: detailMotionTargetDetail})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard-prepare command, actual %d", len(actual))
	}
	command, ok := actual[0].(prepareSelectedDetailClipboardWriteCmd)
	if !ok {
		t.Fatalf("expected a prepareSelectedDetailClipboardWriteCmd, actual %T", actual[0])
	}
	if actual := command.Target; actual != FocusDetailView {
		t.Fatalf("expected clipboard target %v, actual %v", FocusDetailView, actual)
	}
}

func TestUpdate_GivenMsgDetailYankRequestedWithPendingBuildPopupYank_WhenApplying_ThenItReturnsATypedFinishMotionCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestBuildRunPopup = &pullRequestBuildRunPopupState{}
	subject.pullRequestBuildRunPopup.viewState.armPendingYank()

	actual := Update(subject, MsgDetailYankRequested{Target: detailMotionTargetBuildPopup})

	if len(actual) != 1 {
		t.Fatalf("expected one detail-motion command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailMotionCmd)
	if !ok {
		t.Fatalf("expected a detailMotionCmd, actual %T", actual[0])
	}
	if actual := command.Target; actual != detailMotionTargetBuildPopup {
		t.Fatalf("expected target %v, actual %v", detailMotionTargetBuildPopup, actual)
	}
	if actual := command.Operation; actual != detailMotionOperationFinishPendingYank {
		t.Fatalf("expected operation %v, actual %v", detailMotionOperationFinishPendingYank, actual)
	}
	if actual := command.SelectionKind; actual != detailYankMotionLinewise {
		t.Fatalf("expected selection kind %v, actual %v", detailYankMotionLinewise, actual)
	}
}
