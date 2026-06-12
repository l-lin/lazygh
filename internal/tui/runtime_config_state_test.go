package tui

import (
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

func TestRuntimeConfigState_GivenKeymapsSearchesDisplayAndStoryReviewConfig_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := runtimeConfigState{
		keymapOverrides:     appconfig.KeymapOverrides{"global": {"open_search": {"/"}}},
		pullRequestSearches: []appconfig.PullRequestSearch{{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open"}}},
		displayConfig:       appconfig.DisplayConfig{RepositoryStyle: appconfig.RepositoryStyleOwnerName},
		storyReviewConfig:   story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}, Prompt: "Original prompt"},
	}
	given_overrides := appconfig.KeymapOverrides{"global": {"open_search": {"s"}}}
	given_searches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{" search ", " prs ", " --review-requested ", " @me "}}}
	given_displayConfig := appconfig.DisplayConfig{RepositoryStyle: " NAME "}
	given_storyConfig := story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "}

	keymapsUpdated := subject.withKeymapOverrides(given_overrides)
	searchesUpdated := keymapsUpdated.withPullRequestSearches(given_searches)
	displayUpdated := searchesUpdated.withDisplayConfig(given_displayConfig)
	storyUpdated := displayUpdated.withStoryReviewConfig(given_storyConfig)
	given_overrides["global"]["open_search"][0] = "x"
	given_searches[0].Label = "Broken"
	given_displayConfig.RepositoryStyle = "mutated"
	given_storyConfig.AgentCommand[0] = "mutated"

	if actual := keymapsUpdated.keymapOverrides["global"]["open_search"][0]; actual != "s" {
		t.Fatalf("expected copied keymap override %q, actual %q", "s", actual)
	}
	expectedSearches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{"search", "prs", "--review-requested", "@me"}}}
	if !reflect.DeepEqual(searchesUpdated.pullRequestSearches, expectedSearches) {
		t.Fatalf("expected normalized pull request searches %+v, actual %+v", expectedSearches, searchesUpdated.pullRequestSearches)
	}
	expectedDisplayConfig := appconfig.DisplayConfig{RepositoryStyle: appconfig.RepositoryStyleName}
	if !reflect.DeepEqual(displayUpdated.displayConfig, expectedDisplayConfig) {
		t.Fatalf("expected resolved display config %+v, actual %+v", expectedDisplayConfig, displayUpdated.displayConfig)
	}
	expectedStoryConfig := story.ResolveConfig(story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "})
	if !reflect.DeepEqual(storyUpdated.storyReviewConfig, expectedStoryConfig) {
		t.Fatalf("expected resolved story review config %+v, actual %+v", expectedStoryConfig, storyUpdated.storyReviewConfig)
	}
	if actual := subject.keymapOverrides["global"]["open_search"][0]; actual != "/" {
		t.Fatalf("expected the original keymap override %q, actual %q", "/", actual)
	}
	if !reflect.DeepEqual(subject.pullRequestSearches, []appconfig.PullRequestSearch{{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open"}}}) {
		t.Fatalf("expected the original pull request searches to stay intact, actual %+v", subject.pullRequestSearches)
	}
	if !reflect.DeepEqual(subject.displayConfig, appconfig.DisplayConfig{RepositoryStyle: appconfig.RepositoryStyleOwnerName}) {
		t.Fatalf("expected the original display config to stay intact, actual %+v", subject.displayConfig)
	}
	if !reflect.DeepEqual(subject.storyReviewConfig, story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}, Prompt: "Original prompt"}) {
		t.Fatalf("expected the original story review config to stay intact, actual %+v", subject.storyReviewConfig)
	}
}

func TestUpdate_GivenMsgKeymapOverridesApplied_WhenApplying_ThenItCopiesTheOverridesIntoRuntimeConfig(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	given_overrides := appconfig.KeymapOverrides{"global": {"open_search": {"s"}}}

	Update(subject, MsgKeymapOverridesApplied{Overrides: given_overrides})
	given_overrides["global"]["open_search"][0] = "x"

	if actual := subject.runtimeConfig.keymapOverrides["global"]["open_search"][0]; actual != "s" {
		t.Fatalf("expected copied keymap override %q, actual %q", "s", actual)
	}
}

func TestUpdate_GivenMsgDisplayConfigApplied_WhenApplying_ThenItStoresTheResolvedDisplayConfigAndRebuildsStoredRows(t *testing.T) {
	given_pullRequest := githubdomain.PullRequest{Title: "Ship notifications", Number: 42, Repository: githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"}, State: "OPEN"}
	given_notification := githubdomain.Notification{ID: "thread-pr", Repository: githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"}, Subject: githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypePullRequest, Title: "Ship notifications", URL: "https://api.github.com/repos/acme/widgets/pulls/42"}}
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{{Item: Item{Title: "stale pull request title"}, Summary: &given_pullRequest}})
	model.SetNotificationRows([]NotificationRow{{Item: Item{Title: "stale notification title"}, Notification: &given_notification}})
	subject := NewProgramWithModel(model)
	given_config := appconfig.DisplayConfig{}

	Update(subject, MsgDisplayConfigApplied{Config: given_config})
	given_config.RepositoryStyle = "name"

	expectedDisplayConfig := appconfig.DisplayConfig{RepositoryStyle: appconfig.RepositoryStyleOwnerName}
	if !reflect.DeepEqual(subject.runtimeConfig.displayConfig, expectedDisplayConfig) {
		t.Fatalf("expected resolved display config %+v, actual %+v", expectedDisplayConfig, subject.runtimeConfig.displayConfig)
	}
	actualPullRequestRows := subject.model.PullRequestRows(MyPullRequestsTab)
	expectedPullRequestRows := []PullRequestRow{pullRequestRow(given_pullRequest)}
	if !reflect.DeepEqual(actualPullRequestRows, expectedPullRequestRows) {
		t.Fatalf("expected rebuilt pull request rows %+v, actual %+v", expectedPullRequestRows, actualPullRequestRows)
	}
	actualNotificationRows := subject.model.NotificationRows()
	expectedNotificationRows := []NotificationRow{notificationRow(given_notification)}
	if !reflect.DeepEqual(actualNotificationRows, expectedNotificationRows) {
		t.Fatalf("expected rebuilt notification rows %+v, actual %+v", expectedNotificationRows, actualNotificationRows)
	}
}

func TestUpdate_GivenMsgStoryReviewConfigApplied_WhenApplying_ThenItStoresTheResolvedStoryReviewConfigAndClearsCachedStories(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.storyReviewCache["acme/widgets#42"] = storyReviewResult{story: reviewStoryData{Summary: "Cached story"}, pendingReviewID: "PRR_story"}
	given_config := story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "}

	Update(subject, MsgStoryReviewConfigApplied{Config: given_config})
	given_config.AgentCommand[0] = "mutated"

	expected := story.ResolveConfig(story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "})
	if !reflect.DeepEqual(subject.runtimeConfig.storyReviewConfig, expected) {
		t.Fatalf("expected resolved story review config %+v, actual %+v", expected, subject.runtimeConfig.storyReviewConfig)
	}
	if actual := len(subject.storyReviewCache); actual != 0 {
		t.Fatalf("expected the story review cache length %d, actual %d", 0, actual)
	}
}

func TestUpdate_GivenMsgPullRequestSearchesApplied_WhenApplying_ThenItCopiesTheSearchesIntoRuntimeConfigAndRefreshesTheTabs(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	given_searches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{" search ", " prs ", " --review-requested ", " @me "}}}

	Update(subject, MsgPullRequestSearchesApplied{Searches: given_searches})
	given_searches[0].Label = "Broken"

	expectedSearches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{"search", "prs", "--review-requested", "@me"}}}
	if !reflect.DeepEqual(subject.runtimeConfig.pullRequestSearches, expectedSearches) {
		t.Fatalf("expected copied pull request searches %+v, actual %+v", expectedSearches, subject.runtimeConfig.pullRequestSearches)
	}
	if actual := subject.model.PullRequestTabLabel(MyPullRequestsTab); actual != "Team" {
		t.Fatalf("expected the first pull request tab label %q, actual %q", "Team", actual)
	}
}
