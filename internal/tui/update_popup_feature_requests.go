package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyNotificationReadRequested(message MsgNotificationReadRequested) []Cmd {
	target, ok := program.resolveNotificationRequestTarget(message.Target)
	if !ok {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	if !target.notification.Unread {
		return program.applyNotificationFeedbackAndClose(notificationAlreadyReadMessage)
	}

	optimisticNotifications := program.loadedNotifications()
	if !markNotificationReadState(optimisticNotifications, target.threadID, false) {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationReadLoadingMessage, notificationMarkedReadMessage, notificationReadMutationRequest{threadID: target.threadID})
}

func (program *Program) applyNotificationDoneRequested(message MsgNotificationDoneRequested) []Cmd {
	target, ok := program.resolveNotificationRequestTarget(message.Target)
	if !ok {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}

	optimisticNotifications, removed := removeNotificationWithThreadID(program.loadedNotifications(), target.threadID)
	if !removed {
		return program.handleNotificationRequestUnavailable(errActionsPopupActionUnavailable.Error())
	}
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationDoneLoadingMessage, notificationMarkedDoneMessage, notificationDoneMutationRequest{threadID: target.threadID, notification: target.notification})
}

func (program *Program) resolveNotificationRequestTarget(target notificationActionTarget) (notificationActionTarget, bool) {
	if strings.TrimSpace(target.threadID) != "" {
		return target, true
	}
	return program.selectedNotificationActionTarget()
}

func (program *Program) applyAllNotificationsReadRequested() []Cmd {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		return program.applyNotificationFeedbackAndClose(notificationNoNotificationsLoadedMessage)
	}

	optimisticNotifications := append([]githubdomain.Notification(nil), loadedNotifications...)
	markAllNotificationsRead(optimisticNotifications)
	return program.beginNotificationMutation(notificationRows(optimisticNotifications), notificationAllReadLoadingMessage, notificationMarkedAllReadMessage, allNotificationsReadMutationRequest{})
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
	return program.beginNotificationMutation(notificationRows(nil), loadingMessage, notificationMarkedAllDoneMessage, allNotificationsDoneMutationRequest{notifications: append([]githubdomain.Notification(nil), loadedNotifications...)})
}

func formatNotificationDoneLoadingMessage(count int) string {
	return fmt.Sprintf("Marking %d notifications as done...", count)
}

func (program *Program) beginNotificationMutation(optimisticRows []NotificationRow, loadingMessage string, successFeedbackMessage string, request notificationMutationRequest) []Cmd {
	if !program.hasNotificationMutations() {
		return program.handleNotificationRequestUnavailable("github loader is unavailable")
	}

	snapshot := program.captureNotificationMutationSnapshot()
	program.applyNotificationMutationStarted(MsgNotificationMutationStarted{OptimisticRows: optimisticRows, LoadingMessage: loadingMessage})
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{notificationMutationCmd{Snapshot: snapshot, SuccessFeedbackMessage: successFeedbackMessage, request: request}}
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
	return []Cmd{storyReviewPrepareCmd{request: pullRequestStoryReviewPrepareRequest{summary: message.Summary}}}
}
