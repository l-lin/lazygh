package tui

import "testing"

func TestDetailViewState_GivenWrappedDetailText_WhenMovingWithHJKL_ThenItMovesAcrossRenderedRowsAndClampsTheViewport(t *testing.T) {
	document := newDetailDocument("abcd1234\nefgh\nijkl", 4)
	subject := newDetailViewState()

	subject.sync(document, 2)
	subject.moveDown(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 4})
	then_detailOriginRowIs(t, subject, 0)

	subject.moveDown(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 0})
	then_detailOriginRowIs(t, subject, 1)

	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 3})
	then_detailOriginRowIs(t, subject, 1)

	subject.moveDown(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 2, column: 3})
	then_detailOriginRowIs(t, subject, 2)

	subject.moveUp(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 3})
	then_detailOriginRowIs(t, subject, 2)

	subject.moveUp(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 7})
	then_detailOriginRowIs(t, subject, 1)

	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 3})
	then_detailOriginRowIs(t, subject, 0)
}

func TestDetailViewState_GivenRenderedDetailText_WhenUsingVimMotions_ThenItNavigatesRowsAndWords(t *testing.T) {
	document := newDetailDocument("abc def\nghi jkl", 4)
	subject := newDetailViewState()

	subject.moveToWordEnd(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 2})

	subject.moveToWordEnd(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 6})

	subject.moveToNextWord(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 0})

	subject.moveToPreviousWord(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 4})

	subject.moveDown(document, 3)
	subject.moveDown(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 4})

	subject.moveToRowEnd(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 6})

	subject.moveToRowStart(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 4})

	subject.handleGoToTopPrefix(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 4})

	subject.handleGoToTopPrefix(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 0})

	subject.moveToBottom(document, 3)
	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 4})
}

func TestDetailViewState_GivenVisualMode_WhenMovingTheCursor_ThenTheAnchorStaysFixedWhileTheSelectionGrowsAndShrinks(t *testing.T) {
	document := newDetailDocument("abcdef", 4)
	subject := newDetailViewState()

	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	subject.enterVisualMode()

	if subject.mode != detailVisualMode {
		t.Fatalf("expected mode %v, actual %v", detailVisualMode, subject.mode)
	}
	then_detailSelectionIs(t, document, subject, detailPosition{line: 0, column: 2}, detailPosition{line: 0, column: 2})

	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	then_detailSelectionIs(t, document, subject, detailPosition{line: 0, column: 2}, detailPosition{line: 0, column: 5})

	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	subject.moveLeft(document, 2)
	then_detailSelectionIs(t, document, subject, detailPosition{line: 0, column: 1}, detailPosition{line: 0, column: 2})
}

func TestDetailViewState_GivenLinewiseVisualMode_WhenMovingTheCursor_ThenItSelectsEntireRenderedRows(t *testing.T) {
	document := newDetailDocument("abcd1234\nefgh", 4)
	subject := newDetailViewState()

	subject.moveDown(document, 2)
	subject.enterLineVisualMode(document)
	if subject.mode != detailLineVisualMode {
		t.Fatalf("expected mode %v, actual %v", detailLineVisualMode, subject.mode)
	}
	if actual := subject.selectedText(document); actual != "1234" {
		t.Fatalf("expected selected text %q, actual %q", "1234", actual)
	}

	subject.moveDown(document, 2)
	then_detailSelectedRowsAre(t, document, subject, 1, 2)
	if actual := subject.selectedText(document); actual != "1234\nefgh" {
		t.Fatalf("expected selected text %q, actual %q", "1234\nefgh", actual)
	}
}

func TestLayout_GivenDetailFocus_WhenRendering_ThenItShowsTheCursorInTheDetailPane(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	if !gui.Cursor {
		t.Fatal("expected the terminal cursor to be visible while the detail pane is focused")
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualCursorX, actualCursorY := detailView.Cursor()
	if actualCursorX != 0 || actualCursorY != 0 {
		t.Fatalf("expected detail cursor 0,0, actual %d,%d", actualCursorX, actualCursorY)
	}
}

func TestLayout_GivenDetailVisualSelectionOverASearchMatch_WhenRendering_ThenTheSelectionBackgroundOverridesTheSearchHighlight(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	model.StartSearch()
	model.UpdateSearchDraft("Alpha")
	model.SubmitSearch()
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
	actualErr = subject.enterDetailVisualMode(gui, detailView)
	then_noError(t, actualErr)
	for range 4 {
		actualErr = subject.moveDetailCursorRight(gui, detailView)
		then_noError(t, actualErr)
	}

	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewDetailName, 3, "Alpha")
}

func TestCopyPullRequestURL_GivenDetailLineVisualMode_WhenYankingSelectedText_ThenItUsesTheClipboardAndReturnsToNormalMode(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha\nBeta\nGamma"}}})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	for range 4 {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	actualErr = subject.enterDetailLineVisualMode(gui, detailView)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, detailView)
	then_noError(t, actualErr)

	actualErr = subject.copyPullRequestURL(gui, detailView)
	then_noError(t, actualErr)

	if actual := clipboardWriter.writes; len(actual) != 1 || actual[0] != "Beta\nGamma" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"Beta\nGamma"}, actual)
	}
	if subject.detailViewState.mode != detailNormalMode {
		t.Fatalf("expected mode %v after yanking, actual %v", detailNormalMode, subject.detailViewState.mode)
	}

	then_statusLineContains(t, gui, detailYankSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestCopyPullRequestURL_GivenDetailVisualMode_WhenYankingSelectedText_ThenItUsesTheClipboardAndReturnsToNormalMode(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
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
	actualErr = subject.enterDetailVisualMode(gui, detailView)
	then_noError(t, actualErr)
	for range 4 {
		actualErr = subject.moveDetailCursorRight(gui, detailView)
		then_noError(t, actualErr)
	}

	actualErr = subject.copyPullRequestURL(gui, detailView)
	then_noError(t, actualErr)

	if actual := clipboardWriter.writes; len(actual) != 1 || actual[0] != "Alpha" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"Alpha"}, actual)
	}
	if subject.detailViewState.mode != detailNormalMode {
		t.Fatalf("expected mode %v after yanking, actual %v", detailNormalMode, subject.detailViewState.mode)
	}

	then_statusLineContains(t, gui, detailYankSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestCopyPullRequestURL_GivenDetailVisualModeAndClipboardFailure_WhenYankingSelectedText_ThenItShowsFailureFeedbackAndReturnsToNormalMode(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{writeErr: ErrClipboardUnavailable}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
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
	actualErr = subject.enterDetailVisualMode(gui, detailView)
	then_noError(t, actualErr)
	for range 4 {
		actualErr = subject.moveDetailCursorRight(gui, detailView)
		then_noError(t, actualErr)
	}

	actualErr = subject.copyPullRequestURL(gui, detailView)
	then_noError(t, actualErr)

	if subject.detailViewState.mode != detailNormalMode {
		t.Fatalf("expected mode %v after a failed yank, actual %v", detailNormalMode, subject.detailViewState.mode)
	}

	then_statusLineContains(t, gui, detailYankFailureMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestCloseDetail_GivenDetailVisualMode_WhenHandlingEscape_ThenItLeavesVisualModeBeforeClosingThePane(t *testing.T) {
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

	actualErr = subject.enterDetailVisualMode(gui, detailView)
	then_noError(t, actualErr)
	actualErr = subject.closeDetail(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	if subject.detailViewState.mode != detailNormalMode {
		t.Fatalf("expected mode %v after the first escape, actual %v", detailNormalMode, subject.detailViewState.mode)
	}

	actualErr = subject.closeDetail(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
}

func then_detailCursorIs(t *testing.T, subject detailViewState, expected detailPosition) {
	t.Helper()

	if subject.cursor != expected {
		t.Fatalf("expected detail cursor %+v, actual %+v", expected, subject.cursor)
	}
}

func then_detailOriginRowIs(t *testing.T, subject detailViewState, expected int) {
	t.Helper()

	if subject.originRow != expected {
		t.Fatalf("expected detail origin row %d, actual %d", expected, subject.originRow)
	}
}

func then_detailSelectionIs(t *testing.T, document detailDocument, subject detailViewState, expectedStart detailPosition, expectedEnd detailPosition) {
	t.Helper()

	actualStart, actualEnd, actualOk := subject.visualSelection(document)
	if !actualOk {
		t.Fatal("expected a visual selection")
	}
	if actualStart != expectedStart || actualEnd != expectedEnd {
		t.Fatalf("expected detail selection %+v..%+v, actual %+v..%+v", expectedStart, expectedEnd, actualStart, actualEnd)
	}
}

func then_detailSelectedRowsAre(t *testing.T, document detailDocument, subject detailViewState, expectedStart int, expectedEnd int) {
	t.Helper()

	actualStart, actualEnd, actualOk := subject.visualRowSelection(document)
	if !actualOk {
		t.Fatal("expected a linewise visual selection")
	}
	if actualStart != expectedStart || actualEnd != expectedEnd {
		t.Fatalf("expected selected rows %d..%d, actual %d..%d", expectedStart, expectedEnd, actualStart, actualEnd)
	}
}
