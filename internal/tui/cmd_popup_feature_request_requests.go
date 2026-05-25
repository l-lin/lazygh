package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

type notificationReadMutationRequest struct {
	threadID string
}

func (request notificationReadMutationRequest) run(program *Program) error {
	return normalizedNotificationMutationError(program.notificationMutations.MarkNotificationRead(request.threadID))
}

type notificationDoneMutationRequest struct {
	threadID     string
	notification githubdomain.Notification
}

func (request notificationDoneMutationRequest) run(program *Program) error {
	if err := normalizedNotificationMutationError(program.notificationMutations.MarkNotificationDone(request.threadID)); err != nil {
		return err
	}
	program.hideDoneNotificationsBestEffort([]githubdomain.Notification{request.notification})
	return nil
}

type allNotificationsReadMutationRequest struct{}

func (allNotificationsReadMutationRequest) run(program *Program) error {
	_, err := program.notificationMutations.MarkAllNotificationsRead()
	return normalizedNotificationMutationError(err)
}

type allNotificationsDoneMutationRequest struct {
	notifications []githubdomain.Notification
}

func (request allNotificationsDoneMutationRequest) run(program *Program) error {
	_, err := program.notificationMutations.MarkAllNotificationsDone(request.notifications)
	if err = normalizedNotificationMutationError(err); err != nil {
		return err
	}
	program.hideDoneNotificationsBestEffort(request.notifications)
	return nil
}

type pullRequestStoryReviewPrepareRequest struct {
	summary githubdomain.PullRequest
}

func (request pullRequestStoryReviewPrepareRequest) run(program *Program) (preparedStoryReview, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(request.summary.Repository))
	if repository == "" || repository == "-" || request.summary.Number <= 0 {
		return preparedStoryReview{}, errors.New("missing pull request identity")
	}

	detail, detailOK := storyReviewDetail(program, request.summary)
	rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, request.summary.Number)
	if err != nil {
		return preparedStoryReview{}, newTransientErrorPopupActionError(err)
	}

	generatedStory, err := program.storyGenerator.Generate(program.runtimeConfig.storyReviewConfig, story.Request{
		Metadata:  buildStoryReviewMetadata(request.summary, detail, detailOK, rawDiff),
		DiffItems: buildStoryReviewDiffItems(rawDiff.Files),
		DiffText:  rawDiff.UnifiedDiff,
	})
	if err != nil {
		return preparedStoryReview{}, err
	}

	pendingReviewID, err := program.reviewMutations.StartPendingPullRequestReview(repository, request.summary.Number)
	if err != nil {
		return preparedStoryReview{}, newTransientErrorPopupActionError(err)
	}

	diffData := buildReviewDiffData(rawDiff)
	return preparedStoryReview{
		summary:         request.summary,
		detail:          detail,
		detailOK:        detailOK,
		diffData:        diffData,
		storyData:       buildReviewStoryData(generatedStory, diffData.Files),
		pendingReviewID: pendingReviewID,
	}, nil
}

func storyReviewDetail(program *Program, summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, bool) {
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return result.detail, true
	}
	if !program.hasDetailQueries() {
		return githubdomain.PullRequestDetail{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return githubdomain.PullRequestDetail{}, false
	}
	detail, err := program.detailQueries.GetPullRequestDetail(repository, summary.Number)
	if err != nil {
		return githubdomain.PullRequestDetail{}, false
	}
	return detail, true
}
