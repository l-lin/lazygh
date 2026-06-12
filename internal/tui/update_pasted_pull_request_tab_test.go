package tui

import (
	"strings"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestSyncPastedPullRequestTab_GivenDefaultDisplayConfig_WhenHydratingRows_ThenItUsesFullRepositoryLabels(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestAdded(githubdomain.PullRequest{Title: "Widgets PR", Number: 13, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/13", State: "OPEN"})
	})

	actualTab, ok := subject.syncPastedPullRequestTab()
	if !ok {
		t.Fatal("expected the pasted pull request tab to exist")
	}

	actualRows := subject.model.PullRequestRows(actualTab)
	if len(actualRows) != 1 {
		t.Fatalf("expected 1 pasted pull request row, actual %+v", actualRows)
	}
	if !strings.Contains(actualRows[0].Item.Title, "acme/widgets#13") {
		t.Fatalf("expected pasted pull request title to contain %q, actual %q", "acme/widgets#13", actualRows[0].Item.Title)
	}
}

func TestSyncPastedPullRequestTab_GivenShortRepositoryStyle_WhenHydratingRows_ThenItUsesShortRepositoryLabels(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.ApplyDisplayConfig(appconfig.DisplayConfig{RepositoryStyle: appconfig.RepositoryStyleName})
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestAdded(githubdomain.PullRequest{Title: "Widgets PR", Number: 13, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/13", State: "OPEN"})
	})

	actualTab, ok := subject.syncPastedPullRequestTab()
	if !ok {
		t.Fatal("expected the pasted pull request tab to exist")
	}

	actualRows := subject.model.PullRequestRows(actualTab)
	if len(actualRows) != 1 {
		t.Fatalf("expected 1 pasted pull request row, actual %+v", actualRows)
	}
	if !strings.Contains(actualRows[0].Item.Title, "widgets#13") {
		t.Fatalf("expected pasted pull request title to contain %q, actual %q", "widgets#13", actualRows[0].Item.Title)
	}
}
