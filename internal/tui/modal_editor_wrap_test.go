package tui

import (
	"reflect"
	"testing"
)

func TestWrappedInputCursorPositionForText_GivenSoftWrappedBoundary_WhenCursorStartsTheNextWord_ThenItMapsToTheNextVisibleRow(t *testing.T) {
	actual := wrappedInputCursorPositionForText("alpha beta gamma", len([]rune("alpha beta ")), 10)

	expected := wrappedInputCursorPosition{column: 0, row: 1}
	if actual != expected {
		t.Fatalf("expected wrapped cursor position %+v, actual %+v", expected, actual)
	}
}

func TestWrappedInputCursorPositionForText_GivenWrappedPreviousLines_WhenCursorMovesOntoALaterLogicalLine_ThenItIncludesTheEarlierWrappedRows(t *testing.T) {
	actual := wrappedInputCursorPositionForText("alpha beta gamma\nze", len([]rune("alpha beta gamma\nze")), 10)

	expected := wrappedInputCursorPosition{column: 2, row: 2}
	if actual != expected {
		t.Fatalf("expected wrapped cursor position %+v, actual %+v", expected, actual)
	}
}

func TestModalEditor_GivenLongParagraph_WhenRendering_ThenItUsesVisibleWordWrapAndPlacesTheCursorOnTheWrappedRow(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 28, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openModalEditor(gui, "Compose", "alpha beta gamma delta zeta", nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualErr = gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	if !modalView.Wrap {
		t.Fatal("expected the multiline modal editor to enable wrapping")
	}
	actualVisibleLines := modalView.ViewBufferLines()
	expectedVisibleLines := []string{"alpha beta gamma delta", "zeta"}
	if len(actualVisibleLines) < len(expectedVisibleLines) {
		t.Fatalf("expected at least %d visible lines, actual %v", len(expectedVisibleLines), actualVisibleLines)
	}
	if !reflect.DeepEqual(actualVisibleLines[:len(expectedVisibleLines)], expectedVisibleLines) {
		t.Fatalf("expected visible wrapped lines %v, actual %v", expectedVisibleLines, actualVisibleLines)
	}

	actualCursorX, actualCursorY := modalView.Cursor()
	if actualCursorY != 1 || actualCursorX != 4 {
		t.Fatalf("expected wrapped cursor 4,1 at the end of the paragraph, actual %d,%d", actualCursorX, actualCursorY)
	}
}
