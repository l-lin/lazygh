package tui

import (
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/story"
)

func (state runtimeConfigState) withKeymapOverrides(overrides appconfig.KeymapOverrides) runtimeConfigState {
	state.keymapOverrides = copyKeymapOverrides(overrides)
	return state
}

func (state runtimeConfigState) withPullRequestConfig(config appconfig.PullRequestConfig) runtimeConfigState {
	config.Searches = normalizedRuntimePullRequestSearches(config.Searches)
	state.pullRequestConfig = config
	return state
}

func normalizedRuntimePullRequestSearches(searches []appconfig.PullRequestSearch) []appconfig.PullRequestSearch {
	resolved := appconfig.ResolvePullRequestSearches(searches)
	for index := range resolved {
		resolved[index].Label = strings.TrimSpace(resolved[index].Label)
		resolved[index].Command = normalizeRuntimeCommand(resolved[index].Command)
	}
	return resolved
}

func normalizeRuntimeCommand(command []string) []string {
	if len(command) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(command))
	for _, argument := range command {
		argument = strings.TrimSpace(argument)
		if argument == "" {
			return nil
		}
		normalized = append(normalized, argument)
	}
	return normalized
}

func (state runtimeConfigState) withDisplayConfig(config appconfig.DisplayConfig) runtimeConfigState {
	state.displayConfig = appconfig.ResolveDisplayConfig(config)
	return state
}

func (state runtimeConfigState) withStoryReviewConfig(config story.Config) runtimeConfigState {
	state.storyReviewConfig = story.ResolveConfig(config)
	return state
}
