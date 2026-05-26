package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestReactionActionTarget struct {
	repository        string
	number            int
	subjectID         string
	reactionGroups    []githubdomain.ReactionGroup
	invalidateDiff    bool
	actionsPopupGroup string
}

func (target pullRequestReactionActionTarget) popupGroup() string {
	group := strings.TrimSpace(target.actionsPopupGroup)
	if group != "" {
		return group
	}
	return actionsPopupGroupReview
}

type reactionPickerState struct {
	target pullRequestReactionActionTarget
}

func (program *Program) selectedPullRequestReactionActionTarget() (pullRequestReactionActionTarget, bool) {
	if !program.isPullRequestContext() {
		return pullRequestReactionActionTarget{}, false
	}

	if program.reviewModeActive() {
		if program.model.Focus() == FocusDetailView {
			if target, ok := program.selectedReviewDiffReactionActionTarget(); ok {
				return target, true
			}
			if !program.reviewSessionShowsDescription() {
				return pullRequestReactionActionTarget{}, false
			}
		}
		summary, ok := program.currentPullRequestSummary()
		if !ok {
			return pullRequestReactionActionTarget{}, false
		}
		return program.selectedPullRequestReactionTargetFromSummary(summary)
	}

	switch program.model.Focus() {
	case FocusDetailView:
		summary, ok := program.selectedPullRequestSummaryForDetail()
		if !ok {
			return pullRequestReactionActionTarget{}, false
		}
		switch program.detailState.activeTab {
		case CommentsDetailTab:
			return program.selectedBrowserCommentReactionActionTarget(summary)
		case ChangesDetailTab:
			return program.selectedBrowserChangesReactionActionTarget(summary)
		case DescriptionDetailTab:
			return program.selectedPullRequestReactionTargetFromSummary(summary)
		default:
			return pullRequestReactionActionTarget{}, false
		}
	case FocusPullRequestsView:
		if program.detailState.activeTab != DescriptionDetailTab {
			return pullRequestReactionActionTarget{}, false
		}
		summary, ok := program.model.SelectedPullRequestSummary()
		if !ok {
			return pullRequestReactionActionTarget{}, false
		}
		return program.selectedPullRequestReactionTargetFromSummary(summary)
	default:
		return pullRequestReactionActionTarget{}, false
	}
}

func (program *Program) selectedPullRequestReactionTargetFromSummary(summary githubdomain.PullRequest) (pullRequestReactionActionTarget, bool) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 {
		return pullRequestReactionActionTarget{}, false
	}

	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil || strings.TrimSpace(result.detail.ID) == "" {
		return pullRequestReactionActionTarget{}, false
	}

	return pullRequestReactionActionTarget{
		repository:        repository,
		number:            summary.Number,
		subjectID:         strings.TrimSpace(result.detail.ID),
		reactionGroups:    append([]githubdomain.ReactionGroup(nil), result.detail.ReactionGroups...),
		actionsPopupGroup: actionsPopupGroupPullRequest,
	}, true
}

func (program *Program) selectedBrowserCommentReactionActionTarget(summary githubdomain.PullRequest) (pullRequestReactionActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() || program.detailState.activeTab != CommentsDetailTab {
		return pullRequestReactionActionTarget{}, false
	}

	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReactionActionTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailState.wrapWidth, program.detailState.viewState.cursor.line)
	if !ok {
		return pullRequestReactionActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 {
		return pullRequestReactionActionTarget{}, false
	}

	if sectionAtCursor.section.comment != nil {
		comment := *sectionAtCursor.section.comment
		if !hasUsablePullRequestMutationID(comment.ID) {
			return pullRequestReactionActionTarget{}, false
		}
		return pullRequestReactionActionTarget{
			repository:     repository,
			number:         summary.Number,
			subjectID:      strings.TrimSpace(comment.ID),
			reactionGroups: append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...),
		}, true
	}

	if sectionAtCursor.section.inlineComment != nil {
		comment := *sectionAtCursor.section.inlineComment
		if !hasUsablePullRequestMutationID(comment.ID) {
			return pullRequestReactionActionTarget{}, false
		}
		return pullRequestReactionActionTarget{
			repository:     repository,
			number:         summary.Number,
			subjectID:      strings.TrimSpace(comment.ID),
			reactionGroups: append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...),
			invalidateDiff: true,
		}, true
	}

	comment, ok := browserConversationInlineThreadCommentAtCursor(sectionAtCursor)
	if !ok || !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestReactionActionTarget{}, false
	}
	return pullRequestReactionActionTarget{
		repository:     repository,
		number:         summary.Number,
		subjectID:      strings.TrimSpace(comment.ID),
		reactionGroups: append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...),
		invalidateDiff: true,
	}, true
}

func (program *Program) selectedBrowserChangesReactionActionTarget(summary githubdomain.PullRequest) (pullRequestReactionActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() || program.detailState.activeTab != ChangesDetailTab {
		return pullRequestReactionActionTarget{}, false
	}

	context, ok := program.currentBrowserChangesCursorContext()
	if !ok {
		return pullRequestReactionActionTarget{}, false
	}
	_, comment, ok := reviewDiffCommentAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok || !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestReactionActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 {
		return pullRequestReactionActionTarget{}, false
	}

	return pullRequestReactionActionTarget{
		repository:     repository,
		number:         summary.Number,
		subjectID:      strings.TrimSpace(comment.ID),
		reactionGroups: append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...),
		invalidateDiff: true,
	}, true
}

func (program *Program) selectedReviewDiffReactionActionTarget() (pullRequestReactionActionTarget, bool) {
	if !program.reviewModeActive() || program.model.Focus() != FocusDetailView {
		return pullRequestReactionActionTarget{}, false
	}

	context, ok := program.currentReviewDiffCursorContext()
	if !ok {
		return pullRequestReactionActionTarget{}, false
	}
	_, comment, ok := reviewDiffCommentAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok || !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestReactionActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 {
		return pullRequestReactionActionTarget{}, false
	}

	return pullRequestReactionActionTarget{
		repository:     repository,
		number:         context.summary.Number,
		subjectID:      strings.TrimSpace(comment.ID),
		reactionGroups: append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...),
		invalidateDiff: true,
	}, true
}
