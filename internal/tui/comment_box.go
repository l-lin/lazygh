package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

const commentBoxHorizontalPadding = 1

type commentMetadataBadge struct {
	Label         string
	ForegroundHex string
	BackgroundHex string
}

func renderPullRequestCommentSection(comment githubdomain.PullRequestComment, body string, width int) string {
	return renderPullRequestCommentSectionForViewer(comment, body, width, "")
}

func renderPullRequestCommentSectionForViewer(comment githubdomain.PullRequestComment, body string, width int, connectedUserLogin string) string {
	return renderCommentBoxWithMetadataForViewer(comment.Author, comment.CreatedAt, comment.ReactionGroups, body, width, connectedUserLogin)
}

func renderCommentBoxWithMetadata(author *githubdomain.PullRequestCommentAuthor, createdAt string, reactionGroups []githubdomain.ReactionGroup, body string, width int) string {
	return renderCommentBoxWithMetadataForViewer(author, createdAt, reactionGroups, body, width, "")
}

func renderCommentBoxWithMetadataForViewer(author *githubdomain.PullRequestCommentAuthor, createdAt string, reactionGroups []githubdomain.ReactionGroup, body string, width int, connectedUserLogin string) string {
	return renderCommentBoxWithMetadataBadgesForViewer(author, createdAt, nil, reactionGroups, body, width, connectedUserLogin)
}

func renderCommentBoxWithMetadataBadgesForViewer(author *githubdomain.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubdomain.ReactionGroup, body string, width int, connectedUserLogin string) string {
	metadataLine := renderCommentBoxMetadataLineForViewer(author, createdAt, badges, reactionGroups, connectedUserLogin)

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
	return renderRoundedCommentBoxWithInnerWidth(strings.Join(contentLines, "\n"), commentBoxInnerWidth(width))
}

func renderRoundedCommentBox(text string, width int) string {
	return renderRoundedCommentBoxWithInnerWidth(text, commentBoxInnerWidth(width))
}

func renderRoundedCommentBoxWithInnerWidth(text string, innerWidth int) string {
	if innerWidth < 1 {
		innerWidth = 1
	}

	styledLines := wrapCommentBoxStyledLines(splitStyledTextLines(text), innerWidth)
	innerWidth = maxInt(innerWidth, maxStyledTextLineWidthFromLines(styledLines))
	boxLines := make([]string, 0, len(styledLines)+2)
	boxLines = append(boxLines, styleCommentBorder("╭"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╮"))
	for _, line := range styledLines {
		paddingPrefix := styledTextLinePaddingPrefix(line, innerWidth)
		boxLines = append(boxLines, styleCommentBorder("│")+renderStyledPadding(paddingPrefix, commentBoxHorizontalPadding)+renderStyledTextLineWithWidth(line, innerWidth)+renderStyledPadding(paddingPrefix, commentBoxHorizontalPadding)+styleCommentBorder("│"))
	}
	boxLines = append(boxLines, styleCommentBorder("╰"+strings.Repeat("─", innerWidth+(commentBoxHorizontalPadding*2))+"╯"))
	return strings.Join(boxLines, "\n")
}

func wrapCommentBoxStyledLines(lines []styledTextLine, innerWidth int) []styledTextLine {
	if innerWidth < 1 {
		innerWidth = 1
	}

	wrappedLines := make([]styledTextLine, 0, len(lines))
	for _, line := range lines {
		if !styledLineUsesCommentBoxCodeBlockBackground(line) || len(line.runes) <= innerWidth {
			wrappedLines = append(wrappedLines, line)
			continue
		}
		wrappedLines = append(wrappedLines, wrapStyledTextLineAtWordBoundaries(line, innerWidth)...)
	}
	return wrappedLines
}

func wrapStyledTextLineAtWordBoundaries(line styledTextLine, innerWidth int) []styledTextLine {
	segments := wrappedInputSegmentsForLine(line.runes, innerWidth)
	wrappedLines := make([]styledTextLine, 0, len(segments))
	for _, segment := range segments {
		wrappedLines = append(wrappedLines, sliceStyledTextLine(line, segment.start, segment.end))
	}
	return wrappedLines
}

func sliceStyledTextLine(line styledTextLine, start int, end int) styledTextLine {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end > len(line.runes) {
		end = len(line.runes)
	}

	slicedLine := styledTextLine{
		runes:            append([]rune(nil), line.runes[start:end]...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes[start:end]...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets[start:end]...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	for _, control := range line.controls {
		if control.column < start || control.column > end {
			continue
		}
		slicedLine.controls = append(slicedLine.controls, styledTextControl{column: control.column - start, image: control.image})
	}
	return slicedLine
}

func maxStyledTextLineWidthFromLines(lines []styledTextLine) int {
	maximumWidth := 0
	for _, line := range lines {
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

func renderCommentBoxMetadataLineForViewer(author *githubdomain.PullRequestCommentAuthor, createdAt string, badges []commentMetadataBadge, reactionGroups []githubdomain.ReactionGroup, connectedUserLogin string) string {
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

func renderCommentAuthorBadgeForViewer(author *githubdomain.PullRequestCommentAuthor, connectedUserLogin string) string {
	badgeText := commentAuthorBadgeText(author)
	if badgeText == "" {
		return ""
	}
	return styleCommentAuthorBadgeTextForViewer(badgeText, commentAuthorLogin(author), connectedUserLogin)
}

func commentAuthorLogin(author *githubdomain.PullRequestCommentAuthor) string {
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

func commentAuthorBadgeText(author *githubdomain.PullRequestCommentAuthor) string {
	return strings.TrimSpace(detailCommentsIcon) + "  " + pullRequestCommentAuthorLogin(author)
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
