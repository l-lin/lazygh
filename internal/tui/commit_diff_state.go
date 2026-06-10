package tui

import "strings"

type commitDiffTabState struct {
	visible        bool
	pullRequestKey string
	commitOID      string
	shortLabel     string
}

type commitDiffResult struct {
	data reviewDiffData
	err  error
}

func (state commitDiffTabState) visibleForPullRequestKey(pullRequestKey string) bool {
	return state.visible && strings.TrimSpace(state.pullRequestKey) != "" && strings.TrimSpace(state.pullRequestKey) == strings.TrimSpace(pullRequestKey) && strings.TrimSpace(state.commitOID) != ""
}

func (state commitDiffTabState) withOpened(pullRequestKey string, commitOID string, shortLabel string) commitDiffTabState {
	state.visible = strings.TrimSpace(pullRequestKey) != "" && strings.TrimSpace(commitOID) != ""
	state.pullRequestKey = strings.TrimSpace(pullRequestKey)
	state.commitOID = strings.TrimSpace(commitOID)
	state.shortLabel = strings.TrimSpace(shortLabel)
	return state
}

func (state commitDiffTabState) cleared() commitDiffTabState {
	return commitDiffTabState{}
}

func commitDiffCacheKey(pullRequestKey string, commitOID string) string {
	trimmedPullRequestKey := strings.TrimSpace(pullRequestKey)
	trimmedCommitOID := strings.TrimSpace(commitOID)
	if trimmedPullRequestKey == "" || trimmedCommitOID == "" {
		return ""
	}
	return trimmedPullRequestKey + "@" + trimmedCommitOID
}

func (program *Program) commitDiffResultForTarget(pullRequestKey string, commitOID string) (commitDiffResult, bool) {
	key := commitDiffCacheKey(pullRequestKey, commitOID)
	if key == "" {
		return commitDiffResult{}, false
	}
	result, ok := program.commitDiffCache[key]
	return result, ok
}

func (program *Program) commitDiffLoadInFlightForTarget(pullRequestKey string, commitOID string) bool {
	key := commitDiffCacheKey(pullRequestKey, commitOID)
	return key != "" && program.commitDiffLoadInFlight[key]
}
