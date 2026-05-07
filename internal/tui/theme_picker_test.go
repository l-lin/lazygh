package tui

import (
	"errors"
	"reflect"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/theme"
	"github.com/jesseduffield/gocui"
)

func TestActionsPopup_GivenChangeThemeAction_WhenExecuting_ThenItShowsTheThemePickerPresets(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("change theme", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "change theme"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if popupView.Title != themePickerTitle {
		t.Fatalf("expected popup title %q, actual %q", themePickerTitle, popupView.Title)
	}
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), given_themePickerLabels())
}

func TestChangeTheme_GivenPresetSelection_WhenSubmitting_ThenItSavesThePresetUpdatesTheThemeAndRefreshesTheTUI(t *testing.T) {
	t.Cleanup(theme.ResetPalette)

	store := &fakeThemePresetStore{}
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.themePresetStore = store
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("change theme", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "change theme"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("kanagawa dark", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "kanagawa dark"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(store.savedPresets, []string{"kanagawa-dark"}) {
		t.Fatalf("expected saved presets %v, actual %v", []string{"kanagawa-dark"}, store.savedPresets)
	}
	if theme.BackgroundHex != "#1F1F28" {
		t.Fatalf("expected theme background %q, actual %q", "#1F1F28", theme.BackgroundHex)
	}
	expectedBackground := gocui.GetColor("#1F1F28")
	if gui.BgColor != expectedBackground {
		t.Fatalf("expected gui background color %v, actual %v", expectedBackground, gui.BgColor)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.BgColor != expectedBackground {
		t.Fatalf("expected detail view background color %v, actual %v", expectedBackground, detailView.BgColor)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, "Theme changed to Kanagawa Dark")
}

func TestChangeTheme_GivenExistingPullRequestRows_WhenSubmitting_ThenItRestylesThemWithoutWaitingForADataRefresh(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.ResolvePaletteWithPreset("catppuccin-latte", theme.Palette{}))
	oldReferenceHex := theme.PullRequestReferenceHex
	oldTitleHex := theme.PullRequestTitleHex

	store := &fakeThemePresetStore{}
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.themePresetStore = store
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("change theme", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "change theme"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("gruvbox dark", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "gruvbox dark"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewLineSegmentDoesNotHaveForegroundColor(t, gui, viewPullRequestsName, 0, "acme/widgets#42", given_themeColorHex(t, oldReferenceHex), "pull request reference after theme switch")
	then_viewLineSegmentDoesNotHaveForegroundColor(t, gui, viewPullRequestsName, 0, "First PR", given_themeColorHex(t, oldTitleHex), "pull request title after theme switch")
	then_viewLineSegmentHasForegroundContrastAtLeast(t, gui, viewPullRequestsName, 0, "acme/widgets#42", theme.SelectedLineBackgroundHex, 4.5, "pull request reference contrast after theme switch")
	then_viewLineSegmentHasForegroundContrastAtLeast(t, gui, viewPullRequestsName, 0, "First PR", theme.SelectedLineBackgroundHex, 4.5, "pull request title contrast after theme switch")
}

func TestChangeTheme_GivenPresetSaveFailure_WhenSubmitting_ThenItKeepsThePickerOpenAndShowsTheError(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.Palette{BackgroundHex: "#101010"})

	store := &fakeThemePresetStore{saveErr: errors.New("disk full")}
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.themePresetStore = store
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("change theme", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "change theme"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("kanagawa dark", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "kanagawa dark"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if popupView.Title != themePickerTitle+" · disk full" {
		t.Fatalf("expected popup title %q, actual %q", themePickerTitle+" · disk full", popupView.Title)
	}
	if theme.BackgroundHex != "#101010" {
		t.Fatalf("expected theme background to stay %q, actual %q", "#101010", theme.BackgroundHex)
	}
	if gui.BgColor != gocui.GetColor("#101010") {
		t.Fatalf("expected gui background color %v, actual %v", gocui.GetColor("#101010"), gui.BgColor)
	}
}

func given_themePickerLabels() []string {
	presets := theme.Presets()
	labels := make([]string, 0, len(presets))
	for _, preset := range presets {
		labels = append(labels, preset.Label)
	}
	return labels
}

type fakeThemePresetStore struct {
	savedPresets []string
	saveErr      error
}

func (store *fakeThemePresetStore) SaveThemePreset(preset string) error {
	store.savedPresets = append(store.savedPresets, preset)
	return store.saveErr
}
