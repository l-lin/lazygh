package githubcli

import "testing"

func TestRenderMarkdownHTML_GivenRepositoryContext_WhenRendering_ThenItCallsTheMarkdownEndpointWithJSONBody(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("<p><img src=\"https://private-user-images.githubusercontent.com/signed\"></p>\n")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.RenderMarkdownHTML("acme/widgets", `![Architecture](./docs/diagram.png)`)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "markdown", "--method", "POST", "--input", "-"})
	then_stdinIs(t, runner, `{"text":"![Architecture](./docs/diagram.png)","mode":"gfm","context":"acme/widgets"}`)
	if actual != `<p><img src="https://private-user-images.githubusercontent.com/signed"></p>` {
		t.Fatalf("expected rendered HTML %q, actual %q", `<p><img src="https://private-user-images.githubusercontent.com/signed"></p>`, actual)
	}
}

func TestGetAuthToken_GivenGhOutput_WhenFetching_ThenItTrimsTheToken(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("ghp_example-token\n")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetAuthToken()

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"auth", "token"})
	if actual != "ghp_example-token" {
		t.Fatalf("expected auth token %q, actual %q", "ghp_example-token", actual)
	}
}
