package github

import "strings"

type PullRequestMergeQueueEntry struct {
	ID                   string `json:"id,omitempty"`
	State                string `json:"state,omitempty"`
	Position             int    `json:"position,omitempty"`
	EstimatedTimeToMerge int    `json:"estimatedTimeToMerge,omitempty"`
}

func (entry PullRequestMergeQueueEntry) normalized() PullRequestMergeQueueEntry {
	entry.ID = strings.TrimSpace(entry.ID)
	entry.State = strings.TrimSpace(entry.State)
	return entry
}
