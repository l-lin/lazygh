package theme

import (
	"reflect"
	"testing"
)

const (
	darkDefaultActiveTextHex                = "#F0F6FC"
	darkDefaultSelectedLineBackgroundHex    = "#21262D"
	darkDefaultMarkdownHeadingHex           = "#F0F6FC"
	darkDefaultMarkdownHeadingBackgroundHex = "#58A6FF"
	darkDefaultPullRequestReferenceHex      = "#8B949E"
	darkDefaultPullRequestTitleHex          = "#F0F6FC"
	darkDefaultSuccessHex                   = "#3FB950"
	darkDefaultFailureHex                   = "#F85149"
	darkDefaultPendingHex                   = "#8B949E"
	darkDefaultMutedHex                     = "#8B949E"
	darkDefaultDiffAdditionBackgroundHex    = "#033A16"
	lightDefaultActiveTextHex               = "#000000"
	lightDefaultSelectedLineBackgroundHex   = "#E6E6E6"
	lightDefaultPullRequestReferenceHex     = "#656D76"
	lightDefaultPullRequestTitleHex         = "#000000"
	lightDefaultSuccessHex                  = "#1A7F37"
	lightDefaultFailureHex                  = "#CF222E"
	lightDefaultPendingHex                  = "#656D76"
	lightDefaultMutedHex                    = "#636363"
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
		SuccessHex:                "#7FB069",
	})

	expected := DefaultPalette()
	expected.ActiveBorderHex = "#7E9CD8"
	expected.DiffAdditionForegroundHex = "#98BB6C"
	expected.SuccessHex = "#7FB069"
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

func TestMergePalette_GivenSparseOverrides_WhenMerging_ThenItReplacesOnlyConfiguredFieldsAcrossTheWholePalette(t *testing.T) {
	actual := mergePalette(Palette{
		ActiveBorderHex:                    "#111111",
		CommentAuthorBadgeBackgroundHex:    "#222222",
		DiffDeletionHighlightBackgroundHex: "#333333",
	}, Palette{
		ActiveBorderHex:                    "#AAAAAA",
		CommentAuthorBadgeBackgroundHex:    "",
		DiffDeletionHighlightBackgroundHex: "#BBBBBB",
	})

	expected := Palette{
		ActiveBorderHex:                    "#AAAAAA",
		CommentAuthorBadgeBackgroundHex:    "#222222",
		DiffDeletionHighlightBackgroundHex: "#BBBBBB",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected merged palette %+v, actual %+v", expected, actual)
	}
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
		SuccessHex:                "#7FB069",
		FailureHex:                "#E46876",
		PendingHex:                "#727169",
		MutedHex:                  "#8A8980",
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
	if SuccessHex != "#7FB069" || FailureHex != "#E46876" || PendingHex != "#727169" || MutedHex != "#8A8980" {
		t.Fatalf("expected generic status colors to be applied, actual success=%q failure=%q pending=%q muted=%q", SuccessHex, FailureHex, PendingHex, MutedHex)
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

func TestApplyPalette_GivenPullRequestReferenceOverride_WhenApplying_ThenItUpdatesThePackageColor(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{PullRequestReferenceHex: "#54546D"})

	if PullRequestReferenceHex != "#54546D" {
		t.Fatalf("expected pull request reference color %q, actual %q", "#54546D", PullRequestReferenceHex)
	}
	if ActiveTextHex != DefaultPalette().ActiveTextHex {
		t.Fatalf("expected untouched active text color %q, actual %q", DefaultPalette().ActiveTextHex, ActiveTextHex)
	}
}

func TestApplyPalette_GivenPullRequestTitleOverride_WhenApplying_ThenItUpdatesThePackageColor(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{PullRequestTitleHex: "#1F2937"})

	if PullRequestTitleHex != "#1F2937" {
		t.Fatalf("expected pull request title color %q, actual %q", "#1F2937", PullRequestTitleHex)
	}
	if ActiveTextHex != DefaultPalette().ActiveTextHex {
		t.Fatalf("expected untouched active text color %q, actual %q", DefaultPalette().ActiveTextHex, ActiveTextHex)
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
	if actual.MarkdownHeadingBackgroundHex != darkDefaultMarkdownHeadingBackgroundHex {
		t.Fatalf("expected markdown heading background %q, actual %q", darkDefaultMarkdownHeadingBackgroundHex, actual.MarkdownHeadingBackgroundHex)
	}
	if actual.PullRequestReferenceHex != darkDefaultPullRequestReferenceHex {
		t.Fatalf("expected pull request reference color %q, actual %q", darkDefaultPullRequestReferenceHex, actual.PullRequestReferenceHex)
	}
	if actual.PullRequestTitleHex != darkDefaultPullRequestTitleHex {
		t.Fatalf("expected pull request title color %q, actual %q", darkDefaultPullRequestTitleHex, actual.PullRequestTitleHex)
	}
	if actual.SuccessHex != darkDefaultSuccessHex || actual.FailureHex != darkDefaultFailureHex || actual.PendingHex != darkDefaultPendingHex || actual.MutedHex != darkDefaultMutedHex {
		t.Fatalf("expected generic status colors success=%q failure=%q pending=%q muted=%q, actual success=%q failure=%q pending=%q muted=%q", darkDefaultSuccessHex, darkDefaultFailureHex, darkDefaultPendingHex, darkDefaultMutedHex, actual.SuccessHex, actual.FailureHex, actual.PendingHex, actual.MutedHex)
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
	if actual.PullRequestReferenceHex != lightDefaultPullRequestReferenceHex {
		t.Fatalf("expected pull request reference color %q, actual %q", lightDefaultPullRequestReferenceHex, actual.PullRequestReferenceHex)
	}
	if actual.PullRequestTitleHex != lightDefaultPullRequestTitleHex {
		t.Fatalf("expected pull request title color %q, actual %q", lightDefaultPullRequestTitleHex, actual.PullRequestTitleHex)
	}
	if actual.SuccessHex != lightDefaultSuccessHex || actual.FailureHex != lightDefaultFailureHex || actual.PendingHex != lightDefaultPendingHex || actual.MutedHex != lightDefaultMutedHex {
		t.Fatalf("expected generic status colors success=%q failure=%q pending=%q muted=%q, actual success=%q failure=%q pending=%q muted=%q", lightDefaultSuccessHex, lightDefaultFailureHex, lightDefaultPendingHex, lightDefaultMutedHex, actual.SuccessHex, actual.FailureHex, actual.PendingHex, actual.MutedHex)
	}
}
