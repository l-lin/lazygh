package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestViewZeroScroll_GivenPullRequestsFocus_WhenPressingShiftJAndShiftK_ThenItScrollsTheDetailViewportWithoutChangingSelection(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:  "First PR",
		Number: 42,
		Body:   strings.TrimSpace(strings.Repeat("detail line\n", 80)),
	})}
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	initialOriginRow := subject.detailState.viewState.originRow

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr = downHandler(gui, nil)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualOriginRow := subject.detailState.viewState.originRow
	if actualOriginRow != initialOriginRow+1 {
		t.Fatalf("expected detail origin row %d after Shift-j, actual %d", initialOriginRow+1, actualOriginRow)
	}
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 0 {
		t.Fatalf("expected selected pull request index %d, actual %d", 0, subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(gui, nil)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualOriginRow = subject.detailState.viewState.originRow
	if actualOriginRow != initialOriginRow {
		t.Fatalf("expected detail origin row %d after Shift-k, actual %d", initialOriginRow, actualOriginRow)
	}
}

func TestViewZeroScroll_GivenReviewMode_WhenPressingShiftJAndShiftK_ThenItScrollsTheDiffViewportWithoutChangingTheSelectedFile(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_largeReviewSessionPullRequestDiff()), fileTeamOwnersAttempted: true}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_scroll")
	subject.clampReviewSessionSelection()
	expectedSelectedFileTreeRow := subject.navigationState.reviewSession.selectedFileTreeRow
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	initialOriginRow := subject.detailState.viewState.originRow

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr = downHandler(gui, nil)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualOriginRow := subject.detailState.viewState.originRow
	if actualOriginRow != initialOriginRow+1 {
		t.Fatalf("expected detail origin row %d after Shift-j in review mode, actual %d", initialOriginRow+1, actualOriginRow)
	}
	if subject.navigationState.reviewSession.selectedFileTreeRow != expectedSelectedFileTreeRow {
		t.Fatalf("expected selected review file tree row %d, actual %d", expectedSelectedFileTreeRow, subject.navigationState.reviewSession.selectedFileTreeRow)
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(gui, nil)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualOriginRow = subject.detailState.viewState.originRow
	if actualOriginRow != initialOriginRow {
		t.Fatalf("expected detail origin row %d after Shift-k in review mode, actual %d", initialOriginRow, actualOriginRow)
	}
}

func given_largeReviewSessionPullRequestDiff() githubcli.PullRequestDiff {
	lines := []string{
		"diff --git a/internal/tui/render.go b/internal/tui/render.go",
		"index 1111111..2222222 100644",
		"--- a/internal/tui/render.go",
		"+++ b/internal/tui/render.go",
		"@@ -1,0 +1,40 @@",
	}
	for index := range 40 {
		lines = append(lines, fmt.Sprintf("+line %02d", index+1))
	}

	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join(lines, "\n"),
		Files:       []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 40}},
	}
}
