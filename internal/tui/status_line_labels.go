package tui

import (
	"fmt"
	"strconv"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func statusLinePullRequestOperation(action string, summary githubdomain.PullRequest) statusLineOperationDescriptor {
	trimmedAction := strings.TrimSpace(action)
	baseLabel := trimmedAction
	if summary.Number > 0 {
		baseLabel += fmt.Sprintf(" #%d", summary.Number)
	}
	loadingLabel := baseLabel
	if title := statusLineTitle(summary.Title); title != "" {
		loadingLabel += ": " + title
	}
	return statusLineOperationDescriptor{loadingLabel: loadingLabel, failureLabel: baseLabel}
}

func statusLinePullRequestIdentityOperation(program *Program, action string, repository string, number int) statusLineOperationDescriptor {
	summary := githubdomain.PullRequest{
		Repository: githubdomain.RepositoryRef{NameWithOwner: strings.TrimSpace(repository)},
		Number:     number,
	}
	if program != nil {
		if currentSummary, ok := program.currentPullRequestSummary(); ok && samePullRequestIdentity(currentSummary, summary) {
			summary = currentSummary
		}
	}
	return statusLinePullRequestOperation(action, summary)
}

func statusLineTitle(title string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
}

func statusLineRefreshPullRequestListOperation() statusLineOperationDescriptor {
	return statusLineOperationDescriptor{loadingLabel: "Refreshing PR list", failureLabel: "Refreshing PR list"}
}

func statusLineManualRefreshOperation(successMessage string, standardReviewMode bool) statusLineOperationDescriptor {
	if standardReviewMode && strings.TrimSpace(successMessage) == pullRequestRefreshSuccessMessage {
		return statusLineOperationDescriptor{loadingLabel: "Refreshing PR", failureLabel: "Refreshing PR"}
	}
	if strings.TrimSpace(successMessage) == notificationsRefreshSuccessMessage {
		return statusLineRefreshNotificationsOperation()
	}
	return statusLineRefreshPullRequestListOperation()
}

func statusLineRefreshNotificationsOperation() statusLineOperationDescriptor {
	return statusLineOperationDescriptor{loadingLabel: "Refreshing notifications", failureLabel: "Refreshing notifications"}
}

func statusLineRefreshPullRequestKeyOperation(program *Program, key string) statusLineOperationDescriptor {
	if program != nil {
		if summary, ok := program.selectedPullRequestSummaryForDetail(); ok && pullRequestDetailKey(summary.Repository, summary.Number) == strings.TrimSpace(key) {
			return statusLinePullRequestOperation("Refreshing", summary)
		}
		if summary, ok := program.selectedPullRequestSummaryForDiff(); ok && pullRequestDetailKey(summary.Repository, summary.Number) == strings.TrimSpace(key) {
			return statusLinePullRequestOperation("Refreshing", summary)
		}
	}

	_, numberText, ok := strings.Cut(strings.TrimSpace(key), "#")
	if ok {
		if number, err := strconv.Atoi(strings.TrimSpace(numberText)); err == nil {
			return statusLinePullRequestOperation("Refreshing", githubdomain.PullRequest{Number: number})
		}
	}
	return statusLineOperationDescriptor{loadingLabel: "Refreshing", failureLabel: "Refreshing"}
}

func statusLineRefreshIssueOperation(repository string, number int) statusLineOperationDescriptor {
	return statusLineOperationDescriptor{
		loadingLabel: fmt.Sprintf("Refreshing issue #%d", number),
		failureLabel: fmt.Sprintf("Refreshing issue #%d", number),
	}
}

func statusLineRefreshReleaseOperation(repository string, id int) statusLineOperationDescriptor {
	return statusLineOperationDescriptor{
		loadingLabel: fmt.Sprintf("Refreshing release #%d", id),
		failureLabel: fmt.Sprintf("Refreshing release #%d", id),
	}
}

func statusLineOperationForActionsPopupRequest(program *Program, request actionsPopupAsyncRequest) statusLineOperationDescriptor {
	switch actual := request.(type) {
	case startPullRequestReviewPopupRequest:
		return statusLinePullRequestOperation("Creating review", actual.summary)
	case openPullRequestInBrowserPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Opening", actual.repository, actual.number)
	case approvePullRequestPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Approving", actual.repository, actual.number)
	case reRequestPullRequestReviewPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Requesting review for", actual.repository, actual.number)
	case pullRequestLifecycleMutationPopupRequest:
		return statusLinePullRequestOperation(pullRequestLifecycleStatusAction(actual.kind), actual.summary)
	case pullRequestAutoMergeMutationPopupRequest:
		return statusLinePullRequestOperation(pullRequestAutoMergeStatusAction(actual.kind), actual.summary)
	case pullRequestMergeWhenReadyPopupRequest:
		return statusLinePullRequestOperation("Merging", actual.summary)
	case pullRequestMergeQueueMutationPopupRequest:
		return statusLinePullRequestOperation(pullRequestMergeQueueStatusAction(actual.kind), actual.summary)
	case pullRequestBranchUpdatePopupRequest:
		return statusLinePullRequestOperation("Updating branch", actual.summary)
	case updatePullRequestAssigneesPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Updating assignees for", actual.repository, actual.number)
	case addReactionPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Adding reaction to", actual.target.repository, actual.target.number)
	case cancelPendingPullRequestReviewPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Canceling review for", actual.target.repository, actual.target.number)
	case deletePullRequestCommentPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Deleting comment from", actual.target.repository, actual.target.number)
	case deleteInlineCommentPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Deleting comment from", actual.target.repository, actual.target.number)
	case inlineCommentResolutionPopupRequest:
		return statusLinePullRequestIdentityOperation(program, inlineCommentStatusAction(actual.resolved), actual.target.repository, actual.target.number)
	case removeReactionPopupRequest:
		return statusLinePullRequestIdentityOperation(program, "Removing reaction from", actual.target.repository, actual.target.number)
	case pullRequestSquashMergePopupRequest:
		return statusLinePullRequestOperation("Merging", actual.summary)
	default:
		return statusLineOperationDescriptor{}
	}
}

func pullRequestLifecycleStatusAction(kind pullRequestLifecycleMutationKind) string {
	switch kind {
	case pullRequestLifecycleMutationReadyForReview:
		return "Marking ready for review"
	case pullRequestLifecycleMutationConvertToDraft:
		return "Converting to draft"
	case pullRequestLifecycleMutationClose:
		return "Closing"
	case pullRequestLifecycleMutationReopen:
		return "Reopening"
	default:
		return "Updating"
	}
}

func pullRequestAutoMergeStatusAction(kind pullRequestAutoMergeMutationKind) string {
	if kind == pullRequestAutoMergeMutationDisable {
		return "Disabling auto-merge for"
	}
	return "Enabling auto-merge for"
}

func pullRequestMergeQueueStatusAction(kind pullRequestMergeQueueMutationKind) string {
	if kind == pullRequestMergeQueueMutationDequeue {
		return "Removing from merge queue"
	}
	return "Adding to merge queue"
}

func inlineCommentStatusAction(resolved bool) string {
	if resolved {
		return "Resolving comment on"
	}
	return "Unresolving comment on"
}

func statusLineOperationForModalEditorRequest(program *Program, request modalEditorSubmitRequest) statusLineOperationDescriptor {
	switch actual := request.(type) {
	case openPullRequestByURLSubmitRequest:
		return statusLineOperationDescriptor{loadingLabel: "Opening pull request", failureLabel: "Opening pull request"}
	case pullRequestCustomSearchSubmitRequest:
		return statusLineOperationDescriptor{loadingLabel: "Searching pull requests", failureLabel: "Searching pull requests"}
	case pullRequestCommentSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Commenting", actual.target.repository, actual.target.number)
	case pullRequestReviewCommentSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Commenting", actual.target.repository, actual.target.number)
	case pullRequestRequestChangesSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Requesting changes on", actual.target.repository, actual.target.number)
	case pullRequestTitleEditSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Editing", actual.target.repository, actual.target.number)
	case pullRequestDescriptionEditSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Editing", actual.target.repository, actual.target.number)
	case pullRequestCommentUpdateSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Editing comment on", actual.target.repository, actual.target.number)
	case inlineCommentUpdateSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Editing comment on", actual.target.repository, actual.target.number)
	case inlineCommentReplySubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Replying on", actual.target.repository, actual.target.number)
	case reviewInlineCommentSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Commenting on", actual.target.repository, actual.target.number)
	case preparedReviewInlineCommentSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, "Commenting on", actual.target.repository, actual.target.number)
	case pendingPullRequestReviewSubmitRequest:
		return statusLinePullRequestIdentityOperation(program, pendingReviewStatusAction(actual.event), actual.target.repository, actual.target.number)
	default:
		return statusLineOperationDescriptor{}
	}
}

func pendingReviewStatusAction(event githubdomain.PullRequestReviewEvent) string {
	switch event {
	case githubdomain.PullRequestReviewEventApprove:
		return "Approving"
	case githubdomain.PullRequestReviewEventRequestChanges:
		return "Requesting changes on"
	default:
		return "Commenting on"
	}
}
