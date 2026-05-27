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

func TestUpdate_GivenMsgSearchEditorInputRequested_WhenSearchEditorIsMissing_ThenItBuildsEditorStateAndUpdatesTheDraft(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.model.StartSearch()
	subject.model.UpdateSearchDraft("ab")

	Update(subject, MsgSearchEditorInputRequested{Intent: newLineEditorInsertRuneIntent('c')})

	if actual := subject.model.SearchDraft(); actual != "abc" {
		t.Fatalf("expected search draft %q, actual %q", "abc", actual)
	}
	if !subject.searchWidget.hasEditor() {
		t.Fatal("expected the reducer to open the search editor before applying the input intent")
	}
	if actual := subject.searchWidget.editor.Text(); actual != "abc" {
		t.Fatalf("expected search editor text %q, actual %q", "abc", actual)
	}
}

func TestUpdate_GivenMsgActionsPopupSearchInputRequested_WhenAssigneePickerIsVisible_ThenItUpdatesTheSearchStateAndQueuesADebouncedSearchCmd(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.model.UpdateActionsPopupSearch("cha", nil)
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.actionsPopupWidget.assigneePickerSearchDebounceDelay = 123 * time.Millisecond
	subject.actionsPopupWidget.assigneePicker = newAssigneePickerState(pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}, "viewer", "Viewer")

	actual := Update(subject, MsgActionsPopupSearchInputRequested{Intent: newLineEditorInsertRuneIntent('r')})

	if actualQuery := subject.model.ActionsPopupSearchQuery(); actualQuery != "char" {
		t.Fatalf("expected actions popup search query %q, actual %q", "char", actualQuery)
	}
	if !subject.actionsPopupWidget.hasSearchEditor() {
		t.Fatal("expected the reducer to own opening the actions popup search editor before applying the input intent")
	}
	if actualText := subject.actionsPopupWidget.searchEditor.Text(); actualText != "char" {
		t.Fatalf("expected popup search editor text %q, actual %q", "char", actualText)
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

func TestUpdate_GivenMsgModalEditorLineInputRequested_WhenModalEditorIsVisible_ThenItMutatesTheTextAndClearsTheErrorMessage(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.modalEditor = newLineModalEditorState("Prompt", "draft")
	subject.overlayState.modalEditor.errorMessage = "boom"

	Update(subject, MsgModalEditorLineInputRequested{Intent: newLineEditorInsertRuneIntent('!')})

	if actual := subject.overlayState.modalEditor.Text(); actual != "draft!" {
		t.Fatalf("expected modal editor text %q, actual %q", "draft!", actual)
	}
	if actual := subject.overlayState.modalEditor.errorMessage; actual != "" {
		t.Fatalf("expected modal editor error message %q, actual %q", "", actual)
	}
}
