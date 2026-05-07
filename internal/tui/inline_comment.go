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
	return renderPullRequestInlineCommentThreadContent(thread, renderer, width)
}

func renderPullRequestInlineCommentThreadContent(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int) string {
	comment := pullRequestInlineCommentFromThread(thread)
	lines := []string{renderPullRequestInlineCommentLocationLine(comment)}
	threadWidth := normalizedInlineThreadCommentBoxWidth(width)
	if len(thread.Comments) == 0 {
		lines = append(lines, renderRoundedCommentBox("No comments in thread.", threadWidth))
		return strings.Join(lines, "\n")
	}

	for _, commentBox := range renderInlineThreadCommentBoxes(thread.Comments, renderer, threadWidth, inlineThreadMetadataBadges(thread.IsResolved, thread.IsOutdated, thread.Comments)) {
		lines = append(lines, commentBox)
	}
	return strings.Join(lines, "\n")
}

func renderInlineThreadCommentBoxes(comments []githubcli.PullRequestComment, renderer MarkdownRenderer, width int, badges []commentMetadataBadge) []string {
	threadWidth := normalizedInlineThreadCommentBoxWidth(width)
	commentBodyWidth := commentBoxInnerWidth(threadWidth)
	renderedComments := make([]string, 0, len(comments))
	for index, threadComment := range comments {
		body := renderInlineCommentBody(threadComment.Body, renderer, commentBodyWidth)
		commentBadges := []commentMetadataBadge(nil)
		if index == 0 {
			commentBadges = badges
		}
		renderedComments = append(renderedComments, renderCommentBoxWithMetadataBadges(threadComment.Author, threadComment.CreatedAt, commentBadges, body, threadWidth))
	}
	return renderedComments
}

func normalizedInlineThreadCommentBoxWidth(width int) int {
	if width < minimumMarkdownRenderWidth {
		return defaultDetailWrapWidth
	}
	return width
}

func inlineThreadMetadataBadges(resolved bool, outdated bool, comments []githubcli.PullRequestComment) []commentMetadataBadge {
	badges := make([]commentMetadataBadge, 0, 3)
	if pullRequestCommentsContainPendingState(comments) {
		badges = append(badges, pendingInlineThreadMetadataBadge())
	}
	badges = append(badges, inlineThreadResolutionMetadataBadge(resolved))
	if outdated {
		badges = append(badges, outdatedInlineThreadMetadataBadge())
	}
	return badges
}

func pullRequestCommentsContainPendingState(comments []githubcli.PullRequestComment) bool {
	for _, comment := range comments {
		if strings.EqualFold(strings.TrimSpace(comment.State), "PENDING") {
			return true
		}
	}
	return false
}

func pendingInlineThreadMetadataBadge() commentMetadataBadge {
	return commentMetadataBadge{Label: "Pending", ForegroundHex: theme.PendingHex, BackgroundHex: theme.SelectedLineBackgroundHex}
}

func inlineThreadResolutionMetadataBadge(resolved bool) commentMetadataBadge {
	if resolved {
		return commentMetadataBadge{Label: "Resolved", ForegroundHex: theme.DiffAdditionHex, BackgroundHex: theme.DiffAdditionBackgroundHex}
	}
	return commentMetadataBadge{Label: "Unresolved", ForegroundHex: theme.DiffDeletionHex, BackgroundHex: theme.DiffDeletionBackgroundHex}
}

func outdatedInlineThreadMetadataBadge() commentMetadataBadge {
	return commentMetadataBadge{Label: "Outdated", ForegroundHex: theme.DiffHunkHeaderHex, BackgroundHex: theme.SelectedLineBackgroundHex}
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
	additionText := styleText(fmt.Sprintf("+%d", additions), foregroundColorEscape(theme.DiffAdditionHex))
	deletionText := styleText(fmt.Sprintf("-%d", deletions), foregroundColorEscape(theme.DiffDeletionHex))

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
