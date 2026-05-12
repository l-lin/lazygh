package github

import (
	"encoding/json"
	"sort"
	"strings"
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
		Content:          NormalizeReactionContent(payload.Content),
		TotalCount:       totalCount,
		ViewerHasReacted: payload.ViewerHasReacted,
	}
	return nil
}

func NormalizeReactionContent(value string) ReactionContent {
	trimmedValue := strings.TrimSpace(value)
	switch strings.ToUpper(trimmedValue) {
	case "THUMBS_UP", "+1":
		return ReactionContentThumbsUp
	case "THUMBS_DOWN", "-1":
		return ReactionContentThumbsDown
	case "LAUGH":
		return ReactionContentLaugh
	case "HOORAY":
		return ReactionContentHooray
	case "CONFUSED":
		return ReactionContentConfused
	case "HEART":
		return ReactionContentHeart
	case "ROCKET":
		return ReactionContentRocket
	case "EYES":
		return ReactionContentEyes
	default:
		return ReactionContent(strings.ToLower(trimmedValue))
	}
}

func normalizeReactionGroups(groups []ReactionGroup) []ReactionGroup {
	if len(groups) == 0 {
		return nil
	}

	normalizedGroups := make([]ReactionGroup, 0, len(groups))
	for _, group := range groups {
		group.Content = NormalizeReactionContent(string(group.Content))
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
		if supportedContent == NormalizeReactionContent(string(content)) {
			return index
		}
	}
	return len(SupportedReactionContents)
}
