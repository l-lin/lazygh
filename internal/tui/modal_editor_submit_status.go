package tui

import (
	"fmt"
	"strings"
)

func (openPullRequestByURLSubmitRequest) statusCommand() string {
	return ""
}

func (openPullRequestByURLSubmitRequest) asyncRequested() bool {
	return false
}

func (pullRequestCustomSearchSubmitRequest) statusCommand() string {
	return ""
}

func (pullRequestCustomSearchSubmitRequest) asyncRequested() bool {
	return false
}

func (request pullRequestCommentSubmitRequest) statusCommand() string {
	return pullRequestCommentCommand(request.target.repository, request.target.number)
}

func (pullRequestCommentSubmitRequest) asyncRequested() bool {
	return true
}

func (request pullRequestReviewCommentSubmitRequest) statusCommand() string {
	return pullRequestReviewWithCommentCommand(request.target.repository, request.target.number)
}

func (pullRequestReviewCommentSubmitRequest) asyncRequested() bool {
	return true
}

func (request pullRequestRequestChangesSubmitRequest) statusCommand() string {
	return pullRequestRequestChangesCommand(request.target.repository, request.target.number)
}

func (pullRequestRequestChangesSubmitRequest) asyncRequested() bool {
	return true
}

func (request pullRequestTitleEditSubmitRequest) statusCommand() string {
	return pullRequestTitleEditCommand(request.target.repository, request.target.number, request.title)
}

func (pullRequestTitleEditSubmitRequest) asyncRequested() bool {
	return true
}

func (request pullRequestDescriptionEditSubmitRequest) statusCommand() string {
	return pullRequestDescriptionEditCommand(request.target.repository, request.target.number)
}

func (pullRequestDescriptionEditSubmitRequest) asyncRequested() bool {
	return true
}

func (pullRequestCommentUpdateSubmitRequest) statusCommand() string {
	return graphQLMutationCommand()
}

func (pullRequestCommentUpdateSubmitRequest) asyncRequested() bool {
	return true
}

func (inlineCommentUpdateSubmitRequest) statusCommand() string {
	return graphQLMutationCommand()
}

func (inlineCommentUpdateSubmitRequest) asyncRequested() bool {
	return true
}

func (inlineCommentReplySubmitRequest) statusCommand() string {
	return graphQLMutationCommand()
}

func (inlineCommentReplySubmitRequest) asyncRequested() bool {
	return true
}

func (request reviewInlineCommentSubmitRequest) statusCommand() string {
	if strings.TrimSpace(request.target.pendingReview) != "" {
		return ""
	}
	return graphQLMutationCommand()
}

func (request reviewInlineCommentSubmitRequest) asyncRequested() bool {
	return strings.TrimSpace(request.target.pendingReview) == ""
}

func (preparedReviewInlineCommentSubmitRequest) statusCommand() string {
	return graphQLMutationCommand()
}

func (preparedReviewInlineCommentSubmitRequest) asyncRequested() bool {
	return true
}

func (pendingPullRequestReviewSubmitRequest) statusCommand() string {
	return graphQLMutationCommand()
}

func (pendingPullRequestReviewSubmitRequest) asyncRequested() bool {
	return true
}

func pullRequestCommentCommand(repository string, number int) string {
	if !statusPullRequestIdentityValid(repository, number) {
		return ""
	}
	return formatStatusLineCommand("gh", "pr", "comment", fmt.Sprintf("%d", number), "-R", strings.TrimSpace(repository), "--body-file", "-")
}

func pullRequestReviewWithCommentCommand(repository string, number int) string {
	if !statusPullRequestIdentityValid(repository, number) {
		return ""
	}
	return formatStatusLineCommand("gh", "pr", "review", fmt.Sprintf("%d", number), "-R", strings.TrimSpace(repository), "--comment", "--body-file", "-")
}

func pullRequestRequestChangesCommand(repository string, number int) string {
	if !statusPullRequestIdentityValid(repository, number) {
		return ""
	}
	return formatStatusLineCommand("gh", "pr", "review", fmt.Sprintf("%d", number), "-R", strings.TrimSpace(repository), "--request-changes", "--body-file", "-")
}

func pullRequestTitleEditCommand(repository string, number int, title string) string {
	if !statusPullRequestIdentityValid(repository, number) {
		return ""
	}
	return formatStatusLineCommand("gh", "pr", "edit", fmt.Sprintf("%d", number), "-R", strings.TrimSpace(repository), "--title", title)
}

func pullRequestDescriptionEditCommand(repository string, number int) string {
	if !statusPullRequestIdentityValid(repository, number) {
		return ""
	}
	return formatStatusLineCommand("gh", "pr", "edit", fmt.Sprintf("%d", number), "-R", strings.TrimSpace(repository), "--body-file", "-")
}

func graphQLMutationCommand() string {
	return formatStatusLineCommand("gh", "api", "graphql")
}

func statusPullRequestIdentityValid(repository string, number int) bool {
	return strings.TrimSpace(repository) != "" && number > 0
}
