package tui

import (
	glamouransi "github.com/charmbracelet/glamour/ansi"
	glamourstyles "github.com/charmbracelet/glamour/styles"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func prettyMarkdownStyle() glamouransi.StyleConfig {
	light := glamourstyles.LightStyleConfig

	style := glamouransi.StyleConfig{
		Document: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr(theme.ActiveTextHex),
			},
			Margin: uintPtr(0),
		},
		BlockQuote: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				Color: stringPtr(theme.InactiveBorderHex),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		Paragraph: light.Paragraph,
		List: glamouransi.StyleList{
			StyleBlock:  light.List.StyleBlock,
			LevelIndent: light.List.LevelIndent,
		},
		Heading: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr(theme.MarkdownHeadingHex),
				Bold:        boolPtr(true),
			},
		},
		H1: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				Underline: boolPtr(true),
			},
		},
		H2: glamouransi.StyleBlock{},
		H3: glamouransi.StyleBlock{},
		H4: glamouransi.StyleBlock{},
		H5: glamouransi.StyleBlock{},
		H6: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				Bold: boolPtr(false),
			},
		},
		Text: glamouransi.StylePrimitive{},
		Strikethrough: glamouransi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: glamouransi.StylePrimitive{
			Italic: boolPtr(true),
		},
		Strong: glamouransi.StylePrimitive{
			Bold: boolPtr(true),
		},
		HorizontalRule: glamouransi.StylePrimitive{
			Color:  stringPtr(theme.InactiveBorderHex),
			Format: "\n────────────────────\n",
		},
		Item: glamouransi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: glamouransi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: light.Task,
		Link: glamouransi.StylePrimitive{
			Color:     stringPtr(theme.MarkdownLinkHex),
			Underline: boolPtr(true),
		},
		LinkText: glamouransi.StylePrimitive{
			Color: stringPtr(theme.MarkdownLinkHex),
			Bold:  boolPtr(true),
		},
		Image:     light.Image,
		ImageText: light.ImageText,
		Code: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{
				Prefix:          "\u00a0",
				Suffix:          "\u00a0",
				Color:           stringPtr(theme.MarkdownCodeHex),
				BackgroundColor: stringPtr(theme.SelectedLineBackgroundHex),
			},
		},
		CodeBlock:             light.CodeBlock,
		Table:                 light.Table,
		DefinitionList:        light.DefinitionList,
		DefinitionTerm:        light.DefinitionTerm,
		DefinitionDescription: light.DefinitionDescription,
		HTMLBlock:             light.HTMLBlock,
		HTMLSpan:              light.HTMLSpan,
	}

	style.CodeBlock.StyleBlock.Margin = uintPtr(0)
	style.CodeBlock.StyleBlock.StylePrimitive.BackgroundColor = stringPtr(theme.SelectedLineBackgroundHex)
	return style
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func uintPtr(value uint) *uint {
	return &value
}
