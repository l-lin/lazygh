package tui

import (
	"reflect"
	"testing"
)

func TestRefactorGuard_GivenProgramType_WhenInspecting_ThenSearchAndActionsPopupRuntimeStateUseChildModels(t *testing.T) {
	programType := reflect.TypeOf(Program{})

	for _, fieldName := range []string{
		"detailSearchReversed",
		"searchEditor",
		"actionsPopupSearchEditor",
		"actionsPopupErrorMessage",
		"actionsPopupPendingConfirmationActionID",
		"reactionPicker",
		"themePicker",
		"assigneePicker",
		"assigneePickerSearchDebounceDelay",
		"assigneePickerLoad",
	} {
		if _, exists := programType.FieldByName(fieldName); exists {
			t.Fatalf("expected Program field %q to move under a widget child model", fieldName)
		}
	}

	for _, fieldName := range []string{"searchWidget", "actionsPopupWidget"} {
		if _, exists := programType.FieldByName(fieldName); !exists {
			t.Fatalf("expected Program to expose child model field %q", fieldName)
		}
	}
}
