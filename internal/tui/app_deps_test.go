package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

var (
	_ SessionQueries         = (*githubcli.SessionService)(nil)
	_ PullRequestListQueries = (*githubcli.PullRequestListService)(nil)
	_ NotificationQueries    = (*githubcli.NotificationService)(nil)
	_ DetailQueries          = (*githubcli.PullRequestDetailService)(nil)
	_ PullRequestMutations   = (*githubcli.PullRequestMutationService)(nil)
	_ ReviewMutations        = (*githubcli.ReviewService)(nil)
	_ NotificationMutations  = (*githubcli.NotificationService)(nil)
	_ ReactionMutations      = (*githubcli.ReactionService)(nil)
	_ BuildQueries           = (*githubcli.BuildService)(nil)
	_ MarkdownHTMLRenderer   = (*githubcli.MarkdownService)(nil)
	_ AuthTokenProvider      = (*githubcli.AuthService)(nil)

	_ SessionQueries         = (*githubcli.Client)(nil)
	_ PullRequestListQueries = (*githubcli.Client)(nil)
	_ NotificationQueries    = (*githubcli.Client)(nil)
	_ DetailQueries          = (*githubcli.Client)(nil)
	_ PullRequestMutations   = (*githubcli.Client)(nil)
	_ ReviewMutations        = (*githubcli.Client)(nil)
	_ NotificationMutations  = (*githubcli.Client)(nil)
	_ ReactionMutations      = (*githubcli.Client)(nil)
	_ BuildQueries           = (*githubcli.Client)(nil)
	_ MarkdownHTMLRenderer   = (*githubcli.Client)(nil)
	_ AuthTokenProvider      = (*githubcli.Client)(nil)
)

func TestNewProgram_GivenAGitHubCliClient_WhenCreating_ThenItWiresTheCapabilityPortsAndDefaultShellDeps(t *testing.T) {
	subject := NewProgram(githubcli.NewClient())

	if !subject.hasSessionQueries() || !subject.hasPullRequestListQueries() || !subject.hasNotificationQueries() || !subject.hasDetailQueries() {
		t.Fatal("expected the program to wire the GitHub capability query ports")
	}
	if !subject.hasPullRequestMutations() || !subject.hasReviewMutations() || !subject.hasNotificationMutations() || !subject.hasReactionMutations() || !subject.hasBuildQueries() {
		t.Fatal("expected the program to wire the GitHub capability mutation ports")
	}
	if !subject.hasMarkdownHTMLRenderer() || !subject.hasAuthTokenProvider() {
		t.Fatal("expected the program to wire the markdown and auth capability ports")
	}
	if subject.clipboardWriter == nil || subject.externalEditor == nil || subject.linkOpener == nil || subject.themePresetStore == nil {
		t.Fatal("expected the default shell dependencies to stay wired")
	}
}

func TestLoadConnectedUser_GivenSessionQueriesOnly_WhenLoading_ThenItUsesTheSessionPort(t *testing.T) {
	loader := &fakeSessionQueries{user: githubcli.ConnectedUser{Login: "octocat"}}
	subject := NewProgramWithModelAndLoader(given_model(), loader)
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	subject.loadConnectedUser(gui)

	if loader.calls != 1 {
		t.Fatalf("expected one session query call, actual %d", loader.calls)
	}
	if subject.connectedUserLogin != "octocat" {
		t.Fatalf("expected connected user login %q, actual %q", "octocat", subject.connectedUserLogin)
	}
}

func TestLoadPullRequests_GivenPullRequestListQueriesOnly_WhenLoading_ThenItUsesTheListPort(t *testing.T) {
	loader := &fakePullRequestListQueries{pullRequests: []githubcli.PullRequest{{Title: "Ship notifications", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", State: "OPEN"}}}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	subject.loadPullRequests(gui, MyPullRequestsTab)

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
	subject := NewProgramWithModelAndLoader(given_model(), renderer)
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	applied := ""

	subject.loadCurrentDetailImageHTML(gui, detailImageHTMLSource{
		key:        "detail#42",
		repository: "acme/widgets",
		markdown:   "![Architecture](./docs/diagram.png)",
		applyRenderedHTML: func(_ *Program, renderedHTML string) {
			applied = renderedHTML
		},
	})

	if renderer.calls != 1 {
		t.Fatalf("expected one markdown HTML render call, actual %d", renderer.calls)
	}
	if applied != "<p>resolved</p>" {
		t.Fatalf("expected rendered HTML %q, actual %q", "<p>resolved</p>", applied)
	}
}

func TestDetailImageAuthToken_GivenAuthTokenProviderOnly_WhenLoading_ThenItCachesTheToken(t *testing.T) {
	provider := &fakeAuthTokenProvider{token: " ghp_secret-token "}
	subject := NewProgramWithModelAndLoader(given_model(), provider)

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
	user  githubcli.ConnectedUser
	calls int
}

func (loader *fakeSessionQueries) GetConnectedUser() (githubcli.ConnectedUser, error) {
	loader.calls++
	return loader.user, nil
}

type fakePullRequestListQueries struct {
	pullRequests []githubcli.PullRequest
	calls        int
}

func (loader *fakePullRequestListQueries) ListPullRequests(_ []string) ([]githubcli.PullRequest, error) {
	loader.calls++
	return append([]githubcli.PullRequest(nil), loader.pullRequests...), nil
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
