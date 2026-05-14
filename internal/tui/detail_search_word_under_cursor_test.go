package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestDetailSearch_GivenBrowserDetailFocusAndWordUnderCursor_WhenPressingStar_ThenItAppliesTheSearchWithoutOpeningThePromptAndNMovesToTheNextOccurrence(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	given_detailCursorOnSegment(t, gui, subject, "Alpha")
	expectedCurrent := given_detailPositionOfSegmentOccurrence(t, gui, subject, "Alpha", 0)
	expectedNext := given_detailPositionOfSegmentOccurrence(t, gui, subject, "Alpha", 1)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '*')(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewSearchName)
	then_detailCursorIs(t, subject.detailViewState, expectedCurrent)
	if actual := subject.model.DetailSearchQuery(); actual != "Alpha" {
		t.Fatalf("expected applied detail search query %q, actual %q", "Alpha", actual)
	}
	if subject.model.SearchActive() {
		t.Fatal("expected direct word search to avoid leaving the search prompt active")
	}

	actualErr = subject.nextDetailSearchMatch(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, expectedNext)
}

func TestDetailSearch_GivenReviewDetailFocusAndWordUnderCursor_WhenPressingPound_ThenItAppliesTheSearchBackwardsAndFlipsNAndNWithoutOpeningThePrompt(t *testing.T) {
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
	given_detailCursorOnSegmentOccurrence(t, gui, subject, "line", 2)
	expectedPrevious := given_detailPositionOfSegmentOccurrence(t, gui, subject, "line", 1)
	expectedForward := expectedPrevious

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '#')(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewSearchName)
	if actual := subject.model.DetailSearchQuery(); actual != "line" {
		t.Fatalf("expected applied detail search query %q, actual %q", "line", actual)
	}
	if subject.model.SearchActive() {
		t.Fatal("expected reverse word search to avoid leaving the search prompt active")
	}
	then_detailCursorIs(t, subject.detailViewState, expectedPrevious)

	actualErr = subject.nextDetailSearchMatch(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, given_detailPositionOfSegmentOccurrence(t, gui, subject, "line", 0))

	actualErr = subject.previousDetailSearchMatch(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, expectedForward)
}

func given_detailCursorOnSegment(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()
	given_detailCursorOnSegmentOccurrence(t, gui, subject, segment, 0)
}

func given_detailCursorOnSegmentOccurrence(t *testing.T, gui *gocui.Gui, subject *Program, segment string, expectedOccurrence int) {
	t.Helper()

	target := given_detailPositionOfSegmentOccurrence(t, gui, subject, segment, expectedOccurrence)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.detailViewState.cursor = target
	subject.detailViewState.preferredColumn = document.screenColumnForPosition(subject.detailViewState.cursor)
	subject.detailViewState.sync(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)
}

func given_detailPositionOfSegmentOccurrence(t *testing.T, gui *gocui.Gui, subject *Program, segment string, expectedOccurrence int) detailPosition {
	t.Helper()

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.syncDetailViewState(document, detailView.InnerHeight())

	occurrenceCount := 0
	segmentRunes := []rune(segment)
	for lineIndex, lineRunes := range document.lines {
		segmentStart := 0
		for {
			relativeColumn := given_optionalRuneIndexInRunes(lineRunes[segmentStart:], segmentRunes)
			if relativeColumn < 0 {
				break
			}
			column := segmentStart + relativeColumn
			if occurrenceCount == expectedOccurrence {
				return detailPosition{line: lineIndex, column: column}
			}
			occurrenceCount++
			segmentStart = column + len(segmentRunes)
		}
	}

	t.Fatalf("expected occurrence %d for segment %q", expectedOccurrence, segment)
	return detailPosition{}
}

func given_optionalRuneIndexInRunes(textRunes []rune, segmentRunes []rune) int {
	if len(segmentRunes) == 0 || len(segmentRunes) > len(textRunes) {
		return -1
	}
	for index := 0; index <= len(textRunes)-len(segmentRunes); index++ {
		matched := true
		for segmentIndex := range segmentRunes {
			if textRunes[index+segmentIndex] != segmentRunes[segmentIndex] {
				matched = false
				break
			}
		}
		if matched {
			return index
		}
	}
	return -1
}
