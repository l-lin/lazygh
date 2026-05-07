package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const inlineThreadConversationMinimumDiffLines = 5

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
	header := renderPullRequestInlineCommentThreadHeader(thread, false, width)
	body := renderPullRequestInlineCommentThreadBody(thread, renderer, width)
	if strings.TrimSpace(body) == "" {
		return header
	}
	return header + "\n" + body
}

func renderPullRequestInlineCommentThreadHeader(thread githubcli.PullRequestReviewThread, collapsed bool, width int) string {
	lines := []string{
		renderReviewDiffThreadHorizontalBorder(width),
		renderInlineThreadHeaderLine(
			pullRequestInlineCommentLocation(pullRequestInlineCommentFromThread(thread)),
			collapsed,
			inlineThreadStatusBadges(thread.IsResolved, thread.IsOutdated, thread.Comments),
		),
	}
	if collapsed {
		lines = append(lines, renderReviewDiffThreadHorizontalBorder(width))
	}
	return strings.Join(lines, "\n")
}

func renderPullRequestInlineCommentThreadBody(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int) string {
	threadWidth := normalizedInlineThreadCommentBoxWidth(width)
	lines := make([]string, 0, len(thread.Comments)+2)
	if diffPreview := renderPullRequestInlineCommentThreadDiffPreview(pullRequestInlineCommentFromThread(thread)); diffPreview != "" {
		lines = append(lines, diffPreview)
	}
	if len(thread.Comments) == 0 {
		lines = append(lines, renderRoundedCommentBox("No comments in thread.", threadWidth), renderReviewDiffThreadHorizontalBorder(width))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, renderInlineThreadCommentBoxes(thread.Comments, renderer, threadWidth)...)
	lines = append(lines, renderReviewDiffThreadHorizontalBorder(width))
	return strings.Join(lines, "\n")
}

func renderPullRequestInlineCommentThreadDiffPreview(comment githubcli.PullRequestInlineComment) string {
	previewLines := parseDiffPreviewLines(comment.DiffHunk)
	if len(previewLines) == 0 {
		return ""
	}

	markTargetDiffPreviewLines(previewLines, comment)
	previewLines = trimDiffPreviewLinesForConversation(previewLines, comment, inlineThreadConversationMinimumDiffLines)
	changedRangesByLine := diffPreviewChangedStyleRanges(previewLines)
	numberWidth := diffPreviewLineNumberWidth(previewLines)
	renderedLines := make([]string, 0, len(previewLines))
	for lineIndex, previewLine := range previewLines {
		renderedLines = append(renderedLines, renderDiffPreviewLine(comment.Path, previewLine, numberWidth, changedRangesByLine[lineIndex]))
	}
	return strings.Join(renderedLines, "\n")
}

func renderInlineThreadCommentBoxes(comments []githubcli.PullRequestComment, renderer MarkdownRenderer, width int) []string {
	threadWidth := normalizedInlineThreadCommentBoxWidth(width)
	commentBodyWidth := commentBoxInnerWidth(threadWidth)
	renderedComments := make([]string, 0, len(comments))
	for _, threadComment := range comments {
		body := renderInlineCommentBody(threadComment.Body, renderer, commentBodyWidth)
		renderedComments = append(renderedComments, renderCommentBoxWithMetadata(threadComment.Author, threadComment.CreatedAt, body, threadWidth))
	}
	return renderedComments
}

func normalizedInlineThreadCommentBoxWidth(width int) int {
	if width < minimumMarkdownRenderWidth {
		return defaultDetailWrapWidth
	}
	return width
}

func inlineThreadStatusBadges(resolved bool, outdated bool, comments []githubcli.PullRequestComment) []commentMetadataBadge {
	badges := make([]commentMetadataBadge, 0, 3)
	if pullRequestCommentsContainPendingState(comments) {
		badges = append(badges, pendingInlineThreadStatusBadge())
	}
	badges = append(badges, inlineThreadResolutionStatusBadge(resolved))
	if outdated {
		badges = append(badges, outdatedInlineThreadStatusBadge())
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

func pendingInlineThreadStatusBadge() commentMetadataBadge {
	return commentMetadataBadge{Label: "Pending", ForegroundHex: theme.PendingHex, BackgroundHex: theme.PendingBackgroundHex}
}

func inlineThreadResolutionStatusBadge(resolved bool) commentMetadataBadge {
	if resolved {
		return commentMetadataBadge{Label: "Resolved", ForegroundHex: theme.DiffAdditionHex, BackgroundHex: theme.DiffAdditionBackgroundHex}
	}
	return commentMetadataBadge{Label: "Unresolved", ForegroundHex: theme.DiffDeletionHex, BackgroundHex: theme.DiffDeletionBackgroundHex}
}

func outdatedInlineThreadStatusBadge() commentMetadataBadge {
	return commentMetadataBadge{Label: "Outdated", ForegroundHex: theme.DiffHunkHeaderHex, BackgroundHex: theme.SelectedLineBackgroundHex}
}

func renderInlineThreadHeaderLine(location string, collapsed bool, badges []commentMetadataBadge) string {
	chevron := browserDetailExpandedChevron
	if collapsed {
		chevron = browserDetailCollapsedChevron
	}

	segments := []string{styleText(chevron, foregroundColorEscape(theme.DiffHunkHeaderHex))}
	if trimmedLocation := strings.TrimSpace(location); trimmedLocation != "" {
		segments = append(segments, styleText(trimmedLocation, foregroundColorEscape(theme.DiffHunkHeaderHex)))
	}
	for _, badge := range badges {
		if renderedBadge := renderInlineThreadHeaderBadge(badge); renderedBadge != "" {
			segments = append(segments, renderedBadge)
		}
	}
	return strings.Join(segments, " ")
}

func renderInlineThreadHeaderBadge(badge commentMetadataBadge) string {
	label := strings.TrimSpace(badge.Label)
	if label == "" {
		return ""
	}
	if strings.TrimSpace(badge.ForegroundHex) == "" || strings.TrimSpace(badge.BackgroundHex) == "" {
		return label
	}
	return renderRoundedPill(label, badge.ForegroundHex, badge.BackgroundHex)
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

func pullRequestInlineCommentFromReviewDiffThread(thread reviewDiffThread) githubcli.PullRequestInlineComment {
	comment := githubcli.PullRequestInlineComment{
		Path:              strings.TrimSpace(thread.Path),
		Line:              thread.Line,
		OriginalLine:      thread.OriginalLine,
		StartLine:         thread.StartLine,
		OriginalStartLine: thread.OriginalStartLine,
		Side:              string(thread.Side),
		StartSide:         string(thread.StartSide),
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

func pullRequestInlineCommentSideLabel(comment githubcli.PullRequestInlineComment) string {
	_, _, side := pullRequestInlineCommentTargetRange(comment)
	switch side {
	case "LEFT":
		return "L"
	case "RIGHT":
		return "R"
	default:
		return "?"
	}
}

func pullRequestInlineCommentDisplayLine(comment githubcli.PullRequestInlineComment) int {
	_, endLine, _ := pullRequestInlineCommentTargetRange(comment)
	return endLine
}
