package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type modalEditorSubmitRequest interface {
	run(*Program) (modalEditorSubmitSuccess, error)
}

type openPullRequestByURLSubmitRequest struct {
	rawURL string
}

func (request openPullRequestByURLSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	summary, err := pullRequestSummaryForURL(program, request.rawURL)
	if err != nil {
		return nil, err
	}
	return openPullRequestByURLSubmitSuccess{Summary: summary}, nil
}

type pullRequestCustomSearchSubmitRequest struct {
	criteria string
}

func (request pullRequestCustomSearchSubmitRequest) run(*Program) (modalEditorSubmitSuccess, error) {
	if len(pullRequestCustomSearchCommand(request.criteria)) == 0 {
		return nil, errors.New("search criteria cannot be empty")
	}
	return pullRequestCustomSearchSubmitSuccess{Criteria: request.criteria}, nil
}

type pullRequestCommentSubmitRequest struct {
	target         pullRequestCommentTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestCommentSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.CommentOnPullRequest(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestCommentSubmitSuccess{Target: request.target, Body: request.body, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestReviewCommentSubmitRequest struct {
	target pullRequestActionTarget
	body   string
}

func (request pullRequestReviewCommentSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.ReviewPullRequestWithComment(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestReviewCommentSubmitSuccess{Repository: request.target.repository, Number: request.target.number}, nil
}

type pullRequestRequestChangesSubmitRequest struct {
	target pullRequestActionTarget
	body   string
}

func (request pullRequestRequestChangesSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.RequestChangesOnPullRequest(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestRequestChangesSubmitSuccess{Repository: request.target.repository, Number: request.target.number}, nil
}

type pullRequestTitleEditSubmitRequest struct {
	target         pullRequestActionTarget
	title          string
	feedbackTarget Focus
}

func (request pullRequestTitleEditSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.EditPullRequestTitle(request.target.repository, request.target.number, request.title); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestTitleEditSubmitSuccess{Target: request.target, Title: request.title, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestDescriptionEditSubmitRequest struct {
	target         pullRequestActionTarget
	body           string
	feedbackTarget Focus
}

func (request pullRequestDescriptionEditSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.EditPullRequestDescription(request.target.repository, request.target.number, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestDescriptionEditSubmitSuccess{Target: request.target, Body: request.body, FeedbackTarget: request.feedbackTarget}, nil
}

type pullRequestCommentUpdateSubmitRequest struct {
	target pullRequestCommentEditActionTarget
	body   string
}

func (request pullRequestCommentUpdateSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.commentID) == "" {
		return nil, errors.New("missing pull request comment identity")
	}
	if !program.hasPullRequestMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.UpdatePullRequestComment(request.target.commentID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return pullRequestCommentUpdateSubmitSuccess{Target: request.target, Body: request.body}, nil
}

type inlineCommentUpdateSubmitRequest struct {
	target pullRequestReviewCommentActionTarget
	body   string
}

func (request inlineCommentUpdateSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.commentID) == "" {
		return nil, errors.New("missing inline comment identity")
	}
	if !program.hasReviewMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.UpdatePullRequestReviewComment(request.target.commentID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return inlineCommentUpdateSubmitSuccess{Target: request.target, Body: request.body}, nil
}

type inlineCommentReplySubmitRequest struct {
	target pullRequestReviewThreadReplyTarget
	body   string
}

func (request inlineCommentReplySubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.threadID) == "" {
		return nil, errors.New("missing inline comment thread identity")
	}
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return nil, errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.AddPullRequestReviewThreadReply(request.target.pendingReview, request.target.threadID, request.body); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return inlineCommentReplySubmitSuccess{Target: request.target, Body: request.body}, nil
}

type reviewInlineCommentSubmitRequest struct {
	target pullRequestInlineCommentTarget
	body   string
}

func (request reviewInlineCommentSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 {
		return nil, errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return nil, errors.New("github loader is unavailable")
	}

	submittedTarget := request.target
	pendingReviewID, err := pendingReviewIDForInlineCommentMutation(program, submittedTarget)
	if err != nil {
		return nil, err
	}
	if err := program.reviewMutations.AddPullRequestReviewThread(pendingReviewID, request.body, submittedTarget.threadTarget); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	submittedTarget.pendingReview = pendingReviewID
	return reviewInlineCommentSubmitSuccess{Target: submittedTarget, Body: request.body}, nil
}

type pendingPullRequestReviewSubmitRequest struct {
	target         pendingPullRequestReviewTarget
	event          githubdomain.PullRequestReviewEvent
	body           string
	feedbackTarget Focus
}

func (request pendingPullRequestReviewSubmitRequest) run(program *Program) (modalEditorSubmitSuccess, error) {
	if strings.TrimSpace(request.target.repository) == "" || request.target.number <= 0 || strings.TrimSpace(request.target.pendingReviewID) == "" {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, errors.New("missing pull request review context"))
	}
	if !program.hasReviewMutations() {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, errors.New("github loader is unavailable"))
	}
	if err := program.reviewMutations.SubmitPullRequestReview(request.target.pendingReviewID, request.event, request.body); err != nil {
		return nil, pendingReviewSubmitError(request.event, request.feedbackTarget, newTransientErrorPopupActionError(err))
	}
	return pendingPullRequestReviewSubmitSuccess{Target: request.target}, nil
}

func pendingReviewIDForInlineCommentMutation(program *Program, target pullRequestInlineCommentTarget) (string, error) {
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

func pullRequestSummaryForURL(program *Program, rawURL string) (githubdomain.PullRequest, error) {
	if !program.hasDetailQueries() && !program.hasPullRequestListQueries() {
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
