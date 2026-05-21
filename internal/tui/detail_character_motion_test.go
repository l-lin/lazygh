package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestDetailViewState_GivenRenderedDetailText_WhenUsingCharacterMotions_ThenItMovesLikeVimFindAndTill(t *testing.T) {
	document := newDetailDocument("cab bag", 40)
	subject := newDetailViewState()

	subject.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)
	if !subject.consumePendingCharacterMotion(document, 3, 'b') {
		t.Fatal("expected forward find to consume its pending target")
	}
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 2})

	subject.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionBeforeMatch)
	if !subject.consumePendingCharacterMotion(document, 3, 'a') {
		t.Fatal("expected forward till to consume its pending target")
	}
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 4})

	subject.cursor = detailPosition{line: 0, column: 6}
	subject.preferredColumn = 6
	subject.armCharacterMotion(detailCharacterMotionDirectionBackward, detailCharacterMotionMatch)
	if !subject.consumePendingCharacterMotion(document, 3, 'b') {
		t.Fatal("expected backward find to consume its pending target")
	}
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 4})

	subject.armCharacterMotion(detailCharacterMotionDirectionBackward, detailCharacterMotionAfterMatch)
	if !subject.consumePendingCharacterMotion(document, 3, 'a') {
		t.Fatal("expected backward till to consume its pending target")
	}
	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 2})
}

func TestDetailViewState_GivenCharacterMotionHistory_WhenRepeating_ThenSemicolonKeepsTheDirectionAndCommaReversesIt(t *testing.T) {
	document := newDetailDocument("ab,cd,ef", 40)

	find := newDetailViewState()
	find.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)
	if !find.consumePendingCharacterMotion(document, 3, ',') {
		t.Fatal("expected the initial find motion to be consumed")
	}
	if !find.repeatCharacterMotion(document, 3, false) {
		t.Fatal("expected semicolon to repeat the last find motion")
	}
	then_detailCursorIs(t, find, detailPosition{line: 0, column: 5})
	if !find.repeatCharacterMotion(document, 3, true) {
		t.Fatal("expected comma to repeat the last find motion in reverse")
	}
	then_detailCursorIs(t, find, detailPosition{line: 0, column: 2})

	till := newDetailViewState()
	till.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionBeforeMatch)
	if !till.consumePendingCharacterMotion(document, 3, ',') {
		t.Fatal("expected the initial till motion to be consumed")
	}
	if !till.repeatCharacterMotion(document, 3, false) {
		t.Fatal("expected semicolon to repeat the last till motion")
	}
	then_detailCursorIs(t, till, detailPosition{line: 0, column: 4})
	if !till.repeatCharacterMotion(document, 3, true) {
		t.Fatal("expected comma to reverse the last till motion")
	}
	then_detailCursorIs(t, till, detailPosition{line: 0, column: 3})
}

func TestDetailCharacterMotion_GivenDetailVisualMode_WhenPressingVFA_ThenItKeepsVisualModeAndExtendsTheSelectionToTheNextA(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "banana"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	given_detailCursorOnSegment(t, gui, subject, "banana")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	start := given_detailPositionOfSegmentOccurrence(t, gui, subject, "banana", 0)

	registeredBindings := subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'v')(gui, detailView)
	then_noError(t, actualErr)
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'a')(gui, detailView)
	then_noError(t, actualErr)

	if subject.detailViewState.mode != detailVisualMode {
		t.Fatalf("expected detail mode %v, actual %v", detailVisualMode, subject.detailViewState.mode)
	}
	then_detailSelectionIs(t, subject.currentDetailDocument(detailView), subject.detailViewState, start, detailPosition{line: start.line, column: start.column + 1})
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
}

func TestDetailCharacterMotion_GivenBrowserDetailTabs_WhenPressingForwardFindWithAnUnboundTarget_ThenItNavigatesEachRenderedTab(t *testing.T) {
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

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	testCases := []struct {
		name          string
		tab           DetailTab
		startSegment  string
		targetSegment string
	}{
		{name: "description", tab: DescriptionDetailTab, startSegment: "Description.zeta", targetSegment: "zeta"},
		{name: "comments", tab: CommentsDetailTab, startSegment: "Comment.zeta", targetSegment: "zeta"},
		{name: "commits", tab: CommitsDetailTab, startSegment: "Commit.zeta", targetSegment: "zeta"},
		{name: "changes", tab: ChangesDetailTab, startSegment: "Change.zeta", targetSegment: "zeta"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject.activeDetailTab = testCase.tab
			actualErr = subject.refreshDetailView(gui)
			then_noError(t, actualErr)

			given_detailCursorOnSegment(t, gui, subject, testCase.startSegment)
			expected := given_detailPositionOfSegmentOccurrence(t, gui, subject, testCase.targetSegment, 0)
			detailView, actualErr := gui.View(viewDetailName)
			then_noError(t, actualErr)

			registeredBindings := subject.registeredKeybindingSpecs()
			actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
			then_noError(t, actualErr)
			registeredBindings = subject.registeredKeybindingSpecs()
			actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'z')(gui, detailView)
			then_noError(t, actualErr)

			then_detailCursorIs(t, subject.detailViewState, expected)
		})
	}
}

func TestDetailCharacterMotion_GivenReviewModeDescription_WhenPressingForwardFindWithAnUnboundTarget_ThenItNavigatesTheRenderedDescription(t *testing.T) {
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

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)

	given_detailCursorOnSegment(t, gui, subject, "Body.zeta")
	expected := given_detailPositionOfSegmentOccurrence(t, gui, subject, "zeta", 0)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	registeredBindings := subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'z')(gui, detailView)
	then_noError(t, actualErr)

	then_detailCursorIs(t, subject.detailViewState, expected)
}

func TestDetailCharacterMotion_GivenReviewModeViewZero_WhenPressingForwardFindWithAnUnboundTarget_ThenItNavigatesTheRenderedDiff(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, document, "internal/tui/render.go")
	segmentColumn := given_runeIndexInString(t, visibleLine, "internal/tui/render.go")
	expectedColumn := segmentColumn + given_runeIndexInString(t, "internal/tui/render.go", ".")
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: segmentColumn}
	subject.detailViewState.preferredColumn = segmentColumn
	subject.syncDetailViewState(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)

	registeredBindings := subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, '.')(gui, detailView)
	then_noError(t, actualErr)

	then_detailCursorIs(t, subject.detailViewState, detailPosition{line: lineIndex, column: expectedColumn})
}

func TestDetailCharacterMotion_GivenBrowserDetailFocus_WhenPressingSemicolonAndComma_ThenItRepeatsTheLastCharacterMotion(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Description.zeta.zulu": "Description.zeta.zulu"}}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Description.zeta.zulu", State: "OPEN"})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_detailCursorOnSegment(t, gui, subject, "Description.zeta.zulu")
	firstMatch := given_detailPositionOfSegmentOccurrence(t, gui, subject, "zeta", 0)
	secondMatch := given_detailPositionOfSegmentOccurrence(t, gui, subject, "zulu", 0)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	registeredBindings := subject.registeredKeybindingSpecs()

	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'z')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, firstMatch)

	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, ';')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, secondMatch)

	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, ',')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, firstMatch)
}

func TestDetailCharacterMotion_GivenReviewModeViewZero_WhenPressingFCommaSemicolonAndComma_ThenTheTargetAndRepeatsStayInTheDiff(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, document, "@@ -1,2 +1,3 @@")
	startColumn := given_runeIndexInString(t, visibleLine, "@@ -1,2 +1,3 @@")
	firstComma := startColumn + given_runeIndexInString(t, "@@ -1,2 +1,3 @@", ",")
	secondComma := firstComma + 5
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: startColumn}
	subject.detailViewState.preferredColumn = startColumn
	subject.syncDetailViewState(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)
	registeredBindings := subject.registeredKeybindingSpecs()

	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, 'f')(gui, detailView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, ',')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, detailPosition{line: lineIndex, column: firstComma})

	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, ';')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, detailPosition{line: lineIndex, column: secondComma})

	actualErr = given_handlerForBinding(t, registeredBindings, viewDetailName, ',')(gui, detailView)
	then_noError(t, actualErr)
	then_detailCursorIs(t, subject.detailViewState, detailPosition{line: lineIndex, column: firstComma})
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenPressingForwardFindWithAnUnboundTarget_ThenItMovesThePopupCursor(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Build.zeta"})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	registeredBindings := subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, 'f')(gui, popupView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, 'z')(gui, popupView)
	then_noError(t, actualErr)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 6}) {
		t.Fatalf("expected popup cursor %+v, actual %+v", detailPosition{line: 0, column: 6}, actual)
	}
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenPressingSemicolonAndComma_ThenItRepeatsTheLastCharacterMotion(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Build.zeta.zulu"})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	registeredBindings := subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, 'f')(gui, popupView)
	then_noError(t, actualErr)
	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, 'z')(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 6}) {
		t.Fatalf("expected popup cursor %+v after the initial find, actual %+v", detailPosition{line: 0, column: 6}, actual)
	}

	registeredBindings = subject.registeredKeybindingSpecs()
	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, ';')(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 11}) {
		t.Fatalf("expected popup cursor %+v after semicolon, actual %+v", detailPosition{line: 0, column: 11}, actual)
	}

	actualErr = given_handlerForBinding(t, registeredBindings, viewPullRequestBuildInfoName, ',')(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor; actual != (detailPosition{line: 0, column: 6}) {
		t.Fatalf("expected popup cursor %+v after comma, actual %+v", detailPosition{line: 0, column: 6}, actual)
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingCharacterMotionBindings_ThenDetailAndBuildPopupExposeTheVimFindPrefixes(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, key := range []rune{'f', 'F', 't', 'T'} {
		then_bindingKeyExists(t, actual, viewDetailName, key)
		then_bindingKeyExists(t, actual, viewPullRequestBuildInfoName, key)
	}
	for _, key := range []rune{';', ','} {
		then_bindingDoesNotExist(t, actual, viewDetailName, key)
		then_bindingDoesNotExist(t, actual, viewPullRequestBuildInfoName, key)
	}
}
