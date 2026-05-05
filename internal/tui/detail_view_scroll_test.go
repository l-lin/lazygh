package tui

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestViewZeroScroll_GivenPullRequestsFocus_WhenPressingShiftJAndShiftK_ThenItScrollsTheDetailViewportWithoutChangingSelection(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{
		Title:  "First PR",
		Number: 42,
		Body:   strings.TrimSpace(strings.Repeat("detail line\n", 80)),
	}}
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, initialOriginY := detailView.Origin()

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr = downHandler(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, actualOriginY := detailView.Origin()
	if actualOriginY != initialOriginY+1 {
		t.Fatalf("expected detail origin y %d after Shift-j, actual %d", initialOriginY+1, actualOriginY)
	}
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 0 {
		t.Fatalf("expected selected pull request index %d, actual %d", 0, subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, actualOriginY = detailView.Origin()
	if actualOriginY != initialOriginY {
		t.Fatalf("expected detail origin y %d after Shift-k, actual %d", initialOriginY, actualOriginY)
	}
}

func TestViewZeroScroll_GivenReviewMode_WhenPressingShiftJAndShiftK_ThenItScrollsTheDiffViewportWithoutChangingTheSelectedFile(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_largeReviewSessionPullRequestDiff())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_scroll")
	subject.clampReviewSessionSelection()
	expectedSelectedFileTreeRow := subject.reviewSession.selectedFileTreeRow
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, initialOriginY := detailView.Origin()

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr = downHandler(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, actualOriginY := detailView.Origin()
	if actualOriginY != initialOriginY+1 {
		t.Fatalf("expected detail origin y %d after Shift-j in review mode, actual %d", initialOriginY+1, actualOriginY)
	}
	if subject.reviewSession.selectedFileTreeRow != expectedSelectedFileTreeRow {
		t.Fatalf("expected selected review file tree row %d, actual %d", expectedSelectedFileTreeRow, subject.reviewSession.selectedFileTreeRow)
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	_, actualOriginY = detailView.Origin()
	if actualOriginY != initialOriginY {
		t.Fatalf("expected detail origin y %d after Shift-k in review mode, actual %d", initialOriginY, actualOriginY)
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
