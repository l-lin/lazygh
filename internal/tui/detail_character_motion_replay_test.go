package tui

import (
	"errors"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jesseduffield/gocui"
)

func TestLayout_GivenDetailFocus_WhenRendering_ThenTheDetailPaneStaysReadOnly(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Editable {
		t.Fatal("expected the detail pane to stay read-only")
	}
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenRendering_ThenItStaysReadOnly(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Line 1"}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if popupView.Editable {
		t.Fatal("expected the build popup to stay read-only")
	}
}

func TestDetailCharacterMotion_GivenDetailFocus_WhenReplayingForwardFindSpace_ThenItMovesToTheNextSpace(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "foo bar baz"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_replayHeadlessGui(t)
	defer gui.Close()

	then_noError(t, subject.start(gui))
	then_noError(t, subject.focusDetailView(gui, nil))
	given_detailCursorOnSegment(t, gui, subject, "foo")
	expected := given_detailPositionOfSegmentOccurrence(t, gui, subject, "foo ", 0)
	then_currentViewNameIs(t, gui, viewDetailName)

	errCh, idleCh := given_runningReplayGui(t, gui)
	defer then_replayMainLoopStops(t, gui, errCh)

	when_replayingKeyRunes(t, gui, idleCh, 'f', ' ')

	then_detailCursorIs(t, subject.detailState.viewState, detailPosition{line: expected.line, column: expected.column + 3})
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenReplayingForwardFindSpace_ThenItMovesThePopupCursorToTheNextSpace(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_replayHeadlessGui(t)
	defer gui.Close()

	then_noError(t, subject.start(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Build log line"}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 0, column: 0}
	subject.pullRequestBuildRunPopup.viewState.preferredColumn = 0
	subject.syncPullRequestBuildRunPopupViewState(document, popupView.InnerHeight())
	then_noError(t, subject.refreshViews(gui))
	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)

	errCh, idleCh := given_runningReplayGui(t, gui)
	defer then_replayMainLoopStops(t, gui, errCh)

	when_replayingKeyRunes(t, gui, idleCh, 'f', ' ')

	expected := detailPosition{line: 0, column: 5}
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != expected {
		t.Fatalf("expected popup cursor %+v, actual %+v", expected, actual)
	}
}

func given_replayHeadlessGui(t *testing.T) *gocui.Gui {
	t.Helper()

	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode:    gocui.OutputTrue,
		PlayRecording: true,
		Headless:      true,
		Width:         120,
		Height:        30,
	})
	then_noError(t, actualErr)
	return gui
}

func given_runningReplayGui(t *testing.T, gui *gocui.Gui) (<-chan error, chan struct{}) {
	t.Helper()

	idleCh := make(chan struct{}, 10)
	gui.AddIdleListener(idleCh)
	errCh := make(chan error, 1)
	go func() {
		errCh <- gui.MainLoop()
	}()
	return errCh, idleCh
}

func when_replayingKeyRunes(t *testing.T, gui *gocui.Gui, idleCh chan struct{}, keyRunes ...rune) {
	t.Helper()

	for _, keyRune := range keyRunes {
		drainReplayIdleSignals(idleCh)
		when_replayingKeyEvent(t, gui, idleCh, tcell.NewEventKey(tcell.KeyRune, keyRune, tcell.ModNone))
	}
}

func when_replayingKeyEvent(t *testing.T, gui *gocui.Gui, idleCh chan struct{}, event *tcell.EventKey) {
	t.Helper()

	sentCh := make(chan struct{})
	go func() {
		gui.ReplayedEvents.Keys <- gocui.NewTcellKeyEventWrapper(event, time.Now().UnixNano())
		close(sentCh)
	}()

	select {
	case <-sentCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected replayed key event to be accepted")
	}

	select {
	case <-idleCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected replayed key event to finish processing")
	}
}

func drainReplayIdleSignals(idleCh chan struct{}) {
	for {
		select {
		case <-idleCh:
		default:
			return
		}
	}
}

func then_replayMainLoopStops(t *testing.T, gui *gocui.Gui, errCh <-chan error) {
	t.Helper()

	select {
	case actualErr := <-errCh:
		if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
			t.Fatalf("expected no error, actual %v", actualErr)
		}
		return
	default:
	}

	gui.Update(func(*gocui.Gui) error {
		return gocui.ErrQuit
	})

	select {
	case actualErr := <-errCh:
		if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
			t.Fatalf("expected no error, actual %v", actualErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected replay gui main loop to stop")
	}
}
