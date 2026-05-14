package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestLineEditor_GivenCursorInMiddle_WhenInsertingText_ThenTextIsInsertedAtTheCursor(t *testing.T) {
	subject := newLineEditor("abef")
	subject.MoveCursorLeft()
	subject.MoveCursorLeft()

	actual := subject.HandleKey(0, 'c', gocui.ModNone)
	if !actual {
		t.Fatal("expected key handling to report success")
	}
	actual = subject.HandleKey(0, 'd', gocui.ModNone)
	if !actual {
		t.Fatal("expected key handling to report success")
	}

	expectedText := "abcdef"
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != 4 {
		t.Fatalf("expected cursor 4, actual %d", subject.Cursor())
	}
}

func TestLineEditor_GivenCursorMovementKeys_WhenMovingAround_ThenTheCursorStaysWithinBounds(t *testing.T) {
	subject := newLineEditor("search")

	subject.HandleKey(gocui.KeyCtrlA, 0, gocui.ModNone)
	if subject.Cursor() != 0 {
		t.Fatalf("expected cursor 0 after ctrl-a, actual %d", subject.Cursor())
	}

	subject.HandleKey(gocui.KeyCtrlB, 0, gocui.ModNone)
	if subject.Cursor() != 0 {
		t.Fatalf("expected cursor to stay at 0 after ctrl-b, actual %d", subject.Cursor())
	}

	subject.HandleKey(gocui.KeyArrowRight, 0, gocui.ModNone)
	subject.HandleKey(gocui.KeyCtrlF, 0, gocui.ModNone)
	if subject.Cursor() != 2 {
		t.Fatalf("expected cursor 2 after right and ctrl-f, actual %d", subject.Cursor())
	}

	subject.HandleKey(gocui.KeyCtrlE, 0, gocui.ModNone)
	if subject.Cursor() != len([]rune("search")) {
		t.Fatalf("expected cursor at end, actual %d", subject.Cursor())
	}

	subject.HandleKey(gocui.KeyArrowRight, 0, gocui.ModNone)
	if subject.Cursor() != len([]rune("search")) {
		t.Fatalf("expected cursor to stay at end, actual %d", subject.Cursor())
	}
}

func TestLineEditor_GivenWordsAndSpaces_WhenDeletingPreviousWord_ThenItDeletesTheWordToTheLeft(t *testing.T) {
	subject := newLineEditor("alpha beta gamma")

	actual := subject.HandleKey(gocui.KeyCtrlW, 0, gocui.ModNone)
	if !actual {
		t.Fatal("expected ctrl-w to be handled")
	}

	expectedText := "alpha beta "
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != len([]rune(expectedText)) {
		t.Fatalf("expected cursor at %d, actual %d", len([]rune(expectedText)), subject.Cursor())
	}
}

func TestLineEditor_GivenCursorAfterCharacter_WhenDeletingBackwardCharacter_ThenItDeletesTheCharacterToTheLeft(t *testing.T) {
	subject := newLineEditor("abcd")
	subject.MoveCursorLeft()
	subject.MoveCursorLeft()

	actual := subject.HandleKey(gocui.KeyCtrlH, 0, gocui.ModNone)
	if !actual {
		t.Fatal("expected ctrl-h to be handled")
	}

	expectedText := "acd"
	if subject.Text() != expectedText {
		t.Fatalf("expected text %q, actual %q", expectedText, subject.Text())
	}
	if subject.Cursor() != 1 {
		t.Fatalf("expected cursor 1, actual %d", subject.Cursor())
	}
}

func TestLineEditor_GivenCursorBeforeCharacter_WhenDeletingForwardCharacter_ThenItDeletesTheCharacterToTheRight(t *testing.T) {
	subject := newLineEditor("abcd")
	subject.MoveCursorToStart()
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

func TestLineEditor_GivenCursorAtStartOfLastWord_WhenDeletingForwardWord_ThenItDeletesTheWordToTheRight(t *testing.T) {
	subject := newLineEditor("alpha beta gamma")
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

func TestLineEditor_GivenAltWordMovement_WhenMovingByWord_ThenTheCursorJumpsAcrossWords(t *testing.T) {
	subject := newLineEditor("alpha beta gamma")

	actual := subject.HandleKey(0, 'b', gocui.ModAlt)
	if !actual {
		t.Fatal("expected alt-b to be handled")
	}
	if subject.Cursor() != len([]rune("alpha beta ")) {
		t.Fatalf("expected cursor at %d, actual %d", len([]rune("alpha beta ")), subject.Cursor())
	}

	actual = subject.HandleKey(0, 'f', gocui.ModAlt)
	if !actual {
		t.Fatal("expected alt-f to be handled")
	}
	if subject.Cursor() != len([]rune("alpha beta gamma")) {
		t.Fatalf("expected cursor at end, actual %d", subject.Cursor())
	}
}
