package githubcli

import (
	"encoding/json"
	"sort"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type ReactionContent string

const (
	ReactionContentThumbsUp   ReactionContent = "+1"
	ReactionContentThumbsDown ReactionContent = "-1"
	ReactionContentLaugh      ReactionContent = "laugh"
	ReactionContentHooray     ReactionContent = "hooray"
	ReactionContentConfused   ReactionContent = "confused"
	ReactionContentHeart      ReactionContent = "heart"
	ReactionContentRocket     ReactionContent = "rocket"
	ReactionContentEyes       ReactionContent = "eyes"
)

var SupportedReactionContents = []ReactionContent{
	ReactionContentThumbsUp,
	ReactionContentThumbsDown,
	ReactionContentLaugh,
	ReactionContentHooray,
	ReactionContentConfused,
	ReactionContentHeart,
	ReactionContentRocket,
	ReactionContentEyes,
}

type ReactionGroup struct {
	Content          ReactionContent `json:"content"`
	TotalCount       int             `json:"totalCount"`
	ViewerHasReacted bool            `json:"viewerHasReacted"`
}

func (group *ReactionGroup) UnmarshalJSON(data []byte) error {
	var payload struct {
		Content          string `json:"content"`
		TotalCount       int    `json:"totalCount"`
		ViewerHasReacted bool   `json:"viewerHasReacted"`
		Users            *struct {
			TotalCount int `json:"totalCount"`
		} `json:"users"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	totalCount := payload.TotalCount
	if payload.Users != nil {
		totalCount = payload.Users.TotalCount
	}
	*group = ReactionGroup{
		Content:          normalizeReactionContent(payload.Content),
		TotalCount:       totalCount,
		ViewerHasReacted: payload.ViewerHasReacted,
	}
	return nil
}

func normalizeReactionContent(value string) ReactionContent {
	return ReactionContentFromDomain(githubdomain.NormalizeReactionContent(value))
}

func normalizeReactionGroups(groups []ReactionGroup) []ReactionGroup {
	if len(groups) == 0 {
		return nil
	}

	normalizedGroups := make([]ReactionGroup, 0, len(groups))
	for _, group := range groups {
		group.Content = normalizeReactionContent(string(group.Content))
		if strings.TrimSpace(string(group.Content)) == "" {
			continue
		}
		if group.TotalCount < 0 {
			group.TotalCount = 0
		}
		normalizedGroups = append(normalizedGroups, group)
	}
	if len(normalizedGroups) == 0 {
		return nil
	}

	sort.SliceStable(normalizedGroups, func(left int, right int) bool {
		return reactionContentSortIndex(normalizedGroups[left].Content) < reactionContentSortIndex(normalizedGroups[right].Content)
	})
	return normalizedGroups
}

func reactionContentSortIndex(content ReactionContent) int {
	for index, supportedContent := range SupportedReactionContents {
		if supportedContent == normalizeReactionContent(string(content)) {
			return index
		}
	}
	return len(SupportedReactionContents)
}
