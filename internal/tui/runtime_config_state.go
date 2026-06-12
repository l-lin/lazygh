package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/story"
)

func (state runtimeConfigState) withKeymapOverrides(overrides appconfig.KeymapOverrides) runtimeConfigState {
	state.keymapOverrides = copyKeymapOverrides(overrides)
	return state
}

func (state runtimeConfigState) withPullRequestSearches(searches []appconfig.PullRequestSearch) runtimeConfigState {
	state.pullRequestSearches = appconfig.ResolvePullRequestSearches(searches)
	return state
}

func (state runtimeConfigState) withDisplayConfig(config appconfig.DisplayConfig) runtimeConfigState {
	state.displayConfig = appconfig.ResolveDisplayConfig(config)
	return state
}

func (state runtimeConfigState) withStoryReviewConfig(config story.Config) runtimeConfigState {
	state.storyReviewConfig = story.ResolveConfig(config)
	return state
}
