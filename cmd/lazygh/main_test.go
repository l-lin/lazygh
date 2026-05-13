package main

import (
	"errors"
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
	apptheme "github.com/l-lin/lazygh/internal/theme"
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

func TestNewAppDepsWithRunner_GivenRunnerResponses_WhenCallingSelectedCapabilityPorts_ThenItReturnsProviderNeutralValuesFromFocusedAdapters(t *testing.T) {
	runner := &fakeCommandRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"login":"octocat","name":"Octo Cat","html_url":"https://github.com/octocat"}`)},
		{stdout: []byte(`[{"title":"Ship notifications","number":42,"repository":{"name":"widgets","nameWithOwner":"acme/widgets"},"url":"https://github.com/acme/widgets/pull/42","body":"Ship it","state":"OPEN","updatedAt":"2026-05-05T10:00:00Z"}]`)},
		{stdout: []byte(`[[{"id":"1001","unread":true,"reason":"review_requested","updated_at":"2026-05-08T16:53:11Z","repository":{"name":"widgets","nameWithOwner":"acme/widgets"},"subject":{"title":"Ship notifications","type":"PullRequest","url":"https://api.github.com/repos/acme/widgets/pulls/42"}}]]`)},
	}}

	deps := newAppDepsWithRunner(runner)

	if deps.SessionQueries == nil || deps.PullRequestList == nil || deps.NotificationQueries == nil || deps.DetailQueries == nil || deps.PullRequestMutations == nil || deps.ReviewMutations == nil || deps.NotificationMutations == nil || deps.ReactionMutations == nil || deps.BuildQueries == nil || deps.MarkdownHTMLRenderer == nil || deps.AuthTokenProvider == nil {
		t.Fatal("expected the composition root to wire every capability port")
	}

	actualUser, actualErr := deps.SessionQueries.GetConnectedUser()
	then_noError(t, actualErr)
	if actualUser.Login != "octocat" || actualUser.URL != "https://github.com/octocat" {
		t.Fatalf("expected connected user %+v, actual %+v", githubdomain.ConnectedUser{Login: "octocat", URL: "https://github.com/octocat"}, actualUser)
	}

	actualPullRequests, actualErr := deps.PullRequestList.ListPullRequests([]string{"search", "prs", "--author", "@me"})
	then_noError(t, actualErr)
	if len(actualPullRequests) != 1 || actualPullRequests[0].Repository.NameWithOwner != "acme/widgets" || actualPullRequests[0].Number != 42 {
		t.Fatalf("expected provider-neutral pull requests %+v, actual %+v", []githubdomain.PullRequestSummary{{Number: 42, Repository: githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"}}}, actualPullRequests)
	}

	actualNotifications, actualErr := deps.NotificationQueries.ListNotifications()
	then_noError(t, actualErr)
	if len(actualNotifications) != 1 || actualNotifications[0].Subject.Type != githubdomain.NotificationSubjectTypePullRequest {
		t.Fatalf("expected provider-neutral notifications %+v, actual %+v", []string{githubdomain.NotificationSubjectTypePullRequest}, actualNotifications)
	}
}

func TestNewAppDepsWithRunner_GivenNotificationMutation_WhenMarkingItRead_ThenItUsesTheFocusedNotificationAdapter(t *testing.T) {
	runner := &fakeCommandRunner{}
	deps := newAppDepsWithRunner(runner)

	actualErr := deps.NotificationMutations.MarkNotificationRead("1001")

	then_noError(t, actualErr)
	if !reflect.DeepEqual(runner.calls, []fakeCommandCall{{name: "gh", args: []string{"api", "/notifications/threads/1001", "--method", "PATCH"}}}) {
		t.Fatalf("expected notification mutation calls %+v, actual %+v", []fakeCommandCall{{name: "gh", args: []string{"api", "/notifications/threads/1001", "--method", "PATCH"}}}, runner.calls)
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

func TestRun_GivenViewSubcommand_WhenStartingTheProgram_ThenItOpensTheRequestedURLBeforeRunning(t *testing.T) {
	expectedURL := "https://github.com/acme/widgets/pull/99"
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"view", expectedURL},
		func() (appconfig.Config, error) {
			return appconfig.Config{}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if runner.viewURL != expectedURL {
		t.Fatalf("expected view url %q, actual %q", expectedURL, runner.viewURL)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "view", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "view", "run"}, runner.calls)
	}
}

func TestRun_GivenStoryReviewSubcommand_WhenStartingTheProgram_ThenItOpensTheRequestedURLBeforeRunning(t *testing.T) {
	expectedURL := "https://github.com/acme/widgets/pulls/123"
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"story-review", expectedURL},
		func() (appconfig.Config, error) {
			return appconfig.Config{}, nil
		},
		func() configurableRunner {
			return runner
		},
	)

	then_noError(t, actualErr)
	if runner.storyReviewURL != expectedURL {
		t.Fatalf("expected story review url %q, actual %q", expectedURL, runner.storyReviewURL)
	}
	if !reflect.DeepEqual(runner.calls, []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "story_review", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "apply_pull_request_searches", "apply_links", "apply_story_review", "apply_cache", "story_review", "run"}, runner.calls)
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

func TestRun_GivenViewSubcommandWithoutURL_WhenStartingTheProgram_ThenItReturnsAnArgumentError(t *testing.T) {
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"view"},
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
	if actualErr.Error() != "view expects exactly one pull request URL" {
		t.Fatalf("expected argument error %q, actual %q", "view expects exactly one pull request URL", actualErr.Error())
	}
	if runner.runCalled {
		t.Fatal("expected the runner not to be called")
	}
	if runner.viewCalled {
		t.Fatal("expected the view subcommand not to reach the runner")
	}
}

func TestRun_GivenStoryReviewSubcommandWithoutURL_WhenStartingTheProgram_ThenItReturnsAnArgumentError(t *testing.T) {
	runner := &fakeConfigurableRunner{}

	actualErr := run(
		[]string{"story-review"},
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
	if actualErr.Error() != "story-review expects exactly one pull request URL" {
		t.Fatalf("expected argument error %q, actual %q", "story-review expects exactly one pull request URL", actualErr.Error())
	}
	if runner.runCalled {
		t.Fatal("expected the runner not to be called")
	}
	if runner.storyReviewCalled {
		t.Fatal("expected the story-review subcommand not to reach the runner")
	}
}

type fakeConfigurableRunner struct {
	appliedOverrides           appconfig.KeymapOverrides
	appliedPullRequestSearches []appconfig.PullRequestSearch
	appliedLinksConfig         appconfig.LinksConfig
	appliedStoryReviewConfig   story.Config
	appliedCacheConfig         appconfig.CacheConfig
	reviewURL                  string
	viewURL                    string
	storyReviewURL             string
	runCalled                  bool
	reviewCalled               bool
	viewCalled                 bool
	storyReviewCalled          bool
	runErr                     error
	reviewErr                  error
	viewErr                    error
	storyReviewErr             error
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

func (runner *fakeConfigurableRunner) OpenPullRequestByURL(url string) error {
	runner.viewCalled = true
	runner.viewURL = url
	runner.calls = append(runner.calls, "view")
	return runner.viewErr
}

func (runner *fakeConfigurableRunner) OpenStoryReviewByURL(url string) error {
	runner.storyReviewCalled = true
	runner.storyReviewURL = url
	runner.calls = append(runner.calls, "story_review")
	return runner.storyReviewErr
}

func (runner *fakeConfigurableRunner) Run() error {
	runner.runCalled = true
	runner.calls = append(runner.calls, "run")
	return runner.runErr
}

type fakeCommandResponse struct {
	stdout []byte
	stderr []byte
	err    error
}

type fakeCommandCall struct {
	name string
	args []string
}

type fakeCommandRunner struct {
	responses []fakeCommandResponse
	calls     []fakeCommandCall
}

func (runner *fakeCommandRunner) Run(name string, args ...string) (githubcli.CommandResult, error) {
	runner.calls = append(runner.calls, fakeCommandCall{name: name, args: append([]string(nil), args...)})
	if len(runner.responses) == 0 {
		return githubcli.CommandResult{}, nil
	}
	response := runner.responses[0]
	runner.responses = runner.responses[1:]
	return githubcli.CommandResult{Stdout: response.stdout, Stderr: response.stderr}, response.err
}

func (runner *fakeCommandRunner) RunWithInput(name string, input []byte, args ...string) (githubcli.CommandResult, error) {
	return runner.Run(name, args...)
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
