package tui

import (
	"math"
	"strconv"

	glamouransi "github.com/charmbracelet/glamour/ansi"
	glamourstyles "github.com/charmbracelet/glamour/styles"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	markdownReadableDarkHex  = "#000000"
	markdownReadableLightHex = "#FFFFFF"
)

func prettyMarkdownStyle() glamouransi.StyleConfig {
	style := baseMarkdownStyle()

	style.Document.StylePrimitive.BlockPrefix = "\n"
	style.Document.StylePrimitive.BlockSuffix = "\n"
	style.Document.StylePrimitive.Color = stringPtr(theme.ActiveTextHex)
	style.Document.Margin = uintPtr(0)

	style.BlockQuote.StylePrimitive.Color = stringPtr(theme.InactiveTitleHex)
	style.BlockQuote.Indent = uintPtr(1)
	style.BlockQuote.IndentToken = stringPtr("│ ")

	style.Heading.StylePrimitive.BlockSuffix = "\n"
	style.Heading.StylePrimitive.Color = stringPtr(theme.MarkdownHeadingHex)
	style.Heading.StylePrimitive.BackgroundColor = stringPtr(theme.MarkdownHeadingBackgroundHex)
	style.Heading.StylePrimitive.Bold = boolPtr(true)

	style.H1.Prefix = "# "
	style.H1.StylePrimitive.Color = stringPtr(theme.MarkdownHeadingHex)
	style.H1.StylePrimitive.BackgroundColor = stringPtr(theme.MarkdownHeadingBackgroundHex)

	style.HorizontalRule.Color = stringPtr(theme.InactiveBorderHex)

	style.Link.Color = stringPtr(theme.MarkdownLinkHex)
	style.Link.Underline = boolPtr(true)
	style.LinkText.Color = stringPtr(theme.MarkdownLinkHex)
	style.LinkText.Bold = boolPtr(true)
	style.LinkText.Underline = boolPtr(false)
	style.LinkText.Prefix = "󰌹 "

	style.Image.Color = stringPtr(theme.MarkdownLinkHex)
	style.Image.Underline = boolPtr(true)
	style.ImageText.Color = stringPtr(theme.InactiveTitleHex)
	style.ImageText.Format = "{{.text}}"
	style.ImageText.Prefix = " "
	style.ImageText.Bold = boolPtr(true)

	style.Code.StylePrimitive.Color = stringPtr(theme.MarkdownCodeHex)
	style.Code.StylePrimitive.BackgroundColor = stringPtr(theme.SelectedLineBackgroundHex)

	style.CodeBlock.StyleBlock.Margin = uintPtr(0)
	style.CodeBlock.StyleBlock.StylePrimitive.Color = stringPtr(theme.ActiveTextHex)
	style.CodeBlock.StyleBlock.StylePrimitive.BackgroundColor = stringPtr(theme.SelectedLineBackgroundHex)
	applyMarkdownCodeBlockTheme(&style)

	return style
}

func baseMarkdownStyle() glamouransi.StyleConfig {
	if useDarkMarkdownStyle(theme.ActiveTextHex) {
		return glamourstyles.DarkStyleConfig
	}

	return glamourstyles.LightStyleConfig
}

func applyMarkdownCodeBlockTheme(style *glamouransi.StyleConfig) {
	if style.CodeBlock.Chroma == nil {
		style.CodeBlock.Chroma = &glamouransi.Chroma{}
	} else {
		chroma := *style.CodeBlock.Chroma
		style.CodeBlock.Chroma = &chroma
	}

	style.CodeBlock.Chroma.Text.Color = stringPtr(theme.ActiveTextHex)
	style.CodeBlock.Chroma.Error.Color = stringPtr(readableMarkdownForegroundHex(theme.DiffDeletionBackgroundHex))
	style.CodeBlock.Chroma.Error.BackgroundColor = stringPtr(theme.DiffDeletionBackgroundHex)
	style.CodeBlock.Chroma.Comment.Color = stringPtr(theme.SyntaxCommentHex)
	style.CodeBlock.Chroma.CommentPreproc.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.Keyword.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.KeywordReserved.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.KeywordNamespace.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.KeywordType.Color = stringPtr(theme.SyntaxTypeHex)
	style.CodeBlock.Chroma.Operator.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.Punctuation.Color = stringPtr(theme.ActiveTextHex)
	style.CodeBlock.Chroma.Name.Color = stringPtr(theme.ActiveTextHex)
	style.CodeBlock.Chroma.NameBuiltin.Color = stringPtr(theme.SyntaxFunctionHex)
	style.CodeBlock.Chroma.NameTag.Color = stringPtr(theme.SyntaxTypeHex)
	style.CodeBlock.Chroma.NameAttribute.Color = stringPtr(theme.SyntaxPropertyHex)
	style.CodeBlock.Chroma.NameClass.Color = stringPtr(theme.SyntaxTypeHex)
	style.CodeBlock.Chroma.NameConstant.Color = stringPtr(theme.SyntaxPropertyHex)
	style.CodeBlock.Chroma.NameDecorator.Color = stringPtr(theme.SyntaxKeywordHex)
	style.CodeBlock.Chroma.NameException.Color = stringPtr(theme.SyntaxTypeHex)
	style.CodeBlock.Chroma.NameFunction.Color = stringPtr(theme.SyntaxFunctionHex)
	style.CodeBlock.Chroma.NameOther.Color = stringPtr(theme.SyntaxPropertyHex)
	style.CodeBlock.Chroma.Literal.Color = stringPtr(theme.SyntaxStringHex)
	style.CodeBlock.Chroma.LiteralNumber.Color = stringPtr(theme.SyntaxNumberHex)
	style.CodeBlock.Chroma.LiteralDate.Color = stringPtr(theme.SyntaxNumberHex)
	style.CodeBlock.Chroma.LiteralString.Color = stringPtr(theme.SyntaxStringHex)
	style.CodeBlock.Chroma.LiteralStringEscape.Color = stringPtr(theme.SyntaxStringHex)
	style.CodeBlock.Chroma.GenericDeleted.Color = stringPtr(theme.DiffDeletionForegroundHex)
	style.CodeBlock.Chroma.GenericEmph.Italic = boolPtr(true)
	style.CodeBlock.Chroma.GenericInserted.Color = stringPtr(theme.DiffAdditionForegroundHex)
	style.CodeBlock.Chroma.GenericStrong.Bold = boolPtr(true)
	style.CodeBlock.Chroma.GenericSubheading.Color = stringPtr(theme.InactiveTitleHex)
	style.CodeBlock.Chroma.Background.BackgroundColor = stringPtr(theme.SelectedLineBackgroundHex)
	applyMarkdownCodeBlockBackground(style.CodeBlock.Chroma, theme.SelectedLineBackgroundHex)
}

func applyMarkdownCodeBlockBackground(chroma *glamouransi.Chroma, backgroundHex string) {
	backgroundStyles := []*glamouransi.StylePrimitive{
		&chroma.Text,
		&chroma.Comment,
		&chroma.CommentPreproc,
		&chroma.Keyword,
		&chroma.KeywordReserved,
		&chroma.KeywordNamespace,
		&chroma.KeywordType,
		&chroma.Operator,
		&chroma.Punctuation,
		&chroma.Name,
		&chroma.NameBuiltin,
		&chroma.NameTag,
		&chroma.NameAttribute,
		&chroma.NameClass,
		&chroma.NameConstant,
		&chroma.NameDecorator,
		&chroma.NameException,
		&chroma.NameFunction,
		&chroma.NameOther,
		&chroma.Literal,
		&chroma.LiteralNumber,
		&chroma.LiteralDate,
		&chroma.LiteralString,
		&chroma.LiteralStringEscape,
		&chroma.GenericDeleted,
		&chroma.GenericEmph,
		&chroma.GenericInserted,
		&chroma.GenericStrong,
		&chroma.GenericSubheading,
	}
	for _, style := range backgroundStyles {
		style.BackgroundColor = stringPtr(backgroundHex)
	}
}

func useDarkMarkdownStyle(activeTextHex string) bool {
	activeTextLuminance, ok := relativeLuminance(activeTextHex)
	if !ok {
		return false
	}

	return contrastRatio(activeTextLuminance, 0) >= contrastRatio(activeTextLuminance, 1)
}

func readableMarkdownForegroundHex(backgroundHex string) string {
	backgroundLuminance, ok := relativeLuminance(backgroundHex)
	if !ok {
		return markdownReadableDarkHex
	}

	if contrastRatio(backgroundLuminance, 0) >= contrastRatio(backgroundLuminance, 1) {
		return markdownReadableDarkHex
	}

	return markdownReadableLightHex
}

func relativeLuminance(hexColor string) (float64, bool) {
	red, green, blue, ok := parseNormalizedHexColor(hexColor)
	if !ok {
		return 0, false
	}

	return 0.2126*linearizedColorChannel(red) + 0.7152*linearizedColorChannel(green) + 0.0722*linearizedColorChannel(blue), true
}

func parseNormalizedHexColor(hexColor string) (float64, float64, float64, bool) {
	if len(hexColor) != 7 || hexColor[0] != '#' {
		return 0, 0, 0, false
	}

	red, ok := parseHexColorChannel(hexColor[1:3])
	if !ok {
		return 0, 0, 0, false
	}
	green, ok := parseHexColorChannel(hexColor[3:5])
	if !ok {
		return 0, 0, 0, false
	}
	blue, ok := parseHexColorChannel(hexColor[5:7])
	if !ok {
		return 0, 0, 0, false
	}

	return red, green, blue, true
}

func parseHexColorChannel(value string) (float64, bool) {
	parsedValue, actualErr := strconv.ParseUint(value, 16, 8)
	if actualErr != nil {
		return 0, false
	}

	return float64(parsedValue) / 255, true
}

func linearizedColorChannel(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}

	return math.Pow((value+0.055)/1.055, 2.4)
}

func contrastRatio(left float64, right float64) float64 {
	lighter := math.Max(left, right)
	darker := math.Min(left, right)
	return (lighter + 0.05) / (darker + 0.05)
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
