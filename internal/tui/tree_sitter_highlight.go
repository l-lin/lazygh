package tui

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"codeberg.org/l-lin/lazygh/internal/theme"

	tree_sitter_toml "github.com/tree-sitter-grammars/tree-sitter-toml/bindings/go"
	tree_sitter_yaml "github.com/tree-sitter-grammars/tree-sitter-yaml/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_json "github.com/tree-sitter/tree-sitter-json/bindings/go"
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
}

var (
	treeSitterBashRuntime = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_bash.Language()), highlightQuerySource: treeSitterBashHighlightsQuery}
	treeSitterGoRuntime   = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_go.Language()), highlightQuerySource: treeSitterGoHighlightsQuery}
	treeSitterJavaRuntime = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_java.Language()), highlightQuerySource: treeSitterJavaHighlightsQuery}
	treeSitterJSONRuntime = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_json.Language()), highlightQuerySource: treeSitterJSONHighlightsQuery}
	treeSitterTOMLRuntime = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_toml.Language()), highlightQuerySource: treeSitterTOMLHighlightsQuery}
	treeSitterYAMLRuntime = &treeSitterLanguageRuntime{language: tree_sitter.NewLanguage(tree_sitter_yaml.Language()), highlightQuerySource: treeSitterYAMLHighlightsQuery}
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
	case ".java":
		return treeSitterJavaRuntime, true
	case ".json":
		return treeSitterJSONRuntime, true
	case ".toml":
		return treeSitterTOMLRuntime, true
	case ".yaml", ".yml":
		return treeSitterYAMLRuntime, true
	case ".bash", ".sh", ".zsh":
		return treeSitterBashRuntime, true
	}

	switch filepath.Base(normalizedPath) {
	case ".bash_profile", ".bashrc", ".profile", ".zshrc":
		return treeSitterBashRuntime, true
	}

	return nil, false
}

func (runtime *treeSitterLanguageRuntime) highlightQuery() (*tree_sitter.Query, bool) {
	runtime.queryOnce.Do(func() {
		runtime.query, runtime.queryErr = tree_sitter.NewQuery(runtime.language, runtime.highlightQuerySource)
	})
	return runtime.query, runtime.queryErr == nil && runtime.query != nil
}

func (runtime *treeSitterLanguageRuntime) syntaxRanges(text string) []styledRuneRange {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil
	}

	query, ok := runtime.highlightQuery()
	if !ok {
		return nil
	}

	source := []byte(text)
	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(runtime.language); err != nil {
		return nil
	}

	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil
	}
	defer tree.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()
	captures := cursor.Captures(query, tree.RootNode(), source)

	rangesBySpan := make(map[[2]int]capturedStyleRange)
	for match, captureIndex := captures.Next(); match != nil; match, captureIndex = captures.Next() {
		if int(captureIndex) >= len(match.Captures) {
			continue
		}

		capture := match.Captures[captureIndex]
		if int(capture.Index) >= len(query.CaptureNames()) {
			continue
		}

		captureStyle, ok := treeSitterCaptureStyleForName(query.CaptureNames()[capture.Index])
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
		return newTreeSitterCaptureStyle(theme.SyntaxPropertyHex, 70), true
	case "text.literal", "text.reference", "text.uri":
		return newTreeSitterCaptureStyle(theme.SyntaxStringHex, 60), true
	case "text.title":
		return newTreeSitterCaptureStyle(theme.MarkdownHeadingHex, 80), true
	}

	primaryName := trimmedName
	if separatorIndex := strings.IndexByte(trimmedName, '.'); separatorIndex >= 0 {
		primaryName = trimmedName[:separatorIndex]
	}

	switch primaryName {
	case "attribute", "label", "type":
		return newTreeSitterCaptureStyle(theme.SyntaxTypeHex, 80), true
	case "boolean", "constant", "number":
		return newTreeSitterCaptureStyle(theme.SyntaxNumberHex, 70), true
	case "comment":
		return newTreeSitterCaptureStyle(theme.SyntaxCommentHex, 20), true
	case "constructor", "function":
		return newTreeSitterCaptureStyle(theme.SyntaxFunctionHex, 90), true
	case "escape", "string":
		return newTreeSitterCaptureStyle(theme.SyntaxStringHex, 75), true
	case "keyword", "operator":
		return newTreeSitterCaptureStyle(theme.SyntaxKeywordHex, 85), true
	case "property":
		return newTreeSitterCaptureStyle(theme.SyntaxPropertyHex, 80), true
	}

	return treeSitterCaptureStyle{}, false
}

func newTreeSitterCaptureStyle(hexColor string, priority int) treeSitterCaptureStyle {
	return treeSitterCaptureStyle{prefix: foregroundColorEscape(hexColor), priority: priority}
}
