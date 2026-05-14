package theme

import "strings"

const (
	SystemPresetName = "system"
	LightPresetName  = "light"
	DarkPresetName   = "dark"
)

type Preset struct {
	Name  string
	Label string
}

type presetBase int

const (
	presetBaseSystem presetBase = iota
	presetBaseLight
	presetBaseDark
)

type presetDefinition struct {
	Preset
	base      presetBase
	overrides Palette
}

var presetOrder = []string{
	SystemPresetName,
	LightPresetName,
	DarkPresetName,
	"catppuccin-latte",
	"catppuccin-frappe",
	"catppuccin-macchiato",
	"catppuccin-mocha",
	"gruvbox-dark",
	"gruvbox-light",
	"kanagawa-dark",
	"kanagawa-light",
	"nord",
	"tokyonight-dark",
	"tokyonight-light",
}

var presetDefinitions = map[string]presetDefinition{
	SystemPresetName: {
		Preset: Preset{Name: SystemPresetName, Label: "System (auto)"},
		base:   presetBaseSystem,
	},
	LightPresetName: {
		Preset: Preset{Name: LightPresetName, Label: "Light"},
		base:   presetBaseLight,
	},
	DarkPresetName: {
		Preset: Preset{Name: DarkPresetName, Label: "Dark"},
		base:   presetBaseDark,
	},
	"catppuccin-latte": {
		Preset: Preset{Name: "catppuccin-latte", Label: "Catppuccin Latte"},
		base:   presetBaseLight,
		overrides: Palette{
			BackgroundHex:                        "#EFF1F5",
			ActiveBorderHex:                      "#1E66F5",
			ActiveTextHex:                        "#4C4F69",
			InactiveBorderHex:                    "#8C8FA1",
			InactiveTextHex:                      "#5C5F77",
			InactiveTitleHex:                     "#9CA0B0",
			SuccessHex:                           "#40A02B",
			SuccessBackgroundHex:                 "#DCEFD5",
			FailureHex:                           "#D20F39",
			FailureBackgroundHex:                 "#F5D8DE",
			PendingHex:                           "#5C5F77",
			PendingBackgroundHex:                 "#E6E9EF",
			MutedHex:                             "#9CA0B0",
			WarningHex:                           "#FE640B",
			PullRequestReferenceHex:              "#9CA0B0",
			PullRequestTitleHex:                  "#4C4F69",
			SelectedLineBackgroundHex:            "#E6E9EF",
			SearchHighlightHex:                   "#DF8E1D",
			MarkdownHeadingHex:                   "#1E66F5",
			MarkdownHeadingBackgroundHex:         "#DCE0E8",
			MarkdownLinkHex:                      "#209FB5",
			MarkdownCodeHex:                      "#FE640B",
			SyntaxKeywordHex:                     "#8839EF",
			SyntaxFunctionHex:                    "#1E66F5",
			SyntaxTypeHex:                        "#179299",
			SyntaxPropertyHex:                    "#209FB5",
			SyntaxStringHex:                      "#40A02B",
			SyntaxNumberHex:                      "#FE640B",
			SyntaxCommentHex:                     "#8C8FA1",
			CommentAuthorBadgeHex:                "#1E66F5",
			CommentAuthorBadgeBackgroundHex:      "#DCE0E8",
			PullRequestStatusMergedHex:           "#8839EF",
			PullRequestStatusMergedBackgroundHex: "#E8DBFB",
			DiffAdditionBackgroundHex:            "#EAF6E5",
			DiffAdditionHighlightBackgroundHex:   "#DCEFD5",
			DiffDeletionBackgroundHex:            "#FBE9ED",
			DiffDeletionHighlightBackgroundHex:   "#F5D8DE",
			DiffLineNumberHex:                    "#8C8FA1",
			TeamOwnershipHex:                     "#8C8FA1",
			DiffHunkHeaderHex:                    "#209FB5",
		},
	},
	"catppuccin-frappe": {
		Preset: Preset{Name: "catppuccin-frappe", Label: "Catppuccin Frappé"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#303446",
			ActiveBorderHex:                      "#8CAAEE",
			ActiveTextHex:                        "#C6D0F5",
			InactiveBorderHex:                    "#737994",
			InactiveTextHex:                      "#B5BFE2",
			InactiveTitleHex:                     "#A5ADCE",
			SuccessHex:                           "#A6D189",
			SuccessBackgroundHex:                 "#A6D189",
			FailureHex:                           "#E78284",
			FailureBackgroundHex:                 "#E78284",
			PendingHex:                           "#C6D0F5",
			PendingBackgroundHex:                 "#414559",
			MutedHex:                             "#A5ADCE",
			WarningHex:                           "#EF9F76",
			PullRequestReferenceHex:              "#A5ADCE",
			PullRequestTitleHex:                  "#C6D0F5",
			SelectedLineBackgroundHex:            "#414559",
			SearchHighlightHex:                   "#E5C890",
			MarkdownHeadingHex:                   "#8CAAEE",
			MarkdownHeadingBackgroundHex:         "#414559",
			MarkdownLinkHex:                      "#85C1DC",
			MarkdownCodeHex:                      "#EF9F76",
			SyntaxKeywordHex:                     "#CA9EE6",
			SyntaxFunctionHex:                    "#8CAAEE",
			SyntaxTypeHex:                        "#81C8BE",
			SyntaxPropertyHex:                    "#85C1DC",
			SyntaxStringHex:                      "#A6D189",
			SyntaxNumberHex:                      "#EF9F76",
			SyntaxCommentHex:                     "#737994",
			CommentAuthorBadgeHex:                "#8CAAEE",
			CommentAuthorBadgeBackgroundHex:      "#414559",
			PullRequestStatusOpenHex:             "#A6D189",
			PullRequestStatusClosedHex:           "#E78284",
			PullRequestStatusMergedHex:           "#CA9EE6",
			PullRequestStatusMergedBackgroundHex: "#E8DBFB",
			DiffAdditionBackgroundHex:            "#3A4A3A",
			DiffAdditionHighlightBackgroundHex:   "#496049",
			DiffDeletionBackgroundHex:            "#4A363C",
			DiffDeletionHighlightBackgroundHex:   "#61464F",
			DiffLineNumberHex:                    "#737994",
			TeamOwnershipHex:                     "#737994",
			DiffHunkHeaderHex:                    "#85C1DC",
		},
	},
	"catppuccin-macchiato": {
		Preset: Preset{Name: "catppuccin-macchiato", Label: "Catppuccin Macchiato"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#24273A",
			ActiveBorderHex:                      "#8AADF4",
			ActiveTextHex:                        "#CAD3F5",
			InactiveBorderHex:                    "#6E738D",
			InactiveTextHex:                      "#B8C0E0",
			InactiveTitleHex:                     "#A5ADCB",
			SuccessHex:                           "#A6DA95",
			SuccessBackgroundHex:                 "#A6DA95",
			FailureHex:                           "#ED8796",
			FailureBackgroundHex:                 "#ED8796",
			PendingHex:                           "#CAD3F5",
			PendingBackgroundHex:                 "#363A4F",
			MutedHex:                             "#A5ADCB",
			WarningHex:                           "#F5A97F",
			PullRequestReferenceHex:              "#A5ADCB",
			PullRequestTitleHex:                  "#CAD3F5",
			SelectedLineBackgroundHex:            "#363A4F",
			SearchHighlightHex:                   "#EED49F",
			MarkdownHeadingHex:                   "#8AADF4",
			MarkdownHeadingBackgroundHex:         "#363A4F",
			MarkdownLinkHex:                      "#7DC4E4",
			MarkdownCodeHex:                      "#F5A97F",
			SyntaxKeywordHex:                     "#C6A0F6",
			SyntaxFunctionHex:                    "#8AADF4",
			SyntaxTypeHex:                        "#8BD5CA",
			SyntaxPropertyHex:                    "#7DC4E4",
			SyntaxStringHex:                      "#A6DA95",
			SyntaxNumberHex:                      "#F5A97F",
			SyntaxCommentHex:                     "#6E738D",
			CommentAuthorBadgeHex:                "#8AADF4",
			CommentAuthorBadgeBackgroundHex:      "#363A4F",
			PullRequestStatusOpenHex:             "#A6DA95",
			PullRequestStatusClosedHex:           "#ED8796",
			PullRequestStatusMergedHex:           "#C6A0F6",
			PullRequestStatusMergedBackgroundHex: "#E8DBFB",
			DiffAdditionBackgroundHex:            "#2F4638",
			DiffAdditionHighlightBackgroundHex:   "#3F5D49",
			DiffDeletionBackgroundHex:            "#4A333C",
			DiffDeletionHighlightBackgroundHex:   "#623F4A",
			DiffLineNumberHex:                    "#6E738D",
			TeamOwnershipHex:                     "#6E738D",
			DiffHunkHeaderHex:                    "#7DC4E4",
		},
	},
	"catppuccin-mocha": {
		Preset: Preset{Name: "catppuccin-mocha", Label: "Catppuccin Mocha"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#1E1E2E",
			ActiveBorderHex:                      "#89B4FA",
			ActiveTextHex:                        "#CDD6F4",
			InactiveBorderHex:                    "#6C7086",
			InactiveTextHex:                      "#BAC2DE",
			InactiveTitleHex:                     "#A6ADC8",
			SuccessHex:                           "#A6E3A1",
			SuccessBackgroundHex:                 "#A6E3A1",
			FailureHex:                           "#F38BA8",
			FailureBackgroundHex:                 "#F38BA8",
			PendingHex:                           "#CDD6F4",
			PendingBackgroundHex:                 "#313244",
			MutedHex:                             "#A6ADC8",
			WarningHex:                           "#FAB387",
			PullRequestReferenceHex:              "#A6ADC8",
			PullRequestTitleHex:                  "#CDD6F4",
			SelectedLineBackgroundHex:            "#313244",
			SearchHighlightHex:                   "#F9E2AF",
			MarkdownHeadingHex:                   "#89B4FA",
			MarkdownHeadingBackgroundHex:         "#313244",
			MarkdownLinkHex:                      "#74C7EC",
			MarkdownCodeHex:                      "#FAB387",
			SyntaxKeywordHex:                     "#CBA6F7",
			SyntaxFunctionHex:                    "#89B4FA",
			SyntaxTypeHex:                        "#94E2D5",
			SyntaxPropertyHex:                    "#74C7EC",
			SyntaxStringHex:                      "#A6E3A1",
			SyntaxNumberHex:                      "#FAB387",
			SyntaxCommentHex:                     "#6C7086",
			CommentAuthorBadgeHex:                "#89B4FA",
			CommentAuthorBadgeBackgroundHex:      "#313244",
			PullRequestStatusOpenHex:             "#A6E3A1",
			PullRequestStatusClosedHex:           "#F38BA8",
			PullRequestStatusMergedHex:           "#CBA6F7",
			PullRequestStatusMergedBackgroundHex: "#E8DBFB",
			DiffAdditionBackgroundHex:            "#2B4133",
			DiffAdditionHighlightBackgroundHex:   "#395544",
			DiffDeletionBackgroundHex:            "#4A303B",
			DiffDeletionHighlightBackgroundHex:   "#633D4B",
			DiffLineNumberHex:                    "#6C7086",
			TeamOwnershipHex:                     "#6C7086",
			DiffHunkHeaderHex:                    "#74C7EC",
		},
	},
	"gruvbox-dark": {
		Preset: Preset{Name: "gruvbox-dark", Label: "Gruvbox Dark"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#282828",
			ActiveBorderHex:                      "#83A598",
			ActiveTextHex:                        "#EBDBB2",
			InactiveBorderHex:                    "#665C54",
			InactiveTextHex:                      "#D5C4A1",
			InactiveTitleHex:                     "#928374",
			SuccessHex:                           "#B8BB26",
			SuccessBackgroundHex:                 "#32361A",
			FailureHex:                           "#FB4934",
			FailureBackgroundHex:                 "#3C1F1E",
			PendingHex:                           "#928374",
			PendingBackgroundHex:                 "#3C3836",
			MutedHex:                             "#928374",
			WarningHex:                           "#FE8019",
			PullRequestReferenceHex:              "#928374",
			PullRequestTitleHex:                  "#EBDBB2",
			SelectedLineBackgroundHex:            "#3C3836",
			SearchHighlightHex:                   "#FABD2F",
			MarkdownHeadingHex:                   "#83A598",
			MarkdownHeadingBackgroundHex:         "#3C3836",
			MarkdownLinkHex:                      "#83A598",
			MarkdownCodeHex:                      "#FE8019",
			SyntaxKeywordHex:                     "#FB4934",
			SyntaxFunctionHex:                    "#B8BB26",
			SyntaxTypeHex:                        "#FABD2F",
			SyntaxPropertyHex:                    "#83A598",
			SyntaxStringHex:                      "#B8BB26",
			SyntaxNumberHex:                      "#D3869B",
			SyntaxCommentHex:                     "#928374",
			CommentAuthorBadgeHex:                "#83A598",
			CommentAuthorBadgeBackgroundHex:      "#3C3836",
			PullRequestStatusMergedHex:           "#D3869B",
			PullRequestStatusMergedBackgroundHex: "#3F2D3D",
			DiffAdditionHighlightBackgroundHex:   "#4B5C16",
			DiffDeletionHighlightBackgroundHex:   "#5C2B24",
			DiffLineNumberHex:                    "#928374",
			TeamOwnershipHex:                     "#928374",
			DiffHunkHeaderHex:                    "#83A598",
		},
	},
	"gruvbox-light": {
		Preset: Preset{Name: "gruvbox-light", Label: "Gruvbox Light"},
		base:   presetBaseLight,
		overrides: Palette{
			BackgroundHex:                        "#FBF1C7",
			ActiveBorderHex:                      "#076678",
			ActiveTextHex:                        "#3C3836",
			InactiveBorderHex:                    "#A89984",
			InactiveTextHex:                      "#504945",
			InactiveTitleHex:                     "#A89984",
			SuccessHex:                           "#79740E",
			SuccessBackgroundHex:                 "#E3E7C1",
			FailureHex:                           "#9D0006",
			FailureBackgroundHex:                 "#F0D2CF",
			PendingHex:                           "#928374",
			PendingBackgroundHex:                 "#EBDBB2",
			MutedHex:                             "#A89984",
			WarningHex:                           "#AF3A03",
			PullRequestReferenceHex:              "#A89984",
			PullRequestTitleHex:                  "#3C3836",
			SelectedLineBackgroundHex:            "#EBDBB2",
			SearchHighlightHex:                   "#D79921",
			MarkdownHeadingHex:                   "#076678",
			MarkdownHeadingBackgroundHex:         "#D5C4A1",
			MarkdownLinkHex:                      "#427B58",
			MarkdownCodeHex:                      "#AF3A03",
			SyntaxKeywordHex:                     "#9D0006",
			SyntaxFunctionHex:                    "#427B58",
			SyntaxTypeHex:                        "#B57614",
			SyntaxPropertyHex:                    "#076678",
			SyntaxStringHex:                      "#79740E",
			SyntaxNumberHex:                      "#8F3F71",
			SyntaxCommentHex:                     "#928374",
			CommentAuthorBadgeHex:                "#076678",
			CommentAuthorBadgeBackgroundHex:      "#D5C4A1",
			PullRequestStatusMergedHex:           "#8F3F71",
			PullRequestStatusMergedBackgroundHex: "#E5D3DD",
			DiffAdditionHighlightBackgroundHex:   "#D4DA98",
			DiffDeletionHighlightBackgroundHex:   "#E7B8B2",
			DiffLineNumberHex:                    "#928374",
			TeamOwnershipHex:                     "#928374",
			DiffHunkHeaderHex:                    "#076678",
		},
	},
	"kanagawa-dark": {
		Preset: Preset{Name: "kanagawa-dark", Label: "Kanagawa Dark"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#1F1F28",
			ActiveBorderHex:                      "#7E9CD8",
			ActiveTextHex:                        "#DCD7BA",
			InactiveBorderHex:                    "#54546D",
			InactiveTextHex:                      "#C8C093",
			InactiveTitleHex:                     "#727169",
			SuccessHex:                           "#98BB6C",
			SuccessBackgroundHex:                 "#2B3328",
			FailureHex:                           "#E46876",
			FailureBackgroundHex:                 "#43242B",
			PendingHex:                           "#C8C093",
			PendingBackgroundHex:                 "#363646",
			MutedHex:                             "#727169",
			WarningHex:                           "#FFA066",
			PullRequestReferenceHex:              "#727169",
			PullRequestTitleHex:                  "#DCD7BA",
			SelectedLineBackgroundHex:            "#363646",
			SearchHighlightHex:                   "#2D4F67",
			MarkdownHeadingHex:                   "#7E9CD8",
			MarkdownHeadingBackgroundHex:         "#223249",
			MarkdownLinkHex:                      "#7AA89F",
			MarkdownCodeHex:                      "#FFA066",
			SyntaxKeywordHex:                     "#957FB8",
			SyntaxFunctionHex:                    "#7FB4CA",
			SyntaxTypeHex:                        "#7AA89F",
			SyntaxPropertyHex:                    "#E6C384",
			SyntaxStringHex:                      "#98BB6C",
			SyntaxNumberHex:                      "#D27E99",
			SyntaxCommentHex:                     "#727169",
			CommentAuthorBadgeHex:                "#7E9CD8",
			CommentAuthorBadgeBackgroundHex:      "#223249",
			PullRequestStatusMergedHex:           "#957FB8",
			PullRequestStatusMergedBackgroundHex: "#252535",
			DiffAdditionHighlightBackgroundHex:   "#35513B",
			DiffDeletionHighlightBackgroundHex:   "#5A2E35",
			DiffLineNumberHex:                    "#727169",
			TeamOwnershipHex:                     "#727169",
			DiffHunkHeaderHex:                    "#7E9CD8",
		},
	},
	"kanagawa-light": {
		Preset: Preset{Name: "kanagawa-light", Label: "Kanagawa Light"},
		base:   presetBaseLight,
		overrides: Palette{
			BackgroundHex:                        "#F2ECBC",
			ActiveBorderHex:                      "#4D699B",
			ActiveTextHex:                        "#545464",
			InactiveBorderHex:                    "#8A8980",
			InactiveTextHex:                      "#716E61",
			InactiveTitleHex:                     "#8A8980",
			SuccessHex:                           "#6F894E",
			SuccessBackgroundHex:                 "#D7E3D8",
			FailureHex:                           "#C84053",
			FailureBackgroundHex:                 "#EAD4CF",
			PendingHex:                           "#716E61",
			PendingBackgroundHex:                 "#E5DDB0",
			MutedHex:                             "#8A8980",
			WarningHex:                           "#CC6D00",
			PullRequestReferenceHex:              "#8A8980",
			PullRequestTitleHex:                  "#545464",
			SelectedLineBackgroundHex:            "#E7DBA0",
			SearchHighlightHex:                   "#F9D791",
			MarkdownHeadingHex:                   "#4D699B",
			MarkdownHeadingBackgroundHex:         "#C7D7E0",
			MarkdownLinkHex:                      "#597B75",
			MarkdownCodeHex:                      "#CC6D00",
			SyntaxKeywordHex:                     "#624C83",
			SyntaxFunctionHex:                    "#4E8CA2",
			SyntaxTypeHex:                        "#5E857A",
			SyntaxPropertyHex:                    "#597B75",
			SyntaxStringHex:                      "#6F894E",
			SyntaxNumberHex:                      "#B35B79",
			SyntaxCommentHex:                     "#8A8980",
			CommentAuthorBadgeHex:                "#4D699B",
			CommentAuthorBadgeBackgroundHex:      "#C7D7E0",
			PullRequestStatusMergedHex:           "#624C83",
			PullRequestStatusMergedBackgroundHex: "#E2DCEB",
			DiffAdditionHighlightBackgroundHex:   "#B7D0AE",
			DiffDeletionHighlightBackgroundHex:   "#D9A594",
			DiffLineNumberHex:                    "#8A8980",
			TeamOwnershipHex:                     "#8A8980",
			DiffHunkHeaderHex:                    "#4D699B",
		},
	},
	"nord": {
		Preset: Preset{Name: "nord", Label: "Nord"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#2E3440",
			ActiveBorderHex:                      "#88C0D0",
			ActiveTextHex:                        "#E5E9F0",
			InactiveBorderHex:                    "#434C5E",
			InactiveTextHex:                      "#D8DEE9",
			InactiveTitleHex:                     "#4C566A",
			SuccessHex:                           "#A3BE8C",
			SuccessBackgroundHex:                 "#36413C",
			FailureHex:                           "#BF616A",
			FailureBackgroundHex:                 "#4A343C",
			PendingHex:                           "#D8DEE9",
			PendingBackgroundHex:                 "#3B4252",
			MutedHex:                             "#4C566A",
			WarningHex:                           "#D08770",
			PullRequestReferenceHex:              "#4C566A",
			PullRequestTitleHex:                  "#E5E9F0",
			SelectedLineBackgroundHex:            "#434C5E",
			SearchHighlightHex:                   "#5E81AC",
			MarkdownHeadingHex:                   "#81A1C1",
			MarkdownHeadingBackgroundHex:         "#3B4252",
			MarkdownLinkHex:                      "#88C0D0",
			MarkdownCodeHex:                      "#D08770",
			SyntaxKeywordHex:                     "#81A1C1",
			SyntaxFunctionHex:                    "#88C0D0",
			SyntaxTypeHex:                        "#8FBCBB",
			SyntaxPropertyHex:                    "#81A1C1",
			SyntaxStringHex:                      "#A3BE8C",
			SyntaxNumberHex:                      "#B48EAD",
			SyntaxCommentHex:                     "#4C566A",
			CommentAuthorBadgeHex:                "#88C0D0",
			CommentAuthorBadgeBackgroundHex:      "#3B4252",
			PullRequestStatusMergedHex:           "#B48EAD",
			PullRequestStatusMergedBackgroundHex: "#4A4252",
			DiffAdditionHighlightBackgroundHex:   "#43544C",
			DiffDeletionHighlightBackgroundHex:   "#5A414A",
			DiffLineNumberHex:                    "#4C566A",
			TeamOwnershipHex:                     "#4C566A",
			DiffHunkHeaderHex:                    "#81A1C1",
		},
	},
	"tokyonight-dark": {
		Preset: Preset{Name: "tokyonight-dark", Label: "Tokyo Night Dark"},
		base:   presetBaseDark,
		overrides: Palette{
			BackgroundHex:                        "#1A1B26",
			ActiveBorderHex:                      "#7AA2F7",
			ActiveTextHex:                        "#C0CAF5",
			InactiveBorderHex:                    "#414868",
			InactiveTextHex:                      "#A9B1D6",
			InactiveTitleHex:                     "#565F89",
			SuccessHex:                           "#9ECE6A",
			SuccessBackgroundHex:                 "#164846",
			FailureHex:                           "#F7768E",
			FailureBackgroundHex:                 "#823C41",
			PendingHex:                           "#A9B1D6",
			PendingBackgroundHex:                 "#292E42",
			MutedHex:                             "#565F89",
			WarningHex:                           "#FF9E64",
			PullRequestReferenceHex:              "#565F89",
			PullRequestTitleHex:                  "#C0CAF5",
			SelectedLineBackgroundHex:            "#292E42",
			SearchHighlightHex:                   "#3D59A1",
			MarkdownHeadingHex:                   "#7AA2F7",
			MarkdownHeadingBackgroundHex:         "#3B4261",
			MarkdownLinkHex:                      "#73DACA",
			MarkdownCodeHex:                      "#FF9E64",
			SyntaxKeywordHex:                     "#BB9AF7",
			SyntaxFunctionHex:                    "#7AA2F7",
			SyntaxTypeHex:                        "#2AC3DE",
			SyntaxPropertyHex:                    "#73DACA",
			SyntaxStringHex:                      "#9ECE6A",
			SyntaxNumberHex:                      "#FF9E64",
			SyntaxCommentHex:                     "#565F89",
			CommentAuthorBadgeHex:                "#7AA2F7",
			CommentAuthorBadgeBackgroundHex:      "#3B4261",
			PullRequestStatusMergedHex:           "#BB9AF7",
			PullRequestStatusMergedBackgroundHex: "#3B3052",
			DiffAdditionHighlightBackgroundHex:   "#1F5B57",
			DiffDeletionHighlightBackgroundHex:   "#914C54",
			DiffLineNumberHex:                    "#565F89",
			TeamOwnershipHex:                     "#565F89",
			DiffHunkHeaderHex:                    "#7AA2F7",
		},
	},
	"tokyonight-light": {
		Preset: Preset{Name: "tokyonight-light", Label: "Tokyo Night Light"},
		base:   presetBaseLight,
		overrides: Palette{
			BackgroundHex:                        "#E1E2E7",
			ActiveBorderHex:                      "#2959AA",
			ActiveTextHex:                        "#343B58",
			InactiveBorderHex:                    "#9699A3",
			InactiveTextHex:                      "#343B58",
			InactiveTitleHex:                     "#6C6E75",
			SuccessHex:                           "#587539",
			SuccessBackgroundHex:                 "#DCE6CF",
			FailureHex:                           "#C64343",
			FailureBackgroundHex:                 "#DED2D7",
			PendingHex:                           "#6C6E75",
			PendingBackgroundHex:                 "#D5D6DB",
			MutedHex:                             "#6C6E75",
			WarningHex:                           "#B15C00",
			PullRequestReferenceHex:              "#6C6E75",
			PullRequestTitleHex:                  "#343B58",
			SelectedLineBackgroundHex:            "#D5D6DB",
			SearchHighlightHex:                   "#EACB64",
			MarkdownHeadingHex:                   "#2959AA",
			MarkdownHeadingBackgroundHex:         "#CBD9E0",
			MarkdownLinkHex:                      "#33635C",
			MarkdownCodeHex:                      "#8C6C3E",
			SyntaxKeywordHex:                     "#5A3E8E",
			SyntaxFunctionHex:                    "#2959AA",
			SyntaxTypeHex:                        "#006C86",
			SyntaxPropertyHex:                    "#0F4B6E",
			SyntaxStringHex:                      "#385F0D",
			SyntaxNumberHex:                      "#B15C00",
			SyntaxCommentHex:                     "#6C6E75",
			CommentAuthorBadgeHex:                "#2959AA",
			CommentAuthorBadgeBackgroundHex:      "#CBD9E0",
			PullRequestStatusMergedHex:           "#5A3E8E",
			PullRequestStatusMergedBackgroundHex: "#DDD5E7",
			DiffDeletionBackgroundHex:            "#EDD5DA",
			DiffAdditionHighlightBackgroundHex:   "#C5D8AA",
			DiffDeletionHighlightBackgroundHex:   "#E1B8C0",
			DiffLineNumberHex:                    "#6C6E75",
			TeamOwnershipHex:                     "#6C6E75",
			DiffHunkHeaderHex:                    "#2959AA",
		},
	},
}

func Presets() []Preset {
	presets := make([]Preset, 0, len(presetOrder))
	for _, name := range presetOrder {
		definition, ok := presetDefinitions[name]
		if !ok {
			continue
		}
		presets = append(presets, definition.Preset)
	}
	return presets
}

func NormalizePresetName(name string) string {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if _, ok := presetDefinitions[normalizedName]; !ok {
		return ""
	}
	return normalizedName
}

func PresetOverrides(name string) (Palette, bool) {
	normalizedName := NormalizePresetName(name)
	if normalizedName == "" {
		return Palette{}, false
	}
	return NormalizePalette(presetDefinitions[normalizedName].overrides), true
}

func ResolvePresetPalette(name string) (Palette, bool) {
	normalizedName := NormalizePresetName(name)
	if normalizedName == "" {
		return Palette{}, false
	}
	return resolvedPresetPalette(normalizedName), true
}

func ResolvePaletteWithPreset(presetName string, overrides Palette) Palette {
	resolved := resolvedPresetPalette(presetName)
	normalizedOverrides := NormalizePalette(overrides)
	resolved = mergePalette(resolved, normalizedOverrides)
	cascadePaletteColors(&resolved, normalizedOverrides)
	return resolved
}

func resolvedPresetPalette(presetName string) Palette {
	normalizedName := NormalizePresetName(presetName)
	switch normalizedName {
	case "", SystemPresetName:
		return DefaultPalette()
	case LightPresetName:
		return defaultLightPalette
	case DarkPresetName:
		return defaultDarkPalette
	default:
		definition := presetDefinitions[normalizedName]
		normalizedOverrides := NormalizePalette(definition.overrides)
		resolved := mergePalette(defaultPaletteForPresetBase(definition.base), normalizedOverrides)
		cascadePaletteColors(&resolved, normalizedOverrides)
		return resolved
	}
}

func defaultPaletteForPresetBase(base presetBase) Palette {
	switch base {
	case presetBaseLight:
		return defaultLightPalette
	case presetBaseDark:
		return defaultDarkPalette
	default:
		return DefaultPalette()
	}
}
