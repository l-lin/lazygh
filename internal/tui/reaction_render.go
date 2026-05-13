package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

func renderPullRequestReactionLine(groups any) string {
	renderedGroups := renderReactionGroups(groups)
	if strings.TrimSpace(renderedGroups) == "" {
		return ""
	}
	return renderedGroups
}

func renderReactionGroups(groups any) string {
	domainGroups := toDomainReactionGroups(groups)
	if len(domainGroups) == 0 {
		return ""
	}

	renderedGroups := make([]string, 0, len(domainGroups))
	for _, content := range githubdomain.SupportedReactionContents {
		group, ok := reactionGroupForContent(domainGroups, content)
		if !ok || group.TotalCount <= 0 {
			continue
		}
		if renderedGroup := renderReactionGroup(group); renderedGroup != "" {
			renderedGroups = append(renderedGroups, renderedGroup)
		}
	}
	return strings.Join(renderedGroups, " ")
}

func reactionGroupForContent(groups []githubdomain.ReactionGroup, content githubdomain.ReactionContent) (githubdomain.ReactionGroup, bool) {
	for _, group := range groups {
		if strings.TrimSpace(string(group.Content)) != strings.TrimSpace(string(content)) {
			continue
		}
		return group, true
	}
	return githubdomain.ReactionGroup{}, false
}

func renderReactionGroup(group any) string {
	groupValue, ok := toDomainReactionGroup(group)
	if !ok {
		return ""
	}
	label := reactionGroupLabel(groupValue)
	if label == "" {
		return ""
	}
	if groupValue.ViewerHasReacted {
		return renderRoundedPill(label, theme.CommentAuthorBadgeHex, theme.CommentAuthorBadgeBackgroundHex)
	}
	return renderRoundedPill(label, theme.PendingHex, theme.PendingBackgroundHex)
}

func reactionGroupLabel(group githubdomain.ReactionGroup) string {
	emoji := reactionContentEmoji(group.Content)
	if emoji == "" || group.TotalCount <= 0 {
		return ""
	}
	return fmt.Sprintf("%s %d", emoji, group.TotalCount)
}

func reactionContentEmoji(content githubdomain.ReactionContent) string {
	switch content {
	case githubdomain.ReactionContentThumbsUp:
		return "👍"
	case githubdomain.ReactionContentThumbsDown:
		return "👎"
	case githubdomain.ReactionContentLaugh:
		return "😄"
	case githubdomain.ReactionContentHooray:
		return "🎉"
	case githubdomain.ReactionContentConfused:
		return "😕"
	case githubdomain.ReactionContentHeart:
		return "❤️"
	case githubdomain.ReactionContentRocket:
		return "🚀"
	case githubdomain.ReactionContentEyes:
		return "👀"
	default:
		return ""
	}
}
