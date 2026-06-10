package config

import "testing"

func TestDefaultKeymaps_display_commit_changes_GivenBinding_WhenLoading_ThenItUsesGF(t *testing.T) {
	actual := DefaultKeymaps()

	expected := []string{"gf"}
	actualBindings, ok := actual["cursor"]["display_commit_changes"]
	if !ok {
		t.Fatal("expected the default keymaps to include cursor.display_commit_changes")
	}
	if len(actualBindings) != len(expected) || actualBindings[0] != expected[0] {
		t.Fatalf("expected bindings %v, actual %v", expected, actualBindings)
	}
}
