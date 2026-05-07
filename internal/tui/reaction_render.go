package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func renderPullRequestReactionLine(groups []githubcli.ReactionGroup) string {
	renderedGroups := renderReactionGroups(groups)
	if strings.TrimSpace(renderedGroups) == "" {
		return ""
	}
	return styleCommentMetadataText("Reactions:") + " " + renderedGroups
}

func renderReactionGroups(groups []githubcli.ReactionGroup) string {
	if len(groups) == 0 {
		return ""
	}

	renderedGroups := make([]string, 0, len(groups))
	for _, content := range githubcli.SupportedReactionContents {
		group, ok := reactionGroupForContent(groups, content)
		if !ok || group.TotalCount <= 0 {
			continue
		}
		if renderedGroup := renderReactionGroup(group); renderedGroup != "" {
			renderedGroups = append(renderedGroups, renderedGroup)
		}
	}
	return strings.Join(renderedGroups, " ")
}

func reactionGroupForContent(groups []githubcli.ReactionGroup, content githubcli.ReactionContent) (githubcli.ReactionGroup, bool) {
	for _, group := range groups {
		if strings.TrimSpace(string(group.Content)) != strings.TrimSpace(string(content)) {
			continue
		}
		return group, true
	}
	return githubcli.ReactionGroup{}, false
}

func renderReactionGroup(group githubcli.ReactionGroup) string {
	label := reactionGroupLabel(group)
	if label == "" {
		return ""
	}
	if group.ViewerHasReacted {
		return renderRoundedPill(label, theme.CommentAuthorBadgeHex, theme.CommentAuthorBadgeBackgroundHex)
	}
	return renderRoundedPill(label, theme.PendingHex, theme.PendingBackgroundHex)
}

func reactionGroupLabel(group githubcli.ReactionGroup) string {
	emoji := reactionContentEmoji(group.Content)
	if emoji == "" || group.TotalCount <= 0 {
		return ""
	}
	return fmt.Sprintf("%s %d", emoji, group.TotalCount)
}

func reactionContentEmoji(content githubcli.ReactionContent) string {
	switch content {
	case githubcli.ReactionContentThumbsUp:
		return "👍"
	case githubcli.ReactionContentThumbsDown:
		return "👎"
	case githubcli.ReactionContentLaugh:
		return "😄"
	case githubcli.ReactionContentHooray:
		return "🎉"
	case githubcli.ReactionContentConfused:
		return "😕"
	case githubcli.ReactionContentHeart:
		return "❤️"
	case githubcli.ReactionContentRocket:
		return "🚀"
	case githubcli.ReactionContentEyes:
		return "👀"
	default:
		return ""
	}
}
