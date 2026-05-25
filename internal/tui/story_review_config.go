package tui

import "github.com/l-lin/lazygh/internal/story"

func (program *Program) ApplyStoryReviewConfig(config story.Config) {
	program.runtimeConfig.storyReviewConfig = story.ResolveConfig(config)
}
