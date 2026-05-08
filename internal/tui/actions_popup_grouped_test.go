package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
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

func TestActionsPopup_GivenSearchMatchingOnlyTheGroupName_WhenFiltering_ThenItShowsTheMatchingGroupHeaderAndActions(t *testing.T) {
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
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"Navigation",
		actionsPopupLabel(actionsPopupOpenLinkIcon, "Open link under cursor"),
	})
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
