package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type descriptionCursorActionReadModel struct {
	actionsAvailable    bool
	contextKnown        bool
	selection           detailCursorSelection
	summary             githubdomain.PullRequest
	detail              githubdomain.PullRequestDetail
	overviewCursor      browserDetailSectionCursor
	overviewCursorKnown bool
}

func (model descriptionCursorActionReadModel) buildEntryAtCursor() (pullRequestOverviewEntry, bool) {
	if !model.contextKnown || !model.overviewCursorKnown || !model.overviewCursor.inBody || !strings.EqualFold(strings.TrimSpace(model.overviewCursor.section.overviewBlockTitle), "Builds") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(model.overviewCursor.section, model.overviewCursor.bodyLine)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}

func (model descriptionCursorActionReadModel) buildActionEntryAtCursor() (pullRequestOverviewEntry, bool) {
	if !model.actionsAvailable {
		return pullRequestOverviewEntry{}, false
	}
	return model.buildEntryAtCursor()
}

func (model descriptionCursorActionReadModel) reviewerActionEntryAtCursor() (pullRequestOverviewEntry, bool) {
	if !model.actionsAvailable || !model.contextKnown || !model.overviewCursorKnown || !model.overviewCursor.inBody || !strings.EqualFold(strings.TrimSpace(model.overviewCursor.section.overviewBlockTitle), "Reviewers") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(model.overviewCursor.section, model.overviewCursor.bodyLine)
	if !ok || !entry.CanReRequestReview || strings.TrimSpace(entry.ReviewerLogin) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}

func (model descriptionCursorActionReadModel) buildLinkAtCursor() (string, bool) {
	entry, ok := model.buildEntryAtCursor()
	if !ok {
		return "", false
	}
	actual := strings.TrimSpace(entry.Link)
	if actual == "" {
		return "", false
	}
	return actual, true
}
