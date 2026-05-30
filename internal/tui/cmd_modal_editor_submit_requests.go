package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type modalEditorSubmitCommandDeps struct {
	detailQueries          DetailQueries
	pullRequestListQueries PullRequestListQueries
	pullRequestMutations   PullRequestMutations
	reviewMutations        ReviewMutations
}

type modalEditorSubmitRequest interface {
	statusCommand() string
	asyncRequested() bool
	run(modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error)
}

func newModalEditorSubmitCommandDeps(program *Program) modalEditorSubmitCommandDeps {
	if program == nil {
		return modalEditorSubmitCommandDeps{}
	}
	return modalEditorSubmitCommandDeps{
		detailQueries:          program.detailQueries,
		pullRequestListQueries: program.pullRequestListQueries,
		pullRequestMutations:   program.pullRequestMutations,
		reviewMutations:        program.reviewMutations,
	}
}

type openPullRequestByURLSubmitRequest struct {
	rawURL string
}

func (request openPullRequestByURLSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	summary, err := pullRequestSummaryForURL(deps, request.rawURL)
	if err != nil {
		return nil, err
	}
	return openPullRequestInBrowserViewCompletion{Summary: summary}, nil
}

type pullRequestCustomSearchSubmitRequest struct {
	criteria string
}

func (request pullRequestCustomSearchSubmitRequest) run(modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if len(pullRequestCustomSearchCommand(request.criteria)) == 0 {
		return nil, errors.New("search criteria cannot be empty")
	}
	return pullRequestCustomSearchSubmittedCompletion{Criteria: request.criteria}, nil
}

type pullRequestCommentSubmitRequest struct {
	target         pullRequestCommentTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestCommentSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.CommentOnPullRequest(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestCommentSubmittedCompletion{Target: request.target, Body: request.body, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestReviewCommentSubmitRequest struct {
	target         pullRequestActionTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestReviewCommentSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.ReviewPullRequestWithComment(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestInvalidatedWithFeedbackCompletion{Repository: request.target.repository, Number: request.target.number, InvalidateDiff: true, FeedbackTarget: request.feedbackTarget, Message: pullRequestReviewSuccessMessage}, nil
}

type pullRequestRequestChangesSubmitRequest struct {
	target         pullRequestActionTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestRequestChangesSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.RequestChangesOnPullRequest(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestInvalidatedWithFeedbackCompletion{Repository: request.target.repository, Number: request.target.number, InvalidateDiff: true, FeedbackTarget: request.feedbackTarget, Message: pullRequestReviewSuccessMessage}, nil
}

type pullRequestTitleEditSubmitRequest struct {
	target         pullRequestActionTarget
	title          string
	feedbackTarget Focus
}

func (request pullRequestTitleEditSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.EditPullRequestTitle(request.target.repository, request.target.number, request.title); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestTitleEditAppliedCompletion{Target: request.target, Title: request.title, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestDescriptionEditSubmitRequest struct {
	target         pullRequestActionTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestDescriptionEditSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.EditPullRequestDescription(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestDescriptionEditAppliedCompletion{Target: request.target, Body: request.body, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestCommentUpdateSubmitRequest struct {
	target pullRequestCommentEditActionTarget
	body   string
}

func (request pullRequestCommentUpdateSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.commentID) == "" {
		return nil, errors.New("missing pull request comment identity")
	}
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.UpdatePullRequestComment(request.target.commentID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestCommentUpdatedCompletion{Target: request.target, Body: request.body}, nil
}

type inlineCommentUpdateSubmitRequest struct {
	target pullRequestReviewCommentActionTarget
	body   string
}

func (request inlineCommentUpdateSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.commentID) == "" {
		return nil, errors.New("missing inline comment identity")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.UpdatePullRequestReviewComment(request.target.commentID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return inlineCommentUpdatedCompletion{Target: request.target, Body: request.body}, nil
}

type inlineCommentReplySubmitRequest struct {
	target pullRequestReviewThreadReplyTarget
	body   string
}

func (request inlineCommentReplySubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.threadID) == "" {
		return nil, errors.New("missing inline comment thread identity")
	}
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.AddPullRequestReviewThreadReply(request.target.pendingReview, request.target.threadID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return inlineCommentReplySubmittedCompletion{Target: request.target, Body: request.body}, nil
}

type reviewInlineCommentSubmitRequest struct {
	target pullRequestInlineCommentTarget
	body   string
}

func (request reviewInlineCommentSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}

	preparedTarget := request.target
	pendingReviewID, err := pendingReviewIDForInlineCommentMutation(deps, preparedTarget)
	if err != nil {
		return nil, err
	}
	preparedTarget.pendingReview = pendingReviewID
	return reviewInlineCommentPendingReviewPreparedCompletion{Target: preparedTarget, Body: request.body}, nil
}

type preparedReviewInlineCommentSubmitRequest struct {
	target pullRequestInlineCommentTarget
	body   string
}

func (request preparedReviewInlineCommentSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	pendingReviewID := strings.TrimSpace(request.target.pendingReview)
	if pendingReviewID == "" {
		return nil, errors.New("missing pull request review context")
	}
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.AddPullRequestReviewThread(pendingReviewID, request.body, request.target.threadTarget); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return reviewInlineCommentSubmittedCompletion{Target: request.target, Body: request.body}, nil
}

type pendingPullRequestReviewSubmitRequest struct {
	target         pendingPullRequestReviewTarget
	event          githubdomain.PullRequestReviewEvent
	body           string
	feedbackTarget Focus
}

func (request pendingPullRequestReviewSubmitRequest) run(deps modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 || strings.TrimSpace(request.target.pendingReviewID) == "" {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, errors.New("missing pull request review context"))
	}
	if deps.reviewMutations == nil {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, errors.New("github loader is unavailable"))
	}
	if err := deps.reviewMutations.SubmitPullRequestReview(request.target.pendingReviewID, request.event, request.body); err != nil {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, newTransientErrorPopupActionError(err))
	}
	return pendingPullRequestReviewSubmittedCompletion{Target: request.target}, nil
}

func pendingReviewIDForInlineCommentMutation(deps modalEditorSubmitCommandDeps, target pullRequestInlineCommentTarget) (string, error) {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return "", errors.New("missing pull request identity")
	}
	if strings.TrimSpace(target.pendingReview) != "" {
		return strings.TrimSpace(target.pendingReview), nil
	}
	if deps.reviewMutations == nil {
		return "", errors.New("github loader is unavailable")
	}

	pendingReviewID, err := deps.reviewMutations.StartPendingPullRequestReview(target.repository, target.number)
	if err != nil {
		return "", newTransientErrorPopupActionError(err)
	}
	pendingReviewID = strings.TrimSpace(pendingReviewID)
	if pendingReviewID == "" {
		return "", errors.New("missing pull request review context")
	}
	return pendingReviewID, nil
}

func pullRequestSummaryForURL(deps modalEditorSubmitCommandDeps, rawURL string) (githubdomain.PullRequest, error) {
	if deps.detailQueries == nil && deps.pullRequestListQueries == nil {
		return githubdomain.PullRequest{}, errors.New("github loader is unavailable")
	}

	summary, err := githubdomain.ParsePullRequestURL(rawURL)
	if err != nil {
		return githubdomain.PullRequest{}, err
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return githubdomain.PullRequest{}, errors.New("missing pull request identity")
	}
	summary.Repository.NameWithOwner = repository
	return summary, nil
}
