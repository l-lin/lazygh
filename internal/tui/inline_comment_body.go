package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

const inlineCommentSuggestionLineSentinelRune = '\u200b'

type inlineCommentMarkdownRenderPlan struct {
	markdown         string
	suggestionBlocks []inlineCommentSuggestionBlock
}

type inlineCommentSuggestionBlock struct {
	marker          string
	path            string
	originalLines   []string
	suggestionLines []string
}

func renderInlineCommentBody(markdown string, renderer MarkdownRenderer, width int) string {
	return renderInlineCommentBodyWithHTML(markdown, "", renderer, width)
}

func renderInlineCommentBodyWithHTML(markdown string, renderedHTML string, renderer MarkdownRenderer, width int) string {
	return renderInlineCommentBodyWithSuggestionContext(markdown, renderedHTML, githubdomain.PullRequestInlineComment{}, renderer, width)
}

func renderInlineCommentBodyForInlineComment(comment any, renderer MarkdownRenderer, width int) string {
	commentValue, ok := toDomainPullRequestInlineComment(comment)
	if !ok {
		return renderInlineCommentBodyWithSuggestionContext("", "", githubdomain.PullRequestInlineComment{}, renderer, width)
	}
	return renderInlineCommentBodyWithSuggestionContext(commentValue.Body, commentValue.BodyHTML, commentValue, renderer, width)
}

func renderInlineCommentBodyForThreadComment(comment githubdomain.PullRequestComment, suggestionContext any, renderer MarkdownRenderer, width int) string {
	suggestionContextValue, ok := toDomainPullRequestInlineComment(suggestionContext)
	if !ok {
		suggestionContextValue = githubdomain.PullRequestInlineComment{}
	}
	return renderInlineCommentBodyWithSuggestionContext(comment.Body, comment.BodyHTML, suggestionContextValue, renderer, width)
}

func renderInlineCommentBodyWithSuggestionContext(markdown string, renderedHTML string, suggestionContext githubdomain.PullRequestInlineComment, renderer MarkdownRenderer, width int) string {
	renderPlan := prepareInlineCommentMarkdownRenderPlan(markdown, suggestionContext)
	preparedMarkdown := prepareMarkdownForImageRendering(renderPlan.markdown, renderedHTML)
	renderedBody := renderMarkdownWithFallback(preparedMarkdown, renderer, width, "No comment body.")
	return renderInlineCommentSuggestionBlocks(renderedBody, renderPlan.suggestionBlocks)
}

func prepareInlineCommentMarkdown(markdown string) string {
	return prepareInlineCommentMarkdownWithSuggestionContext(markdown, githubdomain.PullRequestInlineComment{})
}

func prepareInlineCommentMarkdownWithSuggestionContext(markdown string, suggestionContext any) string {
	suggestionContextValue, ok := toDomainPullRequestInlineComment(suggestionContext)
	if !ok {
		suggestionContextValue = githubdomain.PullRequestInlineComment{}
	}
	normalized := strings.ReplaceAll(markdown, "\r", "")
	if !strings.Contains(normalized, "```") {
		return markdown
	}

	lines := strings.Split(normalized, "\n")
	preparedLines := make([]string, 0, len(lines))
	inFence := false
	fenceInfo := ""
	fenceLines := make([]string, 0)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !inFence {
			if strings.HasPrefix(trimmedLine, "```") {
				inFence = true
				fenceInfo = strings.TrimSpace(trimmedLine[3:])
				fenceLines = fenceLines[:0]
				continue
			}
			preparedLines = append(preparedLines, line)
			continue
		}
		if strings.HasPrefix(trimmedLine, "```") {
			preparedLines = append(preparedLines, prepareInlineCommentCodeFence(fenceInfo, fenceLines, suggestionContextValue)...)
			inFence = false
			fenceInfo = ""
			fenceLines = fenceLines[:0]
			continue
		}
		fenceLines = append(fenceLines, line)
	}
	if inFence {
		preparedLines = append(preparedLines, "```"+fenceInfo)
		preparedLines = append(preparedLines, fenceLines...)
	}

	return strings.TrimSpace(strings.Join(preparedLines, "\n"))
}

func prepareInlineCommentMarkdownRenderPlan(markdown string, suggestionContext githubdomain.PullRequestInlineComment) inlineCommentMarkdownRenderPlan {
	normalized := strings.ReplaceAll(markdown, "\r", "")
	if !strings.Contains(normalized, "```") {
		return inlineCommentMarkdownRenderPlan{markdown: markdown}
	}

	originalLines := inlineCommentSuggestionOriginalLines(suggestionContext)
	lines := strings.Split(normalized, "\n")
	preparedLines := make([]string, 0, len(lines))
	suggestionBlocks := make([]inlineCommentSuggestionBlock, 0)
	inFence := false
	fenceInfo := ""
	fenceLines := make([]string, 0)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !inFence {
			if strings.HasPrefix(trimmedLine, "```") {
				inFence = true
				fenceInfo = strings.TrimSpace(trimmedLine[3:])
				fenceLines = fenceLines[:0]
				continue
			}
			preparedLines = append(preparedLines, line)
			continue
		}
		if strings.HasPrefix(trimmedLine, "```") {
			if inlineCommentSuggestionFence(fenceInfo) {
				marker := inlineCommentSuggestionMarker(len(suggestionBlocks))
				preparedLines = append(preparedLines, marker)
				suggestionBlocks = append(suggestionBlocks, inlineCommentSuggestionBlock{
					marker:          marker,
					path:            strings.TrimSpace(suggestionContext.Path),
					originalLines:   append([]string(nil), originalLines...),
					suggestionLines: append([]string(nil), fenceLines...),
				})
			} else {
				preparedLines = append(preparedLines, prepareInlineCommentCodeFence(fenceInfo, fenceLines, suggestionContext)...)
			}
			inFence = false
			fenceInfo = ""
			fenceLines = fenceLines[:0]
			continue
		}
		fenceLines = append(fenceLines, line)
	}
	if inFence {
		preparedLines = append(preparedLines, "```"+fenceInfo)
		preparedLines = append(preparedLines, fenceLines...)
	}

	return inlineCommentMarkdownRenderPlan{markdown: strings.TrimSpace(strings.Join(preparedLines, "\n")), suggestionBlocks: suggestionBlocks}
}

func renderInlineCommentSuggestionBlocks(renderedBody string, suggestionBlocks []inlineCommentSuggestionBlock) string {
	if len(suggestionBlocks) == 0 || strings.TrimSpace(renderedBody) == "" {
		return renderedBody
	}

	replacements := make(map[string]string, len(suggestionBlocks))
	for _, suggestionBlock := range suggestionBlocks {
		replacements[suggestionBlock.marker] = renderInlineCommentSuggestionBlock(suggestionBlock)
	}

	renderedLines := strings.Split(renderedBody, "\n")
	for lineIndex, renderedLine := range renderedLines {
		visibleLine := inlineCommentRenderedLineText(renderedLine)
		replacement, ok := replacements[visibleLine]
		if !ok {
			continue
		}
		renderedLines[lineIndex] = replacement
	}

	return strings.Join(renderedLines, "\n")
}

func inlineCommentRenderedLineText(line string) string {
	styledLines := splitStyledTextLines(line)
	if len(styledLines) == 0 {
		return ""
	}
	return string(styledLines[0].runes)
}

func renderInlineCommentSuggestionBlock(suggestionBlock inlineCommentSuggestionBlock) string {
	deletionRanges, additionRanges := inlineCommentSuggestionChangedStyleRanges(suggestionBlock.originalLines, suggestionBlock.suggestionLines)
	renderedLines := make([]string, 0, len(suggestionBlock.originalLines)+len(suggestionBlock.suggestionLines)+1)
	renderedLines = append(renderedLines, renderInlineCommentSuggestionLabelLine())
	for lineIndex, originalLine := range suggestionBlock.originalLines {
		renderedLines = append(renderedLines, renderInlineCommentSuggestionLine(suggestionBlock.path, '-', originalLine, theme.DiffDeletionHex, deletionRanges[lineIndex]))
	}
	for lineIndex, suggestionLine := range suggestionBlock.suggestionLines {
		renderedLines = append(renderedLines, renderInlineCommentSuggestionLine(suggestionBlock.path, '+', suggestionLine, theme.DiffAdditionHex, additionRanges[lineIndex]))
	}
	return strings.Join(renderedLines, "\n")
}

func inlineCommentSuggestionChangedStyleRanges(originalLines []string, suggestionLines []string) ([][]styledRuneRange, [][]styledRuneRange) {
	deletionRanges := make([][]styledRuneRange, len(originalLines))
	additionRanges := make([][]styledRuneRange, len(suggestionLines))
	for lineIndex := range minInt(len(originalLines), len(suggestionLines)) {
		lineDeletionRanges, lineAdditionRanges := reviewDiffLineChangedStyleRanges(originalLines[lineIndex], suggestionLines[lineIndex])
		deletionRanges[lineIndex] = append(deletionRanges[lineIndex], lineDeletionRanges...)
		additionRanges[lineIndex] = append(additionRanges[lineIndex], lineAdditionRanges...)
	}
	return deletionRanges, additionRanges
}

func renderInlineCommentSuggestionLine(path string, sign rune, text string, foregroundHex string, changedRanges []styledRuneRange) string {
	basePrefix := foregroundColorEscape(foregroundHex) + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	return string(inlineCommentSuggestionLineSentinelRune) + styleText(string(sign), basePrefix) + renderSyntaxHighlightedCode(path, text, basePrefix, changedRanges)
}

func renderInlineCommentSuggestionLabelLine() string {
	return styleText("Suggestion", ansiBold)
}

func inlineCommentSuggestionMarker(index int) string {
	return fmt.Sprintf("§lazyghs%d§", index)
}

func prepareInlineCommentCodeFence(info string, lines []string, suggestionContext githubdomain.PullRequestInlineComment) []string {
	label := inlineCommentCodeBlockLabel(info)
	preparedLines := make([]string, 0, len(lines)+4)
	if label != "" {
		preparedLines = append(preparedLines, label, "")
	}
	if inlineCommentSuggestionFence(info) {
		preparedLines = append(preparedLines, "```diff")
		preparedLines = append(preparedLines, inlineCommentSuggestionDiffLines(lines, suggestionContext)...)
		preparedLines = append(preparedLines, "```")
		return preparedLines
	}

	preparedLines = append(preparedLines, "```"+strings.TrimSpace(info))
	preparedLines = append(preparedLines, lines...)
	preparedLines = append(preparedLines, "```")
	return preparedLines
}

func inlineCommentCodeBlockLabel(info string) string {
	language := inlineCommentCodeBlockLanguage(info)
	if inlineCommentSuggestionFence(info) {
		return "**Suggestion**"
	}
	if language == "" {
		return "**Code block**"
	}
	return fmt.Sprintf("**Code block · %s**", language)
}

func inlineCommentCodeBlockLanguage(info string) string {
	fields := strings.Fields(strings.TrimSpace(info))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func inlineCommentSuggestionFence(info string) bool {
	for _, field := range strings.Fields(strings.ToLower(strings.TrimSpace(info))) {
		if strings.HasPrefix(field, "suggestion") {
			return true
		}
	}
	return false
}

func inlineCommentSuggestionDiffLines(lines []string, suggestionContext githubdomain.PullRequestInlineComment) []string {
	originalLines := inlineCommentSuggestionOriginalLines(suggestionContext)
	if len(originalLines) == 0 {
		return prefixInlineCommentDiffLines(lines, "+")
	}

	diffLines := make([]string, 0, len(originalLines)+len(lines))
	diffLines = append(diffLines, prefixInlineCommentDiffLines(originalLines, "-")...)
	diffLines = append(diffLines, prefixInlineCommentDiffLines(lines, "+")...)
	return diffLines
}

func inlineCommentSuggestionOriginalLines(suggestionContext githubdomain.PullRequestInlineComment) []string {
	previewLines := parseDiffPreviewLines(suggestionContext.DiffHunk)
	if len(previewLines) == 0 {
		return nil
	}

	startLine, endLine, side := pullRequestInlineCommentTargetRange(suggestionContext)
	if startLine <= 0 && endLine <= 0 {
		return nil
	}
	if endLine == 0 {
		endLine = startLine
	}
	if startLine == 0 {
		startLine = endLine
	}
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}

	originalLines := make([]string, 0, endLine-startLine+1)
	for _, previewLine := range previewLines {
		if previewLine.kind == diffPreviewHunkHeaderLine || previewLine.kind == diffPreviewNoteLine {
			continue
		}
		lineNumber := conversationPreviewLineNumber(previewLine, side)
		if lineNumber < startLine || lineNumber > endLine {
			continue
		}
		originalLines = append(originalLines, previewLine.text)
	}
	return originalLines
}

func prefixInlineCommentDiffLines(lines []string, prefix string) []string {
	prefixed := make([]string, 0, len(lines))
	for _, line := range lines {
		prefixed = append(prefixed, prefix+line)
	}
	return prefixed
}
