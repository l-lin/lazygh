package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestSearchWidgetState_GivenCommandMode_WhenOpeningAndEditing_ThenItUsesTheSharedLineEditorWithExplicitCommandOwnership(t *testing.T) {
	subject := searchWidgetState{}.withCommandModeOpened()

	if !subject.commandModeActive() {
		t.Fatal("expected command mode to be active")
	}
	if actual := subject.editor.Text(); actual != "" {
		t.Fatalf("expected an empty command, actual %q", actual)
	}

	updated, ok := subject.withEditorIntentApplied(newLineEditorInsertRuneIntent('m'))
	if !ok {
		t.Fatal("expected command input to use the shared line editor")
	}
	if actual := updated.editor.Text(); actual != "m" {
		t.Fatalf("expected command text %q, actual %q", "m", actual)
	}
	if !updated.commandModeActive() {
		t.Fatal("expected editing to preserve command-mode ownership")
	}
}

func TestUpdate_GivenCommandModeWithRecordedErrors_WhenSubmittingMessagesAndClosingPopup_ThenItOpensTheTimestampedHistoryAndClearsCommandState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	currentTime := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	subject.timingState.now = func() time.Time { return currentTime }
	Update(subject, MsgErrorReported{Message: "First error"})
	currentTime = currentTime.Add(time.Minute)
	Update(subject, MsgErrorReported{Message: "Second error"})

	Update(subject, MsgOpenCommandMode{})
	for _, character := range "messages" {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}
	if !subject.commandModeActive() {
		t.Fatal("expected command mode to remain active before submit")
	}
	if actual := subject.searchWidget.editor.Text(); actual != "messages" {
		t.Fatalf("expected command text %q before submit, actual %q", "messages", actual)
	}
	Update(subject, MsgSubmitCommand{})

	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the messages command to open the recorded-errors popup")
	}
	if actual := subject.pullRequestBuildRunPopup.title; actual != "messages" {
		t.Fatalf("expected popup title %q, actual %q", "messages", actual)
	}
	expectedBody := "12:00:00 First error\n12:01:00 Second error"
	if actual := subject.pullRequestBuildRunPopup.body; actual != expectedBody {
		t.Fatalf("expected timestamped chronological body %q, actual %q", expectedBody, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.widthPercent; actual != 90 {
		t.Fatalf("expected popup width percent %d, actual %d", 90, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.heightPercent; actual != 90 {
		t.Fatalf("expected popup height percent %d, actual %d", 90, actual)
	}
	if !subject.commandModeActive() {
		t.Fatal("expected the submitted command to remain behind the popup")
	}
	if actual := subject.searchWidget.editor.Text(); actual != "messages" {
		t.Fatalf("expected the submitted command to remain visible behind the popup, actual %q", actual)
	}
	if actual := subject.currentViewName(); actual != viewPullRequestBuildInfoName {
		t.Fatalf("expected the popup to own focus above the hidden prompt, actual %q", actual)
	}

	Update(subject, MsgPullRequestBuildRunPopupClosed{})

	if subject.pullRequestBuildRunPopup != nil {
		t.Fatal("expected closing the popup to remove it")
	}
	if subject.commandModeActive() {
		t.Fatal("expected closing the popup to clear command mode")
	}
}

func TestUpdate_GivenCommandMode_WhenSubmittingHistory_ThenItOpensTheHistoryPlaceholder(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	Update(subject, MsgOpenCommandMode{})
	for _, character := range "history" {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}

	Update(subject, MsgSubmitCommand{})

	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the history command to open the placeholder popup")
	}
	if actual := subject.pullRequestBuildRunPopup.title; actual != "history" {
		t.Fatalf("expected popup title %q, actual %q", "history", actual)
	}
}

func TestUpdate_GivenCommandMode_WhenSubmittingUnknownCommandWithOuterWhitespace_ThenItClosesAndReportsTheTrimmedCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	now := time.Date(2026, time.May, 20, 14, 6, 6, 0, time.UTC)
	subject.timingState.now = func() time.Time { return now }
	Update(subject, MsgOpenCommandMode{})
	for _, character := range " nope " {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}

	Update(subject, MsgSubmitCommand{})

	if subject.commandModeActive() {
		t.Fatal("expected unknown command submission to close command mode")
	}
	if subject.pullRequestBuildRunPopup != nil {
		t.Fatal("expected unknown command submission not to open a popup")
	}
	expected := formatStatusLineFailure("Unknown command", errors.New("nope"))
	if actual := subject.statusLinePresenter().Text(); actual != expected {
		t.Fatalf("expected status feedback %q, actual %q", expected, actual)
	}
	expectedRecord := recordedErrorMessage{message: "Unknown command: nope", timestamp: now}
	if len(subject.overlayState.errorMessages) != 1 || subject.overlayState.errorMessages[0] != expectedRecord {
		t.Fatalf("expected recorded error %+v, actual %v", expectedRecord, subject.overlayState.errorMessages)
	}

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	then_noError(t, subject.layout(gui))
	statusView, actualErr := gui.View(viewStatusLineName)
	then_noError(t, actualErr)
	if actual := statusView.FgColor; actual != gocui.GetColor(theme.FailureHex) {
		t.Fatalf("expected unknown command status color %v, actual %v", gocui.GetColor(theme.FailureHex), actual)
	}
}

func TestUpdate_GivenCommandMode_WhenSubmittingEmptyCommand_ThenItCancelsWithoutChangingStatusFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.setFeedback(FocusPullRequestsView, "existing feedback")
	Update(subject, MsgOpenCommandMode{})

	Update(subject, MsgSubmitCommand{})

	if subject.commandModeActive() {
		t.Fatal("expected empty command submission to cancel command mode")
	}
	if actual := subject.statusLinePresenter().Text(); actual != "existing feedback" {
		t.Fatalf("expected status feedback %q, actual %q", "existing feedback", actual)
	}
}

func TestUpdate_GivenBlockingInput_WhenOpeningCommandMode_ThenItLeavesCommandModeInactive(t *testing.T) {
	testCases := []struct {
		name  string
		block func(*Program)
	}{
		{name: "search", block: func(program *Program) { program.model.StartSearch() }},
		{name: "actions popup", block: func(program *Program) { program.model.OpenActionsPopup(1) }},
		{name: "modal editor", block: func(program *Program) { program.overlayState.modalEditor = newLineModalEditorState("Prompt", "") }},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject := NewProgramWithModel(given_model())
			testCase.block(subject)

			Update(subject, MsgOpenCommandMode{})

			if subject.commandModeActive() {
				t.Fatal("expected blocking input to keep command mode inactive")
			}
		})
	}
}

func TestKeybindingActions_GivenCommandMode_WhenListingBindings_ThenItAddsCommandEntryAndRetainsConfigurableColonActivation(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	definition, ok := sharedKeybindingDefinitionFor("command_mode")
	if !ok {
		t.Fatal("expected a shared command_mode definition")
	}
	if actual := definition.bindings; len(actual) != 1 || actual[0] != ":" {
		t.Fatalf("expected command_mode binding [:], actual %v", actual)
	}

	before := subject.keybindingSpecs()
	then_bindingExists(t, before, keybindingSpec{viewName: viewUserName, key: ':', handler: subject.openCommandMode})
	then_bindingExists(t, before, keybindingSpec{viewName: viewPullRequestsName, key: ':', handler: subject.openCommandMode})
	then_bindingExists(t, before, keybindingSpec{viewName: viewDetailName, key: ':', handler: subject.openCommandMode})
	then_bindingDoesNotExist(t, before, viewSearchName, ':')
	then_bindingDoesNotExist(t, before, viewActionsPopupName, ':')

	Update(subject, MsgOpenCommandMode{})
	actual := subject.keybindingSpecs()
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEnter, handler: subject.submitCommand})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyCtrlJ, handler: subject.submitCommand})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEsc, handler: subject.cancelCommand})
}

func TestPullRequestBuildRunPopup_GivenDefaultLayout_WhenCalculatingOverlayFrame_ThenItUsesNinetyPercentDimensions(t *testing.T) {
	actual := pullRequestBuildRunPopupOverlayFrame(pullRequestBuildRunPopupLayoutState{}, 100, 100)

	if actual.x1-actual.x0+1 != 90 {
		t.Fatalf("expected popup width %d, actual %d", 90, actual.x1-actual.x0+1)
	}
	if actual.y1-actual.y0+1 != 90 {
		t.Fatalf("expected popup height %d, actual %d", 90, actual.y1-actual.y0+1)
	}
}

func TestLayout_GivenCommandMode_WhenRendering_ThenItShowsTheColonPromptAtTheBottomWithTheCursor(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	then_noError(t, subject.layout(gui))

	Update(subject, MsgOpenCommandMode{})
	for _, character := range "history" {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}
	then_noError(t, subject.layout(gui))

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	if !strings.Contains(searchView.Buffer(), ":history") {
		t.Fatalf("expected command prompt %q, actual %q", ":history", searchView.Buffer())
	}
	if !gui.Cursor {
		t.Fatal("expected command mode to show the cursor")
	}
	then_viewDoesNotExist(t, gui, viewStatusLineKeyHintsName)
	then_currentViewNameIs(t, gui, viewSearchName)
}
