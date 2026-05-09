package tui

import (
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenGlobalOpenSearchOverride_WhenListingBindings_ThenItAppliesToMainPanesTheActionsPopupAndTheBuildPopup(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"open_search": {"s"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName, viewPullRequestBuildInfoName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 's', handler: subject.openSearch})
		then_bindingDoesNotExist(t, actual, viewName, '/')
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 's', handler: subject.focusActionsPopupSearch})
	then_bindingDoesNotExist(t, actual, viewActionsPopupName, '/')
}

func TestKeybindingSpecs_GivenSideOpenDetailOverride_WhenListingBindings_ThenItAppliesAcrossEverySideView(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"side": {
			"open_detail": {"o"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'o', handler: subject.openDetail})
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyEnter)
	}
}

func TestKeybindingSpecs_GivenConflictingSideOpenDetailOverride_WhenListingBindings_ThenItIgnoresTheBadEntry(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"side": {
			"open_detail": {"?"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyEnter, handler: subject.openDetail})
	}
	actualHandler := given_handlerForBinding(t, actual, viewPullRequestsName, '?')
	if !sameHandler(actualHandler, subject.toggleHelp) {
		t.Fatalf("expected %q to keep the toggle help handler", "?")
	}
}

func TestKeybindingSpecs_GivenFullPageOverride_WhenListingBindings_ThenItKeepsHalfPageBindingsSeparate(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
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

func TestKeybindingSpecs_GivenSearchSubmitOverride_WhenListingBindings_ThenItAppliesToTheActionsPopupSearchPrompt(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"search": {
			"submit": {"x"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: 'x', handler: subject.submitSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: 'x', handler: subject.focusActionsPopupList})
	then_bindingDoesNotExist(t, actual, viewSearchName, gocui.KeyEnter)
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, gocui.KeyEnter)
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, gocui.KeyTab)
}

func TestKeybindingSpecs_GivenGlobalCloseOverride_WhenListingBindings_ThenItAppliesToEveryClosableView(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"close": {"X"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'X', handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: 'X', handler: subject.closeHelp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewModalEditorName, key: 'X', handler: subject.closeModalEditor})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'X', handler: subject.closeDetail})
	then_bindingDoesNotExist(t, actual, viewActionsPopupName, 'q')
	then_bindingDoesNotExist(t, actual, viewHelpName, 'q')
	then_bindingDoesNotExist(t, actual, viewModalEditorName, gocui.KeyEsc)
	then_bindingDoesNotExist(t, actual, viewDetailName, 'q')
	then_bindingDoesNotExist(t, actual, viewDetailName, gocui.KeyEsc)
}

func TestKeybindingSpecs_GivenGlobalFullPageOverride_WhenListingBindings_ThenItAppliesAcrossScopes(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"full_page_down": {")"},
			"full_page_up":   {"("},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: ')', handler: subject.fullPageDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: '(', handler: subject.fullPageUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: ')', handler: subject.fullPageActionsPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: '(', handler: subject.fullPageActionsPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: ')', handler: subject.fullPageHelpDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: '(', handler: subject.fullPageHelpUp})
}

func TestKeybindingSpecs_GivenSelectionOverride_WhenListingBindings_ThenItAppliesToSideViewsAndTheActionsPopup(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"selection": {
			"place_selection_at_viewport_top":    {"xt"},
			"recenter_selection":                 {"xx"},
			"place_selection_at_viewport_bottom": {"xb"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewActionsPopupName} {
		then_bindingKeyExists(t, actual, viewName, 'x')
	}
	for _, viewName := range []string{viewUserName, viewNotificationsName, viewActionsPopupName} {
		then_bindingDoesNotExist(t, actual, viewName, 'z')
	}
}

func TestKeybindingSpecs_GivenPullRequestsCopyOverride_WhenListingBindings_ThenItAppliesToTheUserListAndDetailWithoutSeparateScopes(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"copy_pull_request_url": {"u"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'u', handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'u', handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'u', handler: subject.copyPullRequestURL})
	then_bindingDoesNotExist(t, actual, viewUserName, 'y')
	then_bindingDoesNotExist(t, actual, viewDetailName, 'y')
}

func TestKeybindingSpecs_GivenFocusViewOverrides_WhenListingBindings_ThenTheNumericShortcutsStayFixed(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"global": {
			"focus_user_view":          {"u"},
			"focus_pull_requests_view": {"p"},
			"focus_notifications_view": {"n"},
			"focus_detail_view":        {"d"},
		},
		"main": {
			"focus_user_view":          {"u"},
			"focus_pull_requests_view": {"p"},
			"focus_notifications_view": {"n"},
		},
		"side": {
			"focus_detail_view": {"d"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: '1', handler: subject.focusUserView})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: '2', handler: subject.focusPullRequestsView})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewNotificationsName, key: '3', handler: subject.focusNotificationsView})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: '0', handler: subject.focusDetailView})
	then_bindingDoesNotExist(t, actual, viewUserName, 'u')
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, 'p')
	then_bindingDoesNotExist(t, actual, viewNotificationsName, 'n')
	then_bindingDoesNotExist(t, actual, viewUserName, 'd')
}

func TestKeybindingSpecs_GivenFoldToggleOverride_WhenListingBindings_ThenItAppliesToTheReviewTreeAndDetailSections(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"folds": {
			"toggle_fold": {"o"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'o', handler: subject.togglePullRequestFold})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'o', handler: subject.toggleInlineConversationVisibility})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEnter, handler: subject.toggleInlineConversationVisibility})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'a', handler: subject.openActionsPopup})
}

func TestResolvedKeyLabels_GivenTwoStepTabOverrides_WhenResolving_ThenItKeepsTheConfiguredSequences(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"previous_tab": {"g["},
			"next_tab":     {"g]"},
		},
	})

	actual, ok, hasOverride := subject.resolvedKeyLabels(
		keybindingActionID{scope: keymapScopePullRequests, action: "previous_tab"},
		keybindingActionID{scope: keymapScopePullRequests, action: "next_tab"},
	)
	if !ok {
		t.Fatal("expected the tab labels to resolve")
	}
	if !hasOverride {
		t.Fatal("expected the tab labels to report overrides")
	}
	expected := []string{"g[", "g]"}
	if len(actual) != len(expected) {
		t.Fatalf("expected labels %v, actual %v", expected, actual)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("expected labels %v, actual %v", expected, actual)
		}
	}
}

func given_programWithKeymapOverrides(model *Model, overrides appconfig.KeymapOverrides) *Program {
	subject := NewProgramWithModel(model)
	subject.ApplyKeymapOverrides(overrides)
	return subject
}
