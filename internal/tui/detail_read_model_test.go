package tui

import (
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailReadModel_GivenChangesTabSnapshot_WhenBuildingTheDocument_ThenItWrapsTheDiffRows(t *testing.T) {
	longLine := strings.Repeat("detailreadmodelwrap-", 4)
	file := given_reviewDiffFileWithLongLine(longLine)
	renderedRows := buildReviewDiffRenderedRows(file, nil, 60)
	subject := detailReadModel{
		width:                          32,
		activeTab:                      ChangesDetailTab,
		wordWrapEnabled:                true,
		pullRequestSummary:             githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		pullRequestSummaryKnown:        true,
		pullRequestDetailResult:        pullRequestDetailResult{detail: githubdomain.PullRequestDetail{}},
		pullRequestDetailResultKnown:   true,
		pullRequestDiffResult:          pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{file}}},
		pullRequestDiffResultKnown:     true,
		pullRequestChangesRenderedRows: renderedRows,
		pullRequestChangesKnown:        true,
	}

	actual := subject.document()
	lineIndex, actualLine := given_detailDocumentLineContaining(t, actual, longLine)

	if actual := reviewDiffDocumentRowCountForLine(actual, lineIndex); actual < 2 {
		t.Fatalf("expected the long diff line to wrap across multiple rendered rows, actual %d for %q", actual, actualLine)
	}
}

func TestDetailReadModel_GivenNotificationPullRequestSelection_WhenDerivingIdentity_ThenItKeepsTheNotificationPullRequestNamespace(t *testing.T) {
	subject := detailReadModel{
		currentSideFocus:        FocusNotificationsView,
		activeTab:               CommentsDetailTab,
		pullRequestSummary:      githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		pullRequestSummaryKnown: true,
	}

	actual := subject.identity()

	expected := "notification-pr:acme/widgets#42:tab:1"
	if actual != expected {
		t.Fatalf("expected detail identity %q, actual %q", expected, actual)
	}
}
