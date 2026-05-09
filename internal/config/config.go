package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/story"
	"codeberg.org/l-lin/lazygh/internal/theme"
	"github.com/BurntSushi/toml"
)

const (
	configDirectoryName = "lazygh"
	configFileName      = "config.toml"
	cacheFileName       = "cache.sqlite3"
)

type Config struct {
	Keymaps      KeymapOverrides
	PullRequests []PullRequestSearch
	ThemePreset  string
	Theme        theme.Palette
	Links        LinksConfig
	StoryReview  story.Config
	Cache        CacheConfig
}

type CacheConfig struct {
	Path string
}

type LinksConfig struct {
	OpenCommand []string
}

type KeymapOverrides map[string]map[string][]string

type PullRequestSearch struct {
	Label   string
	Command []string
}

type rawConfig struct {
	Keymaps      map[string]map[string]any `toml:"keymaps"`
	PullRequests rawPullRequestConfig      `toml:"pull_requests"`
	Theme        rawThemeConfig            `toml:"theme"`
	Links        rawLinksConfig            `toml:"links"`
	StoryReview  rawStoryReviewConfig      `toml:"story_review"`
	Cache        rawCacheConfig            `toml:"cache"`
}

type rawThemeConfig struct {
	Preset string `toml:"preset"`
	theme.Palette
}

type rawPullRequestConfig struct {
	Searches []rawPullRequestSearch `toml:"searches"`
}

type rawPullRequestSearch struct {
	Label   string `toml:"label"`
	Command any    `toml:"command"`
}

type rawStoryReviewConfig struct {
	AgentCommand any    `toml:"agent_command"`
	Prompt       string `toml:"prompt"`
}

type rawCacheConfig struct {
	Path any `toml:"path"`
}

type rawLinksConfig struct {
	OpenCommand any `toml:"open_command"`
}

func DefaultPath(homeDirectory string, xdgConfigHome string) string {
	trimmedHomeDirectory := strings.TrimSpace(homeDirectory)
	trimmedXDGConfigHome := strings.TrimSpace(xdgConfigHome)
	if trimmedXDGConfigHome != "" {
		return filepath.Join(trimmedXDGConfigHome, configDirectoryName, configFileName)
	}

	return filepath.Join(trimmedHomeDirectory, ".config", configDirectoryName, configFileName)
}

func LoadDefault() (Config, error) {
	homeDirectory, actualErr := os.UserHomeDir()
	if actualErr != nil {
		return Config{}, actualErr
	}

	return Load(DefaultPath(homeDirectory, os.Getenv("XDG_CONFIG_HOME")))
}

func Load(configPath string) (Config, error) {
	_, actualErr := os.Stat(configPath)
	if errors.Is(actualErr, os.ErrNotExist) {
		return Config{}, nil
	}
	if actualErr != nil {
		return Config{}, actualErr
	}

	var raw rawConfig
	if _, actualErr = toml.DecodeFile(configPath, &raw); actualErr != nil {
		return Config{}, actualErr
	}

	return Config{
		Keymaps:      normalizeKeymapOverrides(raw.Keymaps),
		PullRequests: normalizePullRequestSearches(raw.PullRequests.Searches),
		ThemePreset:  theme.NormalizePresetName(raw.Theme.Preset),
		Theme:        theme.NormalizePalette(raw.Theme.Palette),
		Links:        normalizeLinksConfig(raw.Links),
		StoryReview:  normalizeStoryReviewConfig(raw.StoryReview),
		Cache:        normalizeCacheConfig(raw.Cache),
	}, nil
}

func DefaultPullRequestSearches() []PullRequestSearch {
	return []PullRequestSearch{
		{
			Label:   "My PRs",
			Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"},
		},
		{
			Label:   "My reviews",
			Command: []string{"search", "prs", "--reviewed-by", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"},
		},
		{
			Label:   "Requested",
			Command: []string{"search", "prs", "--review-requested", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"},
		},
	}
}

func (config Config) ResolvedPullRequestSearches() []PullRequestSearch {
	return ResolvePullRequestSearches(config.PullRequests)
}

func (config Config) ResolvedTheme() theme.Palette {
	return theme.ResolvePaletteWithPreset(config.ThemePreset, config.Theme)
}

func (config Config) ResolvedLinks() LinksConfig {
	return ResolveLinksConfig(config.Links)
}

func (config Config) ResolvedStoryReview() story.Config {
	return story.ResolveConfig(config.StoryReview)
}

func (config Config) ResolvedCache() (CacheConfig, error) {
	return ResolveCacheConfig(config.Cache)
}

func DefaultCachePath(homeDirectory string, xdgDataHome string) string {
	trimmedHomeDirectory := strings.TrimSpace(homeDirectory)
	trimmedXDGDataHome := strings.TrimSpace(xdgDataHome)
	if trimmedXDGDataHome != "" {
		return filepath.Join(trimmedXDGDataHome, configDirectoryName, cacheFileName)
	}

	return filepath.Join(trimmedHomeDirectory, ".local", "share", configDirectoryName, cacheFileName)
}

func ResolveCacheConfig(config CacheConfig) (CacheConfig, error) {
	configuredPath := strings.TrimSpace(config.Path)
	if configuredPath != "" {
		return CacheConfig{Path: configuredPath}, nil
	}

	homeDirectory, actualErr := os.UserHomeDir()
	if actualErr != nil {
		return CacheConfig{}, actualErr
	}

	return CacheConfig{Path: DefaultCachePath(homeDirectory, os.Getenv("XDG_DATA_HOME"))}, nil
}

func ResolveLinksConfig(config LinksConfig) LinksConfig {
	return resolveLinksConfigForGOOS(config, runtime.GOOS)
}

func resolveLinksConfigForGOOS(config LinksConfig, goos string) LinksConfig {
	configuredCommand := normalizeCommandArguments(config.OpenCommand)
	if len(configuredCommand) > 0 {
		return LinksConfig{OpenCommand: configuredCommand}
	}

	return LinksConfig{OpenCommand: defaultLinksOpenCommand(goos)}
}

func defaultLinksOpenCommand(goos string) []string {
	switch strings.TrimSpace(strings.ToLower(goos)) {
	case "darwin":
		return []string{"open"}
	case "linux":
		return []string{"xdg-open"}
	default:
		return nil
	}
}

func ResolvePullRequestSearches(searches []PullRequestSearch) []PullRequestSearch {
	normalized := normalizeResolvedPullRequestSearches(searches)
	if len(normalized) > 0 {
		return normalized
	}

	return normalizeResolvedPullRequestSearches(DefaultPullRequestSearches())
}

func normalizeResolvedPullRequestSearches(searches []PullRequestSearch) []PullRequestSearch {
	if len(searches) == 0 {
		return nil
	}

	normalized := make([]PullRequestSearch, 0, len(searches))
	for _, search := range searches {
		label := strings.TrimSpace(search.Label)
		command := normalizeCommandArguments(search.Command)
		if label == "" || len(command) == 0 {
			continue
		}
		normalized = append(normalized, PullRequestSearch{Label: label, Command: command})
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func FormatGHCommand(arguments []string) string {
	normalizedArguments := normalizeCommandArguments(arguments)
	if len(normalizedArguments) == 0 {
		return "gh"
	}

	return "gh " + strings.Join(normalizedArguments, " ")
}

func normalizePullRequestSearches(rawSearches []rawPullRequestSearch) []PullRequestSearch {
	if len(rawSearches) == 0 {
		return nil
	}

	normalized := make([]PullRequestSearch, 0, len(rawSearches))
	for _, rawSearch := range rawSearches {
		label := strings.TrimSpace(rawSearch.Label)
		command := normalizeCommand(rawSearch.Command)
		if label == "" || len(command) == 0 {
			continue
		}
		normalized = append(normalized, PullRequestSearch{Label: label, Command: command})
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func normalizeCommand(rawValue any) []string {
	switch actual := rawValue.(type) {
	case string:
		return normalizeCommandArguments(strings.Fields(actual))
	case []string:
		return normalizeCommandArguments(actual)
	case []any:
		arguments := make([]string, 0, len(actual))
		for _, value := range actual {
			stringValue, ok := value.(string)
			if !ok {
				return nil
			}
			arguments = append(arguments, stringValue)
		}
		return normalizeCommandArguments(arguments)
	default:
		return nil
	}
}

func normalizeCommandArguments(arguments []string) []string {
	if len(arguments) == 0 {
		return nil
	}

	normalizedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		trimmedArgument := strings.TrimSpace(argument)
		if trimmedArgument == "" {
			return nil
		}
		normalizedArguments = append(normalizedArguments, trimmedArgument)
	}

	return normalizedArguments
}

func normalizeStoryReviewConfig(raw rawStoryReviewConfig) story.Config {
	return story.Config{
		AgentCommand: normalizeCommand(raw.AgentCommand),
		Prompt:       strings.TrimSpace(raw.Prompt),
	}
}

func normalizeLinksConfig(raw rawLinksConfig) LinksConfig {
	return LinksConfig{OpenCommand: normalizeCommand(raw.OpenCommand)}
}

func normalizeCacheConfig(raw rawCacheConfig) CacheConfig {
	return CacheConfig{Path: normalizeOptionalString(raw.Path)}
}

func normalizeOptionalString(rawValue any) string {
	stringValue, ok := rawValue.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(stringValue)
}

func normalizeKeymapOverrides(rawScopes map[string]map[string]any) KeymapOverrides {
	if len(rawScopes) == 0 {
		return nil
	}

	normalizedScopes := make(KeymapOverrides)
	for scopeName, rawActions := range rawScopes {
		normalizedActions := normalizeKeymapActions(rawActions)
		if len(normalizedActions) == 0 {
			continue
		}
		normalizedScopes[scopeName] = normalizedActions
	}

	if len(normalizedScopes) == 0 {
		return nil
	}

	return normalizedScopes
}

func normalizeKeymapActions(rawActions map[string]any) map[string][]string {
	normalizedActions := map[string][]string{}
	for actionName, rawValue := range rawActions {
		normalizedValue, ok := normalizeKeymapValue(rawValue)
		if !ok {
			continue
		}
		normalizedActions[actionName] = normalizedValue
	}

	if len(normalizedActions) == 0 {
		return nil
	}

	return normalizedActions
}

func normalizeKeymapValue(rawValue any) ([]string, bool) {
	switch actual := rawValue.(type) {
	case string:
		return normalizeKeymapStrings([]string{actual})
	case []string:
		return normalizeKeymapStrings(actual)
	case []any:
		values := make([]string, 0, len(actual))
		for _, value := range actual {
			stringValue, ok := value.(string)
			if !ok {
				return nil, false
			}
			values = append(values, stringValue)
		}
		return normalizeKeymapStrings(values)
	default:
		return nil, false
	}
}

func normalizeKeymapStrings(values []string) ([]string, bool) {
	if len(values) == 0 {
		return nil, false
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, false
		}
		normalized = append(normalized, value)
	}

	return normalized, true
}
