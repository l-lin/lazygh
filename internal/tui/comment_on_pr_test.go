package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingPullRequestCommentBindings_ThenTheShortcutExistsOnlyInPullRequestContextsAndTheComposerSupportsSubmit(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'c', handler: subject.openPullRequestCommentComposer})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'c', handler: subject.openPullRequestCommentComposer})
	then_bindingDoesNotExist(t, actual, viewUserName, 'c')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewModalEditorName, key: gocui.KeyAltEnter, handler: subject.submitModalEditor})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewModalEditorName, key: gocui.KeyCtrlS, handler: subject.submitModalEditor})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewModalEditorName, key: gocui.KeyEsc, handler: subject.closeModalEditor})
}

func TestPullRequestCommentComposer_GivenPullRequestsView_WhenOpening_ThenItShowsACenteredSevenLinePopupAndTakesFocus(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)
	if !gui.Cursor {
		t.Fatal("expected the cursor to be visible while the composer is focused")
	}

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, "Comment on PR") {
		t.Fatalf("expected composer title to contain %q, actual %q", "Comment on PR", composerView.Title)
	}
	then_viewFooterIsRenderedOnBottomBorder(t, gui, viewModalEditorName, "<Alt-Enter> to submit")

	x0, y0, x1, y1, actualErr := gui.ViewPosition(viewModalEditorName)
	then_noError(t, actualErr)
	if actual := y1 - y0 + 1; actual != 7 {
		t.Fatalf("expected composer height 7, actual %d", actual)
	}
	if actual := x1 - x0 + 1; actual != 60 {
		t.Fatalf("expected composer width 60, actual %d", actual)
	}
	if x0 != 30 || y0 != 11 {
		t.Fatalf("expected composer position (%d,%d), actual (%d,%d)", 30, 11, x0, y0)
	}
}

func TestPullRequestCommentComposer_GivenConnectedUserDetail_WhenOpening_ThenItDoesNothing(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewModalEditorName)
}

func TestPullRequestCommentComposer_GivenOpenComposer_WhenHandlingGlobalNavigation_ThenUnderlyingFocusDoesNotChange(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextSideView(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewModalEditorName)
	if subject.model.Focus() != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, subject.model.Focus())
	}
}

func TestPullRequestCommentComposer_GivenOpenComposer_WhenSubmittingWithAltEnter_ThenItPostsTheCurrentBuffer(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	given_commentComposer(t, gui, subject)
	subject.modalEditor.editor.SetText("Ship it")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.commentCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected comment calls %v, actual %v", []string{"acme/widgets#42"}, loader.commentCalls)
	}
	if !reflect.DeepEqual(loader.commentBodies, []string{"Ship it"}) {
		t.Fatalf("expected comment bodies %v, actual %v", []string{"Ship it"}, loader.commentBodies)
	}
}

func TestPullRequestCommentComposer_GivenOpenComposer_WhenSubmittingWithControlS_ThenItPostsTheCurrentBuffer(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	given_commentComposer(t, gui, subject)
	subject.modalEditor.editor.SetText("Ship it faster")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlS)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.commentCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected comment calls %v, actual %v", []string{"acme/widgets#42"}, loader.commentCalls)
	}
	if !reflect.DeepEqual(loader.commentBodies, []string{"Ship it faster"}) {
		t.Fatalf("expected comment bodies %v, actual %v", []string{"Ship it faster"}, loader.commentBodies)
	}
}

func TestPullRequestCommentComposer_GivenSuccessfulSubmit_WhenSubmitting_ThenItClosesThePopupRefreshesTheDetailAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Ship it")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewModalEditorName)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}

	then_statusLineContains(t, gui, pullRequestCommentSuccessMessage)
	then_footerTextIs(t, gui, viewPullRequestsFooterName, "? Help  / Search  a Actions")
}

func TestPullRequestCommentComposer_GivenSubmitFailure_WhenSubmitting_ThenItKeepsTheDraftVisibleAndShowsTheError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{commentErr: errors.New("boom")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Line one\nLine two")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView := given_commentComposer(t, gui, subject)
	if !strings.Contains(composerView.Buffer(), "Line one\nLine two") {
		t.Fatalf("expected composer buffer to contain the draft, actual %q", composerView.Buffer())
	}
	if !strings.Contains(composerView.Title, "boom") {
		t.Fatalf("expected composer title to contain %q, actual %q", "boom", composerView.Title)
	}
}

func TestPullRequestCommentComposer_GivenPullRequestDetail_WhenSubmitting_ThenItReturnsToDetailAndRefreshesIt(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestCommentComposer(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Ship it from detail")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}

	then_statusLineContains(t, gui, pullRequestCommentSuccessMessage)
	then_footerTextIs(t, gui, viewDetailFooterName, "? Help  / Search  a Actions")
}

func TestHelpPopup_GivenPullRequestContext_WhenTogglingHelp_ThenItListsTheCommentShortcut(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	if !strings.Contains(helpView.Buffer(), "Comment on PR") {
		t.Fatalf("expected help buffer to contain %q, actual %q", "Comment on PR", helpView.Buffer())
	}
}

func given_pullRequestCommentModel() *Model {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}),
	})
	return model
}

func given_pullRequestCommentProgram(model *Model, loader *fakePullRequestDetailLoader) *Program {
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}

func given_commentComposer(t *testing.T, gui *gocui.Gui, subject *Program) *gocui.View {
	t.Helper()

	actual, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	subject.renderModalEditorView(actual)
	return actual
}

func then_bindingDoesNotExist(t *testing.T, specs []keybindingSpec, expectedViewName string, expectedKey any) {
	t.Helper()

	for _, actual := range specs {
		if actual.viewName == expectedViewName && reflect.DeepEqual(actual.key, expectedKey) {
			t.Fatalf("expected binding %q %v to be absent, actual %+v", expectedViewName, expectedKey, actual)
		}
	}
}
