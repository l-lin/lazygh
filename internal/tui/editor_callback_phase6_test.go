package tui

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestEditorCallbacks_GivenEditorSourceFiles_WhenInspecting_ThenTheyDoNotBypassDispatchOrDirectlyRedraw(t *testing.T) {
	for path, forbiddenSnippets := range map[string][]string{
		"search_view.go": {
			"Update(program, MsgSearchDraftChanged",
			"program.configureSearchView(view)",
			"program.renderSearchView(view)",
		},
		"actions_popup_interaction.go": {
			"program.afterStateChange(program.gui)",
			"program.configureActionsPopupSearchView(view)",
			"program.renderActionsPopupSearchView(view)",
			"program.queueAssigneePickerSearch(program.gui, requestID, query)",
		},
		"modal_editor_view.go": {
			"program.configureModalEditorView(view)",
			"program.renderModalEditorView(view)",
		},
	} {
		contents, actualErr := os.ReadFile(path)
		then_noError(t, actualErr)
		actualSource := string(contents)

		for _, forbiddenSnippet := range forbiddenSnippets {
			if strings.Contains(actualSource, forbiddenSnippet) {
				t.Fatalf("expected %q to avoid %q inside editor callbacks, actual source:\n%s", path, forbiddenSnippet, actualSource)
			}
		}
	}
}

func TestUpdate_GivenMsgActionsPopupSearchEdited_WhenAssigneePickerIsVisible_ThenItUpdatesTheSearchStateAndQueuesADebouncedSearchCmd(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.actionsPopupWidget.assigneePickerSearchDebounceDelay = 123 * time.Millisecond
	subject.actionsPopupWidget.assigneePicker = newAssigneePickerState(pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}, "viewer", "Viewer")

	actual := Update(subject, MsgActionsPopupSearchEdited{Query: "char"})

	if actualQuery := subject.model.ActionsPopupSearchQuery(); actualQuery != "char" {
		t.Fatalf("expected actions popup search query %q, actual %q", "char", actualQuery)
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
	if actualRequestID := subject.actionsPopupWidget.assigneePicker.searchRequestID; actualRequestID != 1 {
		t.Fatalf("expected assignee picker request id %d, actual %d", 1, actualRequestID)
	}
	if len(actual) != 1 {
		t.Fatalf("expected one queued assignee picker search command, actual %d", len(actual))
	}
	command, ok := actual[0].(assigneePickerSearchCmd)
	if !ok {
		t.Fatalf("expected an assigneePickerSearchCmd, actual %T", actual[0])
	}
	if command.RequestID != 1 || command.Query != "char" || command.Delay != 123*time.Millisecond || !command.DispatchLoading {
		t.Fatalf("expected queued command %+v, actual %+v", assigneePickerSearchCmd{RequestID: 1, Query: "char", Delay: 123 * time.Millisecond, DispatchLoading: true}, command)
	}
}

func TestUpdate_GivenMsgModalEditorEdited_WhenModalEditorIsVisible_ThenItClearsTheErrorMessage(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.modalEditor = newLineModalEditorState("Prompt", "draft", nil)
	subject.modalEditor.errorMessage = "boom"

	Update(subject, MsgModalEditorEdited{})

	if actual := subject.modalEditor.errorMessage; actual != "" {
		t.Fatalf("expected modal editor error message %q, actual %q", "", actual)
	}
}
