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
	if !program.detailCursorActionsAvailable() {
		return pullRequestReviewerRequestTarget{}, false
	}

	context, ok := program.currentPullRequestDescriptionCursorContext()
	if !ok {
		return pullRequestReviewerRequestTarget{}, false
	}

	entry, ok := program.browserOverviewReviewerEntryAtDetailCursorDocument(context.selection.document)
	if !ok {
		return pullRequestReviewerRequestTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || repository == "-" || context.summary.Number <= 0 {
		return pullRequestReviewerRequestTarget{}, false
	}
	trimmedLogin := strings.TrimSpace(entry.ReviewerLogin)
	if trimmedLogin == "" {
		return pullRequestReviewerRequestTarget{}, false
	}

	return pullRequestReviewerRequestTarget{
		repository:    repository,
		number:        context.summary.Number,
		reviewerLogin: trimmedLogin,
	}, true
}

func (program *Program) browserOverviewReviewerEntryAtDetailCursorDocument(document detailDocument) (pullRequestOverviewEntry, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestOverviewEntry{}, false
	}

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, document.width, program.detailState.viewState.cursor.line)
	if !ok || !sectionAtCursor.inBody || !strings.EqualFold(strings.TrimSpace(sectionAtCursor.section.overviewBlockTitle), "Reviewers") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(sectionAtCursor.section, sectionAtCursor.bodyLine)
	if !ok || !entry.CanReRequestReview || strings.TrimSpace(entry.ReviewerLogin) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}
