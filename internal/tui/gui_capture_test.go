package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestCaptureGUI_GivenStoredAndNewGUI_WhenCapturing_ThenItReturnsTheLatestLiveGUI(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	givenFirstGUI := new(gocui.Gui)
	givenSecondGUI := new(gocui.Gui)

	actualNil := subject.captureGUI(nil)
	actualFirst := subject.captureGUI(givenFirstGUI)
	actualStored := subject.captureGUI(nil)
	actualSecond := subject.captureGUI(givenSecondGUI)
	actualStoredSecond := subject.captureGUI(nil)

	if actualNil != nil {
		t.Fatalf("expected nil before any GUI capture, actual %v", actualNil)
	}
	if actualFirst != givenFirstGUI {
		t.Fatalf("expected the first captured GUI %p, actual %p", givenFirstGUI, actualFirst)
	}
	if actualStored != givenFirstGUI {
		t.Fatalf("expected the stored first GUI %p, actual %p", givenFirstGUI, actualStored)
	}
	if actualSecond != givenSecondGUI {
		t.Fatalf("expected the second captured GUI %p, actual %p", givenSecondGUI, actualSecond)
	}
	if actualStoredSecond != givenSecondGUI {
		t.Fatalf("expected the stored second GUI %p, actual %p", givenSecondGUI, actualStoredSecond)
	}
}
