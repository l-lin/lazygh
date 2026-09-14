package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/l-lin/lazygh/internal/story"
	"github.com/l-lin/lazygh/internal/theme"
)

const (
	configDirectoryName = "lazygh"
	configFileName      = "config.toml"
	cacheFileName       = "cache.sqlite3"
)

var (
	pullRequestSearchCommandPrefix     = []string{"search", "prs"}
	legacyPullRequestListCommandPrefix = []string{"pr", "list"}
)

type Config struct {
	Keymaps      KeymapOverrides
	PullRequests PullRequestConfig
	ThemePreset  string
	Theme        theme.Palette
	Display      DisplayConfig
	Links        LinksConfig
	StoryReview  story.Config
	Cache        CacheConfig
}

type PullRequestConfig struct {
	Searches  []PullRequestSearch
	Refresh   time.Duration
	PastedPRs PastedPullRequestConfig
}

type PastedPullRequestConfig struct {
	AutoRefresh bool

	autoRefreshConfigured bool
}

type CacheConfig struct {
	Path string
}

type DisplayConfig struct {
	RepositoryStyle   string
	NotificationsView bool
}

const (
	RepositoryStyleOwnerName = "owner_name"
	RepositoryStyleName      = "name"
)

type LinksConfig struct {
	OpenCommand []string
}

type KeymapOverrides map[string]map[string][]string

type PullRequestSearch struct {
	Label       string
	Command     []string
	AutoRefresh bool
}

type rawConfig struct {
	Keymaps      map[string]map[string]any `toml:"keymaps"`
	PullRequests rawPullRequestConfig      `toml:"pull_requests"`
	Theme        rawThemeConfig            `toml:"theme"`
	Display      rawDisplayConfig          `toml:"display"`
	Links        rawLinksConfig            `toml:"links"`
	StoryReview  rawStoryReviewConfig      `toml:"story_review"`
	Cache        rawCacheConfig            `toml:"cache"`
}

type rawThemeConfig struct {
	Preset string `toml:"preset"`
	theme.Palette
}

type rawDisplayConfig struct {
	RepositoryStyle   any `toml:"repository_style"`
	NotificationsView any `toml:"notifications_view"`
}

type rawPullRequestConfig struct {
	Searches  []rawPullRequestSearch      `toml:"searches"`
	Refresh   any                         `toml:"refresh"`
	PastedPRs *rawPastedPullRequestConfig `toml:"pasted_prs"`
}

type rawPastedPullRequestConfig struct {
	AutoRefresh any `toml:"auto_refresh"`
}

type rawPullRequestSearch struct {
	Label       string `toml:"label"`
	Flags       any    `toml:"flags"`
	Command     any    `toml:"command"`
	AutoRefresh any    `toml:"auto_refresh"`
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
		PullRequests: normalizePullRequestConfig(raw.PullRequests),
		ThemePreset:  theme.NormalizePresetName(raw.Theme.Preset),
		Theme:        theme.NormalizePalette(raw.Theme.Palette),
		Display:      normalizeDisplayConfig(raw.Display),
		Links:        normalizeLinksConfig(raw.Links),
		StoryReview:  normalizeStoryReviewConfig(raw.StoryReview),
		Cache:        normalizeCacheConfig(raw.Cache),
	}, nil
}

func DefaultPullRequestSearches() []PullRequestSearch {
	return []PullRequestSearch{
		{
			Label:       "My PRs",
			AutoRefresh: true,
			Command: append(
				pullRequestSearchCommandPrefix,
				[]string{"--author", "@me", "--sort", "updated", "--order", "desc", "--state", "open"}...,
			),
		},
		{
			Label:       "My reviews",
			AutoRefresh: true,
			Command: append(
				pullRequestSearchCommandPrefix,
				[]string{"--reviewed-by", "@me", "--limit", "100", "--sort", "updated", "--order", "desc", "--state", "open"}...,
			),
		},
		{
			Label:       "Requested",
			AutoRefresh: true,
			Command: append(
				pullRequestSearchCommandPrefix,
				[]string{"--review-requested", "@me", "--limit", "100", "--sort", "updated", "--order", "desc", "--state", "open"}...,
			),
		},
	}
}

func (config Config) ResolvedPullRequestConfig() PullRequestConfig {
	resolved := PullRequestConfig{
		Searches:  ResolvePullRequestSearches(config.PullRequests.Searches),
		Refresh:   normalizePullRequestRefreshDuration(config.PullRequests.Refresh),
		PastedPRs: config.PullRequests.PastedPRs,
	}
	if !config.PullRequests.PastedPRs.autoRefreshConfigured {
		resolved.PastedPRs.AutoRefresh = true
	}
	return resolved
}

func (config Config) ResolvedPullRequestSearches() []PullRequestSearch {
	return config.ResolvedPullRequestConfig().Searches
}

func (config Config) ResolvedTheme() theme.Palette {
	return theme.ResolvePaletteWithPreset(config.ThemePreset, config.Theme)
}

func (config Config) ResolvedDisplay() DisplayConfig {
	return ResolveDisplayConfig(config.Display)
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

func ResolveDisplayConfig(config DisplayConfig) DisplayConfig {
	style := normalizeRepositoryStyle(config.RepositoryStyle)
	if style == "" {
		style = RepositoryStyleOwnerName
	}
	return DisplayConfig{RepositoryStyle: style, NotificationsView: config.NotificationsView}
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

func DefaultPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Searches:  DefaultPullRequestSearches(),
		PastedPRs: PastedPullRequestConfig{AutoRefresh: true, autoRefreshConfigured: true},
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
		normalized = append(normalized, PullRequestSearch{
			Label:       label,
			Command:     command,
			AutoRefresh: search.AutoRefresh,
		})
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

func normalizePullRequestConfig(raw rawPullRequestConfig) PullRequestConfig {
	pasted := PastedPullRequestConfig{}
	if raw.PastedPRs != nil {
		pasted.AutoRefresh = normalizeAutoRefresh(raw.PastedPRs.AutoRefresh)
		pasted.autoRefreshConfigured = true
	}
	return PullRequestConfig{
		Searches:  normalizePullRequestSearches(raw.Searches),
		Refresh:   normalizePullRequestRefresh(raw.Refresh),
		PastedPRs: pasted,
	}
}

func normalizePullRequestSearches(rawSearches []rawPullRequestSearch) []PullRequestSearch {
	if len(rawSearches) == 0 {
		return nil
	}

	normalized := make([]PullRequestSearch, 0, len(rawSearches))
	for _, rawSearch := range rawSearches {
		label := strings.TrimSpace(rawSearch.Label)
		command := normalizePullRequestSearchCommand(rawSearch)
		if label == "" || len(command) == 0 {
			continue
		}
		normalized = append(normalized, PullRequestSearch{
			Label:       label,
			Command:     command,
			AutoRefresh: normalizeAutoRefresh(rawSearch.AutoRefresh),
		})
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func normalizeAutoRefresh(raw any) bool {
	value, ok := raw.(bool)
	if !ok {
		return true
	}
	return value
}

func normalizePullRequestRefresh(raw any) time.Duration {
	value, ok := raw.(string)
	if !ok {
		return 0
	}

	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "manual") {
		return 0
	}

	refresh, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return normalizePullRequestRefreshDuration(refresh)
}

func normalizePullRequestRefreshDuration(refresh time.Duration) time.Duration {
	if refresh <= 0 {
		return 0
	}
	return refresh
}

func normalizePullRequestSearchCommand(raw rawPullRequestSearch) []string {
	arguments := normalizePullRequestSearchArguments(raw.Flags)
	if len(arguments) == 0 {
		arguments = normalizePullRequestSearchArguments(raw.Command)
	}
	if len(arguments) == 0 {
		return nil
	}

	command := make([]string, 0, len(arguments)+len(pullRequestSearchCommandPrefix))
	command = append(command, pullRequestSearchCommandPrefix...)
	command = append(command, arguments...)
	return command
}

func normalizePullRequestSearchArguments(rawValue any) []string {
	arguments := normalizeCommand(rawValue)
	if len(arguments) == 0 {
		return nil
	}
	return stripPullRequestSearchCommandPrefix(arguments)
}

func stripPullRequestSearchCommandPrefix(arguments []string) []string {
	normalizedArguments := normalizeCommandArguments(arguments)
	if len(normalizedArguments) == 0 {
		return nil
	}
	if hasCommandPrefix(normalizedArguments, pullRequestSearchCommandPrefix) || hasCommandPrefix(normalizedArguments, legacyPullRequestListCommandPrefix) {
		return append([]string(nil), normalizedArguments[2:]...)
	}
	return normalizedArguments
}

func hasCommandPrefix(arguments []string, prefix []string) bool {
	if len(arguments) < len(prefix) || len(prefix) == 0 {
		return false
	}
	for index, part := range prefix {
		if !strings.EqualFold(strings.TrimSpace(arguments[index]), part) {
			return false
		}
	}
	return true
}

func normalizeCommand(rawValue any) []string {
	switch actual := rawValue.(type) {
	case string:
		return ParseCommandLine(actual)
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

func ParseCommandLine(command string) []string {
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return nil
	}

	arguments := make([]string, 0)
	currentArgument := strings.Builder{}
	argumentStarted := false
	quotedBy := rune(0)
	escaped := false
	flushArgument := func(force bool) {
		if !argumentStarted && !force {
			return
		}
		arguments = append(arguments, currentArgument.String())
		currentArgument.Reset()
		argumentStarted = false
	}

	for _, character := range trimmedCommand {
		if escaped {
			currentArgument.WriteRune(character)
			argumentStarted = true
			escaped = false
			continue
		}
		if quotedBy != 0 {
			switch {
			case character == '\\' && quotedBy == '"':
				escaped = true
			case character == quotedBy:
				quotedBy = 0
				argumentStarted = true
			default:
				currentArgument.WriteRune(character)
				argumentStarted = true
			}
			continue
		}

		switch {
		case unicode.IsSpace(character):
			if argumentStarted {
				flushArgument(false)
			}
		case character == '\\':
			escaped = true
			argumentStarted = true
		case character == '"' || character == '\'':
			quotedBy = character
			argumentStarted = true
		default:
			currentArgument.WriteRune(character)
			argumentStarted = true
		}
	}
	if escaped {
		currentArgument.WriteRune('\\')
	}
	if argumentStarted || currentArgument.Len() > 0 {
		flushArgument(true)
	}
	return normalizeCommandArguments(arguments)
}

func normalizeStoryReviewConfig(raw rawStoryReviewConfig) story.Config {
	return story.Config{
		AgentCommand: normalizeCommand(raw.AgentCommand),
		Prompt:       strings.TrimSpace(raw.Prompt),
	}
}

func normalizeDisplayConfig(raw rawDisplayConfig) DisplayConfig {
	notificationsView, _ := raw.NotificationsView.(bool)
	return DisplayConfig{
		RepositoryStyle:   normalizeRepositoryStyle(normalizeOptionalString(raw.RepositoryStyle)),
		NotificationsView: notificationsView,
	}
}

func normalizeRepositoryStyle(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case RepositoryStyleOwnerName:
		return RepositoryStyleOwnerName
	case RepositoryStyleName:
		return RepositoryStyleName
	default:
		return ""
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
