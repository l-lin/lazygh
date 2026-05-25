package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

type notificationMutationCommandDeps struct {
	notificationMutations           NotificationMutations
	hideDoneNotificationsBestEffort func([]githubdomain.Notification)
}

type storyReviewPrepareCommandDeps struct {
	detailQueries               DetailQueries
	reviewMutations             ReviewMutations
	storyGenerator              reviewStoryGenerator
	storyReviewConfig           story.Config
	pullRequestDetailForSummary func(githubdomain.PullRequest) (pullRequestDetailResult, bool)
}

type notificationReadMutationRequest struct {
	threadID string
}

func (request notificationReadMutationRequest) run(deps notificationMutationCommandDeps) error {
	if deps.notificationMutations == nil {
		return normalizedNotificationMutationError(errors.New("github loader is unavailable"))
	}
	return normalizedNotificationMutationError(deps.notificationMutations.MarkNotificationRead(request.threadID))
}

type notificationDoneMutationRequest struct {
	threadID     string
	notification githubdomain.Notification
}

func (request notificationDoneMutationRequest) run(deps notificationMutationCommandDeps) error {
	if deps.notificationMutations == nil {
		return normalizedNotificationMutationError(errors.New("github loader is unavailable"))
	}
	if err := normalizedNotificationMutationError(deps.notificationMutations.MarkNotificationDone(request.threadID)); err != nil {
		return err
	}
	if deps.hideDoneNotificationsBestEffort != nil {
		deps.hideDoneNotificationsBestEffort([]githubdomain.Notification{request.notification})
	}
	return nil
}

type allNotificationsReadMutationRequest struct{}

func (allNotificationsReadMutationRequest) run(deps notificationMutationCommandDeps) error {
	if deps.notificationMutations == nil {
		return normalizedNotificationMutationError(errors.New("github loader is unavailable"))
	}
	_, err := deps.notificationMutations.MarkAllNotificationsRead()
	return normalizedNotificationMutationError(err)
}

type allNotificationsDoneMutationRequest struct {
	notifications []githubdomain.Notification
}

func (request allNotificationsDoneMutationRequest) run(deps notificationMutationCommandDeps) error {
	if deps.notificationMutations == nil {
		return normalizedNotificationMutationError(errors.New("github loader is unavailable"))
	}
	_, err := deps.notificationMutations.MarkAllNotificationsDone(request.notifications)
	if err = normalizedNotificationMutationError(err); err != nil {
		return err
	}
	if deps.hideDoneNotificationsBestEffort != nil {
		deps.hideDoneNotificationsBestEffort(request.notifications)
	}
	return nil
}

type pullRequestStoryReviewPrepareRequest struct {
	summary githubdomain.PullRequest
}

func (request pullRequestStoryReviewPrepareRequest) run(deps storyReviewPrepareCommandDeps) (preparedStoryReview, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(request.summary.Repository))
	if repository == "" || repository == "-" || request.summary.Number <= 0 {
		return preparedStoryReview{}, errors.New("missing pull request identity")
	}
	if deps.detailQueries == nil || deps.reviewMutations == nil || deps.storyGenerator == nil {
		return preparedStoryReview{}, errors.New("github loader is unavailable")
	}

	detail, detailOK := storyReviewDetail(deps, request.summary)
	rawDiff, err := deps.detailQueries.GetPullRequestDiff(repository, request.summary.Number)
	if err != nil {
		return preparedStoryReview{}, newTransientErrorPopupActionError(err)
	}

	generatedStory, err := deps.storyGenerator.Generate(deps.storyReviewConfig, story.Request{
		Metadata:  buildStoryReviewMetadata(request.summary, detail, detailOK, rawDiff),
		DiffItems: buildStoryReviewDiffItems(rawDiff.Files),
		DiffText:  rawDiff.UnifiedDiff,
	})
	if err != nil {
		return preparedStoryReview{}, err
	}

	pendingReviewID, err := deps.reviewMutations.StartPendingPullRequestReview(repository, request.summary.Number)
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

func storyReviewDetail(deps storyReviewPrepareCommandDeps, summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, bool) {
	if deps.pullRequestDetailForSummary != nil {
		if result, ok := deps.pullRequestDetailForSummary(summary); ok && result.err == nil {
			return result.detail, true
		}
	}
	if deps.detailQueries == nil {
		return githubdomain.PullRequestDetail{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return githubdomain.PullRequestDetail{}, false
	}
	detail, err := deps.detailQueries.GetPullRequestDetail(repository, summary.Number)
	if err != nil {
		return githubdomain.PullRequestDetail{}, false
	}
	return detail, true
}
