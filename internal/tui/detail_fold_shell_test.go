package tui

import "testing"

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
