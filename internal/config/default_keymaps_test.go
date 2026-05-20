package config

import (
	"reflect"
	"testing"
)

func TestDefaultKeymaps_GivenEmbeddedDefaults_WhenLoading_ThenItIncludesTheSharedSideViewAliasesAndModalEditorBindings(t *testing.T) {
	actual := DefaultKeymaps()

	expected := KeymapOverrides{
		"global": {
			"next_side_view":     {"tab", "l"},
			"previous_side_view": {"shift+tab", "h"},
		},
		"cursor": {
			"search_word_under_cursor_forward":  {"*"},
			"search_word_under_cursor_backward": {"#"},
			"find_character_forward":            {"f"},
			"find_character_backward":           {"F"},
			"till_character_forward":            {"t"},
			"till_character_backward":           {"T"},
			"repeat_character_motion_forward":   {";"},
			"repeat_character_motion_backward":  {","},
		},
		"modal_editor": {
			"cancel":               {"esc"},
			"open_external_editor": {"ctrl+g"},
		},
		"pull_requests": {
			"copy_pull_request_url":    {"alt+y"},
			"custom_search":            {":"},
			"open_pull_request_by_url": {"ctrl+v"},
			"reply_to_inline_comment":  {"r"},
		},
		"pull_request_build_info": {
			"copy_content": {"alt+y"},
		},
	}
	for scope, expectedActions := range expected {
		actualActions, ok := actual[scope]
		if !ok {
			t.Fatalf("expected default keymaps to contain scope %q", scope)
		}
		for action, expectedBindings := range expectedActions {
			actualBindings, ok := actualActions[action]
			if !ok {
				t.Fatalf("expected default keymaps to contain action %q in scope %q", action, scope)
			}
			if !reflect.DeepEqual(actualBindings, expectedBindings) {
				t.Fatalf("expected default bindings %v for %s.%s, actual %v", expectedBindings, scope, action, actualBindings)
			}
		}
	}
}

func TestDefaultKeymaps_GivenMutatedReturnedValue_WhenLoadingAgain_ThenItReturnsAFreshCopy(t *testing.T) {
	first := DefaultKeymaps()
	first["global"]["quit"][0] = "mutated"

	actual := DefaultKeymaps()

	expected := []string{"ctrl+c"}
	if !reflect.DeepEqual(actual["global"]["quit"], expected) {
		t.Fatalf("expected default quit bindings %v, actual %v", expected, actual["global"]["quit"])
	}
}
