package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenEditTitleActionSelected_WhenExecuting_ThenItOpensTheSingleLineTitleEditor(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if subject.modalEditor.lineEditor == nil {
		t.Fatal("expected the title editor to use the line editor")
	}
	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(titleView.Title, pullRequestTitleEditorTitle) {
		t.Fatalf("expected title editor title to contain %q, actual %q", pullRequestTitleEditorTitle, titleView.Title)
	}
	if strings.TrimSpace(titleView.Buffer()) != "First PR" {
		t.Fatalf("expected title editor buffer %q, actual %q", "First PR", titleView.Buffer())
	}
}

func TestPullRequestTitleEditor_GivenOpenEditor_WhenHandlingLineEditingKeys_ThenItReusesTheSingleLineEditorCore(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandled := subject.editModalEditor(titleView, gocui.KeyCtrlA, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-a to be handled")
	}
	actualHandled = subject.editModalEditor(titleView, 0, 'X', gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected insertion to be handled")
	}
	actualHandled = subject.editModalEditor(titleView, gocui.KeyCtrlE, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-e to be handled")
	}
	actualHandled = subject.editModalEditor(titleView, gocui.KeyCtrlH, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-h to be handled")
	}

	if actual := strings.TrimSpace(titleView.Buffer()); actual != "XFirst P" {
		t.Fatalf("expected title buffer %q, actual %q", "XFirst P", actual)
	}
}

func TestEditPullRequestTitle_GivenSuccessfulSubmit_WhenPressingEnter_ThenItRefreshesTheSelectedPullRequestSummaryAndDetail(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details:        map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}},
		myPullRequests: []githubcli.PullRequest{given_actionsPopupPullRequest()},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.lineEditor.SetText("Renamed PR")

	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)
	actualErr = actualHandler(gui, titleView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.editTitleCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected edit title calls %v, actual %v", []string{"acme/widgets#42"}, loader.editTitleCalls)
	}
	if !reflect.DeepEqual(loader.editTitleValues, []string{"Renamed PR"}) {
		t.Fatalf("expected edit title values %v, actual %v", []string{"Renamed PR"}, loader.editTitleValues)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Renamed PR") {
		t.Fatalf("expected pull requests buffer to contain %q, actual %q", "Renamed PR", pullRequestsView.Buffer())
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Renamed PR") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Renamed PR", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestTitleEditSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestEditPullRequestTitle_GivenSubmitFailure_WhenPressingEnter_ThenItKeepsTheDraftVisible(t *testing.T) {
	loader := &fakePullRequestDetailLoader{editTitleErr: errors.New("boom")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.lineEditor.SetText("Broken title")

	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)
	actualErr = actualHandler(gui, titleView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if !strings.Contains(titleView.Buffer(), "Broken title") {
		t.Fatalf("expected title buffer to contain %q, actual %q", "Broken title", titleView.Buffer())
	}
	if strings.Contains(titleView.Title, "boom") {
		t.Fatalf("expected title editor title to hide %q, actual %q", "boom", titleView.Title)
	}
	then_transientErrorPopupContains(t, gui, "boom")
}

func TestEditPullRequestTitle_GivenControlG_WhenOpeningTheExternalEditor_ThenItReplacesTheDraftWithTheSavedSingleLineText(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.externalEditor = &fakeExternalEditor{editedText: "Edited\nin $EDITOR\n"}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.lineEditor.SetText("Draft title")

	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlG)
	actualErr = actualHandler(gui, titleView)
	then_noError(t, actualErr)
	if subject.externalEditor.(*fakeExternalEditor).receivedText != "Draft title" {
		t.Fatalf("expected external editor input %q, actual %q", "Draft title", subject.externalEditor.(*fakeExternalEditor).receivedText)
	}
	if actual := strings.TrimSpace(titleView.Buffer()); actual != "Edited in $EDITOR" {
		t.Fatalf("expected title editor buffer %q, actual %q", "Edited in $EDITOR", actual)
	}
}

func TestActionsPopup_GivenEditDescriptionActionSelected_WhenExecuting_ThenItOpensTheDescriptionEditorSeededWithTheCurrentBody(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Rich body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	descriptionView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(descriptionView.Title, pullRequestDescriptionEditorTitle) {
		t.Fatalf("expected description editor title to contain %q, actual %q", pullRequestDescriptionEditorTitle, descriptionView.Title)
	}
	if !strings.Contains(descriptionView.Buffer(), "Rich body") {
		t.Fatalf("expected description editor buffer to contain %q, actual %q", "Rich body", descriptionView.Buffer())
	}
	_, y0, _, y1, actualErr := gui.ViewPosition(viewModalEditorName)
	then_noError(t, actualErr)
	if actual := y1 - y0 + 1; actual != 20 {
		t.Fatalf("expected description editor height %d, actual %d", 20, actual)
	}
}

func TestEditPullRequestDescription_GivenControlG_WhenOpeningTheExternalEditor_ThenItReplacesTheDraftWithTheSavedText(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.externalEditor = &fakeExternalEditor{editedText: "Edited in $EDITOR"}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Draft body")

	descriptionView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlG)
	actualErr = actualHandler(gui, descriptionView)
	then_noError(t, actualErr)
	if subject.externalEditor.(*fakeExternalEditor).receivedText != "Draft body" {
		t.Fatalf("expected external editor input %q, actual %q", "Draft body", subject.externalEditor.(*fakeExternalEditor).receivedText)
	}
	if !strings.Contains(descriptionView.Buffer(), "Edited in $EDITOR") {
		t.Fatalf("expected description editor buffer to contain %q, actual %q", "Edited in $EDITOR", descriptionView.Buffer())
	}
}

func TestEditPullRequestDescription_GivenMissingExternalEditor_WhenOpeningIt_ThenItShowsTheErrorAndKeepsTheDraft(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.externalEditor = &fakeExternalEditor{editErr: errors.New("EDITOR is not set")}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Draft body")

	descriptionView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlG)
	actualErr = actualHandler(gui, descriptionView)
	then_noError(t, actualErr)
	if !strings.Contains(descriptionView.Buffer(), "Draft body") {
		t.Fatalf("expected description editor buffer to contain %q, actual %q", "Draft body", descriptionView.Buffer())
	}
	if !strings.Contains(descriptionView.Title, "EDITOR is not set") {
		t.Fatalf("expected description editor title to contain %q, actual %q", "EDITOR is not set", descriptionView.Title)
	}
}

func TestEditPullRequestDescription_GivenExternalEditorFailure_WhenOpeningIt_ThenItShowsTheErrorAndKeepsTheDraft(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.externalEditor = &fakeExternalEditor{editErr: errors.New("exit status 1")}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Draft body")

	descriptionView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlG)
	actualErr = actualHandler(gui, descriptionView)
	then_noError(t, actualErr)
	if !strings.Contains(descriptionView.Buffer(), "Draft body") {
		t.Fatalf("expected description editor buffer to contain %q, actual %q", "Draft body", descriptionView.Buffer())
	}
	if !strings.Contains(descriptionView.Title, "exit status 1") {
		t.Fatalf("expected description editor title to contain %q, actual %q", "exit status 1", descriptionView.Title)
	}
}

func TestEditPullRequestDescription_GivenSuccessfulSubmit_WhenSubmitting_ThenItRefreshesTheSelectedPullRequestSummaryAndDetail(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details:        map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}},
		myPullRequests: []githubcli.PullRequest{given_actionsPopupPullRequest()},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.editDescriptionCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected edit description calls %v, actual %v", []string{"acme/widgets#42"}, loader.editDescriptionCalls)
	}
	if !reflect.DeepEqual(loader.editDescriptionBodies, []string{"Updated body"}) {
		t.Fatalf("expected edit description bodies %v, actual %v", []string{"Updated body"}, loader.editDescriptionBodies)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}

	updatedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	if updatedSummary.Body != "Updated body" {
		t.Fatalf("expected selected pull request body %q, actual %q", "Updated body", updatedSummary.Body)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Updated body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Updated body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestDescriptionEditSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestEditPullRequestDescription_GivenSubmitFailure_WhenSubmitting_ThenItKeepsTheDraftVisible(t *testing.T) {
	loader := &fakePullRequestDetailLoader{editDescriptionErr: errors.New("boom")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestDescriptionEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestDescriptionEditorTitle))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Broken body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	descriptionView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(descriptionView.Buffer(), "Broken body") {
		t.Fatalf("expected description buffer to contain %q, actual %q", "Broken body", descriptionView.Buffer())
	}
	if strings.Contains(descriptionView.Title, "boom") {
		t.Fatalf("expected description editor title to hide %q, actual %q", "boom", descriptionView.Title)
	}
	then_transientErrorPopupContains(t, gui, "boom")
}

type fakeExternalEditor struct {
	receivedText string
	editedText   string
	editErr      error
}

func (editor *fakeExternalEditor) Edit(_ *gocui.Gui, text string) (string, error) {
	editor.receivedText = text
	if editor.editErr != nil {
		return "", editor.editErr
	}
	return editor.editedText, nil
}
