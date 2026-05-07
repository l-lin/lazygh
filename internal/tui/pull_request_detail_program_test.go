package tui

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestLayout_GivenSelectedPullRequestSummary_WhenRendering_ThenItLoadsRichDetailAndShowsDescriptionAndConversationsInSeparateTabs(t *testing.T) {
	firstSummary := githubcli.PullRequest{Title: "First PR", Number: 101, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-1"}
	secondSummary := githubcli.PullRequest{Title: "Second PR", Number: 102, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-2"}
	firstDetail := githubcli.PullRequestDetail{
		Title:            "First PR",
		Number:           101,
		Body:             "Body 101",
		BaseRefName:      "main",
		HeadRefName:      "feature-101",
		State:            "OPEN",
		Mergeable:        "MERGEABLE",
		MergeStateStatus: "CLEAN",
		ReviewRequests: []githubcli.PullRequestReviewRequest{{
			RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"},
		}},
		Reviews: []githubcli.PullRequestReview{{
			Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"},
			State:       "APPROVED",
			SubmittedAt: "2026-04-21T10:00:00Z",
		}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
			Name:         "lint",
			WorkflowName: "CI",
			Status:       "COMPLETED",
			Conclusion:   "SUCCESS",
		}},
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Comment 101",
			CreatedAt: "2026-04-18T10:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:         "Inline 101",
			CreatedAt:    "2026-04-18T10:30:00Z",
			Path:         "internal/tui/render.go",
			Line:         43,
			OriginalLine: 43,
			Side:         "RIGHT",
			DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
		}},
	}
	secondDetail := githubcli.PullRequestDetail{
		Title:       "Second PR",
		Number:      102,
		Body:        "Body 102",
		BaseRefName: "main",
		HeadRefName: "feature-102",
		State:       "OPEN",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"},
			Body:      "Comment 102",
			CreatedAt: "2026-04-18T11:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline-two"},
			Body:         "Inline 102",
			CreatedAt:    "2026-04-18T11:30:00Z",
			Path:         "internal/tui/pull_request_detail_program_test.go",
			Line:         19,
			OriginalLine: 19,
			Side:         "RIGHT",
			DiffHunk:     "@@ -18,2 +18,2 @@\n actualErr := subject.layout(gui)\n-if !strings.Contains(detailView.Buffer(), \"Rendered comment 102\") {\n+if !strings.Contains(detailView.Buffer(), \"Rendered inline 102\") {",
		}},
	}
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(firstSummary),
		myPullRequestRow(secondSummary),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#101": firstDetail,
			"acme/widgets#102": secondDetail,
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 101":    "Rendered body 101",
		"Body 102":    "Rendered body 102",
		"Comment 101": "Rendered comment 101",
		"Comment 102": "Rendered comment 102",
		"Inline 101":  "Rendered inline 101",
		"Inline 102":  "Rendered inline 102",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expectedSeparator := strings.Repeat("─", detailView.InnerWidth())
	for _, expected := range []string{"acme/widgets#101 First PR", " " + pullRequestOverviewPendingIcon + " Reviewers (1/2)", " " + pullRequestOverviewPendingIcon + " Merge Checks", " " + pullRequestOverviewSuccessIcon + " Builds", "Rendered body 101"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the overview tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	for _, hidden := range []string{"@reviewer-requested", "CI / lint (Successful)", "Changes can be cleanly merged."} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected the description tab to keep overview block bodies folded, actual %q", detailView.Buffer())
		}
	}
	headerLineIndex := given_viewLineIndexContaining(t, detailView, detailStatusIcon+" OPEN")
	reviewersLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewPendingIcon+" Reviewers (1/2)")
	mergeChecksLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewPendingIcon+" Merge Checks")
	buildsLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewSuccessIcon+" Builds")
	separatorLineIndex := given_viewLineIndexContaining(t, detailView, expectedSeparator)
	bodyLineIndex := given_viewLineIndexContaining(t, detailView, "Rendered body 101")
	if !(headerLineIndex < reviewersLineIndex && reviewersLineIndex < mergeChecksLineIndex && mergeChecksLineIndex < buildsLineIndex && buildsLineIndex < separatorLineIndex && separatorLineIndex < bodyLineIndex) {
		t.Fatalf("expected header, folded overview, separator, and body to stay ordered, actual %q", detailView.Buffer())
	}
	if mergeChecksLineIndex != reviewersLineIndex+1 || buildsLineIndex != mergeChecksLineIndex+1 {
		t.Fatalf("expected folded overview headers to stay consecutive without blank spacer lines, actual %q", detailView.Buffer())
	}
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, reviewersLineIndex, pullRequestOverviewPendingIcon+" Reviewers (1/2)", given_themeColorHex(t, theme.PendingHex), "reviewers overview header")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, mergeChecksLineIndex, pullRequestOverviewPendingIcon+" Merge Checks", given_themeColorHex(t, theme.PendingHex), "merge checks overview header")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, buildsLineIndex, pullRequestOverviewSuccessIcon+" Builds", given_themeColorHex(t, theme.SuccessHex), "builds overview header")
	if strings.Contains(detailView.Buffer(), "Rendered comment 101") {
		t.Fatalf("expected overview tab to hide comments, actual %q", detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (2)", CommitsDetailTab.Label(), ChangesDetailTab.Label()}, 0)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101"}, loader.detailCalls)
	}

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	for _, hidden := range []string{"acme/widgets#101 First PR", detailStatusIcon + " OPEN", "Reviewers", "Merge Checks", "Builds", "Rendered body 101"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected the conversations tab to hide overview metadata %q, actual %q", hidden, detailView.Buffer())
		}
	}
	for _, expected := range []string{" Comment", "Rendered comment 101", " Comment on line R43", detailInlineCommentLocationIcon + " internal/tui/render.go:43  +1  -1", "Rendered inline 101"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the conversations tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (2)", CommitsDetailTab.Label(), ChangesDetailTab.Label()}, 1)

	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101"}) {
		t.Fatalf("expected cached detail calls %v, actual %v", []string{"acme/widgets#101"}, loader.detailCalls)
	}

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered comment 102") {
		t.Fatalf("expected comments tab to contain %q after selection, actual %q", "Rendered comment 102", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered inline 102") {
		t.Fatalf("expected comments tab to contain %q after selection, actual %q", "Rendered inline 102", detailView.Buffer())
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101", "acme/widgets#102"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101", "acme/widgets#102"}, loader.detailCalls)
	}
}

func TestLayout_GivenPullRequestCommits_WhenRendering_ThenBrowserModeShowsFourDetailTabsAndRendersTheCommitTab(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/commits",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Comment body",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
				Commits: []githubcli.PullRequestCommit{{
					OID:             "e9a3253762e768badaa1d4a5b3d267416d1e42f4",
					MessageHeadline: "reintroduce interactive gh pr",
					MessageBody:     "this commit adds gh pr back",
					AuthoredDate:    "2019-10-04T15:23:39Z",
					CommittedDate:   "2019-10-04T15:57:48Z",
					Authors: []githubcli.PullRequestCommitAuthor{{
						Name:  "nate smith",
						Login: "vilmibm",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":                     "Rendered body 42",
		"this commit adds gh pr back": "Rendered commit body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (1)", CommitsDetailTab.Label(), ChangesDetailTab.Label()}, 0)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommitsDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", CommitsDetailTab, subject.activeDetailTab)
	}
	for _, expected := range []string{"e9a3253", "reintroduce interactive gh pr", "Authors: nate smith", "Authored 2019-10-04 15:23 UTC", "Committed 2019-10-04 15:57 UTC", "Rendered commit body"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the commits tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}

	actualErr = subject.previousDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected active detail tab %v after going backward, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
}

func TestBrowserMode_GivenTheCursorOnAnOverviewSection_WhenPressingEnter_ThenItTogglesTheSectionVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				BaseRefName:    "main",
				HeadRefName:    "feature/overview",
				State:          "OPEN",
				ReviewRequests: []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)") || !strings.Contains(detailView.Buffer(), "@reviewer-requested") {
		t.Fatalf("expected enter to expand the overview section, actual %q", detailView.Buffer())
	}

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)") {
		t.Fatalf("expected enter to collapse the overview section again, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "@reviewer-requested") {
		t.Fatalf("expected the collapsed overview section to hide its body, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenTheCursorOnAConversation_WhenPressingZA_ThenItTogglesTheConversationVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Comment body",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Comment body": "Rendered comment body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, " Comment")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment") {
		t.Fatalf("expected za to collapse the conversation, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered comment body") {
		t.Fatalf("expected the collapsed conversation to hide its body, actual %q", detailView.Buffer())
	}

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment") || !strings.Contains(detailView.Buffer(), "Rendered comment body") {
		t.Fatalf("expected za to expand the conversation again, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenInlineThreadConversations_WhenRendering_ThenItCollapsesResolvedThreadsKeepsSectionsTightAndShowsTheDiffPreview(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{
					{
						ID:         "thread-1",
						IsResolved: true,
						Path:       "internal/tui/render.go",
						Line:       43,
						DiffSide:   "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-resolved"},
							Body:      "Resolved thread body",
							CreatedAt: "2026-04-18T10:00:00Z",
							DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
						}},
					},
					{
						ID:       "thread-2",
						Path:     "internal/tui/model.go",
						Line:     60,
						DiffSide: "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-active"},
							Body:      "Active thread body",
							CreatedAt: "2026-04-18T10:30:00Z",
							DiffHunk:  "@@ -56,6 +56,7 @@\n one\n two\n three\n four\n-old value\n+new value\n tail",
						}},
					},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Resolved thread body": "Rendered resolved thread body",
		"Active thread body":   "Rendered active thread body",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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
	resolvedHeaderLineIndex := given_viewLineIndexContaining(t, detailView, " internal/tui/render.go:43")
	activeHeaderLineIndex := given_viewLineIndexContaining(t, detailView, " internal/tui/model.go:60")
	if activeHeaderLineIndex != resolvedHeaderLineIndex+3 {
		t.Fatalf("expected the collapsed thread to render as a bordered block before the next header block, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered resolved thread body") {
		t.Fatalf("expected resolved threads to stay folded by default, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered active thread body") {
		t.Fatalf("expected unresolved threads to stay visible, actual %q", detailView.Buffer())
	}
	for _, expected := range []string{"@@ -56,6 +56,7 @@", "56 : 56 │ one", "57 : 57 │ two", "58 : 58 │ three", "59 : 59 │ four", "60 :    │ old value", "   : 60 │ new value"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the browser conversations tab to keep the diff preview line %q, actual %q", expected, detailView.Buffer())
		}
	}
	if strings.Contains(detailView.Buffer(), "61 : 61 │ tail") {
		t.Fatalf("expected the browser conversations tab to crop the diff preview after five lines plus the selected line, actual %q", detailView.Buffer())
	}
	if activeHeaderLineIndex == 0 || !strings.HasPrefix(detailView.BufferLines()[activeHeaderLineIndex-1], "────") {
		t.Fatalf("expected the browser conversations tab to render the same top border as review mode, actual %q", detailView.Buffer())
	}
	metadataLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-active")
	if strings.Contains(detailView.BufferLines()[metadataLineIndex], "Unresolved") {
		t.Fatalf("expected the unresolved state to move off the comment metadata line, actual %q", detailView.BufferLines()[metadataLineIndex])
	}
}

func TestLayout_GivenOverviewHeaderMetadata_WhenRendering_ThenTheOverviewTabShowsColoredChurnAndLifecycleTimestamps(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 118, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body", UpdatedAt: "2026-04-18T12:30:00Z"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#118": {Title: "Styled PR", Number: 118, Body: "Body 118", Author: &githubcli.PullRequestAuthor{Login: "octocat"}, BaseRefName: "main", HeadRefName: "feature-118", State: "OPEN", CreatedAt: "2026-04-18T10:00:00Z", UpdatedAt: "2026-04-18T12:30:00Z", Additions: 12, Deletions: 3, ChangedFiles: 5},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 118": "Rendered body 118"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for _, expected := range []string{"Created by", "the 2026-04-18 10:00 UTC", "(last updated at 2026-04-18 12:30 UTC)", "+12", "-3", "Rendered body 118"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected detail buffer to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	countsLineIndex := given_viewLineIndexContaining(t, detailView, "+12")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, countsLineIndex, "+12", given_themeColorHex(t, theme.DiffAdditionHex), "description header additions")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, countsLineIndex, "-3", given_themeColorHex(t, theme.DiffDeletionHex), "description header deletions")
}

func TestLayout_GivenInlineCommentDiff_WhenRendering_ThenTheCommentsTabUsesTreeSitterSyntaxColorsAndExactChangeBackgrounds(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 109, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#109": {Title: "Styled PR", Number: 109, Body: "Body 109", BaseRefName: "main", HeadRefName: "feature-109", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline diff body", CreatedAt: "2026-04-18T10:00:00Z", Path: "src/main/java/com/acme/VersionParser.java", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -43,1 +43,1 @@\n-return Versions.fromString(\"5.0.1\");\n+return Versions.fromString(\"5.1.0\");"}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 109": "Rendered body 109", "Inline diff body": "Rendered inline diff body"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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
	locationLineIndex := given_viewLineIndexContaining(t, detailView, detailInlineCommentLocationIcon+" src/main/java/com/acme/VersionParser.java:43")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "+1", given_themeColorHex(t, theme.DiffAdditionHex), "inline addition count")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "-1", given_themeColorHex(t, theme.DiffDeletionHex), "inline deletion count")

	deletionLineIndex := given_viewLineIndexContaining(t, detailView, `return Versions.fromString("5.0.1");`)
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "return", given_themeColorHex(t, theme.SyntaxKeywordHex), "inline deletion keyword")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "fromString", given_themeColorHex(t, theme.SyntaxFunctionHex), "inline deletion method")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, `"5.0.1"`, given_themeColorHex(t, theme.SyntaxStringHex), "inline deletion string")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, "0.1", given_themeColorHex(t, theme.DiffDeletionHighlightBackgroundHex), "inline deletion changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, `return Versions.fromString("5.`, given_themeColorHex(t, theme.DiffDeletionBackgroundHex), "inline deletion unchanged background")

	additionLineIndex := given_viewLineIndexContaining(t, detailView, `return Versions.fromString("5.1.0");`)
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "return", given_themeColorHex(t, theme.SyntaxKeywordHex), "inline addition keyword")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "fromString", given_themeColorHex(t, theme.SyntaxFunctionHex), "inline addition method")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, `"5.1.0"`, given_themeColorHex(t, theme.SyntaxStringHex), "inline addition string")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, "1.0", given_themeColorHex(t, theme.DiffAdditionHighlightBackgroundHex), "inline addition changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, `return Versions.fromString("5.`, given_themeColorHex(t, theme.DiffAdditionBackgroundHex), "inline addition unchanged background")
}

func TestLayout_GivenInlineComments_WhenRendering_ThenTheCommentsTabUsesAHighlightedAuthorBadgeInsideTheCommentBox(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {Title: "Styled PR", Number: 110, Body: "Body 110", BaseRefName: "main", HeadRefName: "feature-110", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline body", CreatedAt: "2026-04-18T10:00:00Z", Path: "internal/tui/render.go", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\","}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 110": "Rendered body 110", "Inline body": "Rendered inline body"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-inline")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-inline", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "inline comment author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-inline", given_themeColorHex(t, theme.CommentAuthorBadgeHex), "inline comment author badge foreground")
	if !strings.Contains(detailView.BufferLines()[authorLineIndex], "2026-04-18 10:00 UTC") {
		t.Fatalf("expected the inline comment timestamp to stay on the metadata line, actual %q", detailView.BufferLines()[authorLineIndex])
	}
}

func TestLayout_GivenSuggestionFenceInlineComment_WhenRendering_ThenTheCommentsTabFillsTheCommentBoxInteriorWithTheCodeBlockBackground(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 115, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#115": {Title: "Styled PR", Number: 115, Body: "Body 115", BaseRefName: "main", HeadRefName: "feature-115", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "```suggestion\nfmt.Println(\"hello\")\n```", CreatedAt: "2026-04-18T10:00:00Z", Path: "internal/tui/render.go", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -43,1 +43,1 @@\n-fmt.Println(\"goodbye\")\n+fmt.Println(\"hello\")"}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
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
	codeLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, `fmt.Println("hello")`)
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[codeLineIndex-1])); actualInnerText != "" {
		t.Fatalf("expected the suggestion code block top padding line to stay blank inside the comment box, actual %q", actualInnerText)
	}
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[codeLineIndex+1])); actualInnerText != "" {
		t.Fatalf("expected the suggestion code block bottom padding line to stay blank inside the comment box, actual %q", actualInnerText)
	}
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion code block background")
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeLineIndex-1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion code block top padding background")
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeLineIndex+1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion code block bottom padding background")
	then_viewCommentBoxBorderDoesNotHaveBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion code block border background")
}

func TestLayout_GivenMarkdownDescriptionAndComments_WhenRendering_ThenTheDetailPaneShowsAStyledHeadingGreyCommentBorderAndHighlightedCommentAuthorBadge(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {Title: "Styled PR", Number: 110, Body: "## Why\n\nParagraph body", BaseRefName: "main", HeadRefName: "feature-110", State: "OPEN", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Ship it", CreatedAt: "2026-04-18T10:00:00Z"}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headingLineIndex := given_viewLineIndexContaining(t, detailView, "Why")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, headingLineIndex, "Why", given_themeColorHex(t, theme.MarkdownHeadingHex), "markdown heading")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	commentBorderLineIndex := given_viewLineIndexContaining(t, detailView, "╭")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, commentBorderLineIndex, "╭", given_themeColorHex(t, theme.InactiveBorderHex), "comment border")
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "comment author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeHex), "comment author badge foreground")
	if !strings.Contains(detailView.BufferLines()[authorLineIndex], "2026-04-18 10:00 UTC") {
		t.Fatalf("expected the comment timestamp to stay on the metadata line, actual %q", detailView.BufferLines()[authorLineIndex])
	}
}

func TestLayout_GivenMarkdownHeading_WhenRendering_ThenItsBackgroundFillsTheWholeVisibleLine(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 112, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#112": {Title: "Styled PR", Number: 112, Body: "# Ship it", BaseRefName: "main", HeadRefName: "feature-112", State: "OPEN"},
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
	headingLineIndex := given_viewLineIndexContaining(t, detailView, "Ship it")
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, headingLineIndex, given_themeColorHex(t, theme.MarkdownHeadingBackgroundHex), "markdown heading background")
}

func TestLayout_GivenMarkdownCodeBlock_WhenRendering_ThenItsBackgroundFillsTheWholeVisibleLine(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 113, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#113": {Title: "Styled PR", Number: 113, Body: "```go\nfmt.Println(\"hi\")\n```", BaseRefName: "main", HeadRefName: "feature-113", State: "OPEN"},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	codeLineIndex := given_viewLineIndexContaining(t, detailView, `fmt.Println("hi")`)
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block background")
}

func TestLayout_GivenMarkdownCodeBlock_WhenRendering_ThenItAddsBlankBackgroundLinesAboveAndBelowTheCode(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 114, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#114": {Title: "Styled PR", Number: 114, Body: "```go\nfmt.Println(\"hi\")\n```", BaseRefName: "main", HeadRefName: "feature-114", State: "OPEN"},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	codeLineIndex := given_viewLineIndexContaining(t, detailView, `fmt.Println("hi")`)
	if codeLineIndex < 1 || codeLineIndex >= len(detailView.BufferLines())-1 {
		t.Fatalf("expected blank padding lines around the code block, actual lines %q", detailView.BufferLines())
	}
	if strings.Trim(detailView.BufferLines()[codeLineIndex-1], " ⠀") != "" {
		t.Fatalf("expected the line above the code block to stay blank, actual %q", detailView.BufferLines()[codeLineIndex-1])
	}
	if strings.Trim(detailView.BufferLines()[codeLineIndex+1], " ⠀") != "" {
		t.Fatalf("expected the line below the code block to stay blank, actual %q", detailView.BufferLines()[codeLineIndex+1])
	}
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex-1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block top padding background")
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex+1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block bottom padding background")
}

func TestLayout_GivenRenderedMarkdownDescription_WhenRendering_ThenVisibleLinesDoNotLeakRawANSISequences(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 111, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#111": {Title: "Styled PR", Number: 111, Body: "No need to deploy the docker image and publish the doc to S3.\n\nFor now, we are just testing deploying the runtime on GCP, not AWS.", BaseRefName: "main", HeadRefName: "feature-111", State: "OPEN"},
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
	for _, actualLine := range detailView.BufferLines() {
		if strings.Contains(actualLine, "[38;") || strings.Contains(actualLine, "[3;") || strings.Contains(actualLine, "[0m") {
			t.Fatalf("expected the visible detail lines to hide ANSI sequences, actual %q", actualLine)
		}
	}
}

func TestLayout_GivenAnotherSelectedPullRequestAfterScrolling_WhenRendering_ThenTheDetailOriginResets(t *testing.T) {
	longRenderedBody := strings.TrimSpace(strings.Repeat("rendered detail line\n", 80))
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 201, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 201"}),
		myPullRequestRow(githubcli.PullRequest{Title: "Second PR", Number: 202, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 202"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#201": {Title: "First PR", Number: 201, Body: "Body 201", BaseRefName: "main", HeadRefName: "feature-201", State: "OPEN"},
			"acme/widgets#202": {Title: "Second PR", Number: 202, Body: "Body 202", BaseRefName: "main", HeadRefName: "feature-202", State: "OPEN"},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 201": longRenderedBody, "Body 202": longRenderedBody}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for range detailView.InnerHeight() + 4 {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	_, originY := detailView.Origin()
	if originY < 1 {
		t.Fatalf("expected detail origin to move down, actual %d", originY)
	}

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	_, originY = detailView.Origin()
	if originY != 0 {
		t.Fatalf("expected detail origin to reset to 0, actual %d", originY)
	}
	cursorX, cursorY := detailView.Cursor()
	if cursorX != 0 || cursorY != 0 {
		t.Fatalf("expected detail cursor 0,0 after resetting, actual %d,%d", cursorX, cursorY)
	}
}

func TestRefreshViews_GivenInvalidatedPullRequestDetail_WhenGhHasNotReturnedYet_ThenTheDetailPaneShowsALoadingState(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 301, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{}
	asyncRunner := &capturingAsyncRunner{}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#301"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 301, Body: "Cached detail body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	subject.invalidatePullRequestDetail("acme/widgets", 301)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail reload, actual %d", len(asyncRunner.runs))
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail body to show spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), pullRequestDetailLoadingTitle) {
		t.Fatalf("expected detail body to hide %q, actual %q", pullRequestDetailLoadingTitle, detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Running `gh pr view 301 -R acme/widgets --json ...`.") {
		t.Fatalf("expected detail body to keep only the spinner, actual %q", detailView.Buffer())
	}

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	expectedStatus := string(loadingSpinnerFrames[0]) + " Running `gh pr view 301 -R acme/widgets --json ...`."
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expectedStatus {
		t.Fatalf("expected status line %q, actual %q", expectedStatus, actual)
	}
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestReloadActivePullRequestsTab_GivenExistingPullRequests_WhenGhHasNotReturnedYet_ThenTheStatusLineShowsTheLoadingCommand(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequests(MyPullRequestsTab, []Item{{Title: "my-pr-1", Detail: "body-1"}, {Title: "my-pr-2", Detail: "body-2"}})
	loader := &fakePullRequestDetailLoader{}
	asyncRunner := &capturingAsyncRunner{}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	subject.reloadActivePullRequestsTab(gui)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued pull request reload, actual %d", len(asyncRunner.runs))
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "my-pr-1") {
		t.Fatalf("expected the existing pull request list to stay visible, actual %q", pullRequestsView.Buffer())
	}

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	expectedStatus := string(loadingSpinnerFrames[0]) + " " + myPullRequestsLoadingDetail
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expectedStatus {
		t.Fatalf("expected status line %q, actual %q", expectedStatus, actual)
	}
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, a: Action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func given_viewLineIndexContaining(t *testing.T, view *gocui.View, segment string) int {
	t.Helper()

	for lineIndex, line := range view.BufferLines() {
		if strings.Contains(line, segment) {
			return lineIndex
		}
	}

	t.Fatalf("expected a detail line containing %q, actual %q", segment, strings.Join(view.BufferLines(), "\n"))
	return -1
}

type fakePullRequestDetailLoader struct {
	details                   map[string]githubcli.PullRequestDetail
	detailErrors              map[string]error
	detailCalls               []string
	diffs                     map[string]githubcli.PullRequestDiff
	diffErrors                map[string]error
	diffCalls                 []string
	commentCalls              []string
	commentBodies             []string
	commentErr                error
	myPullRequests            []githubcli.PullRequest
	requestedPullRequests     []githubcli.PullRequest
	approveCalls              []string
	approveErr                error
	reviewCommentCalls        []string
	reviewCommentBodies       []string
	reviewCommentErr          error
	requestChangesCalls       []string
	requestChangesBodies      []string
	requestChangesErr         error
	submitReviewIDs           []string
	submitReviewEvents        []githubcli.PullRequestReviewEvent
	submitReviewBodies        []string
	submitReviewErr           error
	reviewThreadReviewIDs     []string
	reviewThreadBodies        []string
	reviewThreadTargets       []githubcli.PullRequestReviewThreadTarget
	reviewThreadErr           error
	updateReviewCommentIDs    []string
	updateReviewCommentBodies []string
	updateReviewCommentErr    error
	deleteReviewCommentIDs    []string
	deleteReviewCommentErr    error
	resolveReviewThreadIDs    []string
	resolveReviewThreadErr    error
	unresolveReviewThreadIDs  []string
	unresolveReviewThreadErr  error
	reviewKeyByPendingID      map[string]string
	openBrowserCalls          []string
	openBrowserErr            error
	editTitleCalls            []string
	editTitleValues           []string
	editTitleErr              error
	editDescriptionCalls      []string
	editDescriptionBodies     []string
	editDescriptionErr        error
	startReviewCalls          []string
	startReviewID             string
	startReviewErr            error
}

func (loader *fakePullRequestDetailLoader) GetConnectedUser() (githubcli.ConnectedUser, error) {
	return githubcli.ConnectedUser{}, nil
}

func (loader *fakePullRequestDetailLoader) ListPullRequests(commandArguments []string) ([]githubcli.PullRequest, error) {
	for _, argument := range commandArguments {
		if argument == "--review-requested" {
			return append([]githubcli.PullRequest(nil), loader.requestedPullRequests...), nil
		}
	}

	return append([]githubcli.PullRequest(nil), loader.myPullRequests...), nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestDetail(repository string, number int) (githubcli.PullRequestDetail, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.detailCalls = append(loader.detailCalls, key)
	if loader.detailErrors != nil {
		if err, ok := loader.detailErrors[key]; ok {
			return githubcli.PullRequestDetail{}, err
		}
	}
	if loader.details != nil {
		if detail, ok := loader.details[key]; ok {
			return detail, nil
		}
	}
	return githubcli.PullRequestDetail{}, nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestDiff(repository string, number int) (githubcli.PullRequestDiff, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.diffCalls = append(loader.diffCalls, key)
	if loader.diffErrors != nil {
		if err, ok := loader.diffErrors[key]; ok {
			return githubcli.PullRequestDiff{}, err
		}
	}
	if loader.diffs != nil {
		if diff, ok := loader.diffs[key]; ok {
			return diff, nil
		}
	}
	return githubcli.PullRequestDiff{}, nil
}

func (loader *fakePullRequestDetailLoader) CommentOnPullRequest(repository string, number int, body string) error {
	loader.commentCalls = append(loader.commentCalls, repository+"#"+strconv.Itoa(number))
	loader.commentBodies = append(loader.commentBodies, body)
	return loader.commentErr
}

func (loader *fakePullRequestDetailLoader) ApprovePullRequest(repository string, number int) error {
	loader.approveCalls = append(loader.approveCalls, repository+"#"+strconv.Itoa(number))
	return loader.approveErr
}

func (loader *fakePullRequestDetailLoader) ReviewPullRequestWithComment(repository string, number int, body string) error {
	loader.reviewCommentCalls = append(loader.reviewCommentCalls, repository+"#"+strconv.Itoa(number))
	loader.reviewCommentBodies = append(loader.reviewCommentBodies, body)
	return loader.reviewCommentErr
}

func (loader *fakePullRequestDetailLoader) RequestChangesOnPullRequest(repository string, number int, body string) error {
	loader.requestChangesCalls = append(loader.requestChangesCalls, repository+"#"+strconv.Itoa(number))
	loader.requestChangesBodies = append(loader.requestChangesBodies, body)
	return loader.requestChangesErr
}

func (loader *fakePullRequestDetailLoader) SubmitPullRequestReview(pullRequestReviewID string, event githubcli.PullRequestReviewEvent, body string) error {
	loader.submitReviewIDs = append(loader.submitReviewIDs, strings.TrimSpace(pullRequestReviewID))
	loader.submitReviewEvents = append(loader.submitReviewEvents, event)
	loader.submitReviewBodies = append(loader.submitReviewBodies, body)
	return loader.submitReviewErr
}

func (loader *fakePullRequestDetailLoader) AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubcli.PullRequestReviewThreadTarget) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	loader.reviewThreadReviewIDs = append(loader.reviewThreadReviewIDs, trimmedReviewID)
	loader.reviewThreadBodies = append(loader.reviewThreadBodies, body)
	loader.reviewThreadTargets = append(loader.reviewThreadTargets, target)
	if loader.reviewThreadErr != nil {
		return loader.reviewThreadErr
	}
	if loader.diffs != nil {
		if key, ok := loader.reviewKeyByPendingID[trimmedReviewID]; ok {
			diff := loader.diffs[key]
			diff.Threads = append(diff.Threads, githubcli.PullRequestReviewThread{
				ID:            "thread-" + strconv.Itoa(len(loader.reviewThreadReviewIDs)),
				Path:          target.Path,
				Line:          target.Line,
				StartLine:     target.StartLine,
				DiffSide:      target.Side,
				StartDiffSide: target.StartSide,
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
					Body:      body,
					CreatedAt: "2026-04-20T12:00:00Z",
				}},
			})
			loader.diffs[key] = diff
		}
	}
	return nil
}

func (loader *fakePullRequestDetailLoader) UpdatePullRequestReviewComment(commentID string, body string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.updateReviewCommentIDs = append(loader.updateReviewCommentIDs, trimmedCommentID)
	loader.updateReviewCommentBodies = append(loader.updateReviewCommentBodies, body)
	if loader.updateReviewCommentErr != nil {
		return loader.updateReviewCommentErr
	}
	loader.updateReviewComment(trimmedCommentID, body)
	return nil
}

func (loader *fakePullRequestDetailLoader) DeletePullRequestReviewComment(commentID string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.deleteReviewCommentIDs = append(loader.deleteReviewCommentIDs, trimmedCommentID)
	if loader.deleteReviewCommentErr != nil {
		return loader.deleteReviewCommentErr
	}
	loader.deleteReviewComment(trimmedCommentID)
	return nil
}

func (loader *fakePullRequestDetailLoader) ResolvePullRequestReviewThread(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.resolveReviewThreadIDs = append(loader.resolveReviewThreadIDs, trimmedThreadID)
	if loader.resolveReviewThreadErr != nil {
		return loader.resolveReviewThreadErr
	}
	loader.updateReviewThreadState(trimmedThreadID, true)
	return nil
}

func (loader *fakePullRequestDetailLoader) UnresolvePullRequestReviewThread(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.unresolveReviewThreadIDs = append(loader.unresolveReviewThreadIDs, trimmedThreadID)
	if loader.unresolveReviewThreadErr != nil {
		return loader.unresolveReviewThreadErr
	}
	loader.updateReviewThreadState(trimmedThreadID, false)
	return nil
}

func (loader *fakePullRequestDetailLoader) OpenPullRequestInBrowser(repository string, number int) error {
	loader.openBrowserCalls = append(loader.openBrowserCalls, repository+"#"+strconv.Itoa(number))
	return loader.openBrowserErr
}

func (loader *fakePullRequestDetailLoader) EditPullRequestTitle(repository string, number int, title string) error {
	loader.editTitleCalls = append(loader.editTitleCalls, repository+"#"+strconv.Itoa(number))
	loader.editTitleValues = append(loader.editTitleValues, title)
	if loader.editTitleErr != nil {
		return loader.editTitleErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.Title = title
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.Title = title
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) EditPullRequestDescription(repository string, number int, body string) error {
	loader.editDescriptionCalls = append(loader.editDescriptionCalls, repository+"#"+strconv.Itoa(number))
	loader.editDescriptionBodies = append(loader.editDescriptionBodies, body)
	if loader.editDescriptionErr != nil {
		return loader.editDescriptionErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.Body = body
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.Body = body
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) StartPendingPullRequestReview(repository string, number int) (string, error) {
	loader.startReviewCalls = append(loader.startReviewCalls, repository+"#"+strconv.Itoa(number))
	if loader.startReviewErr != nil {
		return "", loader.startReviewErr
	}
	reviewID := strings.TrimSpace(loader.startReviewID)
	if reviewID == "" {
		reviewID = "PRR_pending"
	}
	if loader.reviewKeyByPendingID == nil {
		loader.reviewKeyByPendingID = map[string]string{}
	}
	loader.reviewKeyByPendingID[reviewID] = repository + "#" + strconv.Itoa(number)
	return reviewID, nil
}

func (loader *fakePullRequestDetailLoader) updatePullRequestSummary(repository string, number int, update func(*githubcli.PullRequest)) {
	for index := range loader.myPullRequests {
		if loader.myPullRequests[index].Repository.NameWithOwner == repository && loader.myPullRequests[index].Number == number {
			update(&loader.myPullRequests[index])
		}
	}
	for index := range loader.requestedPullRequests {
		if loader.requestedPullRequests[index].Repository.NameWithOwner == repository && loader.requestedPullRequests[index].Number == number {
			update(&loader.requestedPullRequests[index])
		}
	}
}

func (loader *fakePullRequestDetailLoader) updatePullRequestDetail(repository string, number int, update func(*githubcli.PullRequestDetail)) {
	if loader.details == nil {
		return
	}
	key := repository + "#" + strconv.Itoa(number)
	detail, ok := loader.details[key]
	if !ok {
		return
	}
	update(&detail)
	loader.details[key] = detail
}

func (loader *fakePullRequestDetailLoader) updateReviewComment(commentID string, body string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		for threadIndex := range detail.InlineCommentThreads {
			for commentIndex := range detail.InlineCommentThreads[threadIndex].Comments {
				if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
					continue
				}
				detail.InlineCommentThreads[threadIndex].Comments[commentIndex].Body = body
				updated = true
			}
		}
		for commentIndex := range detail.InlineComments {
			if strings.TrimSpace(detail.InlineComments[commentIndex].ID) != trimmedCommentID {
				continue
			}
			detail.InlineComments[commentIndex].Body = body
			updated = true
		}
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		for threadIndex := range diff.Threads {
			for commentIndex := range diff.Threads[threadIndex].Comments {
				if strings.TrimSpace(diff.Threads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
					continue
				}
				diff.Threads[threadIndex].Comments[commentIndex].Body = body
				updated = true
			}
		}
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func (loader *fakePullRequestDetailLoader) deleteReviewComment(commentID string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		filteredThreads := detail.InlineCommentThreads[:0]
		for _, thread := range detail.InlineCommentThreads {
			comments := thread.Comments[:0]
			for _, comment := range thread.Comments {
				if strings.TrimSpace(comment.ID) == trimmedCommentID {
					updated = true
					continue
				}
				comments = append(comments, comment)
			}
			thread.Comments = comments
			if len(thread.Comments) > 0 {
				filteredThreads = append(filteredThreads, thread)
			} else if updated {
				continue
			}
		}
		detail.InlineCommentThreads = filteredThreads
		comments := detail.InlineComments[:0]
		for _, comment := range detail.InlineComments {
			if strings.TrimSpace(comment.ID) == trimmedCommentID {
				updated = true
				continue
			}
			comments = append(comments, comment)
		}
		detail.InlineComments = comments
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		filteredThreads := diff.Threads[:0]
		for _, thread := range diff.Threads {
			comments := thread.Comments[:0]
			for _, comment := range thread.Comments {
				if strings.TrimSpace(comment.ID) == trimmedCommentID {
					updated = true
					continue
				}
				comments = append(comments, comment)
			}
			thread.Comments = comments
			if len(thread.Comments) > 0 {
				filteredThreads = append(filteredThreads, thread)
			} else if updated {
				continue
			}
		}
		diff.Threads = filteredThreads
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func (loader *fakePullRequestDetailLoader) updateReviewThreadState(threadID string, resolved bool) {
	if trimmedThreadID := strings.TrimSpace(threadID); trimmedThreadID == "" {
		return
	} else {
		for key, detail := range loader.details {
			updated := false
			for index := range detail.InlineCommentThreads {
				if strings.TrimSpace(detail.InlineCommentThreads[index].ID) != trimmedThreadID {
					continue
				}
				detail.InlineCommentThreads[index].IsResolved = resolved
				updated = true
			}
			if updated {
				loader.details[key] = detail
			}
		}
		for key, diff := range loader.diffs {
			updated := false
			for index := range diff.Threads {
				if strings.TrimSpace(diff.Threads[index].ID) != trimmedThreadID {
					continue
				}
				diff.Threads[index].IsResolved = resolved
				updated = true
			}
			if updated {
				loader.diffs[key] = diff
			}
		}
	}
}

type inlineAsyncRunner struct{}

func (inlineAsyncRunner) Go(run func()) {
	run()
}

type capturingAsyncRunner struct {
	runs []func()
}

func (runner *capturingAsyncRunner) Go(run func()) {
	runner.runs = append(runner.runs, run)
}

type immediateUIUpdater struct{}

func (immediateUIUpdater) Apply(gui *gocui.Gui, update func(*gocui.Gui) error) {
	_ = update(gui)
}
