package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestUpdate_GivenMsgToggleInlineConversationVisibility_WhenApplying_ThenItReturnsATypedToggleInlineConversationCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgToggleInlineConversationVisibility{})

	if len(actual) != 1 {
		t.Fatalf("expected one toggle-inline-conversation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(toggleInlineConversationVisibilityCmd); !ok {
		t.Fatalf("expected a toggleInlineConversationVisibilityCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgSetAllDetailFolds_WhenApplying_ThenItReturnsATypedBulkDetailFoldCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgSetAllDetailFolds{Collapsed: true})

	if len(actual) != 1 {
		t.Fatalf("expected one set-all-detail-folds command, actual %d", len(actual))
	}
	command, ok := actual[0].(setAllDetailFoldsCmd)
	if !ok {
		t.Fatalf("expected a setAllDetailFoldsCmd, actual %T", actual[0])
	}
	if !command.Collapsed {
		t.Fatal("expected the typed bulk-fold command to preserve the collapsed request")
	}
}

func TestUpdate_GivenMsgDetailViewSyncPlanResolvedWithFocusedLine_WhenApplying_ThenItAppliesTheFocusedDetailState(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha\nBeta\nGamma"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	document := newDetailDocument("Alpha\nBeta\nGamma", 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 0}

	actual := Update(subject, MsgDetailViewSyncPlanResolved{Plan: detailViewSyncPlan{document: document, focusLine: 2, focusLineKnown: true}, ViewportHeight: 2})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actualLine := subject.detailState.viewState.cursor.line; actualLine != 2 {
		t.Fatalf("expected focused cursor line %d, actual %d", 2, actualLine)
	}
	expectedOrigin := visibleViewportOrigin(2, 0, 2, document.rowCount())
	if actualOrigin := subject.detailState.viewState.originRow; actualOrigin != expectedOrigin {
		t.Fatalf("expected synced origin row %d, actual %d", expectedOrigin, actualOrigin)
	}
}

func TestSetAllDetailFoldsCommand_GivenAResolvedSyncPlan_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	expectedDocument := newDetailDocument("Alpha", 40)
	expectedPlan := detailViewSyncPlan{document: expectedDocument, focusLine: 1, focusLineKnown: true}
	actualDispatched := []Msg(nil)

	executeSetAllDetailFoldsCommand(detailFoldCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentDetailDocument: func(view *gocui.View) detailDocument {
			return expectedDocument
		},
		setAllDetailFolds: func(detailDocument detailDocument, collapsed bool) (detailViewSyncPlan, bool) {
			if !collapsed {
				t.Fatal("expected collapsed request to stay true")
			}
			if detailDocument.id != expectedDocument.id {
				t.Fatalf("expected the current detail document %d, actual %d", expectedDocument.id, detailDocument.id)
			}
			return expectedPlan, true
		},
	}, nil, setAllDetailFoldsCmd{Collapsed: true})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgDetailViewSyncPlanResolved)
	if !ok {
		t.Fatalf("expected a MsgDetailViewSyncPlanResolved, actual %T", actualDispatched[0])
	}
	if actual := message.Plan.document.id; actual != expectedPlan.document.id {
		t.Fatalf("expected resolved document %d, actual %d", expectedPlan.document.id, actual)
	}
	if !message.Plan.focusLineKnown || message.Plan.focusLine != expectedPlan.focusLine {
		t.Fatalf("expected resolved focus line %+v, actual %+v", expectedPlan, message.Plan)
	}
	if actual := message.ViewportHeight; actual != 1 {
		t.Fatalf("expected fallback viewport height %d, actual %d", 1, actual)
	}
}
