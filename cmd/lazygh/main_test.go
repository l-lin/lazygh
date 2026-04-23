package main

import (
	"errors"
	"reflect"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
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
	if !reflect.DeepEqual(runner.calls, []string{"apply", "review", "run"}) {
		t.Fatalf("expected runner calls %v, actual %v", []string{"apply", "review", "run"}, runner.calls)
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
	appliedOverrides appconfig.KeymapOverrides
	reviewURL        string
	runCalled        bool
	reviewCalled     bool
	runErr           error
	reviewErr        error
	calls            []string
}

func (runner *fakeConfigurableRunner) ApplyKeymapOverrides(overrides appconfig.KeymapOverrides) {
	runner.appliedOverrides = overrides
	runner.calls = append(runner.calls, "apply")
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
