package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgPageNavigationRequested_WhenApplying_ThenItReturnsATypedPageNavigationCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgPageNavigationRequested{Kind: pageNavigationKindHalfDown})

	if len(actual) != 1 {
		t.Fatalf("expected one page-navigation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(pageNavigationCmd); !ok {
		t.Fatalf("expected a pageNavigationCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgPageNavigationResolvedInBrowserContext_WhenApplying_ThenItMovesSideSelectionAndReturnsAViewportCommand(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 20)}))

	actual := Update(subject, MsgPageNavigationResolved{Kind: pageNavigationKindHalfDown, PageSize: 6})

	if len(actual) != 1 {
		t.Fatalf("expected one viewport command, actual %d", len(actual))
	}
	command, ok := actual[0].(sideListViewportCmd)
	if !ok {
		t.Fatalf("expected a sideListViewportCmd, actual %T", actual[0])
	}
	if actual := command.Placement; actual != viewportPlacementCenter {
		t.Fatalf("expected viewport placement %v, actual %v", viewportPlacementCenter, actual)
	}
	expected := pageDelta(6)
	if actual := subject.model.SelectedUserIndex(); actual != expected {
		t.Fatalf("expected selected user index %d, actual %d", expected, actual)
	}
}

func TestUpdate_GivenMsgPageNavigationResolvedInReviewContext_WhenApplying_ThenItMovesReviewSelectionAndReturnsAViewportCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_page_nav")
	subject.navigationState.reviewSession.selectedFileTreeRow = 0
	selectableRows, ok := subject.reviewSessionSelectableRows()
	if !ok || len(selectableRows) == 0 {
		t.Fatalf("expected selectable review rows, actual %v", selectableRows)
	}
	expected := adjustVisibleSelection(subject.navigationState.reviewSession.selectedFileTreeRow, selectableRows, pageDelta(6))

	actual := Update(subject, MsgPageNavigationResolved{Kind: pageNavigationKindHalfDown, PageSize: 6})

	if len(actual) != 1 {
		t.Fatalf("expected one viewport command, actual %d", len(actual))
	}
	command, ok := actual[0].(sideListViewportCmd)
	if !ok {
		t.Fatalf("expected a sideListViewportCmd, actual %T", actual[0])
	}
	if actual := command.Placement; actual != viewportPlacementCenter {
		t.Fatalf("expected viewport placement %v, actual %v", viewportPlacementCenter, actual)
	}
	if actual := subject.navigationState.reviewSession.selectedFileTreeRow; actual != expected {
		t.Fatalf("expected selected review row %d, actual %d", expected, actual)
	}
}

func TestPageNavigationCommand_GivenAResolvedPageSize_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	gui := given_headlessGuiWithSize(t, 80, 20)
	defer gui.Close()
	view, actualErr := gui.SetView(viewUserName, 0, 0, 39, 9, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	actualDispatched := []Msg(nil)

	executePageNavigationCommand(navigationCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		resolveView: func(gui *gocui.Gui, current *gocui.View, name string) *gocui.View {
			return view
		},
		currentViewName: func() string {
			return viewUserName
		},
	}, gui, pageNavigationCmd{Kind: pageNavigationKindHalfDown})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgPageNavigationResolved)
	if !ok {
		t.Fatalf("expected a MsgPageNavigationResolved, actual %T", actualDispatched[0])
	}
	if actual := message.Kind; actual != pageNavigationKindHalfDown {
		t.Fatalf("expected resolved kind %v, actual %v", pageNavigationKindHalfDown, actual)
	}
	expected := viewPageSize(view)
	if actual := message.PageSize; actual != expected {
		t.Fatalf("expected resolved page size %d, actual %d", expected, actual)
	}
}
