package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenRendering_ThenItStartsDirectlyInBrowserModeOnFullscreenDescription(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected browser mode to stay active")
	}
	if !reflect.DeepEqual(loader.startReviewCalls, []string(nil)) {
		t.Fatalf("expected no pending review start calls, actual %v", loader.startReviewCalls)
	}
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.model.Focus())
	}
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected layout size %v, actual %v", PaneLayoutFullscreen, subject.model.PaneLayoutSize())
	}
	if subject.model.FullscreenPane() != FocusDetailView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusDetailView, subject.model.FullscreenPane())
	}
	if subject.activeDetailTab != DescriptionDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", DescriptionDetailTab, subject.activeDetailTab)
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	if selectedSummary.Repository.NameWithOwner != "acme/rocket" {
		t.Fatalf("expected repository %q, actual %q", "acme/rocket", selectedSummary.Repository.NameWithOwner)
	}
	if selectedSummary.Number != 77 {
		t.Fatalf("expected pull request number %d, actual %d", 77, selectedSummary.Number)
	}
	if selectedSummary.URL != "https://github.com/acme/rocket/pull/77" {
		t.Fatalf("expected pull request url %q, actual %q", "https://github.com/acme/rocket/pull/77", selectedSummary.URL)
	}

	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewPullRequestsName)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 0)
	if !strings.Contains(detailView.Buffer(), "Body 77") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Body 77", detailView.Buffer())
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLAfterLayout_WhenOpening_ThenItShowsThatPullRequestInBrowserMode(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected browser mode to stay active")
	}
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.model.Focus())
	}
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected layout size %v, actual %v", PaneLayoutFullscreen, subject.model.PaneLayoutSize())
	}
	if subject.activeDetailTab != DescriptionDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", DescriptionDetailTab, subject.activeDetailTab)
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	if selectedSummary.Repository.NameWithOwner != "acme/rocket" {
		t.Fatalf("expected repository %q, actual %q", "acme/rocket", selectedSummary.Repository.NameWithOwner)
	}
	if selectedSummary.Number != 77 {
		t.Fatalf("expected pull request number %d, actual %d", 77, selectedSummary.Number)
	}

	then_currentViewNameIs(t, gui, viewDetailName)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Body 77") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Body 77", detailView.Buffer())
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingOpenPullRequestByURLBindings_ThenCtrlVIsAvailableInThePullRequestBrowserViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: gocui.KeyCtrlV, handler: subject.openPullRequestByURLShortcut})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyCtrlV, handler: subject.openPullRequestByURLShortcut})
	then_bindingDoesNotExist(t, actual, viewUserName, gocui.KeyCtrlV)
	then_bindingDoesNotExist(t, actual, viewNotificationsName, gocui.KeyCtrlV)
}

func TestActionsPopup_GivenPullRequestsView_WhenExecutingOpenPullRequestByURL_ThenItOpensTheInputBox(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_pullRequestByURLProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("open pr from url", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "open pr from url"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewModalEditorName)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	if subject.modalEditor == nil || subject.modalEditor.lineEditor == nil {
		t.Fatal("expected the PR URL prompt to use the single-line editor")
	}
	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(modalView.Title, "Open PR from URL") {
		t.Fatalf("expected modal title to contain %q, actual %q", "Open PR from URL", modalView.Title)
	}
	_, y0, _, y1, actualErr := gui.ViewPosition(viewModalEditorName)
	then_noError(t, actualErr)
	if actual := y1 - y0 + 1; actual != 3 {
		t.Fatalf("expected the PR URL prompt height %d, actual %d", 3, actual)
	}
}

func TestOpenPullRequestByURL_GivenTheURLInputPopup_WhenPressingEnter_ThenItSubmitsTheRequestedPullRequest(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	loader.details["acme/widgets#13"] = githubcli.PullRequestDetail{
		Title:       "Widgets PR",
		Number:      13,
		URL:         "https://github.com/acme/widgets/pull/13",
		Body:        "Body 13",
		BaseRefName: "main",
		HeadRefName: "feature/widgets",
		State:       "OPEN",
	}
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_pullRequestByURLProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("open pr from url", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "open pr from url"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if subject.modalEditor == nil || subject.modalEditor.lineEditor == nil {
		t.Fatal("expected the PR URL prompt to use the single-line editor")
	}
	subject.modalEditor.lineEditor.SetText("https://github.com/acme/widgets/pull/13")
	if actual := subject.editModalEditor(modalView, gocui.KeyEnter, 0, gocui.ModNone); !actual {
		t.Fatal("expected Enter to submit the PR URL prompt")
	}

	then_currentViewNameIs(t, gui, viewDetailName)
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.model.Focus())
	}
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected layout size %v, actual %v", PaneLayoutFullscreen, subject.model.PaneLayoutSize())
	}
	if subject.model.FullscreenPane() != FocusDetailView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusDetailView, subject.model.FullscreenPane())
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	if selectedSummary.Repository.NameWithOwner != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", selectedSummary.Repository.NameWithOwner)
	}
	if selectedSummary.Number != 13 {
		t.Fatalf("expected pull request number %d, actual %d", 13, selectedSummary.Number)
	}
	if selectedSummary.URL != "https://github.com/acme/widgets/pull/13" {
		t.Fatalf("expected pull request url %q, actual %q", "https://github.com/acme/widgets/pull/13", selectedSummary.URL)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Body 13") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Body 13", detailView.Buffer())
	}
}

func TestOpenPullRequestByURL_GivenAnInvalidGitHubURL_WhenOpening_ThenItReturnsAValidationError(t *testing.T) {
	subject := given_pullRequestByURLProgram(given_model(), &fakePullRequestDetailLoader{})

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/widgets/issues/42")

	if !errors.Is(actualErr, githubdomain.ErrInvalidPullRequestURL) {
		t.Fatalf("expected error %v, actual %v", githubdomain.ErrInvalidPullRequestURL, actualErr)
	}
	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the validation error")
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenResolvingTheSelectedActionTargetFromViewZero_ThenItUsesTheOpenedPullRequestDetail(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	actual, ok := subject.selectedPullRequestActionTarget()
	if !ok {
		t.Fatal("expected a pull request action target")
	}

	expected := pullRequestActionTarget{repository: "acme/rocket", number: 77, title: "Rocket PR", body: "Body 77"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected action target %+v, actual %+v", expected, actual)
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenYankingFromTheActionsPopup_ThenItUsesTheOpenedPullRequestURL(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	clipboardWriter := &fakeClipboardWriter{}
	subject := given_pullRequestByURLProgram(given_model(), loader)
	subject.clipboardWriter = clipboardWriter

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("clipboard", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clipboard"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(clipboardWriter.writes, []string{"https://github.com/acme/rocket/pull/77"}) {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/rocket/pull/77"}, clipboardWriter.writes)
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenOpeningInBrowserFromTheActionsPopup_ThenItUsesTheOpenedPullRequestIdentity(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "browser"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.openBrowserCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected open browser calls %v, actual %v", []string{"acme/rocket#77"}, loader.openBrowserCalls)
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenStartingReviewFromTheActionsPopup_ThenItUsesTheOpenedPullRequestIdentity(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("start review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "start review"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected start review calls %v, actual %v", []string{"acme/rocket#77"}, loader.startReviewCalls)
	}
	if !subject.reviewSession.active {
		t.Fatal("expected review mode to become active")
	}
	if subject.reviewSession.summary.Repository.NameWithOwner != "acme/rocket" {
		t.Fatalf("expected review repository %q, actual %q", "acme/rocket", subject.reviewSession.summary.Repository.NameWithOwner)
	}
	if subject.reviewSession.summary.Number != 77 {
		t.Fatalf("expected review pull request number %d, actual %d", 77, subject.reviewSession.summary.Number)
	}
}

func TestOpenPullRequestByURL_GivenAValidGitHubPRURLBeforeLayout_WhenRefreshingFromTheActionsPopup_ThenItKeepsTheOpenedPullRequestSelected(t *testing.T) {
	loader := given_pullRequestByURLLoader()
	subject := given_pullRequestByURLProgram(given_model(), loader)

	actualErr := subject.OpenPullRequestByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("refresh", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "refresh"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/rocket#77", "acme/rocket#77"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/rocket#77", "acme/rocket#77"}, loader.detailCalls)
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	if selectedSummary.Repository.NameWithOwner != "acme/rocket" {
		t.Fatalf("expected repository %q, actual %q", "acme/rocket", selectedSummary.Repository.NameWithOwner)
	}
	if selectedSummary.Number != 77 {
		t.Fatalf("expected pull request number %d, actual %d", 77, selectedSummary.Number)
	}
}

func given_pullRequestByURLProgram(model *Model, loader *fakePullRequestDetailLoader) *Program {
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}

func given_pullRequestByURLLoader() *fakePullRequestDetailLoader {
	return &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Unrelated PR",
			Number:     13,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/13",
			Body:       "Unrelated summary body",
			State:      "OPEN",
		}},
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:       "Rocket PR",
				Number:      77,
				URL:         "https://github.com/acme/rocket/pull/77",
				Body:        "Body 77",
				BaseRefName: "main",
				HeadRefName: "feature/view-url",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
		startReviewID: "PRR_pending",
	}
}
