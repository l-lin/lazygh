package tui

import "testing"

func TestKeySequenceState_GivenATarget_WhenArmingAndMatching_ThenTheSecondStepConsumesIt(t *testing.T) {
	subject := keySequenceState{}
	target := keySequenceTarget{viewName: viewDetailName, actionID: keybindingActionID{scope: keymapScopeCursor, action: "move_cursor_to_top"}}

	if subject.armOrConsume(target) {
		t.Fatal("expected the first step to arm the key sequence")
	}
	if !subject.consume(target) {
		t.Fatal("expected the matching target to consume the pending key sequence")
	}
	if subject.consume(target) {
		t.Fatal("expected the consumed key sequence to be cleared")
	}
}

func TestKeySequenceState_GivenAPendingTarget_WhenArmingAnotherTarget_ThenItReplacesThePendingSequence(t *testing.T) {
	subject := keySequenceState{}
	originalTarget := keySequenceTarget{viewName: viewDetailName, actionID: keybindingActionID{scope: keymapScopeCursor, action: "move_cursor_to_top"}}
	replacementTarget := keySequenceTarget{viewName: viewActionsPopupName, actionID: keybindingActionID{scope: keymapScopeActionsPopup, action: "move_selection_to_top"}}

	subject.arm(originalTarget)
	subject.arm(replacementTarget)

	if subject.consume(originalTarget) {
		t.Fatal("expected arming a new target to replace the previous pending key sequence")
	}
	if !subject.consume(replacementTarget) {
		t.Fatal("expected the replacement target to stay pending")
	}
}
