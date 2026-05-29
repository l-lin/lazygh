package tui

import "strings"

func (program *Program) applyConnectedUserLoadPlanned() {
	program.connectedUserLoadStarted = true
}

func (program *Program) applyPullRequestsLoadPlanned(message MsgPullRequestsLoadPlanned) {
	program.setPullRequestsLoadStarted(message.Tab, true)
	program.setPullRequestsLoading(message.Tab, true)
}

func (program *Program) applyNotificationsLoadPlanned() {
	program.planNotificationsLoad()
}

func (program *Program) applyPullRequestDetailLoadPlanned(message MsgPullRequestDetailLoadPlanned) {
	key := strings.TrimSpace(message.Key)
	if key == "" {
		return
	}
	program.pullRequestDetailLoadInFlight[key] = true
}

func (program *Program) applyPullRequestDiffLoadPlanned(message MsgPullRequestDiffLoadPlanned) {
	key := strings.TrimSpace(message.Key)
	if key == "" {
		return
	}
	program.pullRequestDiffLoadInFlight[key] = true
}

func (program *Program) applyIssueDetailLoadPlanned(message MsgIssueDetailLoadPlanned) {
	key := notificationDetailKey(message.Repository, message.Number)
	if key == "" {
		return
	}
	program.issueDetailLoadInFlight[key] = true
}

func (program *Program) applyReleaseDetailLoadPlanned(message MsgReleaseDetailLoadPlanned) {
	key := notificationDetailKey(message.Repository, message.ID)
	if key == "" {
		return
	}
	program.releaseDetailLoadInFlight[key] = true
}

func (program *Program) applyCurrentDetailImageHTMLLoadPlanned(message MsgCurrentDetailImageHTMLLoadPlanned) {
	key := strings.TrimSpace(message.SourceKey)
	if key == "" {
		return
	}
	program.detailImageHTMLLoadInFlight[key] = true
}

func (program *Program) applyCurrentDetailImageLoadPlanned(message MsgCurrentDetailImageLoadPlanned) {
	imageURL := strings.TrimSpace(message.ImageURL)
	if imageURL == "" {
		return
	}
	program.detailImageLoadInFlight[imageURL] = true
}

func (program *Program) applyPullRequestDetailCacheHydrated(message MsgPullRequestDetailCacheHydrated) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}
	if _, ok := program.pullRequestDetailCache[key]; ok {
		return
	}
	program.pullRequestDetailCache[key] = message.Result
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyPullRequestDiffCacheHydrated(message MsgPullRequestDiffCacheHydrated) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}
	if _, ok := program.pullRequestDiffCache[key]; ok {
		return
	}
	program.pullRequestDiffCache[key] = message.Result
	program.invalidateReviewDiffRenderCache()
	program.clampReviewSessionSelection()
}
