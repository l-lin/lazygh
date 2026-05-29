package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/story"
)

func (program *Program) updateRuntimeConfig(transition func(runtimeConfigState) runtimeConfigState) {
	if program == nil {
		return
	}
	program.runtimeConfig = transition(program.runtimeConfig)
}

func (program *Program) setRuntimeKeymapOverrides(overrides appconfig.KeymapOverrides) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withKeymapOverrides(overrides)
	})
}

func (program *Program) setRuntimePullRequestSearches(searches []appconfig.PullRequestSearch) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withPullRequestSearches(searches)
	})
}

func (program *Program) setRuntimeStoryReviewConfig(config story.Config) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withStoryReviewConfig(config)
	})
}
