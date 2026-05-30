package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/story"
)

type MsgKeymapOverridesApplied struct {
	Overrides appconfig.KeymapOverrides
}

type MsgPullRequestSearchesApplied struct {
	Searches []appconfig.PullRequestSearch
}

type MsgLinksConfigApplied struct {
	Config appconfig.LinksConfig
}

type MsgCacheConfigApplied struct {
	PullRequestCache      persistentPullRequestCache
	NotificationDoneStore notificationDoneStore
}

type MsgStoryReviewConfigApplied struct {
	Config story.Config
}

func (MsgKeymapOverridesApplied) isMsg()     {}
func (MsgPullRequestSearchesApplied) isMsg() {}
func (MsgLinksConfigApplied) isMsg()         {}
func (MsgCacheConfigApplied) isMsg()         {}
func (MsgStoryReviewConfigApplied) isMsg()   {}
