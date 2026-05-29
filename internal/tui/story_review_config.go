package tui

import "github.com/l-lin/lazygh/internal/story"

func (program *Program) ApplyStoryReviewConfig(config story.Config) {
	_ = program.dispatchRuntimeMessage(MsgStoryReviewConfigApplied{Config: config})
}
