package main

import (
	"errors"
	"reflect"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/story"
	apptheme "codeberg.org/l-lin/lazygh/internal/theme"
)

func TestRun_GivenLoadedKeymapOverrides_WhenStartingTheProgram_ThenItAppliesThemBeforeRunning(t *testing.T) {
	expectedOverrides := appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"o"},
		},
	}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{Keymaps: expectedOverrides}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.appliedOverrides, expectedOverrides) {
		t.Fatalf("expected overrides %+v, actual %+v", expectedOverrides, runner.appliedOverrides)
	}
	if !runner.runCalled {
		t.Fatal("expected the runner to be called")
	}
}

func TestRun_GivenLoadedPullRequestSearches_WhenStartingTheProgram_ThenItAppliesThemBeforeRunning(t *testing.T) {
	expectedSearches := []appconfig.PullRequestSearch{
		{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Team Review", Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
	}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{PullRequests: expectedSearches}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.appliedPullRequestSearches, expectedSearches) {
		t.Fatalf("expected pull request searches %+v, actual %+v", expectedSearches, runner.appliedPullRequestSearches)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}, runner.calls)
	}
}

func TestRun_GivenLoadedLinksConfig_WhenStartingTheProgram_ThenItAppliesItBeforeRunning(t *testing.T) {
	expectedConfig := appconfig.LinksConfig{OpenCommand: []string{"open", "-a", "Firefox"}}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{Links: expectedConfig}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.appliedLinksConfig, expectedConfig) {
		t.Fatalf("expected links config %+v, actual %+v", expectedConfig, runner.appliedLinksConfig)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}, runner.calls)
	}
}

func TestRun_GivenLoadedThemeOverrides_WhenStartingTheProgram_ThenItAppliesThemBeforeConstructingTheRunner(t *testing.T) {
	t.Cleanup(apptheme.ResetPalette)

	expectedActiveBorderHex := "#7E9CD8"
	runner := &fakeConfigurableRunner{}
	actualThemeDuringRunnerConstruction := ""

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{Theme: apptheme.Palette{ActiveBorderHex: expectedActiveBorderHex}}, nil
		},
		func() configurableRunner {
			actualThemeDuringRunnerConstruction = apptheme.ActiveBorderHex
			return runner
		},
	)

	then_noError(t, actualErr)
	if actualThemeDuringRunnerConstruction != expectedActiveBorderHex {
		t.Fatalf("expected active border color %q during runner construction, actual %q", expectedActiveBorderHex, actualThemeDuringRunnerConstruction)
	}
	if !runner.runCalled {
		t.Fatal("expected the runner to be called")
	}
}

func TestRun_GivenLoadedStoryReviewConfig_WhenStartingTheProgram_ThenItAppliesItBeforeRunning(t *testing.T) {
	expectedConfig := story.Config{
		AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"},
		Prompt:       "Tell the story with dry professionalism.",
	}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{StoryReview: expectedConfig}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.appliedStoryReviewConfig, expectedConfig) {
		t.Fatalf("expected story review config %+v, actual %+v", expectedConfig, runner.appliedStoryReviewConfig)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}, runner.calls)
	}
}

func TestRun_GivenLoadedCacheConfig_WhenStartingTheProgram_ThenItAppliesItBeforeRunning(t *testing.T) {
	expectedConfig := appconfig.CacheConfig{Path: "/tmp/lazygh/prs.sqlite3"}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{Cache: expectedConfig}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.appliedCacheConfig, expectedConfig) {
		t.Fatalf("expected cache config %+v, actual %+v", expectedConfig, runner.appliedCacheConfig)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "run"}, runner.calls)
	}
}

func TestRun_GivenConfigLoadError_WhenStartingTheProgram_ThenItReturnsTheError(t *testing.T) {
	expectedErr := errors.New("boom")
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		nil,
		func() (appconfig.Config, error) {
			return appconfig.Config{}, expectedErr
		},
		func() configurableRunner {
			return runner
		},
	)

	if !errors.Is(actualErr, expectedErr) {
		t.Fatalf("expected error %v, actual %v", expectedErr, actualErr)
	}
	if runner.runCalled {
		t.Fatal("expected the runner not to be called")
	}
}

func TestRun_GivenReviewSubcommand_WhenStartingTheProgram_ThenItOpensTheRequestedURLBeforeRunning(t *testing.T) {
	expectedURL := "https://github.com/acme/widgets/pull/42"
	expectedOverrides := appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"o"},
		},
	}
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"review", expectedURL},
		func() (appconfig.Config, error) {
			return appconfig.Config{Keymaps: expectedOverrides}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if runner.reviewURL != expectedURL {
		t.Fatalf("expected review url %q, actual %q", expectedURL, runner.reviewURL)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "review", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "review", "run"}, runner.calls)
	}
}

func TestRun_GivenReviewSubcommandWithoutURL_WhenStartingTheProgram_ThenItReturnsAnArgumentError(t *testing.T) {
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"review"},
		func() (appconfig.Config, error) {
			return appconfig.Config{}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "review expects exactly one pull request URL" {
		t.Fatalf("expected argument error %q, actual %q", "review expects exactly one pull request URL", actualErr.Error())
	}
	if runner.runCalled {
		t.Fatal("expected the runner not to be called")
	}
	if runner.reviewCalled {
		t.Fatal("expected the review subcommand not to reach the runner")
	}
}

type fakeConfigurableRunner struct {
	appliedOverrides           appconfig.KeymapOverrides
	appliedPullRequestSearches []appconfig.PullRequestSearch
	appliedLinksConfig         appconfig.LinksConfig
	appliedStoryReviewConfig   story.Config
	appliedCacheConfig         appconfig.CacheConfig
	reviewURL                  string
	runCalled                  bool
	reviewCalled               bool
	runErr                     error
	reviewErr                  error
	applyCacheErr              error
	calls                      []string
}

func (runner *fakeConfigurableRunner) ApplyKeymapOverrides(overrides appconfig.KeymapOverrides) {
	runner.appliedOverrides = overrides
	runner.calls = append(runner.calls, "apply")
}

func (runner *fakeConfigurableRunner) ApplyPullRequestSearches(searches []appconfig.PullRequestSearch) {
	runner.appliedPullRequestSearches = append([]appconfig.PullRequestSearch(nil), searches...)
	runner.calls = append(runner.calls, "apply_pull_request_searches")
}

func (runner *fakeConfigurableRunner) ApplyLinksConfig(config appconfig.LinksConfig) {
	runner.appliedLinksConfig = config
	runner.calls = append(runner.calls, "apply_links")
}

func (runner *fakeConfigurableRunner) ApplyStoryReviewConfig(config story.Config) {
	runner.appliedStoryReviewConfig = config
	runner.calls = append(runner.calls, "apply_story_review")
}

func (runner *fakeConfigurableRunner) ApplyCacheConfig(config appconfig.CacheConfig) error {
	runner.appliedCacheConfig = config
	runner.calls = append(runner.calls, "apply_cache")
	return runner.applyCacheErr
}

func (runner *fakeConfigurableRunner) OpenReviewByURL(url string) error {
	runner.reviewCalled = true
	runner.reviewURL = url
	runner.calls = append(runner.calls, "review")
	return runner.reviewErr
}

func (runner *fakeConfigurableRunner) Run() error {
	runner.runCalled = true
	runner.calls = append(runner.calls, "run")
	return runner.runErr
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
