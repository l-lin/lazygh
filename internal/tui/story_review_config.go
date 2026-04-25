package tui

import "codeberg.org/l-lin/lazygh/internal/story"

func (program *Program) ApplyStoryReviewConfig(config story.Config) {
	program.storyReviewConfig = story.ResolveConfig(config)
}
