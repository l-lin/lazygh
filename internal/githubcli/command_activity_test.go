package githubcli

import (
	"errors"
	"strings"
	"sync"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestExecutor_GivenActivityObserver_WhenExecutingCommandsWithAndWithoutInput_ThenReportsSafePairedTimingAndExitOutcome(t *testing.T) {
	for _, withInput := range []bool{false, true} {
		for _, fails := range []bool{false, true} {
			t.Run(strings.Join([]string{map[bool]string{false: "args", true: "stdin"}[withInput], map[bool]string{false: "success", true: "failure"}[fails]}, "/"), func(t *testing.T) {
				// GIVEN
				runner := &fakeRunner{stdout: []byte("result")}
				if fails {
					runner.err = errors.New("transport failure")
				}
				var actual []githubdomain.CommandActivity
				subject := NewExecutor(NewActivityObservingRunner(runner, func(activity githubdomain.CommandActivity) { actual = append(actual, activity) }), nil, nil)
				command := Command{Args: []string{"api", "graphql", "-f", "query=secret"}, DisplayArgs: []string{"api", "graphql"}}
				if withInput {
					command.Stdin = []byte("private body")
				}
				// WHEN
				_, actualErr := subject.Execute(command)
				// THEN
				expected := githubdomain.CommandSucceeded
				if fails {
					expected = githubdomain.CommandFailed
				}
				if len(actual) != 2 || actual[0].Outcome != githubdomain.CommandRunning || actual[1].Outcome != expected {
					t.Fatalf("expected running then %v, actual %+v", expected, actual)
				}
				if actual[0].ID == "" || actual[0].ID != actual[1].ID || actual[0].StartedAt != actual[1].StartedAt || actual[1].FinishedAt.Before(actual[1].StartedAt) {
					t.Fatalf("expected paired monotonic timings, actual %+v", actual)
				}
				if actual[0].Command != "gh api graphql" || actual[1].Command != actual[0].Command {
					t.Fatalf("expected display-safe commands, actual %+v", actual)
				}
				if (actualErr != nil) != fails {
					t.Fatalf("expected failure=%v, actual %v", fails, actualErr)
				}
			})
		}
	}
}

type given_activityFormatter struct{}

func (given_activityFormatter) Format(Command) string { return "gh redacted" }

func TestExecutor_GivenCustomDisplayFormatter_WhenObservingExecution_ThenUsesThatFormatter(t *testing.T) {
	// GIVEN
	var actual []githubdomain.CommandActivity
	subject := NewExecutor(NewActivityObservingRunner(&fakeRunner{}, func(activity githubdomain.CommandActivity) { actual = append(actual, activity) }), given_activityFormatter{}, nil)
	// WHEN
	_, _ = subject.Execute(rawCommand("api", "private"))
	// THEN
	expected := "gh redacted"
	if len(actual) != 2 || actual[0].Command != expected || actual[1].Command != expected {
		t.Fatalf("expected custom display text %q, actual %+v", expected, actual)
	}
}

func TestExecutor_GivenConcurrentCommands_WhenObserved_ThenEachExecutionHasAnIndependentIdentity(t *testing.T) {
	// GIVEN
	var mu sync.Mutex
	actual := map[string][]githubdomain.CommandActivity{}
	subject := NewExecutor(NewActivityObservingRunner(&fakeRunner{}, func(activity githubdomain.CommandActivity) {
		mu.Lock()
		defer mu.Unlock()
		actual[activity.ID] = append(actual[activity.ID], activity)
	}), nil, nil)
	// WHEN
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() { _, _ = subject.Execute(rawCommand("api", "user")) })
	}
	workers.Wait()
	// THEN
	expected := 20
	if len(actual) != expected {
		t.Fatalf("expected %d independent executions, actual %d", expected, len(actual))
	}
	for _, events := range actual {
		if len(events) != 2 || events[0].Outcome != githubdomain.CommandRunning || events[1].Outcome != githubdomain.CommandSucceeded {
			t.Fatalf("expected paired events, actual %+v", events)
		}
	}
}
