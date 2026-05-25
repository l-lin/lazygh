package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgNotificationReadRequested struct {
	Target notificationActionTarget
}

type MsgNotificationDoneRequested struct {
	Target notificationActionTarget
}

type MsgAllNotificationsReadRequested struct{}

type MsgAllNotificationsDoneRequested struct{}

type MsgReviewStoryRequested struct {
	Summary githubdomain.PullRequest
}

func (MsgNotificationReadRequested) isMsg()     {}
func (MsgNotificationDoneRequested) isMsg()     {}
func (MsgAllNotificationsReadRequested) isMsg() {}
func (MsgAllNotificationsDoneRequested) isMsg() {}
func (MsgReviewStoryRequested) isMsg()          {}
