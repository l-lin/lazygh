package tui

import (
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/story"
)

func TestRuntimeConfigState_GivenKeymapsSearchesAndStoryReviewConfig_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := runtimeConfigState{
		keymapOverrides:     appconfig.KeymapOverrides{"global": {"open_search": {"/"}}},
		pullRequestSearches: []appconfig.PullRequestSearch{{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open"}}},
		storyReviewConfig:   story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}, Prompt: "Original prompt"},
	}
	given_overrides := appconfig.KeymapOverrides{"global": {"open_search": {"s"}}}
	given_searches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{" search ", " prs ", " --review-requested ", " @me "}}}
	given_storyConfig := story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "}

	keymapsUpdated := subject.withKeymapOverrides(given_overrides)
	searchesUpdated := keymapsUpdated.withPullRequestSearches(given_searches)
	storyUpdated := searchesUpdated.withStoryReviewConfig(given_storyConfig)
	given_overrides["global"]["open_search"][0] = "x"
	given_searches[0].Label = "Broken"
	given_storyConfig.AgentCommand[0] = "mutated"

	if actual := keymapsUpdated.keymapOverrides["global"]["open_search"][0]; actual != "s" {
		t.Fatalf("expected copied keymap override %q, actual %q", "s", actual)
	}
	expectedSearches := []appconfig.PullRequestSearch{{Label: "Team", Command: []string{"search", "prs", "--review-requested", "@me"}}}
	if !reflect.DeepEqual(searchesUpdated.pullRequestSearches, expectedSearches) {
		t.Fatalf("expected normalized pull request searches %+v, actual %+v", expectedSearches, searchesUpdated.pullRequestSearches)
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

func TestUpdate_GivenMsgStoryReviewConfigApplied_WhenApplying_ThenItStoresTheResolvedStoryReviewConfig(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	given_config := story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "}

	Update(subject, MsgStoryReviewConfigApplied{Config: given_config})
	given_config.AgentCommand[0] = "mutated"

	expected := story.ResolveConfig(story.Config{AgentCommand: []string{" pi ", " -p ", " @{{prompt_file}} "}, Prompt: "  Custom prompt  "})
	if !reflect.DeepEqual(subject.runtimeConfig.storyReviewConfig, expected) {
		t.Fatalf("expected resolved story review config %+v, actual %+v", expected, subject.runtimeConfig.storyReviewConfig)
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
