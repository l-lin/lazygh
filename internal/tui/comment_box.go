package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const commentBoxHorizontalPadding = 1

func renderPullRequestCommentSection(comment githubcli.PullRequestComment, body string, width int) string {
	return renderCommentBoxWithMetadata(comment.Author, comment.CreatedAt, body, width)
}

func renderCommentBoxWithMetadata(author *githubcli.PullRequestCommentAuthor, createdAt string, body string, width int) string {
	innerWidth := commentBoxInnerWidth(width)
	content := strings.Join([]string{
		renderCommentBoxMetadataLine(author, createdAt, innerWidth),
		body,
	}, "\n")
	return renderRoundedCommentBox(content, width)
}

func renderRoundedCommentBox(text string, width int) string {
	innerWidth := commentBoxInnerWidth(width)
	styledLines := splitStyledTextLines(text)

	boxLines := make([]string, 0, len(styledLines)+2)
	boxLines = append(boxLines, styleCommentBorder("╭"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╮"))
	for _, line := range styledLines {
		paddingWidth := innerWidth - len(line.runes)
		if paddingWidth < 0 {
			paddingWidth = 0
		}
		boxLines = append(boxLines, styleCommentBorder("│")+strings.Repeat(" ", commentBoxHorizontalPadding)+renderStyledTextLine(line)+strings.Repeat(" ", paddingWidth)+strings.Repeat(" ", commentBoxHorizontalPadding)+styleCommentBorder("│"))
	}
	boxLines = append(boxLines, styleCommentBorder("╰"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╯"))
	return strings.Join(boxLines, "\n")
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

func renderCommentBoxMetadataLine(author *githubcli.PullRequestCommentAuthor, createdAt string, width int) string {
	badgeText := commentAuthorBadgeText(author)
	metadataWidth := width - roundedPillAdornmentWidth
	if metadataWidth < 1 {
		metadataWidth = width
	}

	metadataText := compactCommentMetadataText(badgeText+"  "+formatTimestamp(createdAt), metadataWidth)
	metadataRunes := []rune(metadataText)
	badgeWidth := minInt(len(metadataRunes), runeCountInt(badgeText))
	renderedBadgeText := string(metadataRunes[:badgeWidth])
	remainderText := string(metadataRunes[badgeWidth:])
	if metadataWidth == width {
		return styleCommentAuthorBadgeTextWithoutRoundedPill(renderedBadgeText) + styleCommentMetadataText(remainderText)
	}

	return styleCommentAuthorBadgeText(renderedBadgeText) + styleCommentMetadataText(remainderText)
}

func commentAuthorBadgeText(author *githubcli.PullRequestCommentAuthor) string {
	return strings.TrimSpace(detailCommentsIcon) + "  " + pullRequestCommentAuthorLogin(author)
}

func styleCommentAuthorBadgeText(text string) string {
	return renderRoundedPill(text, theme.CommentAuthorBadgeForegroundHex, theme.CommentAuthorBadgeBackgroundHex)
}

func styleCommentAuthorBadgeTextWithoutRoundedPill(text string) string {
	return styleText(text, foregroundColorEscape(theme.CommentAuthorBadgeForegroundHex), backgroundColorEscape(theme.CommentAuthorBadgeBackgroundHex))
}

func styleCommentMetadataText(text string) string {
	return styleText(text, foregroundColorEscape(theme.DiffLineNumberHex))
}

func compactCommentMetadataText(text string, width int) string {
	trimmedText := strings.TrimSpace(text)
	if width <= 0 || len([]rune(trimmedText)) <= width {
		return trimmedText
	}
	if width == 1 {
		return "…"
	}

	runes := []rune(trimmedText)
	return string(runes[:width-1]) + "…"
}
