package tui

import (
	"reflect"
	"testing"
)

func TestSharedKeybindingDefinitionFor_GivenMutualizedActions_WhenLookingThemUp_ThenItReturnsTheirSharedDefaults(t *testing.T) {
	testCases := []struct {
		name                   string
		action                 string
		expectedBindings       []string
		expectedAllowSequences bool
	}{
		{name: "open actions popup", action: "open_actions_popup", expectedBindings: []string{"a"}, expectedAllowSequences: true},
		{name: "close all folds", action: "close_all_folds", expectedBindings: []string{"zM"}, expectedAllowSequences: true},
		{name: "open all folds", action: "open_all_folds", expectedBindings: []string{"zR"}, expectedAllowSequences: true},
		{name: "next search match", action: "next_search_match", expectedBindings: []string{"n"}, expectedAllowSequences: true},
		{name: "previous search match", action: "previous_search_match", expectedBindings: []string{"N"}, expectedAllowSequences: true},
		{name: "previous tab", action: "previous_tab", expectedBindings: []string{"["}, expectedAllowSequences: true},
		{name: "next tab", action: "next_tab", expectedBindings: []string{"]"}, expectedAllowSequences: true},
		{name: "copy pull request url", action: "copy_pull_request_url", expectedBindings: []string{"y"}, expectedAllowSequences: true},
		{name: "comment on pull request", action: "comment_on_pull_request", expectedBindings: []string{"c"}, expectedAllowSequences: true},
		{name: "page down", action: "page_down", expectedBindings: []string{"ctrl+d"}, expectedAllowSequences: true},
		{name: "page up", action: "page_up", expectedBindings: []string{"ctrl+u"}, expectedAllowSequences: true},
		{name: "move selection to top", action: "move_selection_to_top", expectedBindings: []string{"gg"}, expectedAllowSequences: true},
		{name: "move selection to bottom", action: "move_selection_to_bottom", expectedBindings: []string{"G"}, expectedAllowSequences: true},
		{name: "place selection at viewport top", action: "place_selection_at_viewport_top", expectedBindings: []string{"zt"}, expectedAllowSequences: true},
		{name: "recenter selection", action: "recenter_selection", expectedBindings: []string{"zz"}, expectedAllowSequences: true},
		{name: "place selection at viewport bottom", action: "place_selection_at_viewport_bottom", expectedBindings: []string{"zb"}, expectedAllowSequences: true},
		{name: "move cursor left", action: "move_cursor_left", expectedBindings: []string{"h", "left"}, expectedAllowSequences: true},
		{name: "move cursor right", action: "move_cursor_right", expectedBindings: []string{"l", "right"}, expectedAllowSequences: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, ok := sharedKeybindingDefinitionFor(testCase.action)
			if !ok {
				t.Fatalf("expected shared definition for %q", testCase.action)
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
