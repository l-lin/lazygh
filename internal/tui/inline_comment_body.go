package tui

import (
	"fmt"
	"strings"
)

func renderInlineCommentBody(markdown string, renderer MarkdownRenderer, _ int) string {
	return renderMarkdownWithFallback(prepareInlineCommentMarkdown(markdown), renderer, disabledMarkdownWordWrap, "No comment body.")
}

func prepareInlineCommentMarkdown(markdown string) string {
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
			preparedLines = append(preparedLines, prepareInlineCommentCodeFence(fenceInfo, fenceLines)...)
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

func prepareInlineCommentCodeFence(info string, lines []string) []string {
	label := inlineCommentCodeBlockLabel(info)
	preparedLines := make([]string, 0, len(lines)+4)
	if label != "" {
		preparedLines = append(preparedLines, label, "")
	}
	if inlineCommentSuggestionFence(info) {
		preparedLines = append(preparedLines, "```diff")
		preparedLines = append(preparedLines, prefixInlineCommentSuggestionLines(lines)...)
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
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(info)), "suggestion")
}

func prefixInlineCommentSuggestionLines(lines []string) []string {
	prefixed := make([]string, 0, len(lines))
	for _, line := range lines {
		prefixed = append(prefixed, "+"+line)
	}
	return prefixed
}
