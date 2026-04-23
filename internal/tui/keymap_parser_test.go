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
		{name: "tab", value: "tab", expectedKey: gocui.KeyTab, expectedLabel: "tab"},
		{name: "shift tab", value: "shift+tab", expectedKey: gocui.KeyBacktab, expectedLabel: "shift+tab"},
		{name: "down", value: "down", expectedKey: gocui.KeyArrowDown, expectedLabel: "<down>"},
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

func TestParseConfiguredKey_GivenUnsupportedKeyString_WhenParsing_ThenItRejectsTheValue(t *testing.T) {
	_, ok := parseConfiguredKey("banana")
	if ok {
		t.Fatal("expected the configured key to be rejected")
	}
}
