package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestFingerprintKeybindingSpecs_GivenSameKeyDifferentHandlers_WhenFingerprinting_ThenItChanges(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	first := fingerprintKeybindingSpecs([]keybindingSpec{{viewName: viewDetailName, key: gocui.KeyCtrlR, handler: subject.toggleHelp}})
	second := fingerprintKeybindingSpecs([]keybindingSpec{{viewName: viewDetailName, key: gocui.KeyCtrlR, handler: subject.toggleInlineConversationVisibility}})

	if first == second {
		t.Fatalf("expected different fingerprints for different handlers, actual %q", first)
	}
}

func TestInlineCommentResolutionShortcut_GivenBrowserChangesCursorOnInlineComment_WhenPressingCtrlR_ThenItShowsGHLoadingFoldsTheThreadBeforeCompletionAndResolvesIt(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	asynchronousRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
	subject.asyncRunner = asynchronousRunner
	subject.uiUpdater = immediateUIUpdater{}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.registeredKeybindingSpecs(), viewDetailName, gocui.KeyCtrlR)
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)

	if len(asynchronousRunner.runs) != 1 {
		t.Fatalf("expected one queued async resolution run, actual %d", len(asynchronousRunner.runs))
	}
	expectedLoading := formatRunningCommandStatus(formatStatusLineCommand("gh", "api", "graphql"))
	if actual := subject.statusLinePresenter().Text(); actual != subject.loadingSpinnerStatus(expectedLoading) {
		t.Fatalf("expected status line %q, actual %q", subject.loadingSpinnerStatus(expectedLoading), actual)
	}
	then_statusLineContains(t, gui, expectedLoading)
	if strings.Contains(detailView.Buffer(), "Rendered original inline body") {
		t.Fatalf("expected the inline thread body to fold before async completion, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Resolved") {
		t.Fatalf("expected the thread header to stay unresolved until async completion, actual %q", detailView.Buffer())
	}

	asynchronousRunner.runs[0]()

	if !reflect.DeepEqual(loader.resolveReviewThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected resolved thread ids %v, actual %v", []string{"thread-1"}, loader.resolveReviewThreadIDs)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, inlineCommentResolvedSuccessMessage)
}

func TestInlineCommentResolutionShortcut_GivenBrowserChangesCursorOnResolvedInlineComment_WhenPressingCtrlR_ThenItShowsGHLoadingUnfoldsTheThreadBeforeCompletionAndMarksItUnresolved(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithResolvedInlineThreadForChangesResolutionTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithResolvedInlineThreadForChangesResolutionTests()},
	}
	asynchronousRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:43")
	subject.asyncRunner = asynchronousRunner
	subject.uiUpdater = immediateUIUpdater{}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.registeredKeybindingSpecs(), viewDetailName, gocui.KeyCtrlR)
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)

	if len(asynchronousRunner.runs) != 1 {
		t.Fatalf("expected one queued async unresolve run, actual %d", len(asynchronousRunner.runs))
	}
	expectedLoading := formatRunningCommandStatus(formatStatusLineCommand("gh", "api", "graphql"))
	then_statusLineContains(t, gui, expectedLoading)
	if !strings.Contains(detailView.Buffer(), "Rendered original inline body") {
		t.Fatalf("expected the resolved thread body to unfold before async completion, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Unresolved") {
		t.Fatalf("expected the thread header to stay resolved until async completion, actual %q", detailView.Buffer())
	}

	asynchronousRunner.runs[0]()

	if !reflect.DeepEqual(loader.unresolveReviewThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected unresolved thread ids %v, actual %v", []string{"thread-1"}, loader.unresolveReviewThreadIDs)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, inlineCommentUnresolvedSuccessMessage)
}

func TestKeybindingSpecs_GivenInlineCommentResolutionShortcut_WhenCursorMovesBetweenSupportedAndUnsupportedTargets_ThenCtrlRAppearsOnlyOnInlineComments(t *testing.T) {
	testCases := []struct {
		name        string
		newSubject  func() *Program
		prepare     func(*testing.T, *gocui.Gui, *Program)
		positionOn  func(*testing.T, *gocui.Gui, *Program)
		positionOff func(*testing.T, *gocui.Gui, *Program)
	}{
		{
			name: "browser changes",
			newSubject: func() *Program {
				loader := &fakePullRequestDetailLoader{
					details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
					diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
				}
				subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
				subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
				return subject
			},
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_browserChangesDetailFocusForInlineComment(t, gui, subject)
			},
			positionOn: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
			},
			positionOff: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")
			},
		},
		{
			name: "browser comments",
			newSubject: func() *Program {
				loader := &fakePullRequestDetailLoader{
					details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
				}
				subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
				subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
					"General feedback":   "Rendered general feedback",
					"Inline thread body": "Rendered inline thread body",
				}}
				return subject
			},
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				then_noError(t, subject.layout(gui))
				then_noError(t, subject.openDetail(gui, nil))
				then_noError(t, subject.nextDetailTab(gui, nil))
			},
			positionOn: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
			},
			positionOff: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered general feedback")
			},
		},
		{
			name: "review diff",
			newSubject: func() *Program {
				loader := &fakePullRequestDetailLoader{
					startReviewID: "PRR_pending",
					diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
				}
				subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
				subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
				return subject
			},
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailFocusForActions(t, gui, subject)
			},
			positionOn: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
			},
			positionOff: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject := testCase.newSubject()
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)
			testCase.prepare(t, gui, subject)

			actionID := keybindingActionID{scope: keymapScopePullRequests, action: "toggle_inline_comment_resolution"}

			testCase.positionOn(t, gui, subject)
			then_bindingKeyExists(t, subject.registeredKeybindingSpecs(), viewDetailName, gocui.KeyCtrlR)
			if actualBindings := subject.resolvedBindingsForActionID(actionID); len(actualBindings) == 0 {
				t.Fatalf("expected the inline-comment resolution action to resolve bindings on inline comments")
			}

			testCase.positionOff(t, gui, subject)
			if actualBindings := subject.resolvedBindingsForActionID(actionID); len(actualBindings) != 0 {
				t.Fatalf("expected the inline-comment resolution action to disappear away from inline comments, actual %+v", actualBindings)
			}
		})
	}
}
