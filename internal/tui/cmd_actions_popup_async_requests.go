package tui

import (
	"errors"
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type actionsPopupAsyncCommandDeps struct {
	pullRequestMutations PullRequestMutations
	reviewMutations      ReviewMutations
	reactionMutations    ReactionMutations
}

type actionsPopupAsyncRequest interface {
	statusCommand() string
	asyncRequested() bool
	run(actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error)
}

func newActionsPopupAsyncCommandDeps(program *Program) actionsPopupAsyncCommandDeps {
	if program == nil {
		return actionsPopupAsyncCommandDeps{}
	}
	return actionsPopupAsyncCommandDeps{
		pullRequestMutations: program.pullRequestMutations,
		reviewMutations:      program.reviewMutations,
		reactionMutations:    program.reactionMutations,
	}
}

type startPullRequestReviewPopupRequest struct {
	summary githubdomain.PullRequest
}

func (request startPullRequestReviewPopupRequest) statusCommand() string {
	repository := strings.TrimSpace(pullRequestRepositoryName(request.summary.Repository))
	if repository == "" || repository == "-" || request.summary.Number <= 0 {
		return ""
	}
	return formatStatusLineCommand("start", "review", repository, fmt.Sprintf("#%d", request.summary.Number))
}

func (startPullRequestReviewPopupRequest) asyncRequested() bool {
	return false
}

func (request startPullRequestReviewPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	pendingReviewID, err := startPendingPullRequestReview(deps, request.summary)
	if err != nil {
		return nil, err
	}
	return actionsPopupAsyncStartReviewSuccess{Summary: request.summary, PendingReviewID: pendingReviewID}, nil
}

type openPullRequestInBrowserPopupRequest struct {
	repository string
	number     int
}

func (request openPullRequestInBrowserPopupRequest) statusCommand() string {
	return openPullRequestInBrowserCommand(request.repository, request.number)
}

func (openPullRequestInBrowserPopupRequest) asyncRequested() bool {
	return true
}

func (request openPullRequestInBrowserPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.OpenPullRequestInBrowser(request.repository, request.number); err != nil {
		return nil, err
	}
	return actionsPopupAsyncFeedbackSuccess{Message: pullRequestBrowserOpenSuccessMessage}, nil
}

type approvePullRequestPopupRequest struct {
	repository string
	number     int
}

func (request approvePullRequestPopupRequest) statusCommand() string {
	return approvePullRequestCommand(request.repository, request.number)
}

func (approvePullRequestPopupRequest) asyncRequested() bool {
	return true
}

func (request approvePullRequestPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.ApprovePullRequest(request.repository, request.number); err != nil {
		return nil, err
	}
	return actionsPopupAsyncInvalidatePullRequestSuccess{
		Repository:     request.repository,
		Number:         request.number,
		InvalidateDiff: true,
		Message:        pullRequestReviewSuccessMessage,
	}, nil
}

type reRequestPullRequestReviewPopupRequest struct {
	repository    string
	number        int
	reviewerLogin string
}

func (request reRequestPullRequestReviewPopupRequest) statusCommand() string {
	return requestPullRequestReviewerCommand(request.repository, request.number, request.reviewerLogin)
}

func (reRequestPullRequestReviewPopupRequest) asyncRequested() bool {
	return true
}

func (request reRequestPullRequestReviewPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.RequestPullRequestReviewer(request.repository, request.number, request.reviewerLogin); err != nil {
		return nil, err
	}
	return actionsPopupAsyncInvalidatePullRequestSuccess{
		Repository: request.repository,
		Number:     request.number,
		Message:    pullRequestReviewReRequestedSuccessMessage,
	}, nil
}

type pullRequestLifecycleMutationPopupRequest struct {
	kind           pullRequestLifecycleMutationKind
	repository     string
	number         int
	summary        githubdomain.PullRequest
	state          string
	isDraft        bool
	successMessage string
}

func (request pullRequestLifecycleMutationPopupRequest) statusCommand() string {
	return pullRequestLifecycleMutationCommand(request.kind, request.repository, request.number)
}

func (pullRequestLifecycleMutationPopupRequest) asyncRequested() bool {
	return true
}

func (request pullRequestLifecycleMutationPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if err := runPullRequestLifecycleMutation(deps, request.kind, request.repository, request.number); err != nil {
		return nil, err
	}
	return actionsPopupAsyncPullRequestLifecycleSuccess{
		Summary: request.summary,
		State:   request.state,
		IsDraft: request.isDraft,
		Message: request.successMessage,
	}, nil
}

type pullRequestAutoMergeMutationPopupRequest struct {
	kind           pullRequestAutoMergeMutationKind
	repository     string
	number         int
	summary        githubdomain.PullRequest
	enabled        bool
	successMessage string
}

func (request pullRequestAutoMergeMutationPopupRequest) statusCommand() string {
	return pullRequestAutoMergeMutationCommand(request.kind, request.repository, request.number)
}

func (pullRequestAutoMergeMutationPopupRequest) asyncRequested() bool {
	return true
}

func (request pullRequestAutoMergeMutationPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if err := runPullRequestAutoMergeMutation(deps, request.kind, request.repository, request.number); err != nil {
		return nil, err
	}
	return actionsPopupAsyncPullRequestAutoMergeSuccess{
		Summary: request.summary,
		Enabled: request.enabled,
		Message: request.successMessage,
	}, nil
}

type pullRequestBranchUpdatePopupRequest struct {
	repository string
	number     int
	summary    githubdomain.PullRequest
}

func (request pullRequestBranchUpdatePopupRequest) statusCommand() string {
	return updatePullRequestBranchCommand(request.repository, request.number)
}

func (pullRequestBranchUpdatePopupRequest) asyncRequested() bool {
	return true
}

func (request pullRequestBranchUpdatePopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := normalizedPullRequestMutationError(deps.pullRequestMutations.UpdatePullRequestBranch(request.repository, request.number), "gh pr update-branch"); err != nil {
		return nil, err
	}
	return actionsPopupAsyncPullRequestBranchUpdateSuccess{Summary: request.summary, Message: pullRequestBranchUpdatedSuccessMessage}, nil
}

type updatePullRequestAssigneesPopupRequest struct {
	repository   string
	number       int
	addLogins    []string
	removeLogins []string
}

func (request updatePullRequestAssigneesPopupRequest) statusCommand() string {
	return updatePullRequestAssigneesCommand(request.repository, request.number, request.addLogins, request.removeLogins)
}

func (updatePullRequestAssigneesPopupRequest) asyncRequested() bool {
	return true
}

func (request updatePullRequestAssigneesPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	err := normalizedAssigneePickerError(deps.pullRequestMutations.UpdatePullRequestAssignees(request.repository, request.number, request.addLogins, request.removeLogins))
	if err != nil {
		return nil, err
	}
	return actionsPopupAsyncPullRequestAssigneesUpdatedSuccess{
		Repository:   request.repository,
		Number:       request.number,
		AddLogins:    request.addLogins,
		RemoveLogins: request.removeLogins,
		Message:      pullRequestAssigneesUpdatedSuccessMessage,
	}, nil
}

type addReactionPopupRequest struct {
	target  pullRequestReactionActionTarget
	content githubdomain.ReactionContent
}

func (request addReactionPopupRequest) statusCommand() string {
	return formatStatusLineCommand("add", "reaction", strings.TrimSpace(string(request.content)))
}

func (addReactionPopupRequest) asyncRequested() bool {
	return false
}

func (request addReactionPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reactionMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reactionMutations.AddReaction(request.target.subjectID, request.content); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncReactionAddedSuccess{Target: request.target, Content: request.content}, nil
}

type cancelPendingPullRequestReviewPopupRequest struct {
	target pendingPullRequestReviewActionTarget
}

func (request cancelPendingPullRequestReviewPopupRequest) statusCommand() string {
	return formatStatusLineCommand("cancel", "pending", "review", request.target.repository, fmt.Sprintf("#%d", request.target.number))
}

func (cancelPendingPullRequestReviewPopupRequest) asyncRequested() bool {
	return false
}

func (request cancelPendingPullRequestReviewPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.DeletePullRequestReview(request.target.pendingReviewID); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncPendingReviewCanceledSuccess{Target: request.target}, nil
}

type deletePullRequestCommentPopupRequest struct {
	target pullRequestCommentEditActionTarget
}

func (deletePullRequestCommentPopupRequest) statusCommand() string {
	return ""
}

func (deletePullRequestCommentPopupRequest) asyncRequested() bool {
	return false
}

func (request deletePullRequestCommentPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.pullRequestMutations.DeletePullRequestComment(request.target.commentID); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncPullRequestCommentDeletedSuccess{Target: request.target}, nil
}

type deleteInlineCommentPopupRequest struct {
	target pullRequestReviewCommentActionTarget
}

func (deleteInlineCommentPopupRequest) statusCommand() string {
	return ""
}

func (deleteInlineCommentPopupRequest) asyncRequested() bool {
	return false
}

func (request deleteInlineCommentPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reviewMutations.DeletePullRequestReviewComment(request.target.commentID); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncInlineCommentDeletedSuccess{Target: request.target}, nil
}

type inlineCommentResolutionPopupRequest struct {
	target   pullRequestReviewThreadActionTarget
	resolved bool
}

func (inlineCommentResolutionPopupRequest) statusCommand() string {
	return ""
}

func (inlineCommentResolutionPopupRequest) asyncRequested() bool {
	return false
}

func (request inlineCommentResolutionPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reviewMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	var err error
	if request.resolved {
		err = deps.reviewMutations.ResolvePullRequestReviewThread(request.target.threadID)
	} else {
		err = deps.reviewMutations.UnresolvePullRequestReviewThread(request.target.threadID)
	}
	if err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncInlineCommentResolutionSuccess{Target: request.target, Resolved: request.resolved}, nil
}

type removeReactionPopupRequest struct {
	target pullRequestReactionRemovalTarget
}

func (removeReactionPopupRequest) statusCommand() string {
	return ""
}

func (removeReactionPopupRequest) asyncRequested() bool {
	return false
}

func (request removeReactionPopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.reactionMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := deps.reactionMutations.RemoveReaction(request.target.subjectID, request.target.content); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncReactionRemovedSuccess{Target: request.target}, nil
}

type pullRequestSquashMergePopupRequest struct {
	repository string
	number     int
	summary    githubdomain.PullRequest
}

func (request pullRequestSquashMergePopupRequest) statusCommand() string {
	return squashMergePullRequestCommand(request.repository, request.number)
}

func (pullRequestSquashMergePopupRequest) asyncRequested() bool {
	return true
}

func (request pullRequestSquashMergePopupRequest) run(deps actionsPopupAsyncCommandDeps) (actionsPopupAsyncSuccess, error) {
	if deps.pullRequestMutations == nil {
		return nil, errors.New("github loader is unavailable")
	}
	if err := normalizedPullRequestMutationError(deps.pullRequestMutations.SquashMergePullRequest(request.repository, request.number), "gh pr merge"); err != nil {
		return nil, newTransientErrorPopupActionError(err)
	}
	return actionsPopupAsyncPullRequestLifecycleSuccess{
		Summary: request.summary,
		State:   "MERGED",
		IsDraft: false,
		Message: pullRequestSquashMergedSuccessMessage,
	}, nil
}

func startPendingPullRequestReview(deps actionsPopupAsyncCommandDeps, summary githubdomain.PullRequest) (string, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return "", errors.New("missing pull request identity")
	}
	if deps.reviewMutations == nil {
		return "", errors.New("github loader is unavailable")
	}

	pendingReviewID, err := deps.reviewMutations.StartPendingPullRequestReview(repository, summary.Number)
	if err != nil {
		return "", newTransientErrorPopupActionError(err)
	}
	return pendingReviewID, nil
}

func runPullRequestLifecycleMutation(deps actionsPopupAsyncCommandDeps, kind pullRequestLifecycleMutationKind, repository string, number int) error {
	if deps.pullRequestMutations == nil {
		return errors.New("github loader is unavailable")
	}
	switch kind {
	case pullRequestLifecycleMutationReadyForReview:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.MarkPullRequestReadyForReview(repository, number), "gh pr ready")
	case pullRequestLifecycleMutationConvertToDraft:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.ConvertPullRequestToDraft(repository, number), "gh pr ready")
	case pullRequestLifecycleMutationClose:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.ClosePullRequest(repository, number), "gh pr close")
	case pullRequestLifecycleMutationReopen:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.ReopenPullRequest(repository, number), "gh pr reopen")
	default:
		return errActionsPopupActionUnavailable
	}
}

func runPullRequestAutoMergeMutation(deps actionsPopupAsyncCommandDeps, kind pullRequestAutoMergeMutationKind, repository string, number int) error {
	if deps.pullRequestMutations == nil {
		return errors.New("github loader is unavailable")
	}
	switch kind {
	case pullRequestAutoMergeMutationEnable:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.EnablePullRequestAutoMerge(repository, number), "gh pr merge")
	case pullRequestAutoMergeMutationDisable:
		return normalizedPullRequestMutationError(deps.pullRequestMutations.DisablePullRequestAutoMerge(repository, number), "gh pr merge")
	default:
		return errActionsPopupActionUnavailable
	}
}
