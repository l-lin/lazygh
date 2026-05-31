package tui

import (
	"errors"
	"fmt"
	"strings"
)

const (
	reRequestPullRequestReviewActionTitlePrefix = "Re-request review from"
	pullRequestReviewReRequestedSuccessMessage  = "Review re-requested"
)

type pullRequestReviewerRequestTarget struct {
	repository    string
	number        int
	reviewerLogin string
}

func (program *Program) currentReRequestPullRequestReviewAction() (actionsPopupAction, bool) {
	target, ok := program.currentPullRequestReviewerRequestTargetAtDetailCursor()
	if !ok {
		return actionsPopupAction{}, false
	}

	trimmedLogin := strings.TrimSpace(target.reviewerLogin)
	var requested Msg = MsgReRequestPullRequestReviewRequested{Target: target}
	if !program.hasPullRequestMutations() {
		requested = actionsPopupErrorRequested(errors.New("github loader is unavailable"))
	}
	return actionsPopupAction{
		id:        "re-request-review-" + strings.ToLower(trimmedLogin),
		title:     reRequestPullRequestReviewActionTitle(trimmedLogin),
		icon:      actionsPopupReRequestReviewIcon,
		requested: requested,
	}, true
}

func reRequestPullRequestReviewActionTitle(reviewerLogin string) string {
	label := formatLogin(reviewerLogin)
	if label == "-" {
		label = "reviewer"
	}
	return reRequestPullRequestReviewActionTitlePrefix + " " + label
}

func requestPullRequestReviewerCommand(repository string, number int, reviewerLogin string) string {
	return formatStatusLineCommand("gh", "pr", "edit", fmt.Sprintf("%d", number), "-R", repository, "--add-reviewer", reviewerLogin)
}

func (program *Program) currentPullRequestReviewerRequestTargetAtDetailCursor() (pullRequestReviewerRequestTarget, bool) {
	model := program.currentDescriptionCursorActionReadModel()
	entry, ok := model.reviewerActionEntryAtCursor()
	if !ok {
		return pullRequestReviewerRequestTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(model.summary.Repository))
	if repository == "" || repository == "-" || model.summary.Number <= 0 {
		return pullRequestReviewerRequestTarget{}, false
	}
	trimmedLogin := strings.TrimSpace(entry.ReviewerLogin)
	if trimmedLogin == "" {
		return pullRequestReviewerRequestTarget{}, false
	}

	return pullRequestReviewerRequestTarget{
		repository:    repository,
		number:        model.summary.Number,
		reviewerLogin: trimmedLogin,
	}, true
}
