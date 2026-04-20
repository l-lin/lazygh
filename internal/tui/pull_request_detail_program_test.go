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
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 101, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-1"}),
		myPullRequestRow(githubcli.PullRequest{Title: "Second PR", Number: 102, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-2"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#101": {Title: "First PR", Number: 101, Body: "Body 101", BaseRefName: "main", HeadRefName: "feature-101", State: "OPEN", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Comment 101", CreatedAt: "2026-04-18T10:00:00Z"}}},
			"acme/widgets#102": {Title: "Second PR", Number: 102, Body: "Body 102", BaseRefName: "main", HeadRefName: "feature-102", State: "OPEN", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Comment 102", CreatedAt: "2026-04-18T11:00:00Z"}}},
		},
	}
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 101": "Rendered body 101", "Body 102": "Rendered body 102", "Comment 101": "Rendered comment 101", "Comment 102": "Rendered comment 102"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered body 101") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered body 101", detailView.Buffer())
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
	if !strings.Contains(detailView.Buffer(), "Rendered comment 101") {
		t.Fatalf("expected comments tab to contain %q, actual %q", "Rendered comment 101", detailView.Buffer())
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
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101", "acme/widgets#102"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101", "acme/widgets#102"}, loader.detailCalls)
	}
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
	if !strings.Contains(detailView.Buffer(), "Loading pull request detail...") {
		t.Fatalf("expected detail body to show a loading message, actual %q", detailView.Buffer())
	}

	detailFooterView, actualErr := gui.View("detail-footer")
	then_noError(t, actualErr)
	if !strings.Contains(detailFooterView.Buffer(), "Loading pull request detail...") {
		t.Fatalf("expected detail footer to show a loading message, actual %q", detailFooterView.Buffer())
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
	if !strings.Contains(pullRequestsFooterView.Buffer(), myPullRequestsLoadingTitle) {
		t.Fatalf("expected pull request footer to show %q, actual %q", myPullRequestsLoadingTitle, pullRequestsFooterView.Buffer())
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
	editTitleCalls        []string
	editTitleValues       []string
	editTitleErr          error
	editDescriptionCalls  []string
	editDescriptionBodies []string
	editDescriptionErr    error
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
