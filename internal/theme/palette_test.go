package theme

import (
	"reflect"
	"testing"
)

func TestResolvePalette_GivenPartialOverrides_WhenResolving_ThenItMergesThemWithTheDefaultPalette(t *testing.T) {
	actual := ResolvePalette(Palette{
		ActiveBorderHex:           " #7E9CD8 ",
		DiffAdditionForegroundHex: "#98BB6C",
	})

	expected := DefaultPalette()
	expected.ActiveBorderHex = "#7E9CD8"
	expected.DiffAdditionForegroundHex = "#98BB6C"
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected resolved palette %+v, actual %+v", expected, actual)
	}
}

func TestResolvePalette_GivenInvalidHexColors_WhenResolving_ThenItIgnoresTheBadEntries(t *testing.T) {
	actual := ResolvePalette(Palette{
		ActiveBorderHex:           "blue",
		InactiveBorderHex:         "#54546D",
		SelectedLineBackgroundHex: "#12",
	})

	expected := DefaultPalette()
	expected.InactiveBorderHex = "#54546D"
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected resolved palette %+v, actual %+v", expected, actual)
	}
}

func TestApplyPalette_GivenOverrides_WhenApplying_ThenItUpdatesThePackageColors(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{
		ActiveBorderHex:           "#7E9CD8",
		InactiveBorderHex:         "#54546D",
		SelectedLineBackgroundHex: "#223249",
	})

	if ActiveBorderHex != "#7E9CD8" {
		t.Fatalf("expected active border color %q, actual %q", "#7E9CD8", ActiveBorderHex)
	}
	if InactiveBorderHex != "#54546D" {
		t.Fatalf("expected inactive border color %q, actual %q", "#54546D", InactiveBorderHex)
	}
	if SelectedLineBackgroundHex != "#223249" {
		t.Fatalf("expected selected line background %q, actual %q", "#223249", SelectedLineBackgroundHex)
	}
	if ActiveTextHex != DefaultPalette().ActiveTextHex {
		t.Fatalf("expected untouched palette values to keep their default, actual %q", ActiveTextHex)
	}
}
