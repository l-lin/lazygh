package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultPath_GivenHomeDirectory_WhenBuildingTheDefaultConfigPath_ThenItUsesTheLazyghConfigLocation(t *testing.T) {
	homeDirectory := filepath.Join(string(filepath.Separator), "tmp", "alice")

	actual := DefaultPath(homeDirectory)

	expected := filepath.Join(homeDirectory, ".config", "lazygh", "config.toml")
	if actual != expected {
		t.Fatalf("expected default path %q, actual %q", expected, actual)
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

func TestConfig_ResolvedPullRequestSearches_GivenAnEmptyConfiguredList_WhenResolving_ThenItFallsBackToDefaults(t *testing.T) {
	subject := Config{}

	actual := subject.ResolvedPullRequestSearches()

	expected := DefaultPullRequestSearches()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected searches %+v, actual %+v", expected, actual)
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
