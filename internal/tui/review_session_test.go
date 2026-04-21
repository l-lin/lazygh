package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestReviewMode_GivenStartReviewActionSelected_WhenExecuting_ThenItRepurposesTheExistingThreePanesForReviewWork(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 5,
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to be active")
	}
	if subject.reviewSession.pendingReviewID != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", subject.reviewSession.pendingReviewID)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.Title != "[1]-Metadata" {
		t.Fatalf("expected metadata view title %q, actual %q", "[1]-Metadata", metadataView.Title)
	}
	if !strings.Contains(metadataView.Buffer(), "Pending review: PRR_pending") {
		t.Fatalf("expected metadata view to contain %q, actual %q", "Pending review: PRR_pending", metadataView.Buffer())
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if filesView.Title != "[2]-Files" {
		t.Fatalf("expected files view title %q, actual %q", "[2]-Files", filesView.Title)
	}
	if !strings.Contains(filesView.Buffer(), "5 files pending diff load") {
		t.Fatalf("expected files view to contain %q, actual %q", "5 files pending diff load", filesView.Buffer())
	}
	if len(filesView.Tabs) != 0 {
		t.Fatalf("expected review files view to hide pull request tabs, actual %v", filesView.Tabs)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != "[0]-Diff" {
		t.Fatalf("expected detail view title %q, actual %q", "[0]-Diff", detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "TODO 19 will replace this placeholder") {
		t.Fatalf("expected detail view to explain the placeholder diff, actual %q", detailView.Buffer())
	}
	if len(detailView.Tabs) != 0 {
		t.Fatalf("expected review diff view to hide browser detail tabs, actual %v", detailView.Tabs)
	}
}

func TestReviewMode_GivenFullscreenPullRequestBrowser_WhenStartingAndExiting_ThenItUsesThreePanesInReviewModeAndRestoresThePreviousFullscreenLayout(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.FocusPullRequestsView()
	model.ShrinkFocusedPane()
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected fullscreen layout before starting review mode, actual %v", subject.model.PaneLayoutSize())
	}

	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() == PaneLayoutFullscreen {
		t.Fatalf("expected review mode to leave fullscreen and show three panes, actual %v", subject.model.PaneLayoutSize())
	}
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEsc)
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected fullscreen layout after exiting review mode, actual %v", subject.model.PaneLayoutSize())
	}
	if subject.model.FullscreenPane() != FocusPullRequestsView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusPullRequestsView, subject.model.FullscreenPane())
	}
	then_viewDoesNotExist(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewDetailName)
}

func TestReviewMode_GivenItStartedFromPullRequestDetail_WhenExiting_ThenItRestoresThePriorBrowserFocusSelectionAndDetailTab(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(RequestedPullRequestsTab, []PullRequestRow{
		requestedPullRequestRow(githubcli.PullRequest{Title: "Requested PR 1", Number: 7, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 7"}),
		requestedPullRequestRow(githubcli.PullRequest{Title: "Requested PR 2", Number: 8, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 8"}),
	})
	model.NextPullRequestTab()
	model.MoveSelectionDown()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#8": {
				Title:       "Requested PR 2",
				Number:      8,
				Body:        "Body 8",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	subject.activeDetailTab = CommentsDetailTab
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEsc)
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected review mode to be inactive after exiting")
	}
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.model.Focus())
	}
	if subject.model.ActivePullRequestTab() != RequestedPullRequestsTab {
		t.Fatalf("expected active tab %v, actual %v", RequestedPullRequestsTab, subject.model.ActivePullRequestTab())
	}
	if subject.model.SelectedPullRequestIndex(RequestedPullRequestsTab) != 1 {
		t.Fatalf("expected requested pull request selection 1, actual %d", subject.model.SelectedPullRequestIndex(RequestedPullRequestsTab))
	}
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected detail tab %v, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
	then_currentViewNameIs(t, gui, viewDetailName)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "No comments yet.") {
		t.Fatalf("expected browser detail content to be restored, actual %q", detailView.Buffer())
	}
}

func given_startingReviewMode(t *testing.T, gui *gocui.Gui, subject *Program) error {
	t.Helper()

	actualErr := subject.openActionsPopup(gui, nil)
	if actualErr != nil {
		return actualErr
	}
	subject.model.UpdateActionsPopupSearch("start review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "start review"))
	actualErr = subject.refreshViews(gui)
	if actualErr != nil {
		return actualErr
	}

	return subject.executeSelectedActionsPopupAction(gui, nil)
}
