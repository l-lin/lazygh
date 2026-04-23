package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenReviewPRFromURLActionSelected_WhenSubmittingAValidURL_ThenItStartsReviewModeForThatPullRequest(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:        "Rocket PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/review-url",
				State:        "OPEN",
				ChangedFiles: 2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("paste", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "paste"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if subject.modalEditor == nil || subject.modalEditor.lineEditor == nil {
		t.Fatal("expected a line modal editor for the pull request URL prompt")
	}
	then_currentViewNameIs(t, gui, viewModalEditorName)
	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(modalView.Title, "Review PR from URL") {
		t.Fatalf("expected modal title to contain %q, actual %q", "Review PR from URL", modalView.Title)
	}

	subject.modalEditor.lineEditor.SetText(" https://github.com/acme/rocket/pull/77/files#diff-1 ")
	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = submitHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected review start calls %v, actual %v", []string{"acme/rocket#77"}, loader.startReviewCalls)
	}
	if !subject.reviewSession.active {
		t.Fatal("expected review mode to be active")
	}
	if subject.reviewSession.summary.Repository.NameWithOwner != "acme/rocket" {
		t.Fatalf("expected review repository %q, actual %q", "acme/rocket", subject.reviewSession.summary.Repository.NameWithOwner)
	}
	if subject.reviewSession.summary.Number != 77 {
		t.Fatalf("expected review pull request number %d, actual %d", 77, subject.reviewSession.summary.Number)
	}
	if subject.reviewSession.summary.URL != "https://github.com/acme/rocket/pull/77" {
		t.Fatalf("expected canonical pull request url %q, actual %q", "https://github.com/acme/rocket/pull/77", subject.reviewSession.summary.URL)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if filesView.Title != reviewModeFilesTitle {
		t.Fatalf("expected files view title %q, actual %q", reviewModeFilesTitle, filesView.Title)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected diff view title %q, actual %q", reviewModeDiffTitle, detailView.Title)
	}
}

func TestActionsPopup_GivenReviewPRFromURLActionSelected_WhenSubmittingAnInvalidURL_ThenItKeepsTheDraftVisibleAndShowsTheValidationError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("paste", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "paste"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	subject.modalEditor.lineEditor.SetText("https://github.com/acme/widgets/issues/42")
	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = submitHandler(gui, nil)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the validation error")
	}
	if len(loader.startReviewCalls) != 0 {
		t.Fatalf("expected no review start calls, actual %v", loader.startReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewModalEditorName)
	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(modalView.Title, githubcli.ErrInvalidPullRequestURL.Error()) {
		t.Fatalf("expected modal title to contain %q, actual %q", githubcli.ErrInvalidPullRequestURL.Error(), modalView.Title)
	}
	if !strings.Contains(modalView.Buffer(), "https://github.com/acme/widgets/issues/42") {
		t.Fatalf("expected modal buffer to contain the invalid url draft, actual %q", modalView.Buffer())
	}
}

func TestOpenReviewByURL_GivenAValidGitHubPRURLBeforeLayout_WhenRendering_ThenItStartsDirectlyInReviewMode(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:        "Rocket PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/review-url",
				State:        "OPEN",
				ChangedFiles: 2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_model(), loader)

	actualErr := subject.OpenReviewByURL("https://github.com/acme/rocket/pull/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to start before the first layout")
	}
	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected review start calls %v, actual %v", []string{"acme/rocket#77"}, loader.startReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.Title != reviewModeMetadataTitle {
		t.Fatalf("expected metadata view title %q, actual %q", reviewModeMetadataTitle, metadataView.Title)
	}
	if !strings.Contains(metadataView.Buffer(), "Rocket PR") {
		t.Fatalf("expected metadata view to contain %q, actual %q", "Rocket PR", metadataView.Buffer())
	}
}
