package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestPullRequestCustomSearch_GivenProgram_WhenListingKeybindings_ThenItBindsColonInThePullRequestsView(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: ':', handler: subject.openPullRequestCustomSearch})
}

func TestActionsPopup_GivenPullRequestsView_WhenOpening_ThenItShowsTheCustomSearchAction(t *testing.T) {
	subject := given_pullRequestCustomSearchProgram(&fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Custom search") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Custom search", popupView.Buffer())
	}
}

func TestActionsPopup_GivenPullRequestsView_WhenExecutingCustomSearch_ThenItOpensThePrefilledSearchPopup(t *testing.T) {
	subject := given_pullRequestCustomSearchProgram(&fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("custom search", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "custom search"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewModalEditorName)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	if subject.modalEditor == nil || subject.modalEditor.lineEditor == nil {
		t.Fatal("expected the custom search popup to use the single-line editor")
	}
	if actual := subject.modalEditor.Text(); actual != "--author @me --state open --sort updated --order desc" {
		t.Fatalf("expected the custom search criteria %q, actual %q", "--author @me --state open --sort updated --order desc", actual)
	}
}

func TestPullRequestCustomSearch_GivenPullRequestsView_WhenOpening_ThenItPrefillsTheActiveSearchCriteriaAndUsesASingleInputLine(t *testing.T) {
	subject := given_pullRequestCustomSearchProgram(&fakePullRequestDetailLoader{})
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, ':')(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if subject.modalEditor == nil || subject.modalEditor.lineEditor == nil {
		t.Fatal("expected the custom search popup to use the single-line editor")
	}
	if actual := subject.modalEditor.Text(); actual != "--author @me --state open --sort updated --order desc" {
		t.Fatalf("expected the custom search criteria %q, actual %q", "--author @me --state open --sort updated --order desc", actual)
	}
	then_statusLineKeyHintsAre(t, gui, "Enter: submit, Ctrl+G: editor, Escape: cancel")
	_, y0, _, y1, actualErr := gui.ViewPosition(viewModalEditorName)
	then_noError(t, actualErr)
	if actual := y1 - y0 + 1; actual != 3 {
		t.Fatalf("expected custom search popup height %d, actual %d", 3, actual)
	}
}

func TestPullRequestCustomSearch_GivenSubmittedCriteria_WhenSubmitting_ThenItCreatesAndReusesTheCustomTab(t *testing.T) {
	loader := &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{
		Title:      "Custom PR",
		Number:     108,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/108",
		State:      "OPEN",
	}}}
	subject := given_pullRequestCustomSearchProgram(loader)
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, ':')(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.lineEditor.SetText("--author @me --state open --label bug")
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)(gui, nil)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"My PRs", "My reviews", "Requested", "Custom (1)"}, 3)
	if actual := subject.model.PullRequestTabLabel(PullRequestTab(3)); actual != "Custom" {
		t.Fatalf("expected the raw custom tab label %q, actual %q", "Custom", actual)
	}
	if actual := subject.model.ActivePullRequestTab(); actual != PullRequestTab(3) {
		t.Fatalf("expected the custom tab index %d, actual %d", 3, actual)
	}
	actualRows := subject.model.PullRequestRows(PullRequestTab(3))
	if len(actualRows) != 1 || actualRows[0].Summary == nil || actualRows[0].Summary.Title != "Custom PR" {
		t.Fatalf("expected the custom tab rows to load the submitted search, actual %+v", actualRows)
	}

	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, ':')(gui, nil)
	then_noError(t, actualErr)
	if actual := subject.modalEditor.Text(); actual != "--author @me --state open --label bug" {
		t.Fatalf("expected the active custom search criteria %q, actual %q", "--author @me --state open --label bug", actual)
	}
	subject.modalEditor.lineEditor.SetText("--author @me --state closed")
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)(gui, nil)
	then_noError(t, actualErr)

	pullRequestsView, actualErr = gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"My PRs", "My reviews", "Requested", "Custom (1)"}, 3)
	if !reflect.DeepEqual(loader.listPullRequestCommands, [][]string{
		{"search", "prs", "--author", "@me", "--state", "open", "--label", "bug"},
		{"search", "prs", "--author", "@me", "--state", "closed"},
	}) {
		t.Fatalf("expected the custom tab commands %v, actual %v", [][]string{
			{"search", "prs", "--author", "@me", "--state", "open", "--label", "bug"},
			{"search", "prs", "--author", "@me", "--state", "closed"},
		}, loader.listPullRequestCommands)
	}
}

func given_pullRequestCustomSearchProgram(loader *fakePullRequestDetailLoader) *Program {
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.ApplyPullRequestSearches(appconfig.DefaultPullRequestSearches())
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.additionalPullRequestsLoadStarted[PullRequestTab(2)] = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}
