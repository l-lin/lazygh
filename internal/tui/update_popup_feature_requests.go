package tui

import (
	"errors"
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

func (program *Program) applyNotificationReadRequested(message MsgNotificationReadRequested) []Cmd {
	target := message.Target
	if strings.TrimSpace(target.threadID) == "" {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	if !target.notification.Unread {
		return program.applyNotificationFeedbackAndClose(notificationAlreadyReadMessage)
	}

	optimisticNotifications := program.loadedNotifications()
	if !markNotificationReadState(optimisticNotifications, target.threadID, false) {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationReadLoadingMessage, notificationMarkedReadMessage, func(program *Program) error {
		return normalizedNotificationMutationError(program.notificationMutations.MarkNotificationRead(target.threadID))
	})
}

func (program *Program) applyNotificationDoneRequested(message MsgNotificationDoneRequested) []Cmd {
	target := message.Target
	if strings.TrimSpace(target.threadID) == "" {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}

	optimisticNotifications, removed := removeNotificationWithThreadID(program.loadedNotifications(), target.threadID)
	if !removed {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationDoneLoadingMessage, notificationMarkedDoneMessage, func(program *Program) error {
		if err := normalizedNotificationMutationError(program.notificationMutations.MarkNotificationDone(target.threadID)); err != nil {
			return err
		}
		program.hideDoneNotificationsBestEffort([]githubdomain.Notification{target.notification})
		return nil
	})
}

func (program *Program) applyAllNotificationsReadRequested() []Cmd {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		return program.applyNotificationFeedbackAndClose(notificationNoNotificationsLoadedMessage)
	}

	optimisticNotifications := append([]githubdomain.Notification(nil), loadedNotifications...)
	markAllNotificationsRead(optimisticNotifications)
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationAllReadLoadingMessage, notificationMarkedAllReadMessage, func(program *Program) error {
		_, err := program.notificationMutations.MarkAllNotificationsRead()
		return normalizedNotificationMutationError(err)
	})
}

func (program *Program) applyAllNotificationsDoneRequested() []Cmd {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		return program.applyNotificationFeedbackAndClose(notificationNoNotificationsLoadedMessage)
	}

	loadingMessage := "Marking 0 notifications as done..."
	if count := len(loadedNotifications); count > 0 {
		loadingMessage = formatNotificationDoneLoadingMessage(count)
	}
	return program.beginNotificationMutation(notificationRows(nil), loadingMessage, notificationMarkedAllDoneMessage, func(program *Program) error {
		_, err := program.notificationMutations.MarkAllNotificationsDone(loadedNotifications)
		if err = normalizedNotificationMutationError(err); err != nil {
			return err
		}
		program.hideDoneNotificationsBestEffort(loadedNotifications)
		return nil
	})
}

func formatNotificationDoneLoadingMessage(count int) string {
	return fmt.Sprintf("Marking %d notifications as done...", count)
}

func (program *Program) beginNotificationMutation(optimisticRows []NotificationRow, loadingMessage string, successFeedbackMessage string, work func(*Program) error) []Cmd {
	if !program.hasNotificationMutations() {
		return program.handleNotificationRequestUnavailable("github loader is unavailable")
	}

	snapshot := program.captureNotificationMutationSnapshot()
	program.applyNotificationMutationStarted(MsgNotificationMutationStarted{OptimisticRows: optimisticRows, LoadingMessage: loadingMessage})
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{notificationMutationCmd{Snapshot: snapshot, SuccessFeedbackMessage: successFeedbackMessage, Work: work}}
}

func (program *Program) handleNotificationRequestUnavailable(message string) []Cmd {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" || program == nil || program.model == nil {
		return nil
	}
	if program.model.ActionsPopupVisible() {
		program.actionsPopupWidget.errorMessage = trimmedMessage
		return nil
	}
	program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: trimmedMessage})
	return nil
}

func (program *Program) applyNotificationFeedbackAndClose(message string) []Cmd {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" || program == nil || program.model == nil {
		return nil
	}
	program.applyFeedbackSet(MsgFeedbackSet{Target: program.model.Focus(), Message: trimmedMessage})
	program.closeActionsPopupForAcceptedRequest()
	return nil
}

func (program *Program) closeActionsPopupForAcceptedRequest() {
	if program == nil || program.model == nil || !program.model.ActionsPopupVisible() {
		return
	}
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
}

func (program *Program) applyReviewStoryRequested(message MsgReviewStoryRequested) []Cmd {
	program.feedbackMessage = ""
	program.storyReviewLoading = true
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{storyReviewPrepareCmd{Summary: message.Summary}}
}

func (program *Program) prepareStoryReview(summary githubdomain.PullRequest) (preparedStoryReview, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return preparedStoryReview{}, errors.New("missing pull request identity")
	}

	detail, detailOK := program.storyReviewDetail(summary)
	rawDiff, actualErr := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
	if actualErr != nil {
		return preparedStoryReview{}, newTransientErrorPopupActionError(actualErr)
	}

	generatedStory, actualErr := program.storyGenerator.Generate(program.runtimeConfig.storyReviewConfig, story.Request{
		Metadata:  buildStoryReviewMetadata(summary, detail, detailOK, rawDiff),
		DiffItems: buildStoryReviewDiffItems(rawDiff.Files),
		DiffText:  rawDiff.UnifiedDiff,
	})
	if actualErr != nil {
		return preparedStoryReview{}, actualErr
	}

	pendingReviewID, actualErr := program.reviewMutations.StartPendingPullRequestReview(repository, summary.Number)
	if actualErr != nil {
		return preparedStoryReview{}, newTransientErrorPopupActionError(actualErr)
	}

	diffData := buildReviewDiffData(rawDiff)
	return preparedStoryReview{
		summary:         summary,
		detail:          detail,
		detailOK:        detailOK,
		diffData:        diffData,
		storyData:       buildReviewStoryData(generatedStory, diffData.Files),
		pendingReviewID: pendingReviewID,
	}, nil
}

func (program *Program) storyReviewDetail(summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, bool) {
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
	detail, actualErr := program.detailQueries.GetPullRequestDetail(repository, summary.Number)
	if actualErr != nil {
		return githubdomain.PullRequestDetail{}, false
	}
	return detail, true
}
