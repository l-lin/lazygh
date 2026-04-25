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

func TestLayout_GivenPullRequestComment_WhenRendering_ThenItShowsTheAuthorAsARoundedColoredPill(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {
				Title:       "Styled PR",
				Number:      110,
				Body:        "Body 110",
				BaseRefName: "main",
				HeadRefName: "feature-110",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Ship it",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
			},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{output: "Rendered comment"}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, "", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "comment author pill left separator")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, " "+detailCommentsIcon+" @reviewer-one ", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "comment author pill background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, " "+detailCommentsIcon+" @reviewer-one ", given_themeColorHex(t, theme.CommentAuthorBadgeForegroundHex), "comment author pill foreground")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, "", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "comment author pill right separator")
}
