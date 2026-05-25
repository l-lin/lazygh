package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyPullRequestCommentSubmitRequested(message MsgPullRequestCommentSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasPullRequestMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.pullRequestMutations.CommentOnPullRequest(message.Target.repository, message.Target.number, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.optimisticallyAppendPullRequestComment(message.Target, message.Body)
			program.applyFeedbackSet(MsgFeedbackSet{Target: message.FeedbackTarget, Message: pullRequestCommentSuccessMessage})
			return nil
		},
	}}
}

func (program *Program) applyPullRequestReviewCommentSubmitRequested(message MsgPullRequestReviewCommentSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasReviewMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.reviewMutations.ReviewPullRequestWithComment(message.Target.repository, message.Target.number, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			return actionsPopupAsyncInvalidatePullRequestSuccess{
				Repository:     message.Target.repository,
				Number:         message.Target.number,
				InvalidateDiff: true,
				Message:        pullRequestReviewSuccessMessage,
			}.apply(program)
		},
	}}
}

func (program *Program) applyPullRequestRequestChangesSubmitRequested(message MsgPullRequestRequestChangesSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasReviewMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.reviewMutations.RequestChangesOnPullRequest(message.Target.repository, message.Target.number, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			return actionsPopupAsyncInvalidatePullRequestSuccess{
				Repository:     message.Target.repository,
				Number:         message.Target.number,
				InvalidateDiff: true,
				Message:        pullRequestReviewSuccessMessage,
			}.apply(program)
		},
	}}
}

func (program *Program) applyPullRequestTitleEditRequested(message MsgPullRequestTitleEditRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Title,
		Submit: func(title string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasPullRequestMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.pullRequestMutations.EditPullRequestTitle(message.Target.repository, message.Target.number, title); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			return program.applyPullRequestTitleEditApplied(MsgPullRequestTitleEditApplied{Target: message.Target, Title: message.Title, FeedbackTarget: message.FeedbackTarget})
		},
	}}
}

func (program *Program) applyPullRequestDescriptionEditRequested(message MsgPullRequestDescriptionEditRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasPullRequestMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.pullRequestMutations.EditPullRequestDescription(message.Target.repository, message.Target.number, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			return program.applyPullRequestDescriptionEditApplied(MsgPullRequestDescriptionEditApplied{Target: message.Target, Body: message.Body, FeedbackTarget: message.FeedbackTarget})
		},
	}}
}

func (program *Program) applyPullRequestCommentUpdateRequested(message MsgPullRequestCommentUpdateRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.commentID) == "" {
				return errors.New("missing pull request comment identity")
			}
			if !program.hasPullRequestMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.pullRequestMutations.UpdatePullRequestComment(message.Target.commentID, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.optimisticallyUpdatePullRequestComment(message.Target, message.Body)
			program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestCommentUpdatedSuccessMessage})
			return nil
		},
	}}
}

func (program *Program) applyPullRequestCommentDeleteRequested(message MsgPullRequestCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return []Cmd{actionsPopupAsyncCmd{request: deletePullRequestCommentPopupRequest{target: message.Target}}}
}

func (program *Program) applyInlineCommentUpdateRequested(message MsgInlineCommentUpdateRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.commentID) == "" {
				return errors.New("missing inline comment identity")
			}
			if !program.hasReviewMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.reviewMutations.UpdatePullRequestReviewComment(message.Target.commentID, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.optimisticallyUpdateReviewComment(message.Target, message.Body)
			program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: inlineCommentUpdatedSuccessMessage})
			return nil
		},
	}}
}

func (program *Program) applyInlineCommentDeleteRequested(message MsgInlineCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasReviewMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return []Cmd{actionsPopupAsyncCmd{request: deleteInlineCommentPopupRequest{target: message.Target}}}
}

func (program *Program) applyInlineCommentReplySubmitRequested(message MsgInlineCommentReplySubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.threadID) == "" {
				return errors.New("missing inline comment thread identity")
			}
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasReviewMutations() {
				return errors.New("github loader is unavailable")
			}
			if err := program.reviewMutations.AddPullRequestReviewThreadReply(message.Target.pendingReview, message.Target.threadID, body); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.optimisticallyAppendInlineCommentReply(message.Target, message.Body)
			program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestInlineCommentReplySuccessMessage})
			return nil
		},
	}}
}

func (program *Program) applyInlineCommentResolutionRequested(message MsgInlineCommentResolutionRequested) []Cmd {
	if strings.TrimSpace(message.Target.threadID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasReviewMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return []Cmd{actionsPopupAsyncCmd{request: inlineCommentResolutionPopupRequest{target: message.Target, resolved: message.Resolved}}}
}

func (program *Program) applyReviewInlineCommentSubmitRequested(message MsgReviewInlineCommentSubmitRequested) []Cmd {
	submittedTarget := message.Target
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(submittedTarget.repository) == "" || submittedTarget.number <= 0 {
				return errors.New("missing pull request identity")
			}
			if !program.hasReviewMutations() {
				return errors.New("github loader is unavailable")
			}
			pendingReviewID, err := program.pendingReviewIDForInlineCommentMutation(submittedTarget)
			if err != nil {
				return err
			}
			if err := program.reviewMutations.AddPullRequestReviewThread(pendingReviewID, body, submittedTarget.threadTarget); err != nil {
				return newTransientErrorPopupActionError(err)
			}
			submittedTarget.pendingReview = pendingReviewID
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.optimisticallyAppendInlineComment(submittedTarget, message.Body)
			program.detailState.viewState.exitVisualMode()
			program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestReviewInlineCommentSuccessMessage})
			return nil
		},
	}}
}

func (program *Program) applyPendingPullRequestReviewSubmitRequested(message MsgPendingPullRequestReviewSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{
		Text: message.Body,
		Submit: func(body string) error {
			if strings.TrimSpace(message.Target.repository) == "" || message.Target.number <= 0 || strings.TrimSpace(message.Target.pendingReviewID) == "" {
				return pendingReviewSubmitError(message.Event, message.FeedbackTarget, errors.New("missing pull request review context"))
			}
			if !program.hasReviewMutations() {
				return pendingReviewSubmitError(message.Event, message.FeedbackTarget, errors.New("github loader is unavailable"))
			}
			if err := program.reviewMutations.SubmitPullRequestReview(message.Target.pendingReviewID, message.Event, body); err != nil {
				return pendingReviewSubmitError(message.Event, message.FeedbackTarget, newTransientErrorPopupActionError(err))
			}
			return nil
		},
		AfterSubmit: func(program *Program) []Cmd {
			program.applyPendingPullRequestReviewSubmitted(MsgPendingPullRequestReviewSubmitted{Target: message.Target})
			return nil
		},
	}}
}

func (program *Program) pendingReviewIDForInlineCommentMutation(target pullRequestInlineCommentTarget) (string, error) {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return "", errors.New("missing pull request identity")
	}
	if strings.TrimSpace(target.pendingReview) != "" {
		program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, target.pendingReview)
		return strings.TrimSpace(target.pendingReview), nil
	}
	if !program.hasReviewMutations() {
		return "", errors.New("github loader is unavailable")
	}

	pendingReviewID, err := program.reviewMutations.StartPendingPullRequestReview(target.repository, target.number)
	if err != nil {
		return "", newTransientErrorPopupActionError(err)
	}
	pendingReviewID = strings.TrimSpace(pendingReviewID)
	if pendingReviewID == "" {
		return "", errors.New("missing pull request review context")
	}
	program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, pendingReviewID)
	return pendingReviewID, nil
}

func pendingReviewSubmitError(event githubdomain.PullRequestReviewEvent, feedbackTarget Focus, err error) error {
	if event == githubdomain.PullRequestReviewEventRequestChanges {
		return newModalEditorStatusLineError(feedbackTarget, err)
	}
	return err
}

func (program *Program) applyReactionRemovalRequested(message MsgReactionRemovalRequested) []Cmd {
	if strings.TrimSpace(message.Target.subjectID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !reactionGroupViewerHasReacted(message.Target.reactionGroups, message.Target.content) {
		return Update(program, MsgActionsPopupClosedWithFeedback{Target: program.model.Focus(), Message: pullRequestReactionAlreadyRemovedMessage})
	}
	if !program.hasReactionMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return []Cmd{actionsPopupAsyncCmd{request: removeReactionPopupRequest{target: message.Target}}}
}

func (program *Program) applyPullRequestSquashMergeRequested(message MsgPullRequestSquashMergeRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return []Cmd{actionsPopupAsyncCmd{request: pullRequestSquashMergePopupRequest{repository: repository, number: number, summary: message.Summary}}}
}
