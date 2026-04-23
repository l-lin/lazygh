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

func TestLayout_GivenSelectedPullRequestSummary_WhenRendering_ThenItLoadsRichDetailAndShowsDescriptionAndCommentsInSeparateTabs(t *testing.T) {
	firstSummary := githubcli.PullRequest{Title: "First PR", Number: 101, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-1"}
	secondSummary := githubcli.PullRequest{Title: "Second PR", Number: 102, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-2"}
	firstDetail := githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      101,
		Body:        "Body 101",
		BaseRefName: "main",
		HeadRefName: "feature-101",
		State:       "OPEN",
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
	expectedSeparator := renderPullRequestDetailSectionSeparator(detailView.InnerWidth())
	if actualDetailLines := detailView.BufferLines(); len(actualDetailLines) < 5 || actualDetailLines[3] != expectedSeparator {
		t.Fatalf("expected the description tab to keep a separator after metadata, actual %q", strings.Join(actualDetailLines, "\n"))
	}
	if !strings.Contains(detailView.Buffer(), renderPullRequestMetaLine(firstSummary, firstDetail)+"\n"+expectedSeparator+"\nRendered body 101") {
		t.Fatalf("expected the description tab to keep a separator after metadata, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered comment 101") {
		t.Fatalf("expected description tab to hide comments, actual %q", detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label()}, 0)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101"}, loader.detailCalls)
	}

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	expectedSeparator = renderPullRequestDetailSectionSeparator(detailView.InnerWidth())
	if actualDetailLines := detailView.BufferLines(); len(actualDetailLines) < 5 || actualDetailLines[3] != expectedSeparator {
		t.Fatalf("expected the comments tab to keep a separator after metadata, actual %q", strings.Join(actualDetailLines, "\n"))
	}
	if !strings.Contains(detailView.Buffer(), renderPullRequestMetaLine(firstSummary, firstDetail)+"\n"+expectedSeparator+"\n"+detailCommentsIcon+" @reviewer-one") {
		t.Fatalf("expected the comments tab to keep a separator after metadata, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered comment 101") {
		t.Fatalf("expected comments tab to contain %q, actual %q", "Rendered comment 101", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), detailInlineCommentLocationIcon+" internal/tui/render.go:43  +1  -1") {
		t.Fatalf("expected comments tab to contain the inline comment location and diff counts, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "   : 43 │ \"model\": \"opus\",") {
		t.Fatalf("expected comments tab to contain the inline diff line, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered inline 101") {
		t.Fatalf("expected comments tab to contain %q, actual %q", "Rendered inline 101", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered body 101") {
		t.Fatalf("expected comments tab to hide description text, actual %q", detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label()}, 1)

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

func TestLayout_GivenInlineCommentDiff_WhenRendering_ThenTheCommentsTabUsesBatLikeDiffColorsAndColoredChangeCounts(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 109, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#109": {Title: "Styled PR", Number: 109, Body: "Body 109", BaseRefName: "main", HeadRefName: "feature-109", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline diff body", CreatedAt: "2026-04-18T10:00:00Z", Path: "internal/tui/render.go", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\","}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 109": "Rendered body 109", "Inline diff body": "Rendered inline diff body"}}
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
	locationLineIndex := given_viewLineIndexContaining(t, detailView, detailInlineCommentLocationIcon+" internal/tui/render.go:43")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "+1", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "inline addition count")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "-1", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "inline deletion count")

	deletionLineIndex := given_viewLineIndexContaining(t, detailView, "\"model\": \"opusplan\",")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "\"model\": \"opusplan\",", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "inline deletion text")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, "\"model\": \"opusplan\",", given_themeColorHex(t, theme.DiffDeletionBackgroundHex), "inline deletion background")

	additionLineIndex := given_viewLineIndexContaining(t, detailView, "\"model\": \"opus\",")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "\"model\": \"opus\",", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "inline addition text")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, "\"model\": \"opus\",", given_themeColorHex(t, theme.DiffAdditionBackgroundHex), "inline addition background")
}

func TestLayout_GivenMarkdownDescriptionAndComments_WhenRendering_ThenTheDetailPaneShowsAStyledHeadingAndGreyCommentBorder(t *testing.T) {
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
	gui := given_headlessGui(t)
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
	if !strings.Contains(detailView.Buffer(), "Running `gh pr view 301 -R acme/widgets --json ...`.") {
		t.Fatalf("expected detail body to keep the gh command context, actual %q", detailView.Buffer())
	}

	detailFooterView, actualErr := gui.View("detail-footer")
	then_noError(t, actualErr)
	if !strings.Contains(detailFooterView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail footer to show spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailFooterView.Buffer())
	}
	if strings.Contains(detailFooterView.Buffer(), pullRequestDetailLoadingTitle) {
		t.Fatalf("expected detail footer to hide %q, actual %q", pullRequestDetailLoadingTitle, detailFooterView.Buffer())
	}
}

func TestReloadActivePullRequestsTab_GivenExistingPullRequests_WhenGhHasNotReturnedYet_ThenThePaneFooterShowsALoadingState(t *testing.T) {
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

	pullRequestsFooterView, actualErr := gui.View("pull-requests-footer")
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsFooterView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected pull request footer to show spinner %q, actual %q", string(loadingSpinnerFrames[0]), pullRequestsFooterView.Buffer())
	}
	if strings.Contains(pullRequestsFooterView.Buffer(), myPullRequestsLoadingTitle) {
		t.Fatalf("expected pull request footer to hide %q, actual %q", myPullRequestsLoadingTitle, pullRequestsFooterView.Buffer())
	}
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
	details               map[string]githubcli.PullRequestDetail
	detailErrors          map[string]error
	detailCalls           []string
	diffs                 map[string]githubcli.PullRequestDiff
	diffErrors            map[string]error
	diffCalls             []string
	commentCalls          []string
	commentBodies         []string
	commentErr            error
	myPullRequests        []githubcli.PullRequest
	requestedPullRequests []githubcli.PullRequest
	approveCalls          []string
	approveErr            error
	reviewCommentCalls    []string
	reviewCommentBodies   []string
	reviewCommentErr      error
	requestChangesCalls   []string
	requestChangesBodies  []string
	requestChangesErr     error
	submitReviewIDs       []string
	submitReviewEvents    []githubcli.PullRequestReviewEvent
	submitReviewBodies    []string
	submitReviewErr       error
	reviewThreadReviewIDs []string
	reviewThreadBodies    []string
	reviewThreadTargets   []githubcli.PullRequestReviewThreadTarget
	reviewThreadErr       error
	reviewKeyByPendingID  map[string]string
	openBrowserCalls      []string
	openBrowserErr        error
	editTitleCalls        []string
	editTitleValues       []string
	editTitleErr          error
	editDescriptionCalls  []string
	editDescriptionBodies []string
	editDescriptionErr    error
	startReviewCalls      []string
	startReviewID         string
	startReviewErr        error
}

func (loader *fakePullRequestDetailLoader) GetConnectedUser() (githubcli.ConnectedUser, error) {
	return githubcli.ConnectedUser{}, nil
}

func (loader *fakePullRequestDetailLoader) ListMyPullRequests() ([]githubcli.PullRequest, error) {
	return append([]githubcli.PullRequest(nil), loader.myPullRequests...), nil
}

func (loader *fakePullRequestDetailLoader) ListRequestedPullRequests() ([]githubcli.PullRequest, error) {
	return append([]githubcli.PullRequest(nil), loader.requestedPullRequests...), nil
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
