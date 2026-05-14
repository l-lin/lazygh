package theme

import (
	"reflect"
	"testing"
)

const (
	darkDefaultBackgroundHex                   = ""
	darkDefaultActiveTextHex                   = "#F0F6FC"
	darkDefaultSelectedLineBackgroundHex       = "#21262D"
	darkDefaultMarkdownHeadingHex              = "#F0F6FC"
	darkDefaultMarkdownHeadingBackgroundHex    = "#58A6FF"
	darkDefaultPullRequestReferenceHex         = "#8B949E"
	darkDefaultPullRequestTitleHex             = "#F0F6FC"
	darkDefaultSuccessHex                      = "#3FB950"
	darkDefaultSuccessBackgroundHex            = "#033A16"
	darkDefaultFailureHex                      = "#F85149"
	darkDefaultFailureBackgroundHex            = "#67060C"
	darkDefaultPendingHex                      = "#8B949E"
	darkDefaultPendingBackgroundHex            = "#30363D"
	darkDefaultMutedHex                        = "#8B949E"
	darkDefaultWarningHex                      = "#D29922"
	darkDefaultActionsPopupGroupForegroundHex  = darkDefaultMarkdownHeadingHex
	darkDefaultTeamOwnershipHex                = "#8B949E"
	darkDefaultDiffAdditionBackgroundHex       = "#033A16"
	lightDefaultBackgroundHex                  = ""
	lightDefaultActiveTextHex                  = "#000000"
	lightDefaultSelectedLineBackgroundHex      = "#E6E6E6"
	lightDefaultActionsPopupGroupForegroundHex = "#000000"
	lightDefaultTeamOwnershipHex               = "#656D76"
	lightDefaultPullRequestReferenceHex        = "#656D76"
	lightDefaultPullRequestTitleHex            = "#000000"
	lightDefaultSuccessHex                     = "#1A7F37"
	lightDefaultSuccessBackgroundHex           = "#DFF3E4"
	lightDefaultFailureHex                     = "#CF222E"
	lightDefaultFailureBackgroundHex           = "#FFE2E5"
	lightDefaultPendingHex                     = "#656D76"
	lightDefaultPendingBackgroundHex           = "#E6E6E6"
	lightDefaultMutedHex                       = "#636363"
	lightDefaultWarningHex                     = "#9A6700"
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
		ActiveBorderHex: " #7E9CD8 ",
		DiffAdditionHex: "#98BB6C",
		SuccessHex:      "#7FB069",
	})

	expected := DefaultPalette()
	expected.ActiveBorderHex = "#7E9CD8"
	expected.DiffAdditionHex = "#98BB6C"
	expected.SuccessHex = "#7FB069"
	expected.PullRequestStatusOpenHex = "#7FB069"
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
		ActiveBorderHex: " #7E9CD8 ",
		DiffAdditionHex: "#98BB6C",
	})

	if actual.ActiveBorderHex != "#7E9CD8" {
		t.Fatalf("expected active border color %q, actual %q", "#7E9CD8", actual.ActiveBorderHex)
	}
	if actual.DiffAdditionHex != "#98BB6C" {
		t.Fatalf("expected addition foreground %q, actual %q", "#98BB6C", actual.DiffAdditionHex)
	}
	then_paletteUsesDarkDefaults(t, actual)
}

func TestResolvePaletteWithPreset_GivenDarkPreset_WhenResolving_ThenItUsesTheDarkPaletteWithoutForcingABackground(t *testing.T) {
	actual := ResolvePaletteWithPreset(DarkPresetName, Palette{})

	then_paletteUsesDarkDefaults(t, actual)
}

func TestResolvePaletteWithPreset_GivenBundledThemePreset_WhenResolving_ThenItUsesThePresetBackgroundColor(t *testing.T) {
	actual := ResolvePaletteWithPreset("kanagawa-dark", Palette{})

	if actual.BackgroundHex != "#1F1F28" {
		t.Fatalf("expected background color %q, actual %q", "#1F1F28", actual.BackgroundHex)
	}
	if actual.ActiveTextHex != "#DCD7BA" {
		t.Fatalf("expected active text color %q, actual %q", "#DCD7BA", actual.ActiveTextHex)
	}
	if actual.TeamOwnershipHex != "#727169" {
		t.Fatalf("expected team ownership color %q, actual %q", "#727169", actual.TeamOwnershipHex)
	}
}

func TestResolvePaletteWithPreset_GivenADarkBundledThemeWithoutPopupOverride_WhenResolving_ThenTheActionsPopupGroupForegroundTracksTheHeadingColor(t *testing.T) {
	actual := ResolvePaletteWithPreset("catppuccin-mocha", Palette{})

	if actual.ActionsPopupGroupForegroundHex != actual.MarkdownHeadingHex {
		t.Fatalf("expected actions popup group foreground %q to match the heading color, actual %q", actual.MarkdownHeadingHex, actual.ActionsPopupGroupForegroundHex)
	}
}

func TestResolvePalette_GivenGenericStatusColors_WhenResolving_ThenItCascadesToOpenDraftClosedAndDiffColors(t *testing.T) {
	actual := ResolvePalette(Palette{
		SuccessHex:           "#7FB069",
		SuccessBackgroundHex: "#D7E8D0",
		FailureHex:           "#E46876",
		FailureBackgroundHex: "#F3D4D9",
		PendingHex:           "#727169",
		PendingBackgroundHex: "#E6E3D8",
	})

	if actual.PullRequestStatusOpenHex != "#7FB069" {
		t.Fatalf("expected open foreground %q, actual %q", "#7FB069", actual.PullRequestStatusOpenHex)
	}
	if actual.PullRequestStatusOpenBackgroundHex != "#D7E8D0" {
		t.Fatalf("expected open background %q, actual %q", "#D7E8D0", actual.PullRequestStatusOpenBackgroundHex)
	}
	if actual.DiffAdditionHex != "#7FB069" {
		t.Fatalf("expected addition foreground %q, actual %q", "#7FB069", actual.DiffAdditionHex)
	}
	if actual.DiffAdditionBackgroundHex != "#D7E8D0" {
		t.Fatalf("expected addition background %q, actual %q", "#D7E8D0", actual.DiffAdditionBackgroundHex)
	}
	if actual.PullRequestStatusClosedHex != "#E46876" {
		t.Fatalf("expected closed foreground %q, actual %q", "#E46876", actual.PullRequestStatusClosedHex)
	}
	if actual.PullRequestStatusClosedBackgroundHex != "#F3D4D9" {
		t.Fatalf("expected closed background %q, actual %q", "#F3D4D9", actual.PullRequestStatusClosedBackgroundHex)
	}
	if actual.DiffDeletionHex != "#E46876" {
		t.Fatalf("expected deletion foreground %q, actual %q", "#E46876", actual.DiffDeletionHex)
	}
	if actual.DiffDeletionBackgroundHex != "#F3D4D9" {
		t.Fatalf("expected deletion background %q, actual %q", "#F3D4D9", actual.DiffDeletionBackgroundHex)
	}
	if actual.PullRequestStatusDraftHex != "#727169" {
		t.Fatalf("expected draft foreground %q, actual %q", "#727169", actual.PullRequestStatusDraftHex)
	}
	if actual.PullRequestStatusDraftBackgroundHex != "#E6E3D8" {
		t.Fatalf("expected draft background %q, actual %q", "#E6E3D8", actual.PullRequestStatusDraftBackgroundHex)
	}
}

func TestResolvePalette_GivenSpecificStatusOverrides_WhenResolving_ThenTheyWinOverGenericStatusColors(t *testing.T) {
	actual := ResolvePalette(Palette{
		SuccessHex:                "#7FB069",
		SuccessBackgroundHex:      "#D7E8D0",
		PullRequestStatusOpenHex:  "#101010",
		DiffAdditionBackgroundHex: "#CFE1C8",
	})

	if actual.PullRequestStatusOpenHex != "#101010" {
		t.Fatalf("expected open foreground %q, actual %q", "#101010", actual.PullRequestStatusOpenHex)
	}
	if actual.PullRequestStatusOpenBackgroundHex != "#D7E8D0" {
		t.Fatalf("expected open background %q, actual %q", "#D7E8D0", actual.PullRequestStatusOpenBackgroundHex)
	}
	if actual.DiffAdditionHex != "#7FB069" {
		t.Fatalf("expected addition foreground %q, actual %q", "#7FB069", actual.DiffAdditionHex)
	}
	if actual.DiffAdditionBackgroundHex != "#CFE1C8" {
		t.Fatalf("expected addition background %q, actual %q", "#CFE1C8", actual.DiffAdditionBackgroundHex)
	}
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
	if PullRequestStatusOpenHex != "#7FB069" || PullRequestStatusClosedHex != "#E46876" || PullRequestStatusDraftHex != "#727169" {
		t.Fatalf("expected derived status colors to be applied, actual open=%q closed=%q draft=%q", PullRequestStatusOpenHex, PullRequestStatusClosedHex, PullRequestStatusDraftHex)
	}
}

func TestResolvePalette_GivenActionsPopupGroupForegroundOverride_WhenResolving_ThenItKeepsTheOverride(t *testing.T) {
	actual := ResolvePalette(Palette{ActionsPopupGroupForegroundHex: "#224466"})

	if actual.ActionsPopupGroupForegroundHex != "#224466" {
		t.Fatalf("expected actions popup group foreground %q, actual %q", "#224466", actual.ActionsPopupGroupForegroundHex)
	}
}

func TestApplyPalette_GivenActionsPopupGroupForegroundOverride_WhenApplying_ThenItUpdatesThePackageColors(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{ActionsPopupGroupForegroundHex: "#224466"})

	if ActionsPopupGroupForegroundHex != "#224466" {
		t.Fatalf("expected actions popup group foreground %q, actual %q", "#224466", ActionsPopupGroupForegroundHex)
	}
}

func TestApplyPalette_GivenWarningOverride_WhenApplying_ThenItUpdatesThePackageColor(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{WarningHex: "#D97706"})

	if WarningHex != "#D97706" {
		t.Fatalf("expected warning color %q, actual %q", "#D97706", WarningHex)
	}
	if ActiveTextHex != DefaultPalette().ActiveTextHex {
		t.Fatalf("expected untouched active text color %q, actual %q", DefaultPalette().ActiveTextHex, ActiveTextHex)
	}
}

func TestApplyPalette_GivenTeamOwnershipOverride_WhenApplying_ThenItUpdatesThePackageColor(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{TeamOwnershipHex: "#54546D"})

	if TeamOwnershipHex != "#54546D" {
		t.Fatalf("expected team ownership color %q, actual %q", "#54546D", TeamOwnershipHex)
	}
	if ActiveTextHex != DefaultPalette().ActiveTextHex {
		t.Fatalf("expected untouched active text color %q, actual %q", DefaultPalette().ActiveTextHex, ActiveTextHex)
	}
}

func TestApplyPalette_GivenGenericStatusBackgroundOverrides_WhenApplying_ThenItUpdatesTheDerivedPackageColors(t *testing.T) {
	t.Cleanup(ResetPalette)

	ApplyPalette(Palette{
		SuccessBackgroundHex: "#D7E8D0",
		FailureBackgroundHex: "#F3D4D9",
		PendingBackgroundHex: "#E6E3D8",
	})

	if SuccessBackgroundHex != "#D7E8D0" || FailureBackgroundHex != "#F3D4D9" || PendingBackgroundHex != "#E6E3D8" {
		t.Fatalf("expected generic status backgrounds to be applied, actual success=%q failure=%q pending=%q", SuccessBackgroundHex, FailureBackgroundHex, PendingBackgroundHex)
	}
	if PullRequestStatusOpenBackgroundHex != "#D7E8D0" || DiffAdditionBackgroundHex != "#D7E8D0" {
		t.Fatalf("expected success background to cascade, actual open=%q addition=%q", PullRequestStatusOpenBackgroundHex, DiffAdditionBackgroundHex)
	}
	if PullRequestStatusClosedBackgroundHex != "#F3D4D9" || DiffDeletionBackgroundHex != "#F3D4D9" {
		t.Fatalf("expected failure background to cascade, actual closed=%q deletion=%q", PullRequestStatusClosedBackgroundHex, DiffDeletionBackgroundHex)
	}
	if PullRequestStatusDraftBackgroundHex != "#E6E3D8" {
		t.Fatalf("expected pending background to cascade, actual %q", PullRequestStatusDraftBackgroundHex)
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

	if actual.BackgroundHex != darkDefaultBackgroundHex {
		t.Fatalf("expected background color %q, actual %q", darkDefaultBackgroundHex, actual.BackgroundHex)
	}
	if actual.ActiveTextHex != darkDefaultActiveTextHex {
		t.Fatalf("expected active text color %q, actual %q", darkDefaultActiveTextHex, actual.ActiveTextHex)
	}
	if actual.SelectedLineBackgroundHex != darkDefaultSelectedLineBackgroundHex {
		t.Fatalf("expected selected line background %q, actual %q", darkDefaultSelectedLineBackgroundHex, actual.SelectedLineBackgroundHex)
	}
	if actual.ActionsPopupGroupForegroundHex != darkDefaultActionsPopupGroupForegroundHex {
		t.Fatalf("expected actions popup group foreground %q, actual %q", darkDefaultActionsPopupGroupForegroundHex, actual.ActionsPopupGroupForegroundHex)
	}
	if actual.WarningHex != darkDefaultWarningHex {
		t.Fatalf("expected warning color %q, actual %q", darkDefaultWarningHex, actual.WarningHex)
	}
	if actual.TeamOwnershipHex != darkDefaultTeamOwnershipHex {
		t.Fatalf("expected team ownership color %q, actual %q", darkDefaultTeamOwnershipHex, actual.TeamOwnershipHex)
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
	if actual.SuccessBackgroundHex != darkDefaultSuccessBackgroundHex || actual.FailureBackgroundHex != darkDefaultFailureBackgroundHex || actual.PendingBackgroundHex != darkDefaultPendingBackgroundHex {
		t.Fatalf("expected generic status backgrounds success=%q failure=%q pending=%q, actual success=%q failure=%q pending=%q", darkDefaultSuccessBackgroundHex, darkDefaultFailureBackgroundHex, darkDefaultPendingBackgroundHex, actual.SuccessBackgroundHex, actual.FailureBackgroundHex, actual.PendingBackgroundHex)
	}
	if actual.DiffAdditionBackgroundHex != darkDefaultDiffAdditionBackgroundHex {
		t.Fatalf("expected diff addition background %q, actual %q", darkDefaultDiffAdditionBackgroundHex, actual.DiffAdditionBackgroundHex)
	}
}

func then_paletteUsesLightDefaults(t *testing.T, actual Palette) {
	t.Helper()

	if actual.BackgroundHex != lightDefaultBackgroundHex {
		t.Fatalf("expected background color %q, actual %q", lightDefaultBackgroundHex, actual.BackgroundHex)
	}
	if actual.ActiveTextHex != lightDefaultActiveTextHex {
		t.Fatalf("expected active text color %q, actual %q", lightDefaultActiveTextHex, actual.ActiveTextHex)
	}
	if actual.SelectedLineBackgroundHex != lightDefaultSelectedLineBackgroundHex {
		t.Fatalf("expected selected line background %q, actual %q", lightDefaultSelectedLineBackgroundHex, actual.SelectedLineBackgroundHex)
	}
	if actual.ActionsPopupGroupForegroundHex != lightDefaultActionsPopupGroupForegroundHex {
		t.Fatalf("expected actions popup group foreground %q, actual %q", lightDefaultActionsPopupGroupForegroundHex, actual.ActionsPopupGroupForegroundHex)
	}
	if actual.WarningHex != lightDefaultWarningHex {
		t.Fatalf("expected warning color %q, actual %q", lightDefaultWarningHex, actual.WarningHex)
	}
	if actual.TeamOwnershipHex != lightDefaultTeamOwnershipHex {
		t.Fatalf("expected team ownership color %q, actual %q", lightDefaultTeamOwnershipHex, actual.TeamOwnershipHex)
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
	if actual.SuccessBackgroundHex != lightDefaultSuccessBackgroundHex || actual.FailureBackgroundHex != lightDefaultFailureBackgroundHex || actual.PendingBackgroundHex != lightDefaultPendingBackgroundHex {
		t.Fatalf("expected generic status backgrounds success=%q failure=%q pending=%q, actual success=%q failure=%q pending=%q", lightDefaultSuccessBackgroundHex, lightDefaultFailureBackgroundHex, lightDefaultPendingBackgroundHex, actual.SuccessBackgroundHex, actual.FailureBackgroundHex, actual.PendingBackgroundHex)
	}
}
