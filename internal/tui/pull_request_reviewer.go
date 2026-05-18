package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
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
	return actionsPopupAction{
		id:    "re-request-review-" + strings.ToLower(trimmedLogin),
		title: reRequestPullRequestReviewActionTitle(trimmedLogin),
		icon:  actionsPopupReRequestReviewIcon,
		execute: func(gui *gocui.Gui) actionsPopupActionResult {
			return program.executeReRequestPullRequestReviewAction(gui, target)
		},
	}, true
}

func reRequestPullRequestReviewActionTitle(reviewerLogin string) string {
	label := formatLogin(reviewerLogin)
	if label == "-" {
		label = "reviewer"
	}
	return reRequestPullRequestReviewActionTitlePrefix + " " + label
}

func (program *Program) executeReRequestPullRequestReviewAction(gui *gocui.Gui, target pullRequestReviewerRequestTarget) actionsPopupActionResult {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 || strings.TrimSpace(target.reviewerLogin) == "" {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !program.hasPullRequestMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	return program.startActionsPopupAsyncGHCommand(gui, requestPullRequestReviewerCommand(target.repository, target.number, target.reviewerLogin), func() error {
		return program.pullRequestMutations.RequestPullRequestReviewer(target.repository, target.number, target.reviewerLogin)
	}, func() {
		program.invalidatePullRequestDetail(target.repository, target.number)
		program.setFeedback(program.model.Focus(), pullRequestReviewReRequestedSuccessMessage)
	})
}

func requestPullRequestReviewerCommand(repository string, number int, reviewerLogin string) string {
	return formatStatusLineCommand("gh", "pr", "edit", fmt.Sprintf("%d", number), "-R", repository, "--add-reviewer", reviewerLogin)
}

func (program *Program) currentPullRequestReviewerRequestTargetAtDetailCursor() (pullRequestReviewerRequestTarget, bool) {
	if !program.detailCursorActionsAvailable() {
		return pullRequestReviewerRequestTarget{}, false
	}

	summary, _, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestReviewerRequestTarget{}, false
	}

	actualView := program.resolveView(program.gui, nil, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	entry, ok := program.browserOverviewReviewerEntryAtDetailCursorDocument(document)
	if !ok {
		return pullRequestReviewerRequestTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return pullRequestReviewerRequestTarget{}, false
	}
	trimmedLogin := strings.TrimSpace(entry.ReviewerLogin)
	if trimmedLogin == "" {
		return pullRequestReviewerRequestTarget{}, false
	}

	return pullRequestReviewerRequestTarget{
		repository:    repository,
		number:        summary.Number,
		reviewerLogin: trimmedLogin,
	}, true
}

func (program *Program) browserOverviewReviewerEntryAtDetailCursorDocument(document detailDocument) (pullRequestOverviewEntry, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestOverviewEntry{}, false
	}

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, document.width, program.detailViewState.cursor.line)
	if !ok || !sectionAtCursor.inBody || !strings.EqualFold(strings.TrimSpace(sectionAtCursor.section.overviewBlockTitle), "Reviewers") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(sectionAtCursor.section, sectionAtCursor.bodyLine)
	if !ok || !entry.CanReRequestReview || strings.TrimSpace(entry.ReviewerLogin) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}
