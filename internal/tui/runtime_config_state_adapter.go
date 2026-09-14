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

func (program *Program) setRuntimePullRequestConfig(config appconfig.PullRequestConfig) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withPullRequestConfig(config)
	})
}

func (program *Program) setRuntimePullRequestSearches(searches []appconfig.PullRequestSearch) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		config := state.pullRequestConfig
		config.Searches = append([]appconfig.PullRequestSearch(nil), searches...)
		return state.withPullRequestConfig(config)
	})
}

func (program *Program) setRuntimeDisplayConfig(config appconfig.DisplayConfig) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withDisplayConfig(config)
	})
}

func (program *Program) setRuntimeStoryReviewConfig(config story.Config) {
	program.updateRuntimeConfig(func(state runtimeConfigState) runtimeConfigState {
		return state.withStoryReviewConfig(config)
	})
}
