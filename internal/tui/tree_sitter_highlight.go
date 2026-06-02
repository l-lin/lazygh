package tui

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/l-lin/lazygh/internal/theme"

	tree_sitter_kotlin "github.com/tree-sitter-grammars/tree-sitter-kotlin/bindings/go"
	tree_sitter_lua "github.com/tree-sitter-grammars/tree-sitter-lua/bindings/go"
	tree_sitter_make "github.com/tree-sitter-grammars/tree-sitter-make/bindings/go"
	tree_sitter_toml "github.com/tree-sitter-grammars/tree-sitter-toml/bindings/go"
	tree_sitter_xml "github.com/tree-sitter-grammars/tree-sitter-xml/bindings/go"
	tree_sitter_yaml "github.com/tree-sitter-grammars/tree-sitter-yaml/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_json "github.com/tree-sitter/tree-sitter-json/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type styledRuneRange struct {
	start  int
	end    int
	prefix string
}

type treeSitterCaptureStyle struct {
	prefix   string
	priority int
}

type capturedStyleRange struct {
	styledRuneRange
	priority int
}

type treeSitterLanguageRuntime struct {
	language             *tree_sitter.Language
	highlightQuerySource string
	queryOnce            sync.Once
	query                *tree_sitter.Query
	queryErr             *tree_sitter.QueryError
	parserPool           sync.Pool
	cursorPool           sync.Pool
	syntaxRangeCache     *syntaxRangeCache
}

const defaultSyntaxRangeCacheCapacity = 1024

func newTreeSitterLanguageRuntime(language *tree_sitter.Language, highlightQuerySource string) *treeSitterLanguageRuntime {
	runtime := &treeSitterLanguageRuntime{
		language:             language,
		highlightQuerySource: highlightQuerySource,
		syntaxRangeCache:     newSyntaxRangeCache(defaultSyntaxRangeCacheCapacity),
	}
	runtime.parserPool.New = func() any {
		return runtime.newParser()
	}
	runtime.cursorPool.New = func() any {
		return tree_sitter.NewQueryCursor()
	}
	return runtime
}

var (
	treeSitterBashRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_bash.Language()), treeSitterBashHighlightsQuery)
	treeSitterGoRuntime         = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_go.Language()), treeSitterGoHighlightsQuery)
	treeSitterHTMLRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_html.Language()), treeSitterHTMLHighlightsQuery)
	treeSitterJavaRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_java.Language()), treeSitterJavaHighlightsQuery)
	treeSitterJavaScriptRuntime = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_javascript.Language()), treeSitterJavaScriptHighlightsQuery)
	treeSitterJSONRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_json.Language()), treeSitterJSONHighlightsQuery)
	treeSitterKotlinRuntime     = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_kotlin.Language()), treeSitterKotlinHighlightsQuery)
	treeSitterLuaRuntime        = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_lua.Language()), treeSitterLuaHighlightsQuery)
	treeSitterMakeRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_make.Language()), treeSitterMakeHighlightsQuery)
	treeSitterPythonRuntime     = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_python.Language()), treeSitterPythonHighlightsQuery)
	treeSitterRubyRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_ruby.Language()), treeSitterRubyHighlightsQuery)
	treeSitterTOMLRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_toml.Language()), treeSitterTOMLHighlightsQuery)
	treeSitterTypeScriptRuntime = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()), treeSitterTypeScriptHighlightsQuery)
	treeSitterXMLRuntime        = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_xml.LanguageXML()), treeSitterXMLHighlightsQuery)
	treeSitterYAMLRuntime       = newTreeSitterLanguageRuntime(tree_sitter.NewLanguage(tree_sitter_yaml.Language()), treeSitterYAMLHighlightsQuery)
)

func renderSyntaxHighlightedCode(path string, text string, basePrefix string, leadingRanges []styledRuneRange) string {
	syntaxRanges := treeSitterSyntaxRanges(path, text)
	if basePrefix == "" && len(leadingRanges) == 0 && len(syntaxRanges) == 0 {
		return text
	}

	combinedRanges := make([]styledRuneRange, 0, len(leadingRanges)+len(syntaxRanges))
	combinedRanges = append(combinedRanges, leadingRanges...)
	combinedRanges = append(combinedRanges, syntaxRanges...)
	return renderTextWithStyleRanges(text, basePrefix, combinedRanges)
}

func renderTextWithStyleRanges(text string, basePrefix string, ranges []styledRuneRange) string {
	if text == "" {
		return ""
	}
	if basePrefix == "" && len(ranges) == 0 {
		return text
	}

	runes := []rune(text)
	stylePrefixes := make([]string, len(runes))
	if basePrefix != "" {
		for index := range stylePrefixes {
			stylePrefixes[index] = basePrefix
		}
	}

	for _, styleRange := range ranges {
		if strings.TrimSpace(styleRange.prefix) == "" {
			continue
		}
		start := maxInt(0, styleRange.start)
		end := minInt(len(runes), styleRange.end)
		if start >= end {
			continue
		}
		for index := start; index < end; index++ {
			stylePrefixes[index] += styleRange.prefix
		}
	}

	return renderStyledTextLine(styledTextLine{runes: runes, stylePrefixes: stylePrefixes})
}

func treeSitterSyntaxRanges(path string, text string) []styledRuneRange {
	runtime, ok := treeSitterRuntimeForPath(path)
	if !ok {
		return nil
	}
	return runtime.syntaxRanges(text)
}

func treeSitterRuntimeForPath(path string) (*treeSitterLanguageRuntime, bool) {
	normalizedPath := strings.ToLower(strings.TrimSpace(path))
	if normalizedPath == "" {
		return nil, false
	}

	switch strings.ToLower(filepath.Ext(normalizedPath)) {
	case ".go":
		return treeSitterGoRuntime, true
	case ".html", ".htm":
		return treeSitterHTMLRuntime, true
	case ".java":
		return treeSitterJavaRuntime, true
	case ".js", ".cjs", ".mjs":
		return treeSitterJavaScriptRuntime, true
	case ".json":
		return treeSitterJSONRuntime, true
	case ".kt", ".kts":
		return treeSitterKotlinRuntime, true
	case ".lua":
		return treeSitterLuaRuntime, true
	case ".mk":
		return treeSitterMakeRuntime, true
	case ".py":
		return treeSitterPythonRuntime, true
	case ".rb":
		return treeSitterRubyRuntime, true
	case ".toml":
		return treeSitterTOMLRuntime, true
	case ".ts", ".cts", ".mts":
		return treeSitterTypeScriptRuntime, true
	case ".xml":
		return treeSitterXMLRuntime, true
	case ".yaml", ".yml":
		return treeSitterYAMLRuntime, true
	case ".bash", ".sh", ".zsh":
		return treeSitterBashRuntime, true
	}

	switch filepath.Base(normalizedPath) {
	case ".bash_profile", ".bashrc", ".profile", ".zshrc":
		return treeSitterBashRuntime, true
	case "makefile", "gnumakefile":
		return treeSitterMakeRuntime, true
	case "gemfile", "rakefile":
		return treeSitterRubyRuntime, true
	}

	return nil, false
}

func (runtime *treeSitterLanguageRuntime) highlightQuery() (*tree_sitter.Query, bool) {
	runtime.queryOnce.Do(func() {
		runtime.query, runtime.queryErr = tree_sitter.NewQuery(runtime.language, runtime.highlightQuerySource)
	})
	return runtime.query, runtime.queryErr == nil && runtime.query != nil
}

func (runtime *treeSitterLanguageRuntime) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()
	if parser == nil {
		return nil
	}
	if err := parser.SetLanguage(runtime.language); err != nil {
		parser.Close()
		return nil
	}
	return parser
}

func (runtime *treeSitterLanguageRuntime) borrowParser() *tree_sitter.Parser {
	parser, _ := runtime.parserPool.Get().(*tree_sitter.Parser)
	if parser != nil {
		return parser
	}
	return runtime.newParser()
}

func (runtime *treeSitterLanguageRuntime) releaseParser(parser *tree_sitter.Parser) {
	if parser == nil {
		return
	}
	parser.Reset()
	runtime.parserPool.Put(parser)
}

func (runtime *treeSitterLanguageRuntime) borrowQueryCursor() *tree_sitter.QueryCursor {
	cursor, _ := runtime.cursorPool.Get().(*tree_sitter.QueryCursor)
	if cursor != nil {
		return cursor
	}
	return tree_sitter.NewQueryCursor()
}

func (runtime *treeSitterLanguageRuntime) releaseQueryCursor(cursor *tree_sitter.QueryCursor) {
	if cursor == nil {
		return
	}
	runtime.cursorPool.Put(cursor)
}

func (runtime *treeSitterLanguageRuntime) syntaxRanges(text string) []styledRuneRange {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil
	}
	if cachedRanges, ok := runtime.syntaxRangeCache.Get(text); ok {
		return cachedRanges
	}

	query, ok := runtime.highlightQuery()
	if !ok {
		return nil
	}

	source := []byte(text)
	parser := runtime.borrowParser()
	if parser == nil {
		return nil
	}
	defer runtime.releaseParser(parser)

	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil
	}
	defer tree.Close()

	cursor := runtime.borrowQueryCursor()
	if cursor == nil {
		return nil
	}
	defer runtime.releaseQueryCursor(cursor)

	captureNames := query.CaptureNames()
	captures := cursor.Captures(query, tree.RootNode(), source)

	rangesBySpan := make(map[[2]int]capturedStyleRange)
	for match, captureIndex := captures.Next(); match != nil; match, captureIndex = captures.Next() {
		if int(captureIndex) >= len(match.Captures) {
			continue
		}

		capture := match.Captures[captureIndex]
		if int(capture.Index) >= len(captureNames) {
			continue
		}

		captureStyle, ok := treeSitterCaptureStyleForName(captureNames[capture.Index])
		if !ok {
			continue
		}

		start, end, ok := runeRangeForByteRange(text, int(capture.Node.StartByte()), int(capture.Node.EndByte()))
		if !ok || start >= end {
			continue
		}

		span := [2]int{start, end}
		existingRange, exists := rangesBySpan[span]
		if exists && existingRange.priority >= captureStyle.priority {
			continue
		}

		rangesBySpan[span] = capturedStyleRange{
			styledRuneRange: styledRuneRange{start: start, end: end, prefix: captureStyle.prefix},
			priority:        captureStyle.priority,
		}
	}

	if len(rangesBySpan) == 0 {
		runtime.syntaxRangeCache.Put(text, nil)
		return nil
	}

	ranges := make([]capturedStyleRange, 0, len(rangesBySpan))
	for _, styleRange := range rangesBySpan {
		ranges = append(ranges, styleRange)
	}
	sort.SliceStable(ranges, func(left int, right int) bool {
		if ranges[left].start != ranges[right].start {
			return ranges[left].start < ranges[right].start
		}
		leftWidth := ranges[left].end - ranges[left].start
		rightWidth := ranges[right].end - ranges[right].start
		if leftWidth != rightWidth {
			return leftWidth > rightWidth
		}
		return ranges[left].priority < ranges[right].priority
	})

	result := make([]styledRuneRange, 0, len(ranges))
	for _, styleRange := range ranges {
		result = append(result, styleRange.styledRuneRange)
	}
	runtime.syntaxRangeCache.Put(text, result)
	return result
}

func runeRangeForByteRange(text string, startByte int, endByte int) (int, int, bool) {
	if startByte < 0 || endByte < startByte || endByte > len(text) {
		return 0, 0, false
	}
	if !utf8.ValidString(text[:startByte]) || !utf8.ValidString(text[:endByte]) {
		return 0, 0, false
	}
	return utf8.RuneCountInString(text[:startByte]), utf8.RuneCountInString(text[:endByte]), true
}

func treeSitterCaptureStyleForName(name string) (treeSitterCaptureStyle, bool) {
	trimmedName := strings.TrimSpace(name)
	switch trimmedName {
	case "string.special.key":
		return newTreeSitterCaptureStyle(theme.SyntaxPropertyHex, 95), true
	case "markup.heading", "text.title":
		return newTreeSitterCaptureStyle(theme.MarkdownHeadingHex, 80), true
	case "markup.link", "markup.raw", "text.literal", "text.reference", "text.uri":
		return newTreeSitterCaptureStyle(theme.SyntaxStringHex, 60), true
	}

	primaryName := trimmedName
	if before, _, ok := strings.Cut(trimmedName, "."); ok {
		primaryName = before
	}

	switch primaryName {
	case "attribute", "label", "tag", "type":
		return newTreeSitterCaptureStyle(theme.SyntaxTypeHex, 80), true
	case "boolean", "constant", "number":
		return newTreeSitterCaptureStyle(theme.SyntaxNumberHex, 70), true
	case "comment":
		return newTreeSitterCaptureStyle(theme.SyntaxCommentHex, 20), true
	case "conditional", "exception", "include", "keyword", "operator", "preproc", "repeat":
		return newTreeSitterCaptureStyle(theme.SyntaxKeywordHex, 85), true
	case "constructor", "function", "method":
		return newTreeSitterCaptureStyle(theme.SyntaxFunctionHex, 90), true
	case "escape", "string":
		return newTreeSitterCaptureStyle(theme.SyntaxStringHex, 75), true
	case "field", "parameter", "property":
		return newTreeSitterCaptureStyle(theme.SyntaxPropertyHex, 80), true
	}

	return treeSitterCaptureStyle{}, false
}

func newTreeSitterCaptureStyle(hexColor string, priority int) treeSitterCaptureStyle {
	return treeSitterCaptureStyle{prefix: foregroundColorEscape(hexColor), priority: priority}
}
