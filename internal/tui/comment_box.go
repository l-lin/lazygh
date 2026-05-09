package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

const commentBoxHorizontalPadding = 1

type commentMetadataBadge struct {
	Label         string
	ForegroundHex string
	BackgroundHex string
}

func renderPullRequestCommentSection(comment githubcli.PullRequestComment, body string, width int) string {
	return renderPullRequestCommentSectionForViewer(comment, body, width, "")
}

func renderPullRequestCommentSectionForViewer(comment githubcli.PullRequestComment, body string, width int, connectedUserLogin string) string {
	return renderCommentBoxWithMetadataForViewer(comment.Author, comment.CreatedAt, comment.ReactionGroups, body, width, connectedUserLogin)
}

func renderCommentBoxWithMetadata(author *githubcli.PullRequestCommentAuthor, createdAt string, reactionGroups []githubcli.ReactionGroup, body string, width int) string {
	return renderCommentBoxWithMetadataForViewer(author, createdAt, reactionGroups, body, width, "")
}

func renderCommentBoxWithMetadataForViewer(author *githubcli.PullRequestCommentAuthor, createdAt string, reactionGroups []githubcli.ReactionGroup, body string, width int, connectedUserLogin string) string {
	return renderCommentBoxWithMetadataBadgesForViewer(author, createdAt, nil, reactionGroups, body, width, connectedUserLogin)
}

func renderCommentBoxWithMetadataBadges(author *githubcli.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubcli.ReactionGroup, body string, width int) string {
	return renderCommentBoxWithMetadataBadgesForViewer(author, createdAt, badges, reactionGroups, body, width, "")
}

func renderCommentBoxWithMetadataBadgesForViewer(author *githubcli.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubcli.ReactionGroup, body string, width int, connectedUserLogin string) string {
	metadataLine := renderCommentBoxMetadataLineForViewer(author, createdAt, badges, reactionGroups, connectedUserLogin)
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
	return renderCommentBoxMetadataLineForViewer(author, createdAt, badges, reactionGroups, "")
}

func renderCommentBoxMetadataLineForViewer(author *githubcli.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubcli.ReactionGroup, connectedUserLogin string) string {
	segments := make([]string, 0, len(badges)+3)
	if authorBadge := renderCommentAuthorBadgeForViewer(author, connectedUserLogin); authorBadge != "" {
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
	return renderCommentAuthorBadgeForViewer(author, "")
}

func renderCommentAuthorBadgeForViewer(author *githubcli.PullRequestCommentAuthor, connectedUserLogin string) string {
	badgeText := commentAuthorBadgeText(author)
	if badgeText == "" {
		return ""
	}
	return styleCommentAuthorBadgeTextForViewer(badgeText, commentAuthorLogin(author), connectedUserLogin)
}

func commentAuthorLogin(author *githubcli.PullRequestCommentAuthor) string {
	if author == nil {
		return ""
	}
	return strings.TrimSpace(author.Login)
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
	return renderRoundedPill(label, badge.ForegroundHex, badge.BackgroundHex)
}

func commentAuthorBadgeText(author *githubcli.PullRequestCommentAuthor) string {
	return strings.TrimSpace(detailCommentsIcon) + "  " + pullRequestCommentAuthorLogin(author)
}

func styleCommentAuthorBadgeText(text string) string {
	return styleCommentAuthorBadgeTextForViewer(text, "", "")
}

func styleCommentAuthorBadgeTextForViewer(text string, authorLogin string, connectedUserLogin string) string {
	foregroundHex, backgroundHex := commentAuthorBadgeColors(authorLogin, connectedUserLogin)
	return renderRoundedPill(text, foregroundHex, backgroundHex)
}

func commentAuthorBadgeColors(authorLogin string, connectedUserLogin string) (string, string) {
	trimmedConnectedUserLogin := strings.TrimSpace(connectedUserLogin)
	if trimmedConnectedUserLogin != "" && !strings.EqualFold(strings.TrimSpace(authorLogin), trimmedConnectedUserLogin) {
		return theme.PendingHex, theme.PendingBackgroundHex
	}
	return theme.CommentAuthorBadgeHex, theme.CommentAuthorBadgeBackgroundHex
}

func styleCommentMetadataText(text string) string {
	return styleText(text, foregroundColorEscape(theme.DiffLineNumberHex))
}
