package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

var (
	_ SessionQueries         = (*githubcli.SessionAdapter)(nil)
	_ PullRequestListQueries = (*githubcli.PullRequestListAdapter)(nil)
	_ NotificationQueries    = (*githubcli.NotificationAdapter)(nil)
	_ DetailQueries          = (*githubcli.PullRequestDetailAdapter)(nil)
	_ PullRequestMutations   = (*githubcli.PullRequestMutationAdapter)(nil)
	_ ReviewMutations        = (*githubcli.ReviewAdapter)(nil)
	_ NotificationMutations  = (*githubcli.NotificationAdapter)(nil)
	_ ReactionMutations      = (*githubcli.ReactionAdapter)(nil)
	_ BuildQueries           = (*githubcli.BuildAdapter)(nil)
	_ MarkdownHTMLRenderer   = (*githubcli.MarkdownService)(nil)
	_ AuthTokenProvider      = (*githubcli.AuthService)(nil)
)

func TestNewProgram_GivenFocusedGitHubCapabilityAdapters_WhenCreating_ThenItWiresTheCapabilityPortsAndDefaultShellDeps(t *testing.T) {
	notifications := githubcli.NewNotificationAdapterWithRunner(nil)
	subject := NewProgramWithModelAndDeps(given_model(), AppDeps{
		SessionQueries:        githubcli.NewSessionAdapterWithRunner(nil),
		PullRequestList:       githubcli.NewPullRequestListAdapterWithRunner(nil),
		NotificationQueries:   notifications,
		DetailQueries:         githubcli.NewPullRequestDetailAdapterWithRunner(nil),
		PullRequestMutations:  githubcli.NewPullRequestMutationAdapterWithRunner(nil),
		ReviewMutations:       githubcli.NewReviewAdapterWithRunner(nil),
		NotificationMutations: notifications,
		ReactionMutations:     githubcli.NewReactionAdapterWithRunner(nil),
		BuildQueries:          githubcli.NewBuildAdapterWithRunner(nil),
		MarkdownHTMLRenderer:  githubcli.NewMarkdownServiceWithRunner(nil),
		AuthTokenProvider:     githubcli.NewAuthServiceWithRunner(nil),
	})

	if !subject.hasSessionQueries() || !subject.hasPullRequestListQueries() || !subject.hasNotificationQueries() || !subject.hasDetailQueries() {
		t.Fatal("expected the program to wire the GitHub capability query ports")
	}
	if !subject.hasPullRequestMutations() || !subject.hasReviewMutations() || !subject.hasNotificationMutations() || !subject.hasReactionMutations() || !subject.hasBuildQueries() {
		t.Fatal("expected the program to wire the GitHub capability mutation ports")
	}
	if !subject.hasMarkdownHTMLRenderer() || !subject.hasAuthTokenProvider() {
		t.Fatal("expected the program to wire the markdown and auth capability ports")
	}
	if subject.clipboardReader == nil || subject.clipboardWriter == nil || subject.externalEditor == nil || subject.linkOpener == nil || subject.themePresetStore == nil {
		t.Fatal("expected the default shell dependencies to stay wired")
	}
}

func TestLoadConnectedUser_GivenSessionQueriesOnly_WhenLoading_ThenItUsesTheSessionPort(t *testing.T) {
	loader := &fakeSessionQueries{user: githubdomain.ConnectedUser{Login: "octocat"}}
	subject := NewProgramWithModelAndDeps(given_model(), AppDeps{SessionQueries: loader})
	subject.uiUpdater = immediateUIUpdater{}

	subject.loadConnectedUser(nil)

	if loader.calls != 1 {
		t.Fatalf("expected one session query call, actual %d", loader.calls)
	}
	if subject.connectedUserLogin != "octocat" {
		t.Fatalf("expected connected user login %q, actual %q", "octocat", subject.connectedUserLogin)
	}
}

func TestLoadPullRequests_GivenPullRequestListQueriesOnly_WhenLoading_ThenItUsesTheListPort(t *testing.T) {
	loader := &fakePullRequestListQueries{pullRequests: []githubdomain.PullRequestSummary{{Title: "Ship notifications", Number: 42, Repository: githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", State: "OPEN"}}}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModelAndDeps(model, AppDeps{PullRequestList: loader})
	subject.uiUpdater = immediateUIUpdater{}

	subject.loadPullRequests(nil, MyPullRequestsTab)

	if loader.calls != 1 {
		t.Fatalf("expected one pull request list call, actual %d", loader.calls)
	}
	rows := subject.model.PullRequestRows(MyPullRequestsTab)
	if len(rows) == 0 || rows[0].Summary == nil || rows[0].Summary.Title != "Ship notifications" {
		t.Fatalf("expected the active tab rows to come from the list capability, actual %+v", rows)
	}
}

func TestLoadCurrentDetailImageHTML_GivenMarkdownHTMLRendererOnly_WhenLoading_ThenItUsesThatPort(t *testing.T) {
	renderer := &fakeMarkdownHTMLRenderer{renderedHTML: "<p>resolved</p>"}
	subject := NewProgramWithModelAndDeps(given_model(), AppDeps{MarkdownHTMLRenderer: renderer})
	subject.uiUpdater = immediateUIUpdater{}
	subject.startupState.appStarted = true
	subject.issueDetailCache["acme/widgets#42"] = issueDetailResult{detail: githubdomain.IssueDetail{Body: "![Architecture](./docs/diagram.png)"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_markdown := "![Architecture](./docs/diagram.png)"
	subject.loadCurrentDetailImageHTML(gui, detailImageHTMLSource{
		key:          "detail#42",
		repository:   "acme/widgets",
		markdown:     given_markdown,
		renderedHTML: "",
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindIssue,
			cacheKey:         "acme/widgets#42",
			markdownRevision: detailImageMarkdownRevision(given_markdown),
		},
	})

	if renderer.calls != 1 {
		t.Fatalf("expected one markdown HTML render call, actual %d", renderer.calls)
	}
	actual := subject.issueDetailCache["acme/widgets#42"].detail.BodyHTML
	expected := "<p>resolved</p>"
	if actual != expected {
		t.Fatalf("expected rendered HTML %q, actual %q", expected, actual)
	}
}

func TestDetailImageAuthToken_GivenAuthTokenProviderOnly_WhenLoading_ThenItCachesTheToken(t *testing.T) {
	provider := &fakeAuthTokenProvider{token: " ghp_secret-token "}
	subject := NewProgramWithModelAndDeps(given_model(), AppDeps{AuthTokenProvider: provider})

	actualFirst := subject.detailImageAuthToken()
	actualSecond := subject.detailImageAuthToken()

	if provider.calls != 1 {
		t.Fatalf("expected one auth token lookup, actual %d", provider.calls)
	}
	if actualFirst != "ghp_secret-token" || actualSecond != "ghp_secret-token" {
		t.Fatalf("expected cached auth token %q, actual %q and %q", "ghp_secret-token", actualFirst, actualSecond)
	}
}

type fakeSessionQueries struct {
	user  githubdomain.ConnectedUser
	calls int
}

func (loader *fakeSessionQueries) GetConnectedUser() (githubdomain.ConnectedUser, error) {
	loader.calls++
	return loader.user, nil
}

type fakePullRequestListQueries struct {
	pullRequests []githubdomain.PullRequestSummary
	calls        int
}

func (loader *fakePullRequestListQueries) ListPullRequests(_ []string) ([]githubdomain.PullRequestSummary, error) {
	loader.calls++
	return append([]githubdomain.PullRequestSummary(nil), loader.pullRequests...), nil
}

type fakeMarkdownHTMLRenderer struct {
	renderedHTML string
	calls        int
}

func (renderer *fakeMarkdownHTMLRenderer) RenderMarkdownHTML(_ string, _ string) (string, error) {
	renderer.calls++
	return renderer.renderedHTML, nil
}

type fakeAuthTokenProvider struct {
	token string
	calls int
}

func (provider *fakeAuthTokenProvider) GetAuthToken() (string, error) {
	provider.calls++
	return provider.token, nil
}
