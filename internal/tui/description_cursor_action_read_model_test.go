package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDescriptionCursorActionReadModel_GivenBuildOverviewEntryAtCursor_WhenResolvingTheBuildLink_ThenItUsesTheSnapshot(t *testing.T) {
	selection := given_detailCursorSelectionForTests("build line", 72)
	subject := descriptionCursorActionReadModel{
		actionsAvailable: true,
		contextKnown:     true,
		selection:        selection,
		summary:          githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		detail:           githubdomain.PullRequestDetail{Body: "Body 42"},
		overviewCursor: browserDetailSectionCursor{
			inBody:   true,
			bodyLine: 1,
			section: browserDetailSection{
				overviewBlockTitle: "Builds",
				overviewEntries:    []pullRequestOverviewEntry{{Label: "CI / test", Link: "https://github.com/acme/widgets/actions/runs/42"}},
			},
		},
		overviewCursorKnown: true,
	}

	actualEntry, ok := subject.buildActionEntryAtCursor()
	if !ok {
		t.Fatal("expected a build action entry at the cursor")
	}
	if actualEntry.Label != "CI / test" {
		t.Fatalf("expected build label %q, actual %q", "CI / test", actualEntry.Label)
	}

	actualLink, actualOK := subject.buildLinkAtCursor()
	if !actualOK {
		t.Fatal("expected a build link at the cursor")
	}
	if actualLink != "https://github.com/acme/widgets/actions/runs/42" {
		t.Fatalf("expected build link %q, actual %q", "https://github.com/acme/widgets/actions/runs/42", actualLink)
	}
}

func TestDescriptionCursorActionReadModel_GivenUnavailableDetailActions_WhenResolvingTheBuildActionEntry_ThenItHidesTheEntry(t *testing.T) {
	subject := descriptionCursorActionReadModel{
		actionsAvailable:    false,
		contextKnown:        true,
		overviewCursorKnown: true,
		overviewCursor: browserDetailSectionCursor{
			inBody:   true,
			bodyLine: 1,
			section: browserDetailSection{
				overviewBlockTitle: "Builds",
				overviewEntries:    []pullRequestOverviewEntry{{Label: "CI / test", Link: "https://github.com/acme/widgets/actions/runs/42"}},
			},
		},
	}

	_, ok := subject.buildActionEntryAtCursor()

	if ok {
		t.Fatal("expected build actions to stay hidden when detail actions are unavailable")
	}
}

func TestDescriptionCursorActionReadModel_GivenAReRequestableReviewerEntryAtCursor_WhenResolvingTheReviewerActionEntry_ThenItUsesTheSnapshot(t *testing.T) {
	subject := descriptionCursorActionReadModel{
		actionsAvailable:    true,
		contextKnown:        true,
		overviewCursorKnown: true,
		overviewCursor: browserDetailSectionCursor{
			inBody:   true,
			bodyLine: 1,
			section: browserDetailSection{
				overviewBlockTitle: "Reviewers",
				overviewEntries:    []pullRequestOverviewEntry{{Label: "@reviewer-one", ReviewerLogin: "reviewer-one", CanReRequestReview: true}},
			},
		},
	}

	actual, ok := subject.reviewerActionEntryAtCursor()

	if !ok {
		t.Fatal("expected a reviewer action entry at the cursor")
	}
	if actual.ReviewerLogin != "reviewer-one" {
		t.Fatalf("expected reviewer login %q, actual %q", "reviewer-one", actual.ReviewerLogin)
	}
}
