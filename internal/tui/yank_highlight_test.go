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

func TestDetailYankHighlight_GivenVisualYank_WhenTheHighlightExpires_ThenItClearsFromTheDetailView(t *testing.T) {
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
	subject.clearExpiredYankHighlights()
	then_noError(t, subject.afterStateChange(gui))

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
