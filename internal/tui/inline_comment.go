package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func renderPullRequestInlineCommentSection(comment githubcli.PullRequestInlineComment, body string, width int) string {
	lines := []string{renderPullRequestInlineCommentLocationLine(comment)}

	diffPreview := renderPullRequestInlineCommentDiffPreview(comment)
	if diffPreview != "" {
		lines = append(lines, diffPreview)
	}
	lines = append(lines, renderCommentBoxWithMetadata(comment.Author, comment.CreatedAt, body, width))
	return strings.Join(lines, "\n")
}

func renderPullRequestInlineCommentThreadSection(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int) string {
	comment := pullRequestInlineCommentFromThread(thread)
	lines := []string{renderPullRequestInlineCommentLocationLine(comment), renderPullRequestInlineCommentThreadStatusLine(thread)}

	diffPreview := renderPullRequestInlineCommentDiffPreview(comment)
	if diffPreview != "" {
		lines = append(lines, diffPreview)
	}

	threadWidth := width
	if threadWidth < minimumMarkdownRenderWidth {
		threadWidth = defaultDetailWrapWidth
	}
	commentBodyWidth := commentBoxInnerWidth(threadWidth)
	if len(thread.Comments) == 0 {
		lines = append(lines, renderRoundedCommentBox("No comments in thread.", threadWidth))
		return strings.Join(lines, "\n")
	}

	for _, threadComment := range thread.Comments {
		body := renderInlineCommentBody(threadComment.Body, renderer, commentBodyWidth)
		lines = append(lines, renderCommentBoxWithMetadata(threadComment.Author, threadComment.CreatedAt, body, threadWidth))
	}
	return strings.Join(lines, "\n")
}

func renderPullRequestInlineCommentThreadStatusLine(thread githubcli.PullRequestReviewThread) string {
	badges := []string{renderPullRequestInlineCommentResolutionBadge(thread.IsResolved)}
	if thread.IsOutdated {
		badges = append(badges, renderPullRequestInlineCommentOutdatedBadge())
	}
	return strings.Join(badges, " ")
}

func renderPullRequestInlineCommentResolutionBadge(resolved bool) string {
	if resolved {
		return styleText(" Resolved ", foregroundColorEscape(theme.DiffAdditionForegroundHex), backgroundColorEscape(theme.DiffAdditionBackgroundHex))
	}
	return styleText(" Unresolved ", foregroundColorEscape(theme.DiffDeletionForegroundHex), backgroundColorEscape(theme.DiffDeletionBackgroundHex))
}

func renderPullRequestInlineCommentOutdatedBadge() string {
	return styleText(" Outdated ", foregroundColorEscape(theme.DiffHunkHeaderHex), backgroundColorEscape(theme.SelectedLineBackgroundHex))
}

func pullRequestInlineCommentFromThread(thread githubcli.PullRequestReviewThread) githubcli.PullRequestInlineComment {
	comment := githubcli.PullRequestInlineComment{
		Path:              strings.TrimSpace(thread.Path),
		Line:              thread.Line,
		OriginalLine:      thread.OriginalLine,
		StartLine:         thread.StartLine,
		OriginalStartLine: thread.OriginalStartLine,
		Side:              strings.TrimSpace(thread.DiffSide),
		StartSide:         strings.TrimSpace(thread.StartDiffSide),
	}
	for _, threadComment := range thread.Comments {
		if strings.TrimSpace(threadComment.DiffHunk) == "" {
			continue
		}
		comment.DiffHunk = threadComment.DiffHunk
		break
	}
	return comment
}

func renderPullRequestInlineCommentLocationLine(comment githubcli.PullRequestInlineComment) string {
	additions, deletions := diffHunkChangeCounts(comment.DiffHunk)
	location := pullRequestInlineCommentLocation(comment)
	additionText := styleText(fmt.Sprintf("+%d", additions), foregroundColorEscape(theme.DiffAdditionForegroundHex))
	deletionText := styleText(fmt.Sprintf("-%d", deletions), foregroundColorEscape(theme.DiffDeletionForegroundHex))

	segments := []string{styleText(detailInlineCommentLocationIcon, foregroundColorEscape(theme.DiffLineNumberHex))}
	if location != "" {
		segments = append(segments, location)
	}
	segments = append(segments, additionText, deletionText)
	if len(segments) == 0 {
		return ""
	}

	return segments[0] + " " + strings.Join(segments[1:], "  ")
}

func renderPullRequestInlineCommentDiffPreview(comment githubcli.PullRequestInlineComment) string {
	previewLines := parseDiffPreviewLines(comment.DiffHunk)
	if len(previewLines) == 0 {
		return styleText("No diff preview available.", foregroundColorEscape(theme.DiffHunkHeaderHex))
	}

	markTargetDiffPreviewLines(previewLines, comment)
	changedRangesByLine := diffPreviewChangedStyleRanges(previewLines)
	numberWidth := diffPreviewLineNumberWidth(previewLines)
	renderedLines := make([]string, 0, len(previewLines))
	for lineIndex, previewLine := range previewLines {
		renderedLines = append(renderedLines, renderDiffPreviewLine(comment.Path, previewLine, numberWidth, changedRangesByLine[lineIndex]))
	}
	return strings.Join(renderedLines, "\n")
}

func pullRequestInlineCommentLocation(comment githubcli.PullRequestInlineComment) string {
	path := strings.TrimSpace(comment.Path)
	if path == "" {
		return ""
	}

	startLine, endLine, _ := pullRequestInlineCommentTargetRange(comment)
	switch {
	case startLine > 0 && endLine > 0 && startLine != endLine:
		return fmt.Sprintf("%s:%d-%d", path, startLine, endLine)
	case endLine > 0:
		return fmt.Sprintf("%s:%d", path, endLine)
	case startLine > 0:
		return fmt.Sprintf("%s:%d", path, startLine)
	default:
		return path
	}
}

func pullRequestInlineCommentTargetRange(comment githubcli.PullRequestInlineComment) (int, int, string) {
	side := strings.ToUpper(strings.TrimSpace(comment.Side))
	if side == "LEFT" {
		startLine := firstPositive(comment.OriginalStartLine, comment.OriginalLine, comment.StartLine, comment.Line)
		endLine := firstPositive(comment.OriginalLine, comment.OriginalStartLine, comment.Line, comment.StartLine)
		return startLine, endLine, side
	}

	startLine := firstPositive(comment.StartLine, comment.Line, comment.OriginalStartLine, comment.OriginalLine)
	endLine := firstPositive(comment.Line, comment.StartLine, comment.OriginalLine, comment.OriginalStartLine)
	return startLine, endLine, "RIGHT"
}
