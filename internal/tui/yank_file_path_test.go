package tui

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func given_pathYankProgram(mode reviewSessionMode) *Program {
	model := given_model()
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	files := []reviewDiffFile{
		{Path: "internal/tui/new name.go", PreviousPath: "internal/tui/old.go", ChangeType: reviewDiffChangeTypeRenamed},
		{Path: "internal/tui/deleted.go", ChangeType: reviewDiffChangeTypeRemoved},
		{Path: "assets/image.png", Placeholder: "Binary file"},
	}
	subject.navigationState.reviewSession = reviewSessionState{active: true, mode: mode, summary: githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}})}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{Files: files, FileTree: buildReviewDiffFileTree(files)}}
	if mode == reviewSessionModeStory {
		subject.navigationState.reviewSession.story = buildReviewStoryData(story.Review{Chapters: []story.Chapter{{ID: "chapter", Title: "Changes", Files: []string{files[0].Path, files[1].Path, files[2].Path}}}}, files)
	}
	return subject
}

func TestCopyReviewFilePath_GivenTreeRows_WhenYanking_ThenItCopiesOnlyTheSelectedPath(t *testing.T) {
	for _, testCase := range []struct {
		name             string
		mode             reviewSessionMode
		row              int
		collapsed        bool
		empty            bool
		missing          bool
		clipboardError   error
		expected         string
		expectedFeedback string
	}{
		{name: "compacted directory", row: 0, expected: "internal/tui/"},
		{name: "collapsed directory", row: 0, collapsed: true, expected: "internal/tui/"},
		{name: "renamed file with spaces", row: 1, expected: "internal/tui/new name.go"},
		{name: "deleted file", row: 2, expected: "internal/tui/deleted.go"},
		{name: "binary file", row: 4, expected: "assets/image.png"},
		{name: "visible row after collapsed directory", row: 2, collapsed: true, expected: "assets/image.png"},
		{name: "story chapter", mode: reviewSessionModeStory, row: 0, expectedFeedback: iconUnavailable + " No path selected"},
		{name: "story renamed file", mode: reviewSessionModeStory, row: 1, expected: "internal/tui/new name.go"},
		{name: "story deleted file", mode: reviewSessionModeStory, row: 2, expected: "internal/tui/deleted.go"},
		{name: "story binary file", mode: reviewSessionModeStory, row: 3, expected: "assets/image.png"},
		{name: "empty tree", empty: true, expectedFeedback: iconUnavailable + " No path selected"},
		{name: "missing diff", missing: true, expectedFeedback: iconUnavailable + " No path selected"},
		{name: "clipboard failure", row: 1, expected: "internal/tui/new name.go", clipboardError: errors.New("clipboard failed"), expectedFeedback: yankFailureMessage},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_pathYankProgram(testCase.mode)
			subject.navigationState.reviewSession.selectedFileTreeRow = testCase.row
			if testCase.collapsed {
				subject.navigationState.reviewSession.collapsedTreeRowIDs = map[string]bool{reviewDiffTreeRowIDForDirectory("internal/tui/"): true}
			}
			if testCase.empty {
				subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{}
			}
			if testCase.missing {
				delete(subject.pullRequestDiffCache, "acme/widgets#42")
			}
			clipboardWriter := &fakeClipboardWriter{writeErr: testCase.clipboardError}
			subject.clipboardWriter = clipboardWriter
			expectedState := subject.navigationState.reviewSession
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)
			then_noError(t, subject.layout(gui))

			then_noError(t, subject.copyReviewFilePath(gui, nil))

			actual := clipboardWriter.writes
			var expected []string
			if testCase.expected != "" {
				expected = []string{testCase.expected}
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("expected clipboard writes %v, actual %v", expected, actual)
			}
			expectedFeedback := testCase.expectedFeedback
			if expectedFeedback == "" {
				expectedFeedback = iconCopy + " Path copied"
			}
			then_statusLineContains(t, gui, expectedFeedback)
			if actual := subject.navigationState.reviewSession; !reflect.DeepEqual(actual, expectedState) {
				t.Fatalf("expected review state %+v, actual %+v", expectedState, actual)
			}
			if actual := subject.model.Focus(); actual != FocusPullRequestsView {
				t.Fatalf("expected file tree focus, actual %v", actual)
			}
		})
	}
}

func TestCopyReviewFilePath_GivenOtherContexts_WhenRequested_ThenItDoesNotWriteTheClipboard(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		setup func(*Program)
	}{
		{"browser", func(p *Program) { p.navigationState.reviewSession.active = false }},
		{"view one", func(p *Program) { p.model.FocusUserView() }},
		{"view zero", func(p *Program) { p.model.OpenDetail() }},
		{"help", func(p *Program) { p.overlayState.helpVisible = true }},
		{"search", func(p *Program) { p.model.StartSearch() }},
		{"actions", func(p *Program) { p.model.OpenActionsPopup(1) }},
		{"editor", func(p *Program) {
			p.overlayState.modalEditor = newModalEditorStateWithSubmitDescriptor("Comment", "", modalEditorSubmitDescriptor{})
		}},
		{"build popup", func(p *Program) { p.pullRequestBuildRunPopup = &pullRequestBuildRunPopupState{} }},
		{"command", func(p *Program) { Update(p, MsgOpenCommandMode{}) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_pathYankProgram(reviewSessionModeDiff)
			testCase.setup(subject)

			actual := Update(subject, MsgCopyReviewFilePathRequested{})

			if len(actual) != 0 {
				t.Fatalf("expected no clipboard commands, actual %v", actual)
			}
		})
	}
}

func TestCopyReviewFilePath_GivenDefaultOrOverriddenKey_WhenPressingIt_ThenItCopiesFromViewTwo(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		overrides appconfig.KeymapOverrides
		key       rune
	}{
		{name: "default", key: 'y'},
		{name: "override", overrides: appconfig.KeymapOverrides{"review": {"copy_file_path": {"u"}}}, key: 'u'},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_pathYankProgram(reviewSessionModeStory)
			configured := given_programWithKeymapOverrides(subject.model, testCase.overrides)
			configured.navigationState = subject.navigationState
			configured.pullRequestDiffCache = subject.pullRequestDiffCache
			configured.navigationState.reviewSession.selectedFileTreeRow = 1
			clipboardWriter := &fakeClipboardWriter{}
			configured.clipboardWriter = clipboardWriter
			gui := given_headlessGui(t)
			defer gui.Close()
			configured.configureGUI(gui)
			then_noError(t, configured.layout(gui))

			handler := given_handlerForBinding(t, configured.keybindingSpecs(), viewPullRequestsName, testCase.key)
			then_noError(t, handler(gui, nil))

			actual := clipboardWriter.writes
			expected := []string{"internal/tui/new name.go"}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("expected clipboard writes %v, actual %v", expected, actual)
			}
			then_bindingExists(t, configured.keybindingSpecs(), keybindingSpec{viewName: viewDetailName, key: 'y', mod: gocui.ModNone, handler: configured.startDetailYank})
		})
	}
}

func TestHelpPresenter_GivenModesAndViews_WhenListingPathYankHelp_ThenItAppearsOnlyInReviewViewTwo(t *testing.T) {
	for _, mode := range []ScreenMode{ScreenModeBrowser, ScreenModeReview, ScreenModeStoryReview} {
		for _, focus := range []Focus{FocusUserView, FocusPullRequestsView, FocusDetailView} {
			subject := helpPresenter{actionContext: ActionContext{Mode: mode, ActiveView: ViewState{Focus: focus}}, keyResolver: newKeybindingLabelResolver([]resolvedKeybindingAction{given_resolvedKeybindingAction(keymapScopeReview, "copy_file_path", "u")})}

			actual := subject.localHelpEntries()

			found := false
			for _, entry := range actual {
				if entry.Description == "Copy relative path" {
					found = true
					if entry.Key != "u" {
						t.Fatalf("expected configured key u, actual %q", entry.Key)
					}
				}
			}
			expected := mode != ScreenModeBrowser && focus == FocusPullRequestsView
			if found != expected {
				t.Fatalf("expected path help %v for mode %v focus %v, actual %v", expected, mode, focus, found)
			}
		}
	}
}
