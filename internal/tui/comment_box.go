package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const commentBoxHorizontalPadding = 1

func renderPullRequestCommentSection(comment githubcli.PullRequestComment, body string, width int) string {
	header := detailCommentsIcon + " " + pullRequestCommentAuthorLogin(comment.Author) + " · " + formatTimestamp(comment.CreatedAt)
	return header + "\n" + renderRoundedCommentBox(body, width)
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
