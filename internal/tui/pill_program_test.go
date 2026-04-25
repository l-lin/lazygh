package tui

import (
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestLayout_GivenOpenPullRequestDetail_WhenRendering_ThenItShowsTheStatusAsARoundedColoredPill(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {Title: "Styled PR", Number: 110, Body: "Body 110", BaseRefName: "main", HeadRefName: "feature-110", State: "OPEN"},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	statusLineIndex := given_viewLineIndexContaining(t, detailView, "OPEN")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, statusLineIndex, "", given_themeColorHex(t, theme.PullRequestStatusOpenBackgroundHex), "status pill left separator")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, statusLineIndex, " "+detailStatusIcon+" OPEN ", given_themeColorHex(t, theme.PullRequestStatusOpenBackgroundHex), "status pill background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, statusLineIndex, " "+detailStatusIcon+" OPEN ", given_themeColorHex(t, theme.PullRequestStatusOpenForegroundHex), "status pill foreground")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, statusLineIndex, "", given_themeColorHex(t, theme.PullRequestStatusOpenBackgroundHex), "status pill right separator")
}
