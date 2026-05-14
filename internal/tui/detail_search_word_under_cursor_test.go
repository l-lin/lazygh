package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestSearchPrompt_GivenBrowserDetailFocusAndWordUnderCursor_WhenOpeningWithStar_ThenItPrefillsTheSearchCriteria(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for range 3 {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}

	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '*')(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewSearchName)

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/Alpha" {
		t.Fatalf("expected search buffer %q, actual %q", "/Alpha", actual)
	}
	if actual := subject.model.SearchDraft(); actual != "Alpha" {
		t.Fatalf("expected search draft %q, actual %q", "Alpha", actual)
	}
	if actual := subject.model.SearchTarget(); actual != FocusDetailView {
		t.Fatalf("expected search target %v, actual %v", FocusDetailView, actual)
	}
}

func TestSearchPrompt_GivenReviewDetailFocusAndWordUnderCursor_WhenOpeningWithPound_ThenItPrefillsTheSearchCriteria(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
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
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_detailCursorOnSegment(t, gui, subject, "new")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '#')(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewSearchName)

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/new" {
		t.Fatalf("expected search buffer %q, actual %q", "/new", actual)
	}
	if actual := subject.model.SearchDraft(); actual != "new" {
		t.Fatalf("expected search draft %q, actual %q", "new", actual)
	}
	if actual := subject.model.SearchTarget(); actual != FocusDetailView {
		t.Fatalf("expected search target %v, actual %v", FocusDetailView, actual)
	}
}

func given_detailCursorOnSegment(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.syncDetailViewState(document, detailView.InnerHeight())
	lineIndex, line := given_detailDocumentLineContaining(t, document, segment)
	column := given_runeIndexInString(t, line, segment)
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: column}
	subject.detailViewState.preferredColumn = document.screenColumnForPosition(subject.detailViewState.cursor)
	subject.detailViewState.sync(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)
}
