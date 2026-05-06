package theme

import (
	"reflect"
	"testing"
)

const (
	darkDefaultActiveTextHex              = "#F0F6FC"
	darkDefaultSelectedLineBackgroundHex  = "#21262D"
	darkDefaultMarkdownHeadingHex         = "#58A6FF"
	darkDefaultDiffAdditionBackgroundHex  = "#033A16"
	lightDefaultActiveTextHex             = "#000000"
	lightDefaultSelectedLineBackgroundHex = "#E6E6E6"
)

func TestDefaultPalette_GivenDarkSystemPolarity_WhenResolving_ThenItUsesDarkDefaults(t *testing.T) {
	given_systemPolarityDetector(t, func() systemPolarity { return systemPolarityDark })

	actual := DefaultPalette()

	then_paletteUsesDarkDefaults(t, actual)
}

func TestDefaultPalette_GivenUnknownSystemPolarity_WhenResolving_ThenItFallsBackToLightDefaults(t *testing.T) {
	given_systemPolarityDetector(t, func() systemPolarity { return systemPolarityUnknown })

	actual := DefaultPalette()

	then_paletteUsesLightDefaults(t, actual)
}

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

func TestResolvePalette_GivenDarkSystemPolarityAndPartialOverrides_WhenResolving_ThenItMergesThemWithTheDarkDefaultPalette(t *testing.T) {
	given_systemPolarityDetector(t, func() systemPolarity { return systemPolarityDark })

	actual := ResolvePalette(Palette{
		ActiveBorderHex:           " #7E9CD8 ",
		DiffAdditionForegroundHex: "#98BB6C",
	})

	if actual.ActiveBorderHex != "#7E9CD8" {
		t.Fatalf("expected active border color %q, actual %q", "#7E9CD8", actual.ActiveBorderHex)
	}
	if actual.DiffAdditionForegroundHex != "#98BB6C" {
		t.Fatalf("expected addition foreground %q, actual %q", "#98BB6C", actual.DiffAdditionForegroundHex)
	}
	then_paletteUsesDarkDefaults(t, actual)
}

func TestResetPalette_GivenDarkSystemPolarity_WhenResetting_ThenItRestoresTheDarkDefaultPalette(t *testing.T) {
	t.Cleanup(ResetPalette)
	given_systemPolarityDetector(t, func() systemPolarity { return systemPolarityDark })

	ApplyPalette(Palette{ActiveTextHex: "#123456", SelectedLineBackgroundHex: "#654321"})
	ResetPalette()

	if ActiveTextHex != darkDefaultActiveTextHex {
		t.Fatalf("expected active text color %q, actual %q", darkDefaultActiveTextHex, ActiveTextHex)
	}
	if SelectedLineBackgroundHex != darkDefaultSelectedLineBackgroundHex {
		t.Fatalf("expected selected line background %q, actual %q", darkDefaultSelectedLineBackgroundHex, SelectedLineBackgroundHex)
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

func TestApplyPalette_GivenMarkdownHeadingBackgroundOverride_WhenApplying_ThenItUpdatesThePackageColor(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{MarkdownHeadingBackgroundHex: "#223249"})

	if MarkdownHeadingBackgroundHex != "#223249" {
		t.Fatalf("expected markdown heading background %q, actual %q", "#223249", MarkdownHeadingBackgroundHex)
	}
	if MarkdownHeadingHex != DefaultPalette().MarkdownHeadingHex {
		t.Fatalf("expected untouched markdown heading color %q, actual %q", DefaultPalette().MarkdownHeadingHex, MarkdownHeadingHex)
	}
}

func given_systemPolarityDetector(t *testing.T, detector func() systemPolarity) {
	t.Helper()

	originalDetector := systemPolarityDetector
	systemPolarityDetector = detector
	t.Cleanup(func() {
		systemPolarityDetector = originalDetector
	})
}

func then_paletteUsesDarkDefaults(t *testing.T, actual Palette) {
	t.Helper()

	if actual.ActiveTextHex != darkDefaultActiveTextHex {
		t.Fatalf("expected active text color %q, actual %q", darkDefaultActiveTextHex, actual.ActiveTextHex)
	}
	if actual.SelectedLineBackgroundHex != darkDefaultSelectedLineBackgroundHex {
		t.Fatalf("expected selected line background %q, actual %q", darkDefaultSelectedLineBackgroundHex, actual.SelectedLineBackgroundHex)
	}
	if actual.MarkdownHeadingHex != darkDefaultMarkdownHeadingHex {
		t.Fatalf("expected markdown heading color %q, actual %q", darkDefaultMarkdownHeadingHex, actual.MarkdownHeadingHex)
	}
	if actual.DiffAdditionBackgroundHex != darkDefaultDiffAdditionBackgroundHex {
		t.Fatalf("expected diff addition background %q, actual %q", darkDefaultDiffAdditionBackgroundHex, actual.DiffAdditionBackgroundHex)
	}
}

func then_paletteUsesLightDefaults(t *testing.T, actual Palette) {
	t.Helper()

	if actual.ActiveTextHex != lightDefaultActiveTextHex {
		t.Fatalf("expected active text color %q, actual %q", lightDefaultActiveTextHex, actual.ActiveTextHex)
	}
	if actual.SelectedLineBackgroundHex != lightDefaultSelectedLineBackgroundHex {
		t.Fatalf("expected selected line background %q, actual %q", lightDefaultSelectedLineBackgroundHex, actual.SelectedLineBackgroundHex)
	}
}
