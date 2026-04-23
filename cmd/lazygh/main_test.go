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

type fakeConfigurableRunner struct {
	appliedOverrides appconfig.KeymapOverrides
	runCalled        bool
	runErr           error
}

func (runner *fakeConfigurableRunner) ApplyKeymapOverrides(overrides appconfig.KeymapOverrides) {
	runner.appliedOverrides = overrides
}

func (runner *fakeConfigurableRunner) Run() error {
	runner.runCalled = true
	return runner.runErr
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
