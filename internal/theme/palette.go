package theme

import "strings"

type Palette struct {
	BackgroundHex                        string `toml:"background"`
	ActiveBorderHex                      string `toml:"active_border"`
	InactiveBorderHex                    string `toml:"inactive_border"`
	ActiveTextHex                        string `toml:"active_text"`
	InactiveTextHex                      string `toml:"inactive_text"`
	InactiveTitleHex                     string `toml:"inactive_title"`
	SuccessHex                           string `toml:"success"`
	SuccessBackgroundHex                 string `toml:"success_background"`
	FailureHex                           string `toml:"failure"`
	FailureBackgroundHex                 string `toml:"failure_background"`
	PendingHex                           string `toml:"pending"`
	PendingBackgroundHex                 string `toml:"pending_background"`
	MutedHex                             string `toml:"muted"`
	WarningHex                           string `toml:"warning"`
	PullRequestReferenceHex              string `toml:"pull_request_reference"`
	PullRequestTitleHex                  string `toml:"pull_request_title"`
	SelectedLineBackgroundHex            string `toml:"selected_line_background"`
	ActionsPopupGroupForegroundHex       string `toml:"actions_popup_group_foreground"`
	SearchHighlightHex                   string `toml:"search_highlight"`
	MarkdownHeadingHex                   string `toml:"markdown_heading"`
	MarkdownHeadingBackgroundHex         string `toml:"markdown_heading_background"`
	MarkdownLinkHex                      string `toml:"markdown_link"`
	MarkdownCodeHex                      string `toml:"markdown_code"`
	SyntaxKeywordHex                     string `toml:"syntax_keyword"`
	SyntaxFunctionHex                    string `toml:"syntax_function"`
	SyntaxTypeHex                        string `toml:"syntax_type"`
	SyntaxPropertyHex                    string `toml:"syntax_property"`
	SyntaxStringHex                      string `toml:"syntax_string"`
	SyntaxNumberHex                      string `toml:"syntax_number"`
	SyntaxCommentHex                     string `toml:"syntax_comment"`
	CommentAuthorBadgeHex                string `toml:"comment_author_badge"`
	CommentAuthorBadgeBackgroundHex      string `toml:"comment_author_badge_background"`
	PullRequestStatusOpenHex             string `toml:"pull_request_status_open"`
	PullRequestStatusOpenBackgroundHex   string `toml:"pull_request_status_open_background"`
	PullRequestStatusDraftHex            string `toml:"pull_request_status_draft"`
	PullRequestStatusDraftBackgroundHex  string `toml:"pull_request_status_draft_background"`
	PullRequestStatusClosedHex           string `toml:"pull_request_status_closed"`
	PullRequestStatusClosedBackgroundHex string `toml:"pull_request_status_closed_background"`
	PullRequestStatusMergedHex           string `toml:"pull_request_status_merged"`
	PullRequestStatusMergedBackgroundHex string `toml:"pull_request_status_merged_background"`
	DiffAdditionHex                      string `toml:"diff_addition"`
	DiffAdditionBackgroundHex            string `toml:"diff_addition_background"`
	DiffAdditionHighlightBackgroundHex   string `toml:"diff_addition_highlight_background"`
	DiffDeletionHex                      string `toml:"diff_deletion"`
	DiffDeletionBackgroundHex            string `toml:"diff_deletion_background"`
	DiffDeletionHighlightBackgroundHex   string `toml:"diff_deletion_highlight_background"`
	DiffLineNumberHex                    string `toml:"diff_line_number"`
	DiffHunkHeaderHex                    string `toml:"diff_hunk_header"`
	TeamOwnershipHex                     string `toml:"team_ownership"`
}

var defaultLightPalette = newDefaultLightPalette()

func newDefaultLightPalette() Palette {
	activeTextHex := "#000000"
	mutedHex := "#636363"
	pendingHex := "#656D76"
	pendingBackgroundHex := "#E6E6E6"
	successHex := "#1A7F37"
	successBackgroundHex := "#DFF3E4"
	failureHex := "#CF222E"
	failureBackgroundHex := "#FFE2E5"
	warningHex := "#9A6700"
	searchHighlightHex := "#F9EAB3"
	commentAuthorBadgeHex := "#0969DA"
	commentAuthorBadgeBackgroundHex := "#DDF4FF"
	mergedHex := "#8250DF"
	mergedBackgroundHex := "#F5EDFF"
	return Palette{
		BackgroundHex:                        "",
		ActiveBorderHex:                      "#000000",
		InactiveBorderHex:                    "#CCCCCC",
		ActiveTextHex:                        activeTextHex,
		InactiveTextHex:                      "#000000",
		InactiveTitleHex:                     mutedHex,
		SuccessHex:                           successHex,
		SuccessBackgroundHex:                 successBackgroundHex,
		FailureHex:                           failureHex,
		FailureBackgroundHex:                 failureBackgroundHex,
		PendingHex:                           pendingHex,
		PendingBackgroundHex:                 pendingBackgroundHex,
		MutedHex:                             mutedHex,
		WarningHex:                           warningHex,
		PullRequestReferenceHex:              pendingHex,
		PullRequestTitleHex:                  activeTextHex,
		SelectedLineBackgroundHex:            pendingBackgroundHex,
		ActionsPopupGroupForegroundHex:       activeTextHex,
		SearchHighlightHex:                   searchHighlightHex,
		MarkdownHeadingHex:                   activeTextHex,
		MarkdownHeadingBackgroundHex:         searchHighlightHex,
		MarkdownLinkHex:                      "#000000",
		MarkdownCodeHex:                      "#B45309",
		SyntaxKeywordHex:                     "#CF222E",
		SyntaxFunctionHex:                    "#8250DF",
		SyntaxTypeHex:                        "#953800",
		SyntaxPropertyHex:                    "#0550AE",
		SyntaxStringHex:                      "#0A3069",
		SyntaxNumberHex:                      "#0550AE",
		SyntaxCommentHex:                     "#6E7781",
		CommentAuthorBadgeHex:                commentAuthorBadgeHex,
		CommentAuthorBadgeBackgroundHex:      commentAuthorBadgeBackgroundHex,
		PullRequestStatusOpenHex:             successHex,
		PullRequestStatusOpenBackgroundHex:   successBackgroundHex,
		PullRequestStatusDraftHex:            pendingHex,
		PullRequestStatusDraftBackgroundHex:  pendingBackgroundHex,
		PullRequestStatusClosedHex:           failureHex,
		PullRequestStatusClosedBackgroundHex: failureBackgroundHex,
		PullRequestStatusMergedHex:           mergedHex,
		PullRequestStatusMergedBackgroundHex: mergedBackgroundHex,
		DiffAdditionHex:                      successHex,
		DiffAdditionBackgroundHex:            successBackgroundHex,
		DiffAdditionHighlightBackgroundHex:   "#ACEEBB",
		DiffDeletionHex:                      failureHex,
		DiffDeletionBackgroundHex:            failureBackgroundHex,
		DiffDeletionHighlightBackgroundHex:   "#FFC1C8",
		DiffLineNumberHex:                    pendingHex,
		DiffHunkHeaderHex:                    pendingHex,
		TeamOwnershipHex:                     pendingHex,
	}
}

var defaultDarkPalette = newDefaultDarkPalette()

func newDefaultDarkPalette() Palette {
	activeTextHex := "#F0F6FC"
	mutedHex := "#8B949E"
	pendingHex := "#8B949E"
	pendingBackgroundHex := "#30363D"
	successHex := "#3FB950"
	successBackgroundHex := "#033A16"
	failureHex := "#F85149"
	failureBackgroundHex := "#67060C"
	warningHex := "#D29922"
	commentAuthorBadgeHex := "#DDF4FF"
	commentAuthorBadgeBackgroundHex := "#1F6FEB"
	mergedHex := "#A371F7"
	mergedBackgroundHex := "#3D2A5C"
	return Palette{
		BackgroundHex:                        "",
		ActiveBorderHex:                      activeTextHex,
		InactiveBorderHex:                    "#30363D",
		ActiveTextHex:                        activeTextHex,
		InactiveTextHex:                      "#E6EDF3",
		InactiveTitleHex:                     mutedHex,
		SuccessHex:                           successHex,
		SuccessBackgroundHex:                 successBackgroundHex,
		FailureHex:                           failureHex,
		FailureBackgroundHex:                 failureBackgroundHex,
		PendingHex:                           pendingHex,
		PendingBackgroundHex:                 pendingBackgroundHex,
		MutedHex:                             mutedHex,
		WarningHex:                           warningHex,
		PullRequestReferenceHex:              pendingHex,
		PullRequestTitleHex:                  activeTextHex,
		SelectedLineBackgroundHex:            "#21262D",
		ActionsPopupGroupForegroundHex:       activeTextHex,
		SearchHighlightHex:                   "#633C01",
		MarkdownHeadingHex:                   activeTextHex,
		MarkdownHeadingBackgroundHex:         "#58A6FF",
		MarkdownLinkHex:                      "#79C0FF",
		MarkdownCodeHex:                      "#FFA657",
		SyntaxKeywordHex:                     "#FF7B72",
		SyntaxFunctionHex:                    "#D2A8FF",
		SyntaxTypeHex:                        "#FFA657",
		SyntaxPropertyHex:                    "#79C0FF",
		SyntaxStringHex:                      "#A5D6FF",
		SyntaxNumberHex:                      "#79C0FF",
		SyntaxCommentHex:                     mutedHex,
		CommentAuthorBadgeHex:                commentAuthorBadgeHex,
		CommentAuthorBadgeBackgroundHex:      commentAuthorBadgeBackgroundHex,
		PullRequestStatusOpenHex:             successHex,
		PullRequestStatusOpenBackgroundHex:   successBackgroundHex,
		PullRequestStatusDraftHex:            pendingHex,
		PullRequestStatusDraftBackgroundHex:  pendingBackgroundHex,
		PullRequestStatusClosedHex:           failureHex,
		PullRequestStatusClosedBackgroundHex: failureBackgroundHex,
		PullRequestStatusMergedHex:           mergedHex,
		PullRequestStatusMergedBackgroundHex: mergedBackgroundHex,
		DiffAdditionHex:                      successHex,
		DiffAdditionBackgroundHex:            successBackgroundHex,
		DiffAdditionHighlightBackgroundHex:   "#0F5323",
		DiffDeletionHex:                      failureHex,
		DiffDeletionBackgroundHex:            failureBackgroundHex,
		DiffDeletionHighlightBackgroundHex:   "#8E1519",
		DiffLineNumberHex:                    pendingHex,
		DiffHunkHeaderHex:                    pendingHex,
		TeamOwnershipHex:                     pendingHex,
	}
}

var systemPolarityDetector = detectSystemPolarity

var initialDefaultPalette = defaultPaletteForPolarity(systemPolarityDetector())

var (
	BackgroundHex                        = initialDefaultPalette.BackgroundHex
	ActiveBorderHex                      = initialDefaultPalette.ActiveBorderHex
	InactiveBorderHex                    = initialDefaultPalette.InactiveBorderHex
	ActiveTextHex                        = initialDefaultPalette.ActiveTextHex
	InactiveTextHex                      = initialDefaultPalette.InactiveTextHex
	InactiveTitleHex                     = initialDefaultPalette.InactiveTitleHex
	SuccessHex                           = initialDefaultPalette.SuccessHex
	SuccessBackgroundHex                 = initialDefaultPalette.SuccessBackgroundHex
	FailureHex                           = initialDefaultPalette.FailureHex
	FailureBackgroundHex                 = initialDefaultPalette.FailureBackgroundHex
	PendingHex                           = initialDefaultPalette.PendingHex
	PendingBackgroundHex                 = initialDefaultPalette.PendingBackgroundHex
	MutedHex                             = initialDefaultPalette.MutedHex
	WarningHex                           = initialDefaultPalette.WarningHex
	PullRequestReferenceHex              = initialDefaultPalette.PullRequestReferenceHex
	PullRequestTitleHex                  = initialDefaultPalette.PullRequestTitleHex
	SelectedLineBackgroundHex            = initialDefaultPalette.SelectedLineBackgroundHex
	ActionsPopupGroupForegroundHex       = initialDefaultPalette.ActionsPopupGroupForegroundHex
	SearchHighlightHex                   = initialDefaultPalette.SearchHighlightHex
	MarkdownHeadingHex                   = initialDefaultPalette.MarkdownHeadingHex
	MarkdownHeadingBackgroundHex         = initialDefaultPalette.MarkdownHeadingBackgroundHex
	MarkdownLinkHex                      = initialDefaultPalette.MarkdownLinkHex
	MarkdownCodeHex                      = initialDefaultPalette.MarkdownCodeHex
	SyntaxKeywordHex                     = initialDefaultPalette.SyntaxKeywordHex
	SyntaxFunctionHex                    = initialDefaultPalette.SyntaxFunctionHex
	SyntaxTypeHex                        = initialDefaultPalette.SyntaxTypeHex
	SyntaxPropertyHex                    = initialDefaultPalette.SyntaxPropertyHex
	SyntaxStringHex                      = initialDefaultPalette.SyntaxStringHex
	SyntaxNumberHex                      = initialDefaultPalette.SyntaxNumberHex
	SyntaxCommentHex                     = initialDefaultPalette.SyntaxCommentHex
	CommentAuthorBadgeHex                = initialDefaultPalette.CommentAuthorBadgeHex
	CommentAuthorBadgeBackgroundHex      = initialDefaultPalette.CommentAuthorBadgeBackgroundHex
	PullRequestStatusOpenHex             = initialDefaultPalette.PullRequestStatusOpenHex
	PullRequestStatusOpenBackgroundHex   = initialDefaultPalette.PullRequestStatusOpenBackgroundHex
	PullRequestStatusDraftHex            = initialDefaultPalette.PullRequestStatusDraftHex
	PullRequestStatusDraftBackgroundHex  = initialDefaultPalette.PullRequestStatusDraftBackgroundHex
	PullRequestStatusClosedHex           = initialDefaultPalette.PullRequestStatusClosedHex
	PullRequestStatusClosedBackgroundHex = initialDefaultPalette.PullRequestStatusClosedBackgroundHex
	PullRequestStatusMergedHex           = initialDefaultPalette.PullRequestStatusMergedHex
	PullRequestStatusMergedBackgroundHex = initialDefaultPalette.PullRequestStatusMergedBackgroundHex
	DiffAdditionHex                      = initialDefaultPalette.DiffAdditionHex
	DiffAdditionBackgroundHex            = initialDefaultPalette.DiffAdditionBackgroundHex
	DiffAdditionHighlightBackgroundHex   = initialDefaultPalette.DiffAdditionHighlightBackgroundHex
	DiffDeletionHex                      = initialDefaultPalette.DiffDeletionHex
	DiffDeletionBackgroundHex            = initialDefaultPalette.DiffDeletionBackgroundHex
	DiffDeletionHighlightBackgroundHex   = initialDefaultPalette.DiffDeletionHighlightBackgroundHex
	DiffLineNumberHex                    = initialDefaultPalette.DiffLineNumberHex
	DiffHunkHeaderHex                    = initialDefaultPalette.DiffHunkHeaderHex
	TeamOwnershipHex                     = initialDefaultPalette.TeamOwnershipHex
)

func DefaultPalette() Palette {
	return defaultPaletteForPolarity(systemPolarityDetector())
}

func defaultPaletteForPolarity(polarity systemPolarity) Palette {
	if polarity == systemPolarityDark {
		return defaultDarkPalette
	}

	return defaultLightPalette
}

func NormalizePalette(overrides Palette) Palette {
	return normalizePalette(overrides)
}

func ResolvePalette(overrides Palette) Palette {
	return ResolvePaletteWithPreset(SystemPresetName, overrides)
}

func cascadePaletteColors(resolved *Palette, overrides Palette) {
	inheritColor(&resolved.ActionsPopupGroupForegroundHex, overrides.ActionsPopupGroupForegroundHex, resolved.MarkdownHeadingHex, overrides.MarkdownHeadingHex)
	inheritColor(&resolved.PullRequestStatusOpenHex, overrides.PullRequestStatusOpenHex, resolved.SuccessHex, overrides.SuccessHex)
	inheritColor(&resolved.PullRequestStatusOpenBackgroundHex, overrides.PullRequestStatusOpenBackgroundHex, resolved.SuccessBackgroundHex, overrides.SuccessBackgroundHex)
	inheritColor(&resolved.PullRequestStatusDraftHex, overrides.PullRequestStatusDraftHex, resolved.PendingHex, overrides.PendingHex)
	inheritColor(&resolved.PullRequestStatusDraftBackgroundHex, overrides.PullRequestStatusDraftBackgroundHex, resolved.PendingBackgroundHex, overrides.PendingBackgroundHex)
	inheritColor(&resolved.PullRequestStatusClosedHex, overrides.PullRequestStatusClosedHex, resolved.FailureHex, overrides.FailureHex)
	inheritColor(&resolved.PullRequestStatusClosedBackgroundHex, overrides.PullRequestStatusClosedBackgroundHex, resolved.FailureBackgroundHex, overrides.FailureBackgroundHex)
	inheritColor(&resolved.DiffAdditionHex, overrides.DiffAdditionHex, resolved.SuccessHex, overrides.SuccessHex)
	inheritColor(&resolved.DiffAdditionBackgroundHex, overrides.DiffAdditionBackgroundHex, resolved.SuccessBackgroundHex, overrides.SuccessBackgroundHex)
	inheritColor(&resolved.DiffDeletionHex, overrides.DiffDeletionHex, resolved.FailureHex, overrides.FailureHex)
	inheritColor(&resolved.DiffDeletionBackgroundHex, overrides.DiffDeletionBackgroundHex, resolved.FailureBackgroundHex, overrides.FailureBackgroundHex)
}

func inheritColor(target *string, targetOverride string, source string, sourceOverride string) {
	if targetOverride != "" || sourceOverride == "" {
		return
	}
	*target = source
}

func mergePalette(base Palette, overrides Palette) Palette {
	resolved := base
	visitPaletteColorPairs(&resolved, &overrides, func(resolvedColorValue *string, overrideColorValue *string) {
		*resolvedColorValue = resolvedColor(*resolvedColorValue, *overrideColorValue)
	})
	return resolved
}

func ApplyPalette(overrides Palette) {
	applyResolvedPalette(ResolvePalette(overrides))
}

func ResetPalette() {
	applyResolvedPalette(DefaultPalette())
}

func applyResolvedPalette(palette Palette) {
	BackgroundHex = palette.BackgroundHex
	ActiveBorderHex = palette.ActiveBorderHex
	InactiveBorderHex = palette.InactiveBorderHex
	ActiveTextHex = palette.ActiveTextHex
	InactiveTextHex = palette.InactiveTextHex
	InactiveTitleHex = palette.InactiveTitleHex
	SuccessHex = palette.SuccessHex
	SuccessBackgroundHex = palette.SuccessBackgroundHex
	FailureHex = palette.FailureHex
	FailureBackgroundHex = palette.FailureBackgroundHex
	PendingHex = palette.PendingHex
	PendingBackgroundHex = palette.PendingBackgroundHex
	MutedHex = palette.MutedHex
	WarningHex = palette.WarningHex
	PullRequestReferenceHex = palette.PullRequestReferenceHex
	PullRequestTitleHex = palette.PullRequestTitleHex
	SelectedLineBackgroundHex = palette.SelectedLineBackgroundHex
	ActionsPopupGroupForegroundHex = palette.ActionsPopupGroupForegroundHex
	SearchHighlightHex = palette.SearchHighlightHex
	MarkdownHeadingHex = palette.MarkdownHeadingHex
	MarkdownHeadingBackgroundHex = palette.MarkdownHeadingBackgroundHex
	MarkdownLinkHex = palette.MarkdownLinkHex
	MarkdownCodeHex = palette.MarkdownCodeHex
	SyntaxKeywordHex = palette.SyntaxKeywordHex
	SyntaxFunctionHex = palette.SyntaxFunctionHex
	SyntaxTypeHex = palette.SyntaxTypeHex
	SyntaxPropertyHex = palette.SyntaxPropertyHex
	SyntaxStringHex = palette.SyntaxStringHex
	SyntaxNumberHex = palette.SyntaxNumberHex
	SyntaxCommentHex = palette.SyntaxCommentHex
	CommentAuthorBadgeHex = palette.CommentAuthorBadgeHex
	CommentAuthorBadgeBackgroundHex = palette.CommentAuthorBadgeBackgroundHex
	PullRequestStatusOpenHex = palette.PullRequestStatusOpenHex
	PullRequestStatusOpenBackgroundHex = palette.PullRequestStatusOpenBackgroundHex
	PullRequestStatusDraftHex = palette.PullRequestStatusDraftHex
	PullRequestStatusDraftBackgroundHex = palette.PullRequestStatusDraftBackgroundHex
	PullRequestStatusClosedHex = palette.PullRequestStatusClosedHex
	PullRequestStatusClosedBackgroundHex = palette.PullRequestStatusClosedBackgroundHex
	PullRequestStatusMergedHex = palette.PullRequestStatusMergedHex
	PullRequestStatusMergedBackgroundHex = palette.PullRequestStatusMergedBackgroundHex
	DiffAdditionHex = palette.DiffAdditionHex
	DiffAdditionBackgroundHex = palette.DiffAdditionBackgroundHex
	DiffAdditionHighlightBackgroundHex = palette.DiffAdditionHighlightBackgroundHex
	DiffDeletionHex = palette.DiffDeletionHex
	DiffDeletionBackgroundHex = palette.DiffDeletionBackgroundHex
	DiffDeletionHighlightBackgroundHex = palette.DiffDeletionHighlightBackgroundHex
	DiffLineNumberHex = palette.DiffLineNumberHex
	DiffHunkHeaderHex = palette.DiffHunkHeaderHex
	TeamOwnershipHex = palette.TeamOwnershipHex
}

func normalizePalette(overrides Palette) Palette {
	normalized := overrides
	visitPaletteColors(&normalized, func(color *string) {
		*color = normalizeHexColor(*color)
	})
	return normalized
}

func visitPaletteColors(palette *Palette, visit func(color *string)) {
	for _, color := range paletteColorPointers(palette) {
		visit(color)
	}
}

func visitPaletteColorPairs(left *Palette, right *Palette, visit func(leftColor *string, rightColor *string)) {
	leftColors := paletteColorPointers(left)
	rightColors := paletteColorPointers(right)
	for index := range leftColors {
		visit(leftColors[index], rightColors[index])
	}
}

func paletteColorPointers(palette *Palette) []*string {
	return []*string{
		&palette.BackgroundHex,
		&palette.ActiveBorderHex,
		&palette.InactiveBorderHex,
		&palette.ActiveTextHex,
		&palette.InactiveTextHex,
		&palette.InactiveTitleHex,
		&palette.SuccessHex,
		&palette.SuccessBackgroundHex,
		&palette.FailureHex,
		&palette.FailureBackgroundHex,
		&palette.PendingHex,
		&palette.PendingBackgroundHex,
		&palette.MutedHex,
		&palette.WarningHex,
		&palette.PullRequestReferenceHex,
		&palette.PullRequestTitleHex,
		&palette.SelectedLineBackgroundHex,
		&palette.ActionsPopupGroupForegroundHex,
		&palette.SearchHighlightHex,
		&palette.MarkdownHeadingHex,
		&palette.MarkdownHeadingBackgroundHex,
		&palette.MarkdownLinkHex,
		&palette.MarkdownCodeHex,
		&palette.SyntaxKeywordHex,
		&palette.SyntaxFunctionHex,
		&palette.SyntaxTypeHex,
		&palette.SyntaxPropertyHex,
		&palette.SyntaxStringHex,
		&palette.SyntaxNumberHex,
		&palette.SyntaxCommentHex,
		&palette.CommentAuthorBadgeHex,
		&palette.CommentAuthorBadgeBackgroundHex,
		&palette.PullRequestStatusOpenHex,
		&palette.PullRequestStatusOpenBackgroundHex,
		&palette.PullRequestStatusDraftHex,
		&palette.PullRequestStatusDraftBackgroundHex,
		&palette.PullRequestStatusClosedHex,
		&palette.PullRequestStatusClosedBackgroundHex,
		&palette.PullRequestStatusMergedHex,
		&palette.PullRequestStatusMergedBackgroundHex,
		&palette.DiffAdditionHex,
		&palette.DiffAdditionBackgroundHex,
		&palette.DiffAdditionHighlightBackgroundHex,
		&palette.DiffDeletionHex,
		&palette.DiffDeletionBackgroundHex,
		&palette.DiffDeletionHighlightBackgroundHex,
		&palette.DiffLineNumberHex,
		&palette.DiffHunkHeaderHex,
		&palette.TeamOwnershipHex,
	}
}

func normalizeHexColor(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if !isHexColor(trimmedValue) {
		return ""
	}

	return trimmedValue
}

func isHexColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}

	for _, runeValue := range value[1:] {
		switch {
		case runeValue >= '0' && runeValue <= '9':
		case runeValue >= 'a' && runeValue <= 'f':
		case runeValue >= 'A' && runeValue <= 'F':
		default:
			return false
		}
	}

	return true
}

func resolvedColor(defaultValue string, overrideValue string) string {
	if overrideValue == "" {
		return defaultValue
	}

	return overrideValue
}
