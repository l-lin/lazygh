package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/story"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestDefaultPath_GivenXDGConfigHome_WhenBuildingTheDefaultConfigPath_ThenItUsesTheXDGConfigDirectory(t *testing.T) {
	homeDirectory := filepath.Join(string(filepath.Separator), "tmp", "alice")
	xdgConfigHome := filepath.Join(string(filepath.Separator), "var", "config-root")

	actual := DefaultPath(homeDirectory, xdgConfigHome)

	expected := filepath.Join(xdgConfigHome, "lazygh", "config.toml")
	if actual != expected {
		t.Fatalf("expected default path %q, actual %q", expected, actual)
	}
}

func TestDefaultPath_GivenNoXDGConfigHome_WhenBuildingTheDefaultConfigPath_ThenItFallsBackToTheHomeConfigDirectory(t *testing.T) {
	homeDirectory := filepath.Join(string(filepath.Separator), "tmp", "alice")

	actual := DefaultPath(homeDirectory, "")

	expected := filepath.Join(homeDirectory, ".config", "lazygh", "config.toml")
	if actual != expected {
		t.Fatalf("expected default path %q, actual %q", expected, actual)
	}
}

func TestDefaultCachePath_GivenXDGDataHome_WhenBuildingTheDefaultCachePath_ThenItUsesTheXDGDataDirectory(t *testing.T) {
	homeDirectory := filepath.Join(string(filepath.Separator), "tmp", "alice")
	xdgDataHome := filepath.Join(string(filepath.Separator), "var", "data")

	actual := DefaultCachePath(homeDirectory, xdgDataHome)

	expected := filepath.Join(xdgDataHome, "lazygh", "cache.sqlite3")
	if actual != expected {
		t.Fatalf("expected cache path %q, actual %q", expected, actual)
	}
}

func TestDefaultCachePath_GivenNoXDGDataHome_WhenBuildingTheDefaultCachePath_ThenItFallsBackToTheHomeDataDirectory(t *testing.T) {
	homeDirectory := filepath.Join(string(filepath.Separator), "tmp", "alice")

	actual := DefaultCachePath(homeDirectory, "")

	expected := filepath.Join(homeDirectory, ".local", "share", "lazygh", "cache.sqlite3")
	if actual != expected {
		t.Fatalf("expected cache path %q, actual %q", expected, actual)
	}
}

func TestLoad_GivenMissingConfigFile_WhenLoading_ThenItReturnsAnEmptyConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	if len(actual.Keymaps) != 0 {
		t.Fatalf("expected no keymaps, actual %+v", actual.Keymaps)
	}
}

func TestLoad_GivenScopedKeymapOverrides_WhenLoading_ThenItPreservesTheConfiguredValues(t *testing.T) {
	configPath := given_configFile(t, `
[keymaps.global]
quit = "ctrl+x"

[keymaps.pull_requests]
open_detail = ["o", "enter"]
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Keymaps: KeymapOverrides{
		"global": {
			"quit": {"ctrl+x"},
		},
		"pull_requests": {
			"open_detail": {"o", "enter"},
		},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenInvalidKeymapEntryTypes_WhenLoading_ThenItIgnoresOnlyTheBadEntries(t *testing.T) {
	configPath := given_configFile(t, `
[keymaps.pull_requests]
open_detail = 1
copy_pull_request_url = ["y", 2]
comment_on_pull_request = "c"
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Keymaps: KeymapOverrides{
		"pull_requests": {
			"comment_on_pull_request": {"c"},
		},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenPullRequestSearches_WhenLoading_ThenItPreservesConfiguredLabelsAndCommands(t *testing.T) {
	configPath := given_configFile(t, `
[[pull_requests.searches]]
label = "Mine"
command = ["pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"]

[[pull_requests.searches]]
label = "Team Review"
command = "search prs --review-requested @me --limit 50 --state open --json title,number,repository,url,body,state,isDraft,updatedAt"
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{PullRequests: []PullRequestSearch{
		{Label: "Mine", Command: []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
		{Label: "Team Review", Command: []string{"search", "prs", "--review-requested", "@me", "--limit", "50", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenThemeOverrides_WhenLoading_ThenItPreservesOnlyTheValidConfiguredColors(t *testing.T) {
	configPath := given_configFile(t, `
[theme]
active_border = " #7E9CD8 "
inactive_border = "#54546D"
selected_line_background = "wrong"
markdown_heading = "#1F1F28"
pull_request_reference = "#656D76"
pull_request_title = "#111827"
success = "#7FB069"
success_background = "#D7E8D0"
comment_author_badge = "#4D699B"
pending_background = "broken"
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Theme: theme.Palette{
		ActiveBorderHex:         "#7E9CD8",
		InactiveBorderHex:       "#54546D",
		MarkdownHeadingHex:      "#1F1F28",
		PullRequestReferenceHex: "#656D76",
		PullRequestTitleHex:     "#111827",
		SuccessHex:              "#7FB069",
		SuccessBackgroundHex:    "#D7E8D0",
		CommentAuthorBadgeHex:   "#4D699B",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenThemePresetAndOverrides_WhenLoading_ThenItPreservesThePresetAndValidConfiguredColors(t *testing.T) {
	configPath := given_configFile(t, `
[theme]
preset = " Kanagawa-Dark "
active_border = " #7E9CD8 "
background = "broken"
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{ThemePreset: "kanagawa-dark", Theme: theme.Palette{ActiveBorderHex: "#7E9CD8"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenAnInvalidThemePreset_WhenLoading_ThenItIgnoresIt(t *testing.T) {
	configPath := given_configFile(t, `
[theme]
preset = "solarized"
active_border = "#7E9CD8"
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Theme: theme.Palette{ActiveBorderHex: "#7E9CD8"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenStoryReviewSettings_WhenLoading_ThenItPreservesTheConfiguredAgentCommandAndPrompt(t *testing.T) {
	configPath := given_configFile(t, `
[story_review]
agent_command = ["pi", "-p", "@{{prompt_file}}"]
prompt = "Tell the story with dry professionalism."
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{StoryReview: story.Config{
		AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"},
		Prompt:       "Tell the story with dry professionalism.",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenLinksOpenCommand_WhenLoading_ThenItPreservesTheConfiguredCommand(t *testing.T) {
	configPath := given_configFile(t, `
[links]
open_command = ["open", "-a", "Safari"]
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Links: LinksConfig{OpenCommand: []string{"open", "-a", "Safari"}}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenCacheSettings_WhenLoading_ThenItPreservesTheConfiguredCachePath(t *testing.T) {
	configPath := given_configFile(t, `
[cache]
path = " /tmp/lazygh/prs.sqlite3 "
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{Cache: CacheConfig{Path: "/tmp/lazygh/prs.sqlite3"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenAnInvalidCachePathType_WhenLoading_ThenItIgnoresTheBadValue(t *testing.T) {
	configPath := given_configFile(t, `
[cache]
path = 7
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	if actual.Cache.Path != "" {
		t.Fatalf("expected an empty cache path, actual %q", actual.Cache.Path)
	}
}

func TestLoad_GivenInvalidPullRequestSearchEntries_WhenLoading_ThenItIgnoresOnlyTheBadEntries(t *testing.T) {
	configPath := given_configFile(t, `
[[pull_requests.searches]]
label = "   "
command = ["search", "prs", "--author", "@me"]

[[pull_requests.searches]]
label = "Valid"
command = ["search", "prs", "--review-requested", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"]

[[pull_requests.searches]]
label = "Broken"
command = 1
`)

	actual, actualErr := when_loading(configPath)

	then_noError(t, actualErr)
	expected := Config{PullRequests: []PullRequestSearch{{
		Label:   "Valid",
		Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"},
	}}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestDefaultPullRequestSearches_WhenReadingDefaults_ThenTheySortByLastUpdatedDescendingWithoutJSONFlags(t *testing.T) {
	actual := DefaultPullRequestSearches()

	expected := []PullRequestSearch{
		{Label: "My PRs", Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "My reviews", Command: []string{"search", "prs", "--reviewed-by", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Requested", Command: []string{"search", "prs", "--review-requested", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected searches %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedPullRequestSearches_GivenAnEmptyConfiguredList_WhenResolving_ThenItFallsBackToDefaults(t *testing.T) {
	subject := Config{}

	actual := subject.ResolvedPullRequestSearches()

	expected := DefaultPullRequestSearches()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected searches %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedTheme_GivenNoConfiguredTheme_WhenResolving_ThenItFallsBackToTheDefaultPalette(t *testing.T) {
	subject := Config{}

	actual := subject.ResolvedTheme()

	expected := theme.DefaultPalette()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected palette %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedTheme_GivenThemePreset_WhenResolving_ThenItUsesThatPresetBase(t *testing.T) {
	subject := Config{ThemePreset: "kanagawa-dark"}

	actual := subject.ResolvedTheme()

	if actual.BackgroundHex != "#1F1F28" {
		t.Fatalf("expected background color %q, actual %q", "#1F1F28", actual.BackgroundHex)
	}
	if actual.ActiveTextHex != "#DCD7BA" {
		t.Fatalf("expected active text color %q, actual %q", "#DCD7BA", actual.ActiveTextHex)
	}
}

func TestResolveLinksConfig_GivenNoConfiguredOpenCommandOnDarwin_WhenResolving_ThenItUsesOpen(t *testing.T) {
	actual := resolveLinksConfigForGOOS(LinksConfig{}, "darwin")

	expected := LinksConfig{OpenCommand: []string{"open"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected links config %+v, actual %+v", expected, actual)
	}
}

func TestResolveLinksConfig_GivenNoConfiguredOpenCommandOnLinux_WhenResolving_ThenItUsesXDGOpen(t *testing.T) {
	actual := resolveLinksConfigForGOOS(LinksConfig{}, "linux")

	expected := LinksConfig{OpenCommand: []string{"xdg-open"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected links config %+v, actual %+v", expected, actual)
	}
}

func TestResolveLinksConfig_GivenAConfiguredOpenCommand_WhenResolving_ThenItKeepsTheConfiguredCommand(t *testing.T) {
	actual := resolveLinksConfigForGOOS(LinksConfig{OpenCommand: []string{"open", "-a", "Firefox"}}, "linux")

	expected := LinksConfig{OpenCommand: []string{"open", "-a", "Firefox"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected links config %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedStoryReview_GivenNoConfiguredPrompt_WhenResolving_ThenItFallsBackToTheDefaultPrompt(t *testing.T) {
	subject := Config{}

	actual := subject.ResolvedStoryReview()

	if len(actual.AgentCommand) != 0 {
		t.Fatalf("expected no agent command, actual %v", actual.AgentCommand)
	}
	if actual.Prompt != story.DefaultPrompt() {
		t.Fatalf("expected default prompt %q, actual %q", story.DefaultPrompt(), actual.Prompt)
	}
}

func TestConfig_ResolvedCache_GivenNoConfiguredPath_WhenResolving_ThenItUsesTheXDGDataDirectory(t *testing.T) {
	t.Setenv("HOME", filepath.Join(string(filepath.Separator), "tmp", "alice"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(string(filepath.Separator), "var", "cache-root"))
	subject := Config{}

	actual, actualErr := subject.ResolvedCache()

	then_noError(t, actualErr)
	expected := CacheConfig{Path: filepath.Join(string(filepath.Separator), "var", "cache-root", "lazygh", "cache.sqlite3")}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cache config %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedCache_GivenNoConfiguredPathAndNoXDGDataHome_WhenResolving_ThenItFallsBackToTheHomeDataDirectory(t *testing.T) {
	t.Setenv("HOME", filepath.Join(string(filepath.Separator), "tmp", "alice"))
	t.Setenv("XDG_DATA_HOME", "")
	subject := Config{}

	actual, actualErr := subject.ResolvedCache()

	then_noError(t, actualErr)
	expected := CacheConfig{Path: filepath.Join(string(filepath.Separator), "tmp", "alice", ".local", "share", "lazygh", "cache.sqlite3")}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cache config %+v, actual %+v", expected, actual)
	}
}

func TestConfig_ResolvedCache_GivenAConfiguredPath_WhenResolving_ThenItKeepsTheConfiguredLocation(t *testing.T) {
	subject := Config{Cache: CacheConfig{Path: "/tmp/lazygh/custom.sqlite3"}}

	actual, actualErr := subject.ResolvedCache()

	then_noError(t, actualErr)
	expected := CacheConfig{Path: "/tmp/lazygh/custom.sqlite3"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cache config %+v, actual %+v", expected, actual)
	}
}

func TestLoadDefault_GivenXDGConfigHome_WhenLoading_ThenItUsesTheXDGConfigLocation(t *testing.T) {
	homeDirectory := t.TempDir()
	xdgConfigHome := filepath.Join(t.TempDir(), "config-root")
	configPath := filepath.Join(xdgConfigHome, "lazygh", "config.toml")
	actualErr := os.MkdirAll(filepath.Dir(configPath), 0o755)
	then_noError(t, actualErr)
	actualErr = os.WriteFile(configPath, []byte("[theme]\npreset = \"dark\"\n"), 0o644)
	then_noError(t, actualErr)
	t.Setenv("HOME", homeDirectory)
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

	actual, actualErr := LoadDefault()

	then_noError(t, actualErr)
	expected := Config{ThemePreset: theme.DarkPresetName}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected config %+v, actual %+v", expected, actual)
	}
}

func TestLoad_GivenMalformedTOML_WhenLoading_ThenItReturnsTheParseError(t *testing.T) {
	configPath := given_configFile(t, "[keymaps.pull_requests\nopen_detail = \"o\"\n")

	_, actualErr := when_loading(configPath)

	if actualErr == nil {
		t.Fatal("expected a parse error")
	}
}

func given_configFile(t *testing.T, contents string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.toml")
	actualErr := os.WriteFile(configPath, []byte(contents), 0o644)
	then_noError(t, actualErr)
	return configPath
}

func when_loading(configPath string) (Config, error) {
	return Load(configPath)
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
