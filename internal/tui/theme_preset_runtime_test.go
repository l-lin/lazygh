package tui

import (
	"reflect"
	"testing"
)

func TestThemeCapabilitySnapshot_GivenThemePresetStore_WhenResolving_ThenItReportsPresetPersistenceAvailability(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.themePresetStore = &fakeThemePresetStore{}

	actual := subject.currentThemeCapabilitySnapshot()

	if !actual.presetPersistenceAvailable {
		t.Fatal("expected theme preset persistence to be available")
	}
}

func TestThemeCapabilitySnapshot_GivenMissingThemePresetStore_WhenResolving_ThenItReportsPresetPersistenceAsUnavailable(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.themePresetStore = nil

	actual := subject.currentThemeCapabilitySnapshot()

	if actual.presetPersistenceAvailable {
		t.Fatal("expected theme preset persistence to stay unavailable")
	}
}

func TestNewThemePresetRuntime_GivenThemePresetStore_WhenSavingThePreset_ThenItDelegatesToTheStore(t *testing.T) {
	store := &fakeThemePresetStore{}
	subject := NewProgramWithModel(given_model())
	subject.themePresetStore = store

	actual := newThemePresetRuntime(subject)
	if actual.saveThemePreset == nil {
		t.Fatal("expected the runtime save hook")
	}
	actualErr := actual.saveThemePreset("kanagawa-dark")
	then_noError(t, actualErr)

	expected := []string{"kanagawa-dark"}
	if actual := store.savedPresets; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected saved presets %v, actual %v", expected, actual)
	}
}
