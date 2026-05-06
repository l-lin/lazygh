package theme

import "strings"

type Palette struct {
	ActiveBorderHex                      string `toml:"active_border"`
	InactiveBorderHex                    string `toml:"inactive_border"`
	ActiveTextHex                        string `toml:"active_text"`
	InactiveTextHex                      string `toml:"inactive_text"`
	InactiveTitleHex                     string `toml:"inactive_title"`
	SelectedLineBackgroundHex            string `toml:"selected_line_background"`
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
	CommentAuthorBadgeForegroundHex      string `toml:"comment_author_badge_foreground"`
	CommentAuthorBadgeBackgroundHex      string `toml:"comment_author_badge_background"`
	PullRequestStatusOpenForegroundHex   string `toml:"pull_request_status_open_foreground"`
	PullRequestStatusOpenBackgroundHex   string `toml:"pull_request_status_open_background"`
	PullRequestStatusDraftForegroundHex  string `toml:"pull_request_status_draft_foreground"`
	PullRequestStatusDraftBackgroundHex  string `toml:"pull_request_status_draft_background"`
	PullRequestStatusClosedForegroundHex string `toml:"pull_request_status_closed_foreground"`
	PullRequestStatusClosedBackgroundHex string `toml:"pull_request_status_closed_background"`
	PullRequestStatusMergedForegroundHex string `toml:"pull_request_status_merged_foreground"`
	PullRequestStatusMergedBackgroundHex string `toml:"pull_request_status_merged_background"`
	DiffAdditionForegroundHex            string `toml:"diff_addition_foreground"`
	DiffAdditionBackgroundHex            string `toml:"diff_addition_background"`
	DiffAdditionHighlightBackgroundHex   string `toml:"diff_addition_highlight_background"`
	DiffDeletionForegroundHex            string `toml:"diff_deletion_foreground"`
	DiffDeletionBackgroundHex            string `toml:"diff_deletion_background"`
	DiffDeletionHighlightBackgroundHex   string `toml:"diff_deletion_highlight_background"`
	DiffLineNumberHex                    string `toml:"diff_line_number"`
	DiffHunkHeaderHex                    string `toml:"diff_hunk_header"`
}

var defaultLightPalette = Palette{
	ActiveBorderHex:                      "#000000",
	InactiveBorderHex:                    "#CCCCCC",
	ActiveTextHex:                        "#000000",
	InactiveTextHex:                      "#000000",
	InactiveTitleHex:                     "#636363",
	SelectedLineBackgroundHex:            "#E6E6E6",
	SearchHighlightHex:                   "#F9EAB3",
	MarkdownHeadingHex:                   "#000000",
	MarkdownHeadingBackgroundHex:         "#F9EAB3",
	MarkdownLinkHex:                      "#000000",
	MarkdownCodeHex:                      "#B45309",
	SyntaxKeywordHex:                     "#CF222E",
	SyntaxFunctionHex:                    "#8250DF",
	SyntaxTypeHex:                        "#953800",
	SyntaxPropertyHex:                    "#0550AE",
	SyntaxStringHex:                      "#0A3069",
	SyntaxNumberHex:                      "#0550AE",
	SyntaxCommentHex:                     "#6E7781",
	CommentAuthorBadgeForegroundHex:      "#0969DA",
	CommentAuthorBadgeBackgroundHex:      "#DDF4FF",
	PullRequestStatusOpenForegroundHex:   "#1A7F37",
	PullRequestStatusOpenBackgroundHex:   "#DFF3E4",
	PullRequestStatusDraftForegroundHex:  "#656D76",
	PullRequestStatusDraftBackgroundHex:  "#E6E6E6",
	PullRequestStatusClosedForegroundHex: "#CF222E",
	PullRequestStatusClosedBackgroundHex: "#FFE2E5",
	PullRequestStatusMergedForegroundHex: "#8250DF",
	PullRequestStatusMergedBackgroundHex: "#F5EDFF",
	DiffAdditionForegroundHex:            "#1A7F37",
	DiffAdditionBackgroundHex:            "#DFF3E4",
	DiffAdditionHighlightBackgroundHex:   "#ACEEBB",
	DiffDeletionForegroundHex:            "#CF222E",
	DiffDeletionBackgroundHex:            "#FFE2E5",
	DiffDeletionHighlightBackgroundHex:   "#FFC1C8",
	DiffLineNumberHex:                    "#656D76",
	DiffHunkHeaderHex:                    "#656D76",
}

var defaultDarkPalette = Palette{
	ActiveBorderHex:                      "#F0F6FC",
	InactiveBorderHex:                    "#30363D",
	ActiveTextHex:                        "#F0F6FC",
	InactiveTextHex:                      "#E6EDF3",
	InactiveTitleHex:                     "#8B949E",
	SelectedLineBackgroundHex:            "#21262D",
	SearchHighlightHex:                   "#633C01",
	MarkdownHeadingHex:                   "#F0F6FC",
	MarkdownHeadingBackgroundHex:         "#58A6FF",
	MarkdownLinkHex:                      "#79C0FF",
	MarkdownCodeHex:                      "#FFA657",
	SyntaxKeywordHex:                     "#FF7B72",
	SyntaxFunctionHex:                    "#D2A8FF",
	SyntaxTypeHex:                        "#FFA657",
	SyntaxPropertyHex:                    "#79C0FF",
	SyntaxStringHex:                      "#A5D6FF",
	SyntaxNumberHex:                      "#79C0FF",
	SyntaxCommentHex:                     "#8B949E",
	CommentAuthorBadgeForegroundHex:      "#DDF4FF",
	CommentAuthorBadgeBackgroundHex:      "#1F6FEB",
	PullRequestStatusOpenForegroundHex:   "#3FB950",
	PullRequestStatusOpenBackgroundHex:   "#033A16",
	PullRequestStatusDraftForegroundHex:  "#8B949E",
	PullRequestStatusDraftBackgroundHex:  "#30363D",
	PullRequestStatusClosedForegroundHex: "#F85149",
	PullRequestStatusClosedBackgroundHex: "#67060C",
	PullRequestStatusMergedForegroundHex: "#A371F7",
	PullRequestStatusMergedBackgroundHex: "#3D2A5C",
	DiffAdditionForegroundHex:            "#3FB950",
	DiffAdditionBackgroundHex:            "#033A16",
	DiffAdditionHighlightBackgroundHex:   "#0F5323",
	DiffDeletionForegroundHex:            "#F85149",
	DiffDeletionBackgroundHex:            "#67060C",
	DiffDeletionHighlightBackgroundHex:   "#8E1519",
	DiffLineNumberHex:                    "#8B949E",
	DiffHunkHeaderHex:                    "#8B949E",
}

var systemPolarityDetector = detectSystemPolarity

var initialDefaultPalette = defaultPaletteForPolarity(systemPolarityDetector())

var (
	ActiveBorderHex                      = initialDefaultPalette.ActiveBorderHex
	InactiveBorderHex                    = initialDefaultPalette.InactiveBorderHex
	ActiveTextHex                        = initialDefaultPalette.ActiveTextHex
	InactiveTextHex                      = initialDefaultPalette.InactiveTextHex
	InactiveTitleHex                     = initialDefaultPalette.InactiveTitleHex
	SelectedLineBackgroundHex            = initialDefaultPalette.SelectedLineBackgroundHex
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
	CommentAuthorBadgeForegroundHex      = initialDefaultPalette.CommentAuthorBadgeForegroundHex
	CommentAuthorBadgeBackgroundHex      = initialDefaultPalette.CommentAuthorBadgeBackgroundHex
	PullRequestStatusOpenForegroundHex   = initialDefaultPalette.PullRequestStatusOpenForegroundHex
	PullRequestStatusOpenBackgroundHex   = initialDefaultPalette.PullRequestStatusOpenBackgroundHex
	PullRequestStatusDraftForegroundHex  = initialDefaultPalette.PullRequestStatusDraftForegroundHex
	PullRequestStatusDraftBackgroundHex  = initialDefaultPalette.PullRequestStatusDraftBackgroundHex
	PullRequestStatusClosedForegroundHex = initialDefaultPalette.PullRequestStatusClosedForegroundHex
	PullRequestStatusClosedBackgroundHex = initialDefaultPalette.PullRequestStatusClosedBackgroundHex
	PullRequestStatusMergedForegroundHex = initialDefaultPalette.PullRequestStatusMergedForegroundHex
	PullRequestStatusMergedBackgroundHex = initialDefaultPalette.PullRequestStatusMergedBackgroundHex
	DiffAdditionForegroundHex            = initialDefaultPalette.DiffAdditionForegroundHex
	DiffAdditionBackgroundHex            = initialDefaultPalette.DiffAdditionBackgroundHex
	DiffAdditionHighlightBackgroundHex   = initialDefaultPalette.DiffAdditionHighlightBackgroundHex
	DiffDeletionForegroundHex            = initialDefaultPalette.DiffDeletionForegroundHex
	DiffDeletionBackgroundHex            = initialDefaultPalette.DiffDeletionBackgroundHex
	DiffDeletionHighlightBackgroundHex   = initialDefaultPalette.DiffDeletionHighlightBackgroundHex
	DiffLineNumberHex                    = initialDefaultPalette.DiffLineNumberHex
	DiffHunkHeaderHex                    = initialDefaultPalette.DiffHunkHeaderHex
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
	normalized := NormalizePalette(overrides)
	resolved := DefaultPalette()
	resolved.ActiveBorderHex = resolvedColor(resolved.ActiveBorderHex, normalized.ActiveBorderHex)
	resolved.InactiveBorderHex = resolvedColor(resolved.InactiveBorderHex, normalized.InactiveBorderHex)
	resolved.ActiveTextHex = resolvedColor(resolved.ActiveTextHex, normalized.ActiveTextHex)
	resolved.InactiveTextHex = resolvedColor(resolved.InactiveTextHex, normalized.InactiveTextHex)
	resolved.InactiveTitleHex = resolvedColor(resolved.InactiveTitleHex, normalized.InactiveTitleHex)
	resolved.SelectedLineBackgroundHex = resolvedColor(resolved.SelectedLineBackgroundHex, normalized.SelectedLineBackgroundHex)
	resolved.SearchHighlightHex = resolvedColor(resolved.SearchHighlightHex, normalized.SearchHighlightHex)
	resolved.MarkdownHeadingHex = resolvedColor(resolved.MarkdownHeadingHex, normalized.MarkdownHeadingHex)
	resolved.MarkdownHeadingBackgroundHex = resolvedColor(resolved.MarkdownHeadingBackgroundHex, normalized.MarkdownHeadingBackgroundHex)
	resolved.MarkdownLinkHex = resolvedColor(resolved.MarkdownLinkHex, normalized.MarkdownLinkHex)
	resolved.MarkdownCodeHex = resolvedColor(resolved.MarkdownCodeHex, normalized.MarkdownCodeHex)
	resolved.SyntaxKeywordHex = resolvedColor(resolved.SyntaxKeywordHex, normalized.SyntaxKeywordHex)
	resolved.SyntaxFunctionHex = resolvedColor(resolved.SyntaxFunctionHex, normalized.SyntaxFunctionHex)
	resolved.SyntaxTypeHex = resolvedColor(resolved.SyntaxTypeHex, normalized.SyntaxTypeHex)
	resolved.SyntaxPropertyHex = resolvedColor(resolved.SyntaxPropertyHex, normalized.SyntaxPropertyHex)
	resolved.SyntaxStringHex = resolvedColor(resolved.SyntaxStringHex, normalized.SyntaxStringHex)
	resolved.SyntaxNumberHex = resolvedColor(resolved.SyntaxNumberHex, normalized.SyntaxNumberHex)
	resolved.SyntaxCommentHex = resolvedColor(resolved.SyntaxCommentHex, normalized.SyntaxCommentHex)
	resolved.CommentAuthorBadgeForegroundHex = resolvedColor(resolved.CommentAuthorBadgeForegroundHex, normalized.CommentAuthorBadgeForegroundHex)
	resolved.CommentAuthorBadgeBackgroundHex = resolvedColor(resolved.CommentAuthorBadgeBackgroundHex, normalized.CommentAuthorBadgeBackgroundHex)
	resolved.PullRequestStatusOpenForegroundHex = resolvedColor(resolved.PullRequestStatusOpenForegroundHex, normalized.PullRequestStatusOpenForegroundHex)
	resolved.PullRequestStatusOpenBackgroundHex = resolvedColor(resolved.PullRequestStatusOpenBackgroundHex, normalized.PullRequestStatusOpenBackgroundHex)
	resolved.PullRequestStatusDraftForegroundHex = resolvedColor(resolved.PullRequestStatusDraftForegroundHex, normalized.PullRequestStatusDraftForegroundHex)
	resolved.PullRequestStatusDraftBackgroundHex = resolvedColor(resolved.PullRequestStatusDraftBackgroundHex, normalized.PullRequestStatusDraftBackgroundHex)
	resolved.PullRequestStatusClosedForegroundHex = resolvedColor(resolved.PullRequestStatusClosedForegroundHex, normalized.PullRequestStatusClosedForegroundHex)
	resolved.PullRequestStatusClosedBackgroundHex = resolvedColor(resolved.PullRequestStatusClosedBackgroundHex, normalized.PullRequestStatusClosedBackgroundHex)
	resolved.PullRequestStatusMergedForegroundHex = resolvedColor(resolved.PullRequestStatusMergedForegroundHex, normalized.PullRequestStatusMergedForegroundHex)
	resolved.PullRequestStatusMergedBackgroundHex = resolvedColor(resolved.PullRequestStatusMergedBackgroundHex, normalized.PullRequestStatusMergedBackgroundHex)
	resolved.DiffAdditionForegroundHex = resolvedColor(resolved.DiffAdditionForegroundHex, normalized.DiffAdditionForegroundHex)
	resolved.DiffAdditionBackgroundHex = resolvedColor(resolved.DiffAdditionBackgroundHex, normalized.DiffAdditionBackgroundHex)
	resolved.DiffAdditionHighlightBackgroundHex = resolvedColor(resolved.DiffAdditionHighlightBackgroundHex, normalized.DiffAdditionHighlightBackgroundHex)
	resolved.DiffDeletionForegroundHex = resolvedColor(resolved.DiffDeletionForegroundHex, normalized.DiffDeletionForegroundHex)
	resolved.DiffDeletionBackgroundHex = resolvedColor(resolved.DiffDeletionBackgroundHex, normalized.DiffDeletionBackgroundHex)
	resolved.DiffDeletionHighlightBackgroundHex = resolvedColor(resolved.DiffDeletionHighlightBackgroundHex, normalized.DiffDeletionHighlightBackgroundHex)
	resolved.DiffLineNumberHex = resolvedColor(resolved.DiffLineNumberHex, normalized.DiffLineNumberHex)
	resolved.DiffHunkHeaderHex = resolvedColor(resolved.DiffHunkHeaderHex, normalized.DiffHunkHeaderHex)
	return resolved
}

func ApplyPalette(overrides Palette) {
	applyResolvedPalette(ResolvePalette(overrides))
}

func ResetPalette() {
	applyResolvedPalette(DefaultPalette())
}

func applyResolvedPalette(palette Palette) {
	ActiveBorderHex = palette.ActiveBorderHex
	InactiveBorderHex = palette.InactiveBorderHex
	ActiveTextHex = palette.ActiveTextHex
	InactiveTextHex = palette.InactiveTextHex
	InactiveTitleHex = palette.InactiveTitleHex
	SelectedLineBackgroundHex = palette.SelectedLineBackgroundHex
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
	CommentAuthorBadgeForegroundHex = palette.CommentAuthorBadgeForegroundHex
	CommentAuthorBadgeBackgroundHex = palette.CommentAuthorBadgeBackgroundHex
	PullRequestStatusOpenForegroundHex = palette.PullRequestStatusOpenForegroundHex
	PullRequestStatusOpenBackgroundHex = palette.PullRequestStatusOpenBackgroundHex
	PullRequestStatusDraftForegroundHex = palette.PullRequestStatusDraftForegroundHex
	PullRequestStatusDraftBackgroundHex = palette.PullRequestStatusDraftBackgroundHex
	PullRequestStatusClosedForegroundHex = palette.PullRequestStatusClosedForegroundHex
	PullRequestStatusClosedBackgroundHex = palette.PullRequestStatusClosedBackgroundHex
	PullRequestStatusMergedForegroundHex = palette.PullRequestStatusMergedForegroundHex
	PullRequestStatusMergedBackgroundHex = palette.PullRequestStatusMergedBackgroundHex
	DiffAdditionForegroundHex = palette.DiffAdditionForegroundHex
	DiffAdditionBackgroundHex = palette.DiffAdditionBackgroundHex
	DiffAdditionHighlightBackgroundHex = palette.DiffAdditionHighlightBackgroundHex
	DiffDeletionForegroundHex = palette.DiffDeletionForegroundHex
	DiffDeletionBackgroundHex = palette.DiffDeletionBackgroundHex
	DiffDeletionHighlightBackgroundHex = palette.DiffDeletionHighlightBackgroundHex
	DiffLineNumberHex = palette.DiffLineNumberHex
	DiffHunkHeaderHex = palette.DiffHunkHeaderHex
}

func normalizePalette(overrides Palette) Palette {
	return Palette{
		ActiveBorderHex:                      normalizeHexColor(overrides.ActiveBorderHex),
		InactiveBorderHex:                    normalizeHexColor(overrides.InactiveBorderHex),
		ActiveTextHex:                        normalizeHexColor(overrides.ActiveTextHex),
		InactiveTextHex:                      normalizeHexColor(overrides.InactiveTextHex),
		InactiveTitleHex:                     normalizeHexColor(overrides.InactiveTitleHex),
		SelectedLineBackgroundHex:            normalizeHexColor(overrides.SelectedLineBackgroundHex),
		SearchHighlightHex:                   normalizeHexColor(overrides.SearchHighlightHex),
		MarkdownHeadingHex:                   normalizeHexColor(overrides.MarkdownHeadingHex),
		MarkdownHeadingBackgroundHex:         normalizeHexColor(overrides.MarkdownHeadingBackgroundHex),
		MarkdownLinkHex:                      normalizeHexColor(overrides.MarkdownLinkHex),
		MarkdownCodeHex:                      normalizeHexColor(overrides.MarkdownCodeHex),
		SyntaxKeywordHex:                     normalizeHexColor(overrides.SyntaxKeywordHex),
		SyntaxFunctionHex:                    normalizeHexColor(overrides.SyntaxFunctionHex),
		SyntaxTypeHex:                        normalizeHexColor(overrides.SyntaxTypeHex),
		SyntaxPropertyHex:                    normalizeHexColor(overrides.SyntaxPropertyHex),
		SyntaxStringHex:                      normalizeHexColor(overrides.SyntaxStringHex),
		SyntaxNumberHex:                      normalizeHexColor(overrides.SyntaxNumberHex),
		SyntaxCommentHex:                     normalizeHexColor(overrides.SyntaxCommentHex),
		CommentAuthorBadgeForegroundHex:      normalizeHexColor(overrides.CommentAuthorBadgeForegroundHex),
		CommentAuthorBadgeBackgroundHex:      normalizeHexColor(overrides.CommentAuthorBadgeBackgroundHex),
		PullRequestStatusOpenForegroundHex:   normalizeHexColor(overrides.PullRequestStatusOpenForegroundHex),
		PullRequestStatusOpenBackgroundHex:   normalizeHexColor(overrides.PullRequestStatusOpenBackgroundHex),
		PullRequestStatusDraftForegroundHex:  normalizeHexColor(overrides.PullRequestStatusDraftForegroundHex),
		PullRequestStatusDraftBackgroundHex:  normalizeHexColor(overrides.PullRequestStatusDraftBackgroundHex),
		PullRequestStatusClosedForegroundHex: normalizeHexColor(overrides.PullRequestStatusClosedForegroundHex),
		PullRequestStatusClosedBackgroundHex: normalizeHexColor(overrides.PullRequestStatusClosedBackgroundHex),
		PullRequestStatusMergedForegroundHex: normalizeHexColor(overrides.PullRequestStatusMergedForegroundHex),
		PullRequestStatusMergedBackgroundHex: normalizeHexColor(overrides.PullRequestStatusMergedBackgroundHex),
		DiffAdditionForegroundHex:            normalizeHexColor(overrides.DiffAdditionForegroundHex),
		DiffAdditionBackgroundHex:            normalizeHexColor(overrides.DiffAdditionBackgroundHex),
		DiffAdditionHighlightBackgroundHex:   normalizeHexColor(overrides.DiffAdditionHighlightBackgroundHex),
		DiffDeletionForegroundHex:            normalizeHexColor(overrides.DiffDeletionForegroundHex),
		DiffDeletionBackgroundHex:            normalizeHexColor(overrides.DiffDeletionBackgroundHex),
		DiffDeletionHighlightBackgroundHex:   normalizeHexColor(overrides.DiffDeletionHighlightBackgroundHex),
		DiffLineNumberHex:                    normalizeHexColor(overrides.DiffLineNumberHex),
		DiffHunkHeaderHex:                    normalizeHexColor(overrides.DiffHunkHeaderHex),
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
