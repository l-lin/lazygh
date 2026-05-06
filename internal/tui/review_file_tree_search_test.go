package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingReviewFileTreeSearchFollowBindings_ThenPullRequestsViewSupportsNAndN(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'n', handler: subject.nextReviewFileTreeSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'N', handler: subject.previousReviewFileTreeSearchMatch})
}

func TestReviewMode_GivenSubmittedFileTreeSearch_WhenRendering_ThenItKeepsTheTreeVisibleHighlightsMatchesAndMovesToTheFirstMatchingFile(t *testing.T) {
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
				ChangedFiles: 2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
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

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	when_typingSearchQuery(t, subject, searchView, "model")

	actualErr = subject.submitSearch(gui, searchView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), "render.go") || !strings.Contains(filesView.Buffer(), "model.go") {
		t.Fatalf("expected the review tree to stay visible after search, actual %q", filesView.Buffer())
	}
	modelLineIndex := given_viewLineIndexContaining(t, filesView, "model.go")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewPullRequestsName, modelLineIndex, "model")
	if subject.reviewSession.selectedFileTreeRow != modelLineIndex {
		t.Fatalf("expected selected review tree row %d, actual %d", modelLineIndex, subject.reviewSession.selectedFileTreeRow)
	}

	footerView, actualErr := gui.View(viewPullRequestsFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != "/model (1 match)  •  ? Help  / Search  a Actions" {
		t.Fatalf("expected review tree footer %q, actual %q", "/model (1 match)  •  ? Help  / Search  a Actions", actual)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "internal/tui/model.go") || !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected detail view to switch to the matching file diff, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenSubmittedFileTreeSearch_WhenRepeatingWithNAndN_ThenItMovesBetweenMatchingFilesWithWraparound(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#77": {
				Title:        "Colorful Tree PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/review-tree-colors",
				State:        "OPEN",
				ChangedFiles: 3,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#77": given_reviewSessionColoredTreePullRequestDiff(),
		},
	}
	model := given_pullRequestCommentModel()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Colorful Tree PR", Number: 77, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/77"}),
	})
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	when_typingSearchQuery(t, subject, searchView, "ed")
	actualErr = subject.submitSearch(gui, searchView)
	then_noError(t, actualErr)

	then_reviewTreeSelectionAndDetailAre(t, gui, subject, "changed.go", "+new changed content")

	actualErr = subject.nextReviewFileTreeSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_reviewTreeSelectionAndDetailAre(t, gui, subject, "added.go", "+added content")

	actualErr = subject.nextReviewFileTreeSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_reviewTreeSelectionAndDetailAre(t, gui, subject, "deleted.go", "-deleted content")

	actualErr = subject.nextReviewFileTreeSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_reviewTreeSelectionAndDetailAre(t, gui, subject, "changed.go", "+new changed content")

	actualErr = subject.previousReviewFileTreeSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_reviewTreeSelectionAndDetailAre(t, gui, subject, "deleted.go", "-deleted content")
}

func then_reviewTreeSelectionAndDetailAre(t *testing.T, gui *gocui.Gui, subject *Program, expectedFile string, expectedDetailSnippet string) {
	t.Helper()

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	expectedRow := given_viewLineIndexContaining(t, filesView, expectedFile)
	if subject.reviewSession.selectedFileTreeRow != expectedRow {
		t.Fatalf("expected selected review tree row %d for %q, actual %d", expectedRow, expectedFile, subject.reviewSession.selectedFileTreeRow)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), expectedFile) || !strings.Contains(detailView.Buffer(), expectedDetailSnippet) {
		t.Fatalf("expected detail view to contain %q and %q, actual %q", expectedFile, expectedDetailSnippet, detailView.Buffer())
	}
}
