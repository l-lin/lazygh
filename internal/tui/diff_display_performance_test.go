package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	largeDiffDisplayFixtureLineCount        = 10000
	largeCachedDiffRefreshAllocationCeiling = 200000.0
	largeCachedDiffFixtureRepositoryName    = "acme/widgets"
	largeCachedDiffFixturePullRequestNumber = 42
	largeCachedDiffFixturePendingReviewID   = "PRR_diff_perf"
	largeCachedDiffFixtureFilePath          = "internal/tui/render.go"
)

func TestRefreshViews_GivenLargeCachedDiff_WhenRefreshingRepeatedly_ThenItStaysBelowTheAllocationCeiling(t *testing.T) {
	subject, gui := given_largeCachedDiffRefreshProgram(t, largeDiffDisplayFixtureLineCount)
	defer gui.Close()

	if actualErr := subject.refreshViews(gui); actualErr != nil {
		t.Fatalf("expected no warm refresh error, actual %v", actualErr)
	}

	actual := testing.AllocsPerRun(3, func() {
		if actualErr := subject.refreshViews(gui); actualErr != nil {
			t.Fatalf("expected no refresh error, actual %v", actualErr)
		}
	})
	if actual > largeCachedDiffRefreshAllocationCeiling {
		t.Fatalf("expected cached diff refresh allocations to stay below %.0f allocs/run, actual %.2f", largeCachedDiffRefreshAllocationCeiling, actual)
	}
}

func given_largeCachedDiffRefreshProgram(tb testing.TB, lineCount int) (*Program, *gocui.Gui) {
	tb.Helper()

	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache[largeCachedDiffFixturePullRequestKey()] = pullRequestDiffResult{
		data:                    buildReviewDiffData(given_largeCachedReviewSessionPullRequestDiff(lineCount)),
		fileTeamOwnersAttempted: true,
	}
	subject.startReviewSession(githubcli.PullRequest{
		Title:      "First PR",
		Number:     largeCachedDiffFixturePullRequestNumber,
		Repository: githubcli.Repository{NameWithOwner: largeCachedDiffFixtureRepositoryName},
	}, largeCachedDiffFixturePendingReviewID)
	subject.clampReviewSessionSelection()

	gui := given_benchmarkHeadlessGuiWithSize(tb, 120, 40)
	subject.configureGUI(gui)
	if actualErr := subject.layout(gui); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no layout error, actual %v", actualErr)
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no detail focus error, actual %v", actualErr)
	}

	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		gui.Close()
		tb.Fatalf("expected a detail view, actual %v", actualErr)
	}
	document := subject.currentDetailDocument(detailView)
	targetRow := minInt(document.rowCount()-1, maxInt(detailView.InnerHeight()+20, document.rowCount()/2))
	subject.detailState.viewState.cursor = document.positionForRow(targetRow, 0)
	subject.detailState.viewState.preferredColumn = 0
	subject.detailState.viewState.sync(document, detailView.InnerHeight())
	if actualErr := subject.refreshDetailView(gui); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no detail refresh error, actual %v", actualErr)
	}

	return subject, gui
}

func given_largeCachedReviewSessionPullRequestDiff(lineCount int) githubcli.PullRequestDiff {
	if lineCount < 1 {
		lineCount = 1
	}

	var unifiedDiff strings.Builder
	unifiedDiff.Grow(lineCount * 48)
	unifiedDiff.WriteString("diff --git a/")
	unifiedDiff.WriteString(largeCachedDiffFixtureFilePath)
	unifiedDiff.WriteString(" b/")
	unifiedDiff.WriteString(largeCachedDiffFixtureFilePath)
	unifiedDiff.WriteString("\n")
	unifiedDiff.WriteString("index 1111111..2222222 100644\n")
	unifiedDiff.WriteString("--- a/")
	unifiedDiff.WriteString(largeCachedDiffFixtureFilePath)
	unifiedDiff.WriteString("\n")
	unifiedDiff.WriteString("+++ b/")
	unifiedDiff.WriteString(largeCachedDiffFixtureFilePath)
	unifiedDiff.WriteString("\n")
	unifiedDiff.WriteString(fmt.Sprintf("@@ -1,0 +1,%d @@\n", lineCount))
	for lineNumber := range lineCount {
		unifiedDiff.WriteString(fmt.Sprintf("+func addedLine%05d() int { return %d }\n", lineNumber+1, lineNumber+1))
	}

	return githubcli.PullRequestDiff{
		UnifiedDiff: unifiedDiff.String(),
		Files: []githubcli.PullRequestDiffFile{{
			Path:       largeCachedDiffFixtureFilePath,
			ChangeType: "modified",
			Additions:  lineCount,
			Deletions:  0,
		}},
	}
}

func largeCachedDiffFixturePullRequestKey() string {
	return fmt.Sprintf("%s#%d", largeCachedDiffFixtureRepositoryName, largeCachedDiffFixturePullRequestNumber)
}
