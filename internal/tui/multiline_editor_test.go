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
