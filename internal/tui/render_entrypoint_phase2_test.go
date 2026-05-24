package tui

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestRenderEntryPoints_GivenRenderSourceFiles_WhenInspecting_ThenTheyDoNotCallMaybeLoadHelpers(t *testing.T) {
	for _, path := range []string{"render.go", "program_view_state.go"} {
		contents, actualErr := os.ReadFile(path)
		then_noError(t, actualErr)

		if strings.Contains(string(contents), "maybeLoad") {
			t.Fatalf("expected %q to stay free of maybeLoad helper calls, actual source:\n%s", path, string(contents))
		}
	}
}

func TestLayout_GivenStartedProgramWithLoadablePullRequestDetail_WhenRendering_ThenItDoesNotPlanTheDetailLoad(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{
		"acme/widgets#42": {Title: "Loaded PR", Number: 42, Body: "Hello from the detail loader"},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.appStarted = true
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	if len(loader.detailCalls) != 0 {
		t.Fatalf("expected layout to avoid planning detail loads, actual calls %v", loader.detailCalls)
	}
	if len(subject.pullRequestDetailLoadInFlight) != 0 {
		t.Fatalf("expected no in-flight detail loads after layout, actual %v", subject.pullRequestDetailLoadInFlight)
	}
}

func TestRefreshViews_GivenVisibleActionsPopup_WhenRefreshing_ThenItDoesNotMutateThePopupSearchState(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.appStarted = true
	subject.model.OpenActionsPopup(3)
	subject.model.UpdateActionsPopupSearch("stale", []int{99})
	expected := []int{99}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.refreshViews(gui)
	then_noError(t, actualErr)

	actual := subject.model.ActionsPopupFilteredActionIndexes()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected refreshViews to keep popup search indexes %v, actual %v", expected, actual)
	}
}
