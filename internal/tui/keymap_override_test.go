package tui

import (
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenMainPaneSearchOverride_WhenListingBindings_ThenItAppliesToEveryMainPane(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"main": {
			"open_search": {"s"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 's', handler: subject.openSearch})
		then_bindingDoesNotExist(t, actual, viewName, '/')
	}
}

func TestKeybindingSpecs_GivenPullRequestsOpenDetailOverride_WhenListingBindings_ThenItReplacesOnlyThatScopedAction(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"o"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'o', handler: subject.openDetail})
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, gocui.KeyEnter)
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: gocui.KeyEnter, handler: subject.openDetail})
}

func TestKeybindingSpecs_GivenConflictingPullRequestsOverride_WhenListingBindings_ThenItIgnoresTheBadEntry(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"y"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: gocui.KeyEnter, handler: subject.openDetail})
	actualHandler := given_handlerForBinding(t, actual, viewPullRequestsName, 'y')
	if !sameHandler(actualHandler, subject.copyPullRequestURL) {
		t.Fatalf("expected %q to keep the copy URL handler", "y")
	}
}

func TestKeybindingSpecs_GivenFullPageOverride_WhenListingBindings_ThenItKeepsHalfPageBindingsSeparate(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"main": {
			"full_page_down": {"pagedown"},
			"full_page_up":   {"pageup"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyPgdn, handler: subject.fullPageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyPgup, handler: subject.fullPageUp})
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyCtrlF)
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyCtrlB)
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlD, handler: subject.pageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlU, handler: subject.pageUp})
	}
}

func TestKeybindingSpecs_GivenGlobalCloseOverride_WhenListingBindings_ThenItAppliesToEveryClosableViewUnlessScopedOverridesWin(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"close": {"x"},
		},
		"detail": {
			"close": {"d"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'x', handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: 'x', handler: subject.closeHelp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewModalEditorName, key: 'x', handler: subject.closeModalEditor})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'd', handler: subject.closeDetail})
	then_bindingDoesNotExist(t, actual, viewActionsPopupName, 'q')
	then_bindingDoesNotExist(t, actual, viewHelpName, 'q')
	then_bindingDoesNotExist(t, actual, viewModalEditorName, gocui.KeyEsc)
	then_bindingDoesNotExist(t, actual, viewDetailName, 'q')
	then_bindingDoesNotExist(t, actual, viewDetailName, gocui.KeyEsc)
}

func TestKeybindingSpecs_GivenGlobalFullPageOverride_WhenListingBindings_ThenItAppliesAcrossScopesUnlessScopedOverridesWin(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"full_page_down": {")"},
			"full_page_up":   {"("},
		},
		"main": {
			"full_page_down": {"pagedown"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: gocui.KeyPgdn, handler: subject.fullPageDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: '(', handler: subject.fullPageUp})
	then_bindingDoesNotExist(t, actual, viewUserName, ')')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: ')', handler: subject.fullPageActionsPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: '(', handler: subject.fullPageActionsPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: ')', handler: subject.fullPageHelpDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: '(', handler: subject.fullPageHelpUp})
}

func TestKeybindingSpecs_GivenPullRequestsToggleFoldOverride_WhenListingBindings_ThenItSupportsSingleCharacterCustomization(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"toggle_fold": {"o"},
		},
	})

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'o')
	if actualHandler == nil {
		t.Fatal("expected a binding for the customized fold toggle")
	}
	then_bindingExists(t, subject.keybindingSpecs(), keybindingSpec{viewName: viewPullRequestsName, key: 'a', handler: subject.openActionsPopup})
}

func given_programWithKeymapOverrides(model *Model, overrides appconfig.KeymapOverrides) *Program {
	subject := NewProgramWithModel(model)
	subject.ApplyKeymapOverrides(overrides)
	return subject
}
