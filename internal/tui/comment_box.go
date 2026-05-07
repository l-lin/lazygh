package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const commentBoxHorizontalPadding = 1

type commentMetadataBadge struct {
	Label         string
	ForegroundHex string
	BackgroundHex string
}

func renderPullRequestCommentSection(comment githubcli.PullRequestComment, body string, width int) string {
	return renderCommentBoxWithMetadata(comment.Author, comment.CreatedAt, comment.ReactionGroups, body, width)
}

func renderCommentBoxWithMetadata(author *githubcli.PullRequestCommentAuthor, createdAt string, reactionGroups []githubcli.ReactionGroup, body string, width int) string {
	return renderCommentBoxWithMetadataBadges(author, createdAt, nil, reactionGroups, body, width)
}

func renderCommentBoxWithMetadataBadges(author *githubcli.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubcli.ReactionGroup, body string, width int) string {
	metadataLine := renderCommentBoxMetadataLine(author, createdAt, badges, reactionGroups)
	innerWidth := maxInt(commentBoxInnerWidth(width), maxStyledTextLineWidth(metadataLine))
	innerWidth = maxInt(innerWidth, maxStyledTextLineWidth(body))

	contentLines := make([]string, 0, 2)
	if metadataLine != "" {
		contentLines = append(contentLines, metadataLine)
	}
	if body != "" {
		contentLines = append(contentLines, body)
	}
	if len(contentLines) == 0 {
		contentLines = append(contentLines, "")
	}
	return renderRoundedCommentBoxWithInnerWidth(strings.Join(contentLines, "\n"), innerWidth)
}

func renderRoundedCommentBox(text string, width int) string {
	innerWidth := maxInt(commentBoxInnerWidth(width), maxStyledTextLineWidth(text))
	return renderRoundedCommentBoxWithInnerWidth(text, innerWidth)
}

func renderRoundedCommentBoxWithInnerWidth(text string, innerWidth int) string {
	if innerWidth < 1 {
		innerWidth = 1
	}

	styledLines := splitStyledTextLines(text)
	boxLines := make([]string, 0, len(styledLines)+2)
	boxLines = append(boxLines, styleCommentBorder("╭"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╮"))
	for _, line := range styledLines {
		paddingPrefix := styledTextLinePaddingPrefix(line, innerWidth)
		boxLines = append(boxLines, styleCommentBorder("│")+renderStyledPadding(paddingPrefix, commentBoxHorizontalPadding)+renderStyledTextLineWithWidth(line, innerWidth)+renderStyledPadding(paddingPrefix, commentBoxHorizontalPadding)+styleCommentBorder("│"))
	}
	boxLines = append(boxLines, styleCommentBorder("╰"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╯"))
	return strings.Join(boxLines, "\n")
}

func maxStyledTextLineWidth(text string) int {
	maximumWidth := 0
	for _, line := range splitStyledTextLines(text) {
		maximumWidth = maxInt(maximumWidth, len(line.runes))
	}
	return maximumWidth
}

func commentBoxInnerWidth(width int) int {
	innerWidth := effectiveMarkdownWidth(width) - ((commentBoxHorizontalPadding * 2) + 2)
	if innerWidth < 1 {
		return 1
	}

	return innerWidth
}

func styleCommentBorder(text string) string {
	return foregroundColorEscape(theme.InactiveBorderHex) + text + ansiReset
}

func renderCommentBoxMetadataLine(author *githubcli.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubcli.ReactionGroup) string {
	segments := make([]string, 0, len(badges)+3)
	if authorBadge := renderCommentAuthorBadge(author); authorBadge != "" {
		segments = append(segments, authorBadge)
	}
	if timestamp := renderCommentMetadataTimestamp(createdAt); timestamp != "" {
		segments = append(segments, timestamp)
	}
	for _, badge := range badges {
		if renderedBadge := renderCommentMetadataBadge(badge); renderedBadge != "" {
			segments = append(segments, renderedBadge)
		}
	}
	if renderedReactionGroups := renderReactionGroups(reactionGroups); renderedReactionGroups != "" {
		segments = append(segments, renderedReactionGroups)
	}
	return strings.Join(segments, "  ")
}

func renderCommentAuthorBadge(author *githubcli.PullRequestCommentAuthor) string {
	badgeText := commentAuthorBadgeText(author)
	if badgeText == "" {
		return ""
	}
	return styleCommentAuthorBadgeText(badgeText)
}

func renderCommentMetadataTimestamp(createdAt string) string {
	timestamp := formatTimestamp(createdAt)
	if timestamp == "" {
		return ""
	}
	return styleCommentMetadataText(timestamp)
}

func renderCommentMetadataBadge(badge commentMetadataBadge) string {
	label := strings.TrimSpace(badge.Label)
	if label == "" {
		return ""
	}
	if strings.TrimSpace(badge.ForegroundHex) == "" || strings.TrimSpace(badge.BackgroundHex) == "" {
		return label
	}
	return styleText(" "+label+" ", foregroundColorEscape(badge.ForegroundHex), backgroundColorEscape(badge.BackgroundHex))
}

func commentAuthorBadgeText(author *githubcli.PullRequestCommentAuthor) string {
	return strings.TrimSpace(detailCommentsIcon) + "  " + pullRequestCommentAuthorLogin(author)
}

func styleCommentAuthorBadgeText(text string) string {
	return renderRoundedPill(text, theme.CommentAuthorBadgeHex, theme.CommentAuthorBadgeBackgroundHex)
}

func styleCommentMetadataText(text string) string {
	return styleText(text, foregroundColorEscape(theme.DiffLineNumberHex))
}
