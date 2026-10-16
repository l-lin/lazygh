package tui

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestCopyPullRequestBranch_GivenPullRequestContexts_WhenYanking_ThenItCopiesOnlyTheSelectedHeadBranch(t *testing.T) {
	testCases := []struct {
		name             string
		detailFocus      bool
		visualMode       bool
		branch           string
		missingDetail    bool
		userFocus        bool
		clipboardError   error
		expectedWrites   []string
		expectedFeedback string
	}{
		{name: "pull request list", branch: "feature/branch", expectedWrites: []string{"feature/branch"}, expectedFeedback: iconCopy + " Branch copied"},
		{name: "detail", detailFocus: true, branch: " feature/branch ", expectedWrites: []string{"feature/branch"}, expectedFeedback: iconCopy + " Branch copied"},
		{name: "visual selection", detailFocus: true, visualMode: true, branch: "feature/branch", expectedWrites: []string{"feature/branch"}, expectedFeedback: iconCopy + " Branch copied"},
		{name: "missing detail", missingDetail: true, expectedFeedback: iconUnavailable + " No PR branch"},
		{name: "empty branch", branch: "  ", expectedFeedback: iconUnavailable + " No PR branch"},
		{name: "user without PR context", userFocus: true, branch: "feature/branch", expectedFeedback: iconUnavailable + " No PR branch"},
		{name: "clipboard failure", branch: "feature/branch", clipboardError: errors.New("clipboard failed"), expectedWrites: []string{"feature/branch"}, expectedFeedback: yankFailureMessage},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := given_model()
			model.FocusPullRequestsView()
			model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
				myPullRequestRow(githubcli.PullRequest{Number: 42, Title: "Selected PR", Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}),
			})
			if testCase.detailFocus {
				model.OpenDetail()
			}
			if testCase.userFocus {
				model.FocusUserView()
			}
			subject := NewProgramWithModel(model)
			clipboardWriter := &fakeClipboardWriter{writeErr: testCase.clipboardError}
			subject.clipboardWriter = clipboardWriter
			if !testCase.missingDetail {
				subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{BaseRefName: "main", HeadRefName: testCase.branch})}
			}
			subject.pullRequestDetailCache["acme/widgets#99"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{HeadRefName: "wrong-branch"})}
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)
			then_noError(t, subject.layout(gui))
			if testCase.visualMode {
				subject.detailState.viewState.mode = detailVisualMode
			}

			then_noError(t, subject.copyPullRequestBranch(gui, nil))

			actual := clipboardWriter.writes
			expected := testCase.expectedWrites
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("expected clipboard writes %v, actual %v", expected, actual)
			}
			then_statusLineContains(t, gui, testCase.expectedFeedback)
		})
	}
}

func TestKeybindingSpecs_GivenDefaultOrOverriddenBranchBinding_WhenListingBindings_ThenItAppliesToAllPRViews(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		overrides appconfig.KeymapOverrides
		key       rune
		mod       gocui.Modifier
	}{
		{name: "default", key: 'y', mod: gocui.ModAlt},
		{name: "override", overrides: appconfig.KeymapOverrides{"pull_requests": {"copy_pull_request_branch": {"u"}}}, key: 'u'},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_programWithKeymapOverrides(given_model(), testCase.overrides)

			actual := subject.keybindingSpecs()

			for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
				then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: testCase.key, mod: testCase.mod, handler: subject.copyPullRequestBranch})
				then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'Y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
			}
		})
	}
}

func TestUpdate_GivenReviewMode_WhenRequestingBranchYankFromAnyPane_ThenItTargetsTheReviewedPR(t *testing.T) {
	for _, focus := range []Focus{FocusUserView, FocusPullRequestsView, FocusDetailView} {
		model := given_model()
		switch focus {
		case FocusPullRequestsView:
			model.FocusPullRequestsView()
		case FocusDetailView:
			model.OpenDetail()
		}
		subject := NewProgramWithModel(model)
		subject.navigationState.reviewSession = reviewSessionState{
			active:  true,
			summary: githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{Number: 99, Repository: githubcli.Repository{NameWithOwner: "acme/reviewed"}}),
		}
		subject.pullRequestDetailCache["acme/reviewed#99"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{HeadRefName: "feature/reviewed"})}

		actual := Update(subject, MsgCopyPullRequestBranchRequested{})

		if len(actual) != 1 {
			t.Fatalf("expected one clipboard command for focus %v, actual %v", focus, actual)
		}
		command, ok := actual[0].(writeClipboardCmd)
		expected := "feature/reviewed"
		if !ok || command.Text != expected {
			t.Fatalf("expected clipboard text %q for focus %v, actual %+v", expected, focus, actual[0])
		}
	}
}

func TestUpdate_GivenHelpOrSearchActive_WhenRequestingBranchYank_ThenItDoesNotWriteTheClipboard(t *testing.T) {
	for _, helpVisible := range []bool{true, false} {
		subject := NewProgramWithModel(given_model())
		subject.overlayState.helpVisible = helpVisible
		if !helpVisible {
			subject.model.StartSearch()
		}

		actual := Update(subject, MsgCopyPullRequestBranchRequested{})

		if len(actual) != 0 {
			t.Fatalf("expected no commands, actual %v", actual)
		}
	}
}
