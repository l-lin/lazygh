package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestActionsPopup_GivenUserView_WhenOpening_ThenItShowsTheGlobalGroupedActionsAndTakesFocus(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.pullRequestCache = &fakePersistentPullRequestCache{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"Theme",
		actionsPopupLabel(actionsPopupChangeThemeIcon, themePickerActionTitle),
		"Cache",
		actionsPopupLabel(iconDelete, "Clear cache"),
	})
	_, actualCursorY := popupView.Cursor()
	if actualCursorY != 1 {
		t.Fatalf("expected the first selectable action to start below the header, actual cursor row %d", actualCursorY)
	}
}

func TestActionsPopup_GivenSearchMatchingOnlyTheGroupName_WhenHighlighting_ThenItKeepsThePopupVisibleAndHighlightsTheMatchingHeader(t *testing.T) {
	subject := given_pullRequestDetailProgramWithRenderedBody("Docs https://example.com/docs")
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailLink(t, subject, detailView, "https://example.com/docs")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "navigation" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{"Navigation", "Open link under cursor", "Open PR in browser"} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to keep %q visible, actual %q", expected, popupView.Buffer())
		}
	}
	matchLineIndex := given_viewLineIndexContaining(t, popupView, "Navigation")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewActionsPopupName, matchLineIndex, "Navigation")
}

func TestActionsPopup_GivenGroupedHeaders_WhenRendering_ThenItCentersTheHeaderAndUsesMarkdownHeadingBackgroundWithItsOwnForegroundColor(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.ResolvePalette(theme.Palette{
		ActionsPopupGroupForegroundHex: "#000000",
		MarkdownHeadingBackgroundHex:   "#204060",
		SelectedLineBackgroundHex:      "#802020",
	}))

	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_viewLineHasBackgroundColor(t, gui, viewActionsPopupName, 0, given_themeColorHex(t, theme.MarkdownHeadingBackgroundHex), "actions popup group header background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewActionsPopupName, 0, actionsPopupGroupPullRequest, given_themeColorHex(t, theme.ActionsPopupGroupForegroundHex), "actions popup group header foreground")
	then_viewLineSegmentIsCenteredInView(t, gui, viewActionsPopupName, 0, actionsPopupGroupPullRequest)
	if actual := strings.TrimSpace(popupView.BufferLines()[0]); actual != actionsPopupGroupPullRequest {
		t.Fatalf("expected grouped header %q, actual %q", actionsPopupGroupPullRequest, actual)
	}
	then_viewLineRuneDoesNotHaveBackgroundColor(t, gui, viewActionsPopupName, 0, 0, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "actions popup group header selected background")
}

func TestActionsPopup_GivenADarkBundledTheme_WhenRenderingGroupedHeaders_ThenItUsesThePresetHeadingForeground(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.ResolvePaletteWithPreset("catppuccin-mocha", theme.Palette{}))

	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_viewLineSegmentHasForegroundColor(t, gui, viewActionsPopupName, 0, actionsPopupGroupPullRequest, given_themeColorHex(t, theme.MarkdownHeadingHex), "actions popup group header preset foreground")
}

func TestActionsPopup_GivenNoPersistentCache_WhenOpening_ThenItHidesTheClearCacheAction(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Clear cache") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Clear cache", popupView.Buffer())
	}
}

func then_viewLineSegmentIsCenteredInView(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	x0, _, _, _, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	_, _, actualX, _ := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)

	expectedStartColumn := maxInt(0, (view.InnerWidth()-utf8.RuneCountInString(segment))/2)
	actualStartColumn := actualX - (x0 + 1)
	if actualStartColumn != expectedStartColumn {
		t.Fatalf("expected %q in %s line %d to start at centered column %d, actual %d", segment, viewName, lineIndex, expectedStartColumn, actualStartColumn)
	}
}
