package tui

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
	"github.com/jesseduffield/gocui"
)

const (
	actionsPopupChangeThemeIcon = ""
	themePickerActionTitle      = "Change theme"
	themePickerTitle            = "Select theme"
)

type themePresetStore interface {
	SaveThemePreset(string) error
}

type defaultThemePresetStore struct {
	save func(string) error
}

type themePickerState struct{}

func (store *defaultThemePresetStore) SaveThemePreset(preset string) error {
	if store == nil || store.save == nil {
		return errors.New("theme preset store is unavailable")
	}
	return store.save(preset)
}

func (program *Program) themePickerVisible() bool {
	return program.themePicker != nil
}

func (program *Program) changeThemeActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "change-theme",
		title:    themePickerActionTitle,
		icon:     actionsPopupChangeThemeIcon,
		keywords: []string{"theme", "color", "colorscheme", "palette", "appearance", "light", "dark"},
		execute:  program.executeOpenThemePickerAction,
	}
}

func (program *Program) executeOpenThemePickerAction(_ *gocui.Gui) actionsPopupActionResult {
	program.themePicker = &themePickerState{}
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
	return actionsPopupActionResult{}
}

func (program *Program) currentThemePickerActions() []actionsPopupAction {
	if !program.themePickerVisible() {
		return nil
	}

	presets := theme.Presets()
	actions := make([]actionsPopupAction, 0, len(presets))
	for _, preset := range presets {
		actions = append(actions, program.themePickerAction(preset))
	}
	return actions
}

func (program *Program) themePickerAction(preset theme.Preset) actionsPopupAction {
	normalizedName := theme.NormalizePresetName(preset.Name)
	return actionsPopupAction{
		id:       "theme-" + normalizedName,
		title:    strings.TrimSpace(preset.Label),
		keywords: themePickerKeywords(preset),
		execute: func(gui *gocui.Gui) actionsPopupActionResult {
			return program.executeThemePickerPresetAction(gui, preset)
		},
	}
}

func themePickerKeywords(preset theme.Preset) []string {
	keywords := []string{strings.TrimSpace(preset.Name), strings.TrimSpace(preset.Label)}
	if strings.EqualFold(strings.TrimSpace(preset.Name), theme.SystemPresetName) {
		keywords = append(keywords, "auto", "default", "system")
	}
	return keywords
}

func (program *Program) executeThemePickerPresetAction(gui *gocui.Gui, preset theme.Preset) actionsPopupActionResult {
	if !program.themePickerVisible() {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.themePresetStore == nil {
		return actionsPopupActionResult{err: errors.New("theme preset store is unavailable")}
	}

	normalizedName := theme.NormalizePresetName(preset.Name)
	if normalizedName == "" {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.themePresetStore.SaveThemePreset(normalizedName); err != nil {
		return actionsPopupActionResult{err: err}
	}

	theme.ApplyPalette(theme.ResolvePaletteWithPreset(normalizedName, theme.Palette{}))
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.actionsPopupErrorMessage = ""
	program.setFeedback(program.model.Focus(), fmt.Sprintf("Theme changed to %s", strings.TrimSpace(preset.Label)))
	if gui != nil {
		program.configureGUI(gui)
	}
	return actionsPopupActionResult{closePopup: true}
}
