package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

type yankMotionTestKeyPress struct {
	key any
	mod gocui.Modifier
}

func TestDetailYank_GivenNormalMode_WhenUsingMotionBindings_ThenItCopiesExpectedTextWithoutMovingTheCursor(t *testing.T) {
	testCases := []struct {
		name              string
		detail            string
		startSegment      string
		startOccurrence   int
		keyPresses        []yankMotionTestKeyPress
		expectedClipboard string
	}{
		{
			name:              "word end",
			detail:            "Alpha Beta",
			startSegment:      "Alpha",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'e'}},
			expectedClipboard: "Alpha",
		},
		{
			name:              "big word end",
			detail:            "Alpha-Beta Gamma",
			startSegment:      "Alpha-Beta",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'E'}},
			expectedClipboard: "Alpha-Beta",
		},
		{
			name:              "word start backward",
			detail:            "Beta",
			startSegment:      "a",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'b'}},
			expectedClipboard: "Beta",
		},
		{
			name:              "big word start backward",
			detail:            "Omega-1",
			startSegment:      "1",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'B'}},
			expectedClipboard: "Omega-1",
		},
		{
			name:              "row start",
			detail:            "Alpha Beta",
			startSegment:      "Beta",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: '0'}},
			expectedClipboard: "Alpha B",
		},
		{
			name:              "row end",
			detail:            "Alpha Beta",
			startSegment:      "Beta",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: '$'}},
			expectedClipboard: "Beta",
		},
		{
			name:              "current line",
			detail:            "Alpha\nBeta\nGamma",
			startSegment:      "Beta",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'y'}},
			expectedClipboard: "Beta",
		},
		{
			name:              "next line",
			detail:            "Alpha\nBeta\nGamma",
			startSegment:      "Beta",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'j'}},
			expectedClipboard: "Beta\nGamma",
		},
		{
			name:         "file top",
			detail:       "Alpha\nBeta\nGamma",
			startSegment: "Gamma",
			keyPresses:   []yankMotionTestKeyPress{{key: 'y'}, {key: 'g'}, {key: 'g'}},
		},
		{
			name:              "character find",
			detail:            "Alpha.Beta",
			startSegment:      "Alpha",
			keyPresses:        []yankMotionTestKeyPress{{key: 'y'}, {key: 'f'}, {key: '.'}},
			expectedClipboard: "Alpha.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: testCase.detail}}})
			model.OpenDetail()
			clipboardWriter := &fakeClipboardWriter{}
			subject := NewProgramWithModel(model)
			subject.clipboardWriter = clipboardWriter
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)

			then_noError(t, subject.layout(gui))
			given_detailCursorOnSegmentOccurrence(t, gui, subject, testCase.startSegment, testCase.startOccurrence)
			detailView, actualErr := gui.View(viewDetailName)
			then_noError(t, actualErr)
			expectedCursor := subject.detailViewState.cursor

			expectedClipboard := testCase.expectedClipboard
			if expectedClipboard == "" {
				document := subject.currentDetailDocument(detailView)
				expectedClipboard = document.rowSelectionText(0, document.rowIndexForPosition(expectedCursor))
			}

			when_pressingBoundKeys(t, gui, subject, detailView, viewDetailName, testCase.keyPresses...)

			if actual := clipboardWriter.writes; len(actual) != 1 || actual[0] != expectedClipboard {
				t.Fatalf("expected clipboard writes %v, actual %v", []string{expectedClipboard}, actual)
			}
			if subject.detailViewState.cursor != expectedCursor {
				t.Fatalf("expected cursor %+v after yanking, actual %+v", expectedCursor, subject.detailViewState.cursor)
			}
			if subject.detailViewState.mode != detailNormalMode {
				t.Fatalf("expected mode %v after yanking, actual %v", detailNormalMode, subject.detailViewState.mode)
			}
			then_statusLineContains(t, gui, detailYankSuccessMessage)
		})
	}
}

func TestCloseDetail_GivenPendingYank_WhenHandlingEscape_ThenItCancelsTheYankBeforeClosingThePane(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	when_pressingBoundKeys(t, gui, subject, detailView, viewDetailName, yankMotionTestKeyPress{key: 'y'})

	actualErr = given_handlerForBinding(t, subject.registeredKeybindingSpecs(), viewDetailName, gocui.KeyEsc)(gui, detailView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	if subject.detailViewState.hasPendingYank() {
		t.Fatal("expected pending yank to be cleared after escape")
	}

	actualErr = given_handlerForBinding(t, subject.registeredKeybindingSpecs(), viewDetailName, gocui.KeyEsc)(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenPressingYE_ThenItCopiesUntilEndOfWordAndKeepsTheCursorInPlace(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	clipboardWriter := &fakeClipboardWriter{}
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       "Alpha Beta",
	}))
	given_pullRequestBuildRunPopupCursorOnLineContaining(t, gui, subject, "Alpha Beta")
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	expectedCursor := subject.pullRequestBuildRunPopup.viewState.cursor

	when_pressingBoundKeys(t, gui, subject, popupView, viewPullRequestBuildInfoName, yankMotionTestKeyPress{key: 'y'}, yankMotionTestKeyPress{key: 'e'})

	if actual := clipboardWriter.writes; len(actual) != 1 || actual[0] != "Alpha" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"Alpha"}, actual)
	}
	if subject.pullRequestBuildRunPopup.viewState.cursor != expectedCursor {
		t.Fatalf("expected popup cursor %+v after yanking, actual %+v", expectedCursor, subject.pullRequestBuildRunPopup.viewState.cursor)
	}
	if subject.pullRequestBuildRunPopup.viewState.mode != detailNormalMode {
		t.Fatalf("expected popup mode %v after yanking, actual %v", detailNormalMode, subject.pullRequestBuildRunPopup.viewState.mode)
	}
	then_statusLineContains(t, gui, detailYankSuccessMessage)
}

func when_pressingBoundKeys(t *testing.T, gui *gocui.Gui, subject *Program, view *gocui.View, viewName string, keyPresses ...yankMotionTestKeyPress) {
	t.Helper()

	for _, keyPress := range keyPresses {
		handler := given_handlerForBindingWithModifier(t, subject.registeredKeybindingSpecs(), viewName, keyPress.key, keyPress.mod)
		then_noError(t, handler(gui, view))
	}
}
