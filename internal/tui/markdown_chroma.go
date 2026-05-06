package tui

import (
	"sync"

	glamouransi "charm.land/glamour/v2/ansi"
	"github.com/alecthomas/chroma/v2"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
)

const markdownChromaStyleTheme = "charm"

var markdownChromaStyleRegistryMutex sync.Mutex

func registerMarkdownChromaStyle(style glamouransi.StyleConfig) {
	if style.CodeBlock.Chroma == nil {
		return
	}

	markdownChromaStyleRegistryMutex.Lock()
	defer markdownChromaStyleRegistryMutex.Unlock()

	chromastyles.Registry[markdownChromaStyleTheme] = chroma.MustNewStyle(markdownChromaStyleTheme, chroma.StyleEntries{
		chroma.Text:                markdownChromaStyleEntry(style.CodeBlock.Chroma.Text),
		chroma.Error:               markdownChromaStyleEntry(style.CodeBlock.Chroma.Error),
		chroma.Comment:             markdownChromaStyleEntry(style.CodeBlock.Chroma.Comment),
		chroma.CommentPreproc:      markdownChromaStyleEntry(style.CodeBlock.Chroma.CommentPreproc),
		chroma.Keyword:             markdownChromaStyleEntry(style.CodeBlock.Chroma.Keyword),
		chroma.KeywordReserved:     markdownChromaStyleEntry(style.CodeBlock.Chroma.KeywordReserved),
		chroma.KeywordNamespace:    markdownChromaStyleEntry(style.CodeBlock.Chroma.KeywordNamespace),
		chroma.KeywordType:         markdownChromaStyleEntry(style.CodeBlock.Chroma.KeywordType),
		chroma.Operator:            markdownChromaStyleEntry(style.CodeBlock.Chroma.Operator),
		chroma.Punctuation:         markdownChromaStyleEntry(style.CodeBlock.Chroma.Punctuation),
		chroma.Name:                markdownChromaStyleEntry(style.CodeBlock.Chroma.Name),
		chroma.NameBuiltin:         markdownChromaStyleEntry(style.CodeBlock.Chroma.NameBuiltin),
		chroma.NameTag:             markdownChromaStyleEntry(style.CodeBlock.Chroma.NameTag),
		chroma.NameAttribute:       markdownChromaStyleEntry(style.CodeBlock.Chroma.NameAttribute),
		chroma.NameClass:           markdownChromaStyleEntry(style.CodeBlock.Chroma.NameClass),
		chroma.NameConstant:        markdownChromaStyleEntry(style.CodeBlock.Chroma.NameConstant),
		chroma.NameDecorator:       markdownChromaStyleEntry(style.CodeBlock.Chroma.NameDecorator),
		chroma.NameException:       markdownChromaStyleEntry(style.CodeBlock.Chroma.NameException),
		chroma.NameFunction:        markdownChromaStyleEntry(style.CodeBlock.Chroma.NameFunction),
		chroma.NameOther:           markdownChromaStyleEntry(style.CodeBlock.Chroma.NameOther),
		chroma.Literal:             markdownChromaStyleEntry(style.CodeBlock.Chroma.Literal),
		chroma.LiteralNumber:       markdownChromaStyleEntry(style.CodeBlock.Chroma.LiteralNumber),
		chroma.LiteralDate:         markdownChromaStyleEntry(style.CodeBlock.Chroma.LiteralDate),
		chroma.LiteralString:       markdownChromaStyleEntry(style.CodeBlock.Chroma.LiteralString),
		chroma.LiteralStringEscape: markdownChromaStyleEntry(style.CodeBlock.Chroma.LiteralStringEscape),
		chroma.GenericDeleted:      markdownChromaStyleEntry(style.CodeBlock.Chroma.GenericDeleted),
		chroma.GenericEmph:         markdownChromaStyleEntry(style.CodeBlock.Chroma.GenericEmph),
		chroma.GenericInserted:     markdownChromaStyleEntry(style.CodeBlock.Chroma.GenericInserted),
		chroma.GenericStrong:       markdownChromaStyleEntry(style.CodeBlock.Chroma.GenericStrong),
		chroma.GenericSubheading:   markdownChromaStyleEntry(style.CodeBlock.Chroma.GenericSubheading),
		chroma.Background:          markdownChromaStyleEntry(style.CodeBlock.Chroma.Background),
	})
}

func markdownChromaStyleEntry(style glamouransi.StylePrimitive) string {
	entry := ""
	if style.Color != nil {
		entry = *style.Color
	}
	if style.BackgroundColor != nil {
		if entry != "" {
			entry += " "
		}
		entry += "bg:" + *style.BackgroundColor
	}
	if style.Italic != nil && *style.Italic {
		if entry != "" {
			entry += " "
		}
		entry += "italic"
	}
	if style.Bold != nil && *style.Bold {
		if entry != "" {
			entry += " "
		}
		entry += "bold"
	}
	if style.Underline != nil && *style.Underline {
		if entry != "" {
			entry += " "
		}
		entry += "underline"
	}

	return entry
}
