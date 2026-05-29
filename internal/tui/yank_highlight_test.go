package tui

import (
	"testing"
	"time"
)

func TestDetailYankHighlight_GivenMotionYank_WhenRendering_ThenItUsesSearchHighlightBackground(t *testing.T) {
	now := time.Date(2026, time.May, 19, 9, 0, 0, 0, time.UTC)
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	subject.timingState.yankHighlightDuration = time.Second
	subject.timingState.now = func() time.Time { return now }
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_detailCursorOnSegment(t, gui, subject, "Alpha")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	when_pressingBoundKeys(t, gui, subject, detailView, viewDetailName,
		yankMotionTestKeyPress{key: 'y'},
		yankMotionTestKeyPress{key: 'e'},
	)

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	lineIndex := given_viewLineIndexContaining(t, detailView, "Alpha Beta")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewDetailName, lineIndex, "Alpha")
}

func TestDetailYankHighlight_GivenVisualYank_WhenTheHighlightExpires_ThenTheLoadingSpinnerTickClearsItFromTheDetailView(t *testing.T) {
	now := time.Date(2026, time.May, 19, 9, 0, 0, 0, time.UTC)
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	subject.timingState.yankHighlightDuration = 200 * time.Millisecond
	subject.timingState.now = func() time.Time { return now }
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_detailCursorOnSegment(t, gui, subject, "Alpha")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_noError(t, subject.enterDetailVisualMode(gui, detailView))
	for range 4 {
		then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	}
	then_noError(t, subject.copyPullRequestURL(gui, detailView))

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	lineIndex := given_viewLineIndexContaining(t, detailView, "Alpha Beta")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewDetailName, lineIndex, "Alpha")

	now = now.Add(subject.timingState.yankHighlightDuration + time.Millisecond)
	Update(subject, MsgLoadingSpinnerTick{})
	then_noError(t, subject.refreshViews(gui))

	then_viewLineRuneDoesNotHaveBackgroundColor(t, gui, viewDetailName, lineIndex, 0, given_searchHighlightColorHex(t), "expired yank highlight")
}

func TestPullRequestBuildRunPopup_GivenVisualYank_WhenRendering_ThenItUsesSearchHighlightBackground(t *testing.T) {
	now := time.Date(2026, time.May, 19, 9, 0, 0, 0, time.UTC)
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	subject.clipboardWriter = clipboardWriter
	subject.timingState.yankHighlightDuration = time.Second
	subject.timingState.now = func() time.Time { return now }
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

	then_noError(t, subject.enterPullRequestBuildRunPopupVisualMode(gui, popupView))
	for range 4 {
		then_noError(t, subject.movePullRequestBuildRunPopupCursorRight(gui, popupView))
	}
	then_noError(t, subject.copyPullRequestBuildRunPopupContent(gui, popupView))

	popupView, actualErr = gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	lineIndex := given_viewLineIndexContaining(t, popupView, "Alpha Beta")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewPullRequestBuildInfoName, lineIndex, "Alpha")
}

func TestUpdate_GivenMsgClipboardWriteFinishedWithDetailSelection_WhenApplying_ThenItActivatesTheDetailYankHighlightThroughChildStateTransitions(t *testing.T) {
	now := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.UTC)
	subject := NewProgramWithModel(given_model())
	subject.timingState.yankHighlightDuration = time.Second
	subject.timingState.now = func() time.Time { return now }
	selection := detailSelectionRange{start: detailPosition{line: 0, column: 0}, end: detailPosition{line: 0, column: 4}}

	actual := Update(subject, MsgClipboardWriteFinished{
		Target:          FocusDetailView,
		SuccessMessage:  detailYankSuccessMessage,
		Selection:       selection,
		SelectionTarget: clipboardWriteSelectionDetail,
	})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.detailState.hasYankHighlight() {
		t.Fatal("expected the detail yank highlight to become active")
	}
	if actual := subject.detailState.viewState.yankHighlight.start; actual != selection.start {
		t.Fatalf("expected detail highlight start %+v, actual %+v", selection.start, actual)
	}
	if actual := subject.detailState.viewState.yankHighlight.end; actual != selection.end {
		t.Fatalf("expected detail highlight end %+v, actual %+v", selection.end, actual)
	}
	if actual := subject.detailState.viewState.yankHighlight.expiresAt; !actual.Equal(now.Add(time.Second)) {
		t.Fatalf("expected detail highlight expiry %v, actual %v", now.Add(time.Second), actual)
	}
}

func TestUpdate_GivenMsgClipboardWriteFinishedWithBuildPopupSelection_WhenApplying_ThenItActivatesThePopupYankHighlightThroughChildStateTransitions(t *testing.T) {
	now := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.UTC)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.timingState.yankHighlightDuration = time.Second
	subject.timingState.now = func() time.Time { return now }
	subject.openPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})
	selection := detailSelectionRange{start: detailPosition{line: 0, column: 0}, end: detailPosition{line: 0, column: 4}}

	actual := Update(subject, MsgClipboardWriteFinished{
		Target:          FocusDetailView,
		SuccessMessage:  detailYankSuccessMessage,
		Selection:       selection,
		SelectionTarget: clipboardWriteSelectionBuildPopup,
	})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the build popup to stay visible")
	}
	if !subject.pullRequestBuildRunPopup.hasYankHighlight() {
		t.Fatal("expected the build-popup yank highlight to become active")
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.yankHighlight.start; actual != selection.start {
		t.Fatalf("expected popup highlight start %+v, actual %+v", selection.start, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.yankHighlight.end; actual != selection.end {
		t.Fatalf("expected popup highlight end %+v, actual %+v", selection.end, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.yankHighlight.expiresAt; !actual.Equal(now.Add(time.Second)) {
		t.Fatalf("expected popup highlight expiry %v, actual %v", now.Add(time.Second), actual)
	}
}

func TestProgram_GivenExpiredDetailAndPopupYankHighlights_WhenClearing_ThenItClearsThemThroughChildStateTransitions(t *testing.T) {
	now := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.UTC)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.timingState.now = func() time.Time { return now }
	selection := detailSelectionRange{start: detailPosition{line: 0, column: 0}, end: detailPosition{line: 0, column: 4}}

	subject.detailState = subject.detailState.withYankHighlightActivated(selection, now.Add(-time.Millisecond))
	subject.openPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})
	subject.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withYankHighlightActivated(selection, now.Add(-time.Millisecond))
	})

	actual := subject.clearExpiredYankHighlights()

	if !actual {
		t.Fatal("expected expired yank highlight cleanup to report a change")
	}
	if subject.detailState.hasYankHighlight() {
		t.Fatalf("expected detail yank highlight to be cleared, actual %+v", subject.detailState.viewState.yankHighlight)
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the build popup to stay visible")
	}
	if subject.pullRequestBuildRunPopup.hasYankHighlight() {
		t.Fatalf("expected popup yank highlight to be cleared, actual %+v", subject.pullRequestBuildRunPopup.viewState.yankHighlight)
	}
}
