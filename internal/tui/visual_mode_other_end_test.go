package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestDetailViewState_GivenCharacterVisualSelection_WhenMovingToTheOtherSelectionEnd_ThenItSwapsTheCursorAndAnchorWithoutChangingTheSelection(t *testing.T) {
	document := newDetailDocument("abcdef", 6)
	subject := newDetailViewState()

	subject.moveRight(document, 2)
	subject.moveRight(document, 2)
	subject.enterVisualMode()
	subject.moveRight(document, 2)
	subject.moveRight(document, 2)

	then_detailSelectionIs(t, document, subject, detailPosition{line: 0, column: 2}, detailPosition{line: 0, column: 4})

	subject.moveToOtherSelectionEnd(document, 2)

	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 2})
	if actual := subject.visualAnchor; actual != (detailPosition{line: 0, column: 4}) {
		t.Fatalf("expected visual anchor %+v, actual %+v", detailPosition{line: 0, column: 4}, actual)
	}
	then_detailSelectionIs(t, document, subject, detailPosition{line: 0, column: 2}, detailPosition{line: 0, column: 4})
}

func TestDetailViewState_GivenLinewiseVisualSelection_WhenMovingToTheOtherSelectionEnd_ThenItSwapsTheCursorAndAnchorWithoutChangingTheSelection(t *testing.T) {
	document := newDetailDocument("abcd\nefgh\nijkl", 4)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 0, column: 2}
	subject.preferredColumn = 2

	subject.enterLineVisualMode(document)
	subject.moveDown(document, 2)

	then_detailSelectedRowsAre(t, document, subject, 0, 1)

	subject.moveToOtherSelectionEnd(document, 2)

	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 0})
	if actual := subject.visualAnchor; actual != (detailPosition{line: 1, column: 2}) {
		t.Fatalf("expected visual anchor %+v, actual %+v", detailPosition{line: 1, column: 2}, actual)
	}
	then_detailSelectedRowsAre(t, document, subject, 0, 1)
}

func TestDetailView_GivenBrowserDetailTabsAndVisualSelection_WhenPressingO_ThenItMovesTheCursorToTheOtherSelectionEnd(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Description.zeta": "Description.zeta",
		"Comment.zeta":     "Comment.zeta",
		"Commit.zeta":      "Commit.zeta",
	}}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:  "First PR",
		Number: 42,
		Body:   "Description.zeta",
		State:  "OPEN",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Comment.zeta",
			CreatedAt: "2026-05-19T10:00:00Z",
		}},
		Commits: []githubcli.PullRequestCommit{{
			OID:             "abcdef1234567890",
			MessageHeadline: "Commit headline",
			MessageBody:     "Commit.zeta",
			AuthoredDate:    "2026-05-19T10:00:00Z",
			Authors:         []githubcli.PullRequestCommitAuthor{{Login: "reviewer-one", Name: "Reviewer One"}},
		}},
	})}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(githubcli.PullRequestDiff{
		UnifiedDiff: "diff --git a/test.txt b/test.txt\nindex 0000000..1111111 100644\n--- a/test.txt\n+++ b/test.txt\n@@ -0,0 +1 @@\n+Change.zeta\n",
		Files:       []githubcli.PullRequestDiffFile{{Path: "test.txt", ChangeType: "added", Additions: 1}},
	}), fileTeamOwnersAttempted: true}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))

	testCases := []struct {
		name    string
		tab     DetailTab
		segment string
	}{
		{name: "description", tab: DescriptionDetailTab, segment: "Description.zeta"},
		{name: "comments", tab: CommentsDetailTab, segment: "Comment.zeta"},
		{name: "commits", tab: CommitsDetailTab, segment: "Commit.zeta"},
		{name: "changes", tab: ChangesDetailTab, segment: "Change.zeta"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject.detailState = subject.detailState.withVisualModeExited()
			subject.detailState.activeTab = testCase.tab
			then_noError(t, subject.refreshDetailView(gui))
			given_detailCursorOnSegment(t, gui, subject, testCase.segment)
			detailView, actualErr := gui.View(viewDetailName)
			then_noError(t, actualErr)

			then_noError(t, subject.enterDetailVisualMode(gui, detailView))
			selectionStart := subject.detailState.viewState.cursor
			then_noError(t, subject.moveDetailCursorRight(gui, detailView))
			then_noError(t, subject.moveDetailCursorRight(gui, detailView))
			then_noError(t, subject.moveDetailCursorRight(gui, detailView))
			then_noError(t, subject.moveDetailCursorRight(gui, detailView))
			selectionEnd := subject.detailState.viewState.cursor

			registeredBindings := subject.registeredKeybindingSpecs()
			then_noError(t, given_handlerForBinding(t, registeredBindings, viewDetailName, 'o')(gui, detailView))

			then_detailCursorIs(t, subject.detailState.viewState, selectionStart)
			if actual := subject.detailState.viewState.visualAnchor; actual != selectionEnd {
				t.Fatalf("expected visual anchor %+v, actual %+v", selectionEnd, actual)
			}
		})
	}
}

func TestReviewMode_GivenDescriptionVisualSelection_WhenPressingO_ThenItMovesTheCursorToTheOtherSelectionEnd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body.zeta",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body.zeta": "Body.zeta"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))
	then_noError(t, subject.focusUserView(gui, nil))
	then_noError(t, subject.focusDetailView(gui, nil))
	given_detailCursorOnSegment(t, gui, subject, "Body.zeta")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	then_noError(t, subject.enterDetailVisualMode(gui, detailView))
	selectionStart := subject.detailState.viewState.cursor
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	selectionEnd := subject.detailState.viewState.cursor

	registeredBindings := subject.registeredKeybindingSpecs()
	then_noError(t, given_handlerForBinding(t, registeredBindings, viewDetailName, 'o')(gui, detailView))

	then_detailCursorIs(t, subject.detailState.viewState, selectionStart)
	if actual := subject.detailState.viewState.visualAnchor; actual != selectionEnd {
		t.Fatalf("expected visual anchor %+v, actual %+v", selectionEnd, actual)
	}
}

func TestStoryReviewMode_GivenChapterVisualSelection_WhenPressingO_ThenItMovesTheCursorToTheOtherSelectionEnd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/story",
				ChangedFiles: 1,
				Author:       &githubcli.PullRequestAuthor{Login: "octocat"},
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Summary: "A calmer way to review the pull request.",
		Chapters: []story.Chapter{{
			ID:        "chapter-1",
			Title:     "The Renderer Wakes",
			Narrative: "Alpha Beta",
			Files:     []string{"internal/tui/render.go"},
		}},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingStoryReviewMode(t, gui, subject))
	then_noError(t, subject.focusDetailView(gui, nil))
	given_detailCursorOnSegment(t, gui, subject, "The Renderer Wakes")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	then_noError(t, subject.enterDetailVisualMode(gui, detailView))
	selectionStart := subject.detailState.viewState.cursor
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	then_noError(t, subject.moveDetailCursorRight(gui, detailView))
	selectionEnd := subject.detailState.viewState.cursor

	registeredBindings := subject.registeredKeybindingSpecs()
	then_noError(t, given_handlerForBinding(t, registeredBindings, viewDetailName, 'o')(gui, detailView))

	then_detailCursorIs(t, subject.detailState.viewState, selectionStart)
	if actual := subject.detailState.viewState.visualAnchor; actual != selectionEnd {
		t.Fatalf("expected visual anchor %+v, actual %+v", selectionEnd, actual)
	}
}

func TestPullRequestBuildRunPopup_GivenVisualSelection_WhenPressingO_ThenItMovesTheCursorToTheOtherSelectionEnd(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)

	then_noError(t, subject.enterPullRequestBuildRunPopupVisualMode(gui, popupView))
	selectionStart := subject.pullRequestBuildRunPopup.viewState.cursor
	then_noError(t, subject.movePullRequestBuildRunPopupCursorRight(gui, popupView))
	then_noError(t, subject.movePullRequestBuildRunPopupCursorRight(gui, popupView))
	then_noError(t, subject.movePullRequestBuildRunPopupCursorRight(gui, popupView))
	then_noError(t, subject.movePullRequestBuildRunPopupCursorRight(gui, popupView))
	selectionEnd := subject.pullRequestBuildRunPopup.viewState.cursor

	registeredBindings := subject.registeredKeybindingSpecs()
	then_noError(t, given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, 'o')(gui, popupView))

	then_detailCursorIs(t, subject.pullRequestBuildRunPopup.viewState, selectionStart)
	if actual := subject.pullRequestBuildRunPopup.viewState.visualAnchor; actual != selectionEnd {
		t.Fatalf("expected popup visual anchor %+v, actual %+v", selectionEnd, actual)
	}
}
