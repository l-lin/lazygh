package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestMultilineEditor_GivenEnterAndControlJ_WhenHandlingInput_ThenTheyInsertNewlines(t *testing.T) {
	subject := newMultilineEditor("first line")

	actual := subject.HandleKey(gocui.KeyEnter, 0, gocui.ModNone)
	if !actual {
		t.Fatal("expected enter to be handled")
	}
	actual = subject.HandleKey(gocui.KeyCtrlJ, 0, gocui.ModNone)
	if !actual {
		t.Fatal("expected ctrl-j to be handled")
	}

	expectedText := "first line\n\n"
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != len([]rune(expectedText)) {
		t.Fatalf("expected cursor %d, actual %d", len([]rune(expectedText)), subject.Cursor())
	}
}

func TestMultilineEditor_GivenAltC_WhenHandlingInput_ThenItInsertsTheFenceSnippetAndLeavesTheCursorAfterTheOpeningFence(t *testing.T) {
	subject := newMultilineEditor("")

	actual := subject.HandleKey(0, 'c', gocui.ModAlt)
	if !actual {
		t.Fatal("expected alt-c to be handled")
	}

	expectedText := "```\n```"
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != len([]rune("```")) {
		t.Fatalf("expected cursor %d, actual %d", len([]rune("```")), subject.Cursor())
	}
}

func TestMultilineEditor_GivenCursorBeforeCharacter_WhenDeletingForwardCharacter_ThenItDeletesTheCharacterToTheRight(t *testing.T) {
	subject := newMultilineEditor("abcd")
	subject.MoveCursorToLineStart()
	subject.MoveCursorRight()

	actual := subject.HandleKey(gocui.KeyCtrlD, 0, gocui.ModNone)
	if !actual {
		t.Fatal("expected ctrl-d to be handled")
	}

	expectedText := "acd"
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != 1 {
		t.Fatalf("expected cursor 1, actual %d", subject.Cursor())
	}
}

func TestMultilineEditor_GivenCursorAtStartOfLastWord_WhenDeletingForwardWord_ThenItDeletesTheWordToTheRight(t *testing.T) {
	subject := newMultilineEditor("alpha beta gamma")
	subject.cursor = len([]rune("alpha beta "))

	actual := subject.HandleKey(0, 'd', gocui.ModAlt)
	if !actual {
		t.Fatal("expected alt-d to be handled")
	}

	expectedText := "alpha beta "
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != len([]rune(expectedText)) {
		t.Fatalf("expected cursor at %d, actual %d", len([]rune(expectedText)), subject.Cursor())
	}
}

func TestMultilineEditor_GivenMixedLineLengths_WhenMovingVertically_ThenItRestoresThePreferredColumnWhenPossible(t *testing.T) {
	subject := newMultilineEditor("12345\n12\n1234")
	subject.cursor = 4

	subject.MoveCursorDown()
	actualColumn, actualRow := subject.CursorXY()
	if actualColumn != 2 || actualRow != 1 {
		t.Fatalf("expected cursor 2,1 after the first move, actual %d,%d", actualColumn, actualRow)
	}

	subject.MoveCursorDown()
	actualColumn, actualRow = subject.CursorXY()
	if actualColumn != 4 || actualRow != 2 {
		t.Fatalf("expected cursor 4,2 after the second move, actual %d,%d", actualColumn, actualRow)
	}
}

func TestMultilineEditor_GivenMultipleLines_WhenDeletingToLineBoundaries_ThenItOnlyRemovesTheCurrentLineSegment(t *testing.T) {
	subject := newMultilineEditor("alpha beta\ngamma delta")
	subject.cursor = len([]rune("alpha beta\ngam"))

	subject.DeleteToLineStart()
	if actual := subject.Text(); actual != "alpha beta\nma delta" {
		t.Fatalf("expected text %q after deleting to line start, actual %q", "alpha beta\nma delta", actual)
	}
	if actual := subject.Cursor(); actual != len([]rune("alpha beta\n")) {
		t.Fatalf("expected cursor %d after deleting to line start, actual %d", len([]rune("alpha beta\n")), actual)
	}

	subject.DeleteToLineEnd()
	if actual := subject.Text(); actual != "alpha beta\n" {
		t.Fatalf("expected text %q after deleting to line end, actual %q", "alpha beta\n", actual)
	}
}
