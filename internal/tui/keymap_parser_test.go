package tui

import (
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestParseConfiguredKey_GivenSupportedKeyStrings_WhenParsing_ThenItReturnsBindingsAndHelpLabels(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		expectedKey   any
		expectedLabel string
	}{
		{name: "single rune", value: "o", expectedKey: 'o', expectedLabel: "o"},
		{name: "uppercase rune", value: "G", expectedKey: 'G', expectedLabel: "G"},
		{name: "enter", value: "enter", expectedKey: gocui.KeyEnter, expectedLabel: "<enter>"},
		{name: "escape", value: "esc", expectedKey: gocui.KeyEsc, expectedLabel: "<esc>"},
		{name: "control bracket", value: "ctrl+[", expectedKey: gocui.KeyCtrlLsqBracket, expectedLabel: "<c-[>"},
		{name: "control b", value: "ctrl+b", expectedKey: gocui.KeyCtrlB, expectedLabel: "<c-b>"},
		{name: "control f", value: "ctrl+f", expectedKey: gocui.KeyCtrlF, expectedLabel: "<c-f>"},
		{name: "tab", value: "tab", expectedKey: gocui.KeyTab, expectedLabel: "tab"},
		{name: "shift tab", value: "shift+tab", expectedKey: gocui.KeyBacktab, expectedLabel: "shift+tab"},
		{name: "down", value: "down", expectedKey: gocui.KeyArrowDown, expectedLabel: "<down>"},
		{name: "pageup", value: "pageup", expectedKey: gocui.KeyPgup, expectedLabel: "pageup"},
		{name: "pagedown", value: "pagedown", expectedKey: gocui.KeyPgdn, expectedLabel: "pagedown"},
		{name: "alt enter", value: "alt+enter", expectedKey: gocui.KeyAltEnter, expectedLabel: "alt+enter"},
		{name: "space", value: "space", expectedKey: ' ', expectedLabel: "space"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, ok := parseConfiguredKey(testCase.value)
			if !ok {
				t.Fatalf("expected %q to parse", testCase.value)
			}
			if !reflect.DeepEqual(actual.value, testCase.expectedKey) {
				t.Fatalf("expected key %v, actual %v", testCase.expectedKey, actual.value)
			}
			if actual.label != testCase.expectedLabel {
				t.Fatalf("expected label %q, actual %q", testCase.expectedLabel, actual.label)
			}
		})
	}
}

func TestParseConfiguredBindings_GivenATwoCharacterSequence_WhenParsing_ThenItReturnsATwoStepBindingWithACombinedHelpLabel(t *testing.T) {
	actual, ok := parseConfiguredBindings([]string{"za"})
	if !ok {
		t.Fatal("expected the configured sequence to parse")
	}
	if len(actual) != 1 {
		t.Fatalf("expected one binding, actual %d", len(actual))
	}
	if actual[0].label != "za" {
		t.Fatalf("expected label %q, actual %q", "za", actual[0].label)
	}
	if len(actual[0].keys) != 2 {
		t.Fatalf("expected two keys, actual %d", len(actual[0].keys))
	}
	if !reflect.DeepEqual(actual[0].keys[0].value, 'z') {
		t.Fatalf("expected first key %q, actual %v", 'z', actual[0].keys[0].value)
	}
	if !reflect.DeepEqual(actual[0].keys[1].value, 'a') {
		t.Fatalf("expected second key %q, actual %v", 'a', actual[0].keys[1].value)
	}
}

func TestParseConfiguredBindings_GivenUnsupportedKeyString_WhenParsing_ThenItRejectsTheValue(t *testing.T) {
	_, ok := parseConfiguredBindings([]string{"banana"})
	if ok {
		t.Fatal("expected the configured key to be rejected")
	}
}
