package tui

import (
	"reflect"
	"testing"
)

func TestSharedKeybindingDefinitionFor_GivenMutualizedActions_WhenLookingThemUp_ThenItReturnsTheirBehaviorFirstScopesAndDefaults(t *testing.T) {
	testCases := []struct {
		name                   string
		action                 string
		expectedScope          string
		expectedBindings       []string
		expectedAllowSequences bool
	}{
		{name: "toggle help", action: "toggle_help", expectedScope: keymapScopeGlobal, expectedBindings: []string{"?"}, expectedAllowSequences: true},
		{name: "open search", action: "open_search", expectedScope: keymapScopeGlobal, expectedBindings: []string{"/"}, expectedAllowSequences: true},
		{name: "move selection down", action: "move_selection_down", expectedScope: keymapScopeGlobal, expectedBindings: []string{"j", "down"}, expectedAllowSequences: true},
		{name: "move selection up", action: "move_selection_up", expectedScope: keymapScopeGlobal, expectedBindings: []string{"k", "up"}, expectedAllowSequences: true},
		{name: "page down", action: "page_down", expectedScope: keymapScopeGlobal, expectedBindings: []string{"ctrl+d"}, expectedAllowSequences: true},
		{name: "page up", action: "page_up", expectedScope: keymapScopeGlobal, expectedBindings: []string{"ctrl+u"}, expectedAllowSequences: true},
		{name: "grow focused pane", action: "grow_focused_pane", expectedScope: keymapScopeGlobal, expectedBindings: []string{"+"}, expectedAllowSequences: true},
		{name: "shrink focused pane", action: "shrink_focused_pane", expectedScope: keymapScopeGlobal, expectedBindings: []string{"-"}, expectedAllowSequences: true},
		{name: "open actions popup", action: "open_actions_popup", expectedScope: keymapScopeGlobal, expectedBindings: []string{"a"}, expectedAllowSequences: true},
		{name: "refresh", action: "refresh", expectedScope: keymapScopeGlobal, expectedBindings: []string{"alt+r"}, expectedAllowSequences: true},
		{name: "open detail", action: "open_detail", expectedScope: keymapScopeSide, expectedBindings: []string{"enter"}, expectedAllowSequences: true},
		{name: "toggle fold", action: "toggle_fold", expectedScope: keymapScopeFolds, expectedBindings: []string{"za"}, expectedAllowSequences: true},
		{name: "close all folds", action: "close_all_folds", expectedScope: keymapScopeFolds, expectedBindings: []string{"zM"}, expectedAllowSequences: true},
		{name: "open all folds", action: "open_all_folds", expectedScope: keymapScopeFolds, expectedBindings: []string{"zR"}, expectedAllowSequences: true},
		{name: "next search match", action: "next_search_match", expectedScope: keymapScopeSearch, expectedBindings: []string{"n"}, expectedAllowSequences: true},
		{name: "previous search match", action: "previous_search_match", expectedScope: keymapScopeSearch, expectedBindings: []string{"N"}, expectedAllowSequences: true},
		{name: "previous tab", action: "previous_tab", expectedScope: keymapScopeGlobal, expectedBindings: []string{"["}, expectedAllowSequences: true},
		{name: "next tab", action: "next_tab", expectedScope: keymapScopeGlobal, expectedBindings: []string{"]"}, expectedAllowSequences: true},
		{name: "copy pull request url", action: "copy_pull_request_url", expectedScope: keymapScopePullRequests, expectedBindings: []string{"alt+y"}, expectedAllowSequences: true},
		{name: "comment on pull request", action: "comment_on_pull_request", expectedScope: keymapScopePullRequests, expectedBindings: []string{"c"}, expectedAllowSequences: true},
		{name: "move selection to top", action: "move_selection_to_top", expectedScope: keymapScopeSelection, expectedBindings: []string{"gg"}, expectedAllowSequences: true},
		{name: "move selection to bottom", action: "move_selection_to_bottom", expectedScope: keymapScopeSelection, expectedBindings: []string{"G"}, expectedAllowSequences: true},
		{name: "place selection at viewport top", action: "place_selection_at_viewport_top", expectedScope: keymapScopeSelection, expectedBindings: []string{"zt"}, expectedAllowSequences: true},
		{name: "recenter selection", action: "recenter_selection", expectedScope: keymapScopeSelection, expectedBindings: []string{"zz"}, expectedAllowSequences: true},
		{name: "place selection at viewport bottom", action: "place_selection_at_viewport_bottom", expectedScope: keymapScopeSelection, expectedBindings: []string{"zb"}, expectedAllowSequences: true},
		{name: "move cursor left", action: "move_cursor_left", expectedScope: keymapScopeCursor, expectedBindings: []string{"h", "left"}, expectedAllowSequences: true},
		{name: "move cursor right", action: "move_cursor_right", expectedScope: keymapScopeCursor, expectedBindings: []string{"l", "right"}, expectedAllowSequences: true},
		{name: "find character forward", action: "find_character_forward", expectedScope: keymapScopeCursor, expectedBindings: []string{"f"}, expectedAllowSequences: false},
		{name: "find character backward", action: "find_character_backward", expectedScope: keymapScopeCursor, expectedBindings: []string{"F"}, expectedAllowSequences: false},
		{name: "till character forward", action: "till_character_forward", expectedScope: keymapScopeCursor, expectedBindings: []string{"t"}, expectedAllowSequences: false},
		{name: "till character backward", action: "till_character_backward", expectedScope: keymapScopeCursor, expectedBindings: []string{"T"}, expectedAllowSequences: false},
		{name: "repeat character motion forward", action: "repeat_character_motion_forward", expectedScope: keymapScopeCursor, expectedBindings: []string{";"}, expectedAllowSequences: false},
		{name: "repeat character motion backward", action: "repeat_character_motion_backward", expectedScope: keymapScopeCursor, expectedBindings: []string{","}, expectedAllowSequences: false},
		{name: "start yank", action: "start_yank", expectedScope: keymapScopeCursor, expectedBindings: []string{"y"}, expectedAllowSequences: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, ok := sharedKeybindingDefinitionFor(testCase.action)
			if !ok {
				t.Fatalf("expected shared definition for %q", testCase.action)
			}
			if actual.scope != testCase.expectedScope {
				t.Fatalf("expected scope %q, actual %q", testCase.expectedScope, actual.scope)
			}
			if !reflect.DeepEqual(actual.bindings, testCase.expectedBindings) {
				t.Fatalf("expected bindings %v, actual %v", testCase.expectedBindings, actual.bindings)
			}
			if actual.allowSequences != testCase.expectedAllowSequences {
				t.Fatalf("expected allow sequences %t, actual %t", testCase.expectedAllowSequences, actual.allowSequences)
			}
		})
	}
}

func TestKeybindingActions_GivenProgram_WhenListingActions_ThenViewFocusShortcutsStayFixedAndNonConfigurable(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actions := subject.keybindingActions()
	expected := map[keybindingActionID]bool{
		{scope: keymapScopeMain, action: "focus_user_view"}:          false,
		{scope: keymapScopeMain, action: "focus_pull_requests_view"}: false,
		{scope: keymapScopeMain, action: "focus_notifications_view"}: false,
		{scope: keymapScopeSide, action: "focus_detail_view"}:        false,
	}

	for _, action := range actions {
		expectedConfigurable, ok := expected[action.id]
		if !ok {
			continue
		}
		if action.configurable != expectedConfigurable {
			t.Fatalf("expected action %v configurable=%t, actual %t", action.id, expectedConfigurable, action.configurable)
		}
	}
}

func TestKeybindingActions_GivenProgram_WhenListingActions_ThenLegacyViewScopesStayRemoved(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	legacyScopes := map[string]struct{}{
		"actions_popup_search": {},
		"detail":               {},
		"help":                 {},
	}
	for _, action := range subject.keybindingActions() {
		if _, ok := legacyScopes[action.id.scope]; ok {
			t.Fatalf("expected action id scope %q to stay removed for %+v", action.id.scope, action.id)
		}
		if action.configID == (keybindingActionID{}) {
			continue
		}
		if _, ok := legacyScopes[action.configID.scope]; ok {
			t.Fatalf("expected config scope %q to stay removed for %+v", action.configID.scope, action.id)
		}
	}
}

func TestKeybindingActions_GivenProgram_WhenListingActions_ThenSideViewSwitchAliasesReuseTheGlobalConfigEntries(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actualConfigScopes := map[keybindingActionID]keybindingActionID{}
	for _, action := range subject.keybindingActions() {
		actualConfigScopes[action.id] = action.configID
	}

	for _, actionID := range []keybindingActionID{
		{scope: keymapScopeSide, action: "next_side_view"},
		{scope: keymapScopeSide, action: "previous_side_view"},
	} {
		actualConfigID, ok := actualConfigScopes[actionID]
		if !ok {
			t.Fatalf("expected action %+v to exist", actionID)
		}
		expected := keybindingActionID{scope: keymapScopeGlobal, action: actionID.action}
		if actualConfigID != expected {
			t.Fatalf("expected action %+v to use config id %+v, actual %+v", actionID, expected, actualConfigID)
		}
	}
}
