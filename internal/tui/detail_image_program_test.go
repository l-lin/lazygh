package tui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestLayout_GivenMarkdownWithoutImages_WhenRendering_ThenItQueuesNoImageWork(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "No diagrams today.", State: "OPEN"}}
	asyncRunner := &recordingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	loader := subject.githubLoader.(*fakePullRequestDetailLoader)
	if len(loader.renderMarkdownHTMLCalls) != 0 {
		t.Fatalf("expected markdown without images to skip GitHub markdown rendering, actual %v", loader.renderMarkdownHTMLCalls)
	}
	if len(asyncRunner.runs) != 0 {
		t.Fatalf("expected markdown without images to queue no async image work, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenAnHTMLImageTag_WhenRendering_ThenItShowsTheFallbackCaptionAndQueuesTheImageLoad(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: `<img src="https://example.com/diagram.png" alt="Architecture">`, State: "OPEN"}}
	asyncRunner := &recordingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.markdownRenderer = glamourMarkdownRenderer{imageStore: subject.detailImageStore, imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}}, terminalCellSize: fixedTerminalCellSize{width: 10, height: 10}}
	subject.detailImageManager = &capturingDetailImageManager{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "[Image: Architecture]") {
		t.Fatalf("expected the detail view to keep the fallback caption for HTML images, actual %q", detailView.Buffer())
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued image load for the HTML image, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenAPublicMarkdownImage_WhenRendering_ThenItShowsFallbackAndSkipsGitHubMarkdownRendering(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "![Architecture](https://example.com/diagram.png)", State: "OPEN"}}
	asyncRunner := &recordingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.markdownRenderer = glamourMarkdownRenderer{imageStore: subject.detailImageStore, imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}}, terminalCellSize: fixedTerminalCellSize{width: 10, height: 10}}
	subject.detailImageManager = &capturingDetailImageManager{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "[Image: Architecture]") {
		t.Fatalf("expected the detail view to keep the fallback caption, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "https://example.com/diagram.png") {
		t.Fatalf("expected the detail view to keep the image URL, actual %q", detailView.Buffer())
	}
	loader := subject.githubLoader.(*fakePullRequestDetailLoader)
	if len(loader.renderMarkdownHTMLCalls) != 0 {
		t.Fatalf("expected public images to skip GitHub markdown rendering, actual %v", loader.renderMarkdownHTMLCalls)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued image load, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenAnImageLoadFailure_WhenAsyncLoading_ThenItKeepsTheFallbackPlaceholderVisible(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "![Architecture](https://example.com/diagram.png)", State: "OPEN"}}
	subject.imageHTTPClient = &http.Client{Transport: &stubImageRoundTripper{statusCode: http.StatusBadGateway, body: []byte("boom")}}
	subject.markdownRenderer = glamourMarkdownRenderer{imageStore: subject.detailImageStore, imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}}, terminalCellSize: fixedTerminalCellSize{width: 10, height: 10}}
	subject.detailImageManager = &capturingDetailImageManager{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "[Image: Architecture]") {
		t.Fatalf("expected the fallback caption to remain visible after the failed image load, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "https://example.com/diagram.png") {
		t.Fatalf("expected the fallback URL to remain visible after the failed image load, actual %q", detailView.Buffer())
	}
	if !subject.detailImageLoadFailed["https://example.com/diagram.png"] {
		t.Fatalf("expected the failed image source to be marked as failed")
	}
}

func TestLayout_GivenAPrivateGitHubMarkdownImage_WhenRendering_ThenItLoadsRenderedHTMLAndDownloadsTheResolvedImageWithAuth(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		renderedMarkdownHTML: map[string]string{
			`acme/widgets|![Architecture](./docs/diagram.png)`: `<p><img src="https://raw.githubusercontent.com/acme/widgets/main/docs/diagram.png" alt="Architecture"></p>`,
		},
		authToken: "ghp_secret-token",
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "![Architecture](./docs/diagram.png)", State: "OPEN"}}
	transport := &stubImageRoundTripper{statusCode: http.StatusOK, body: given_pngImageBytes(t, 40, 20)}
	subject.imageHTTPClient = &http.Client{Transport: transport}
	subject.markdownRenderer = glamourMarkdownRenderer{imageStore: subject.detailImageStore, imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}}, terminalCellSize: fixedTerminalCellSize{width: 10, height: 10}}
	manager := &capturingDetailImageManager{}
	subject.detailImageManager = manager
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	if len(loader.renderMarkdownHTMLCalls) != 1 {
		t.Fatalf("expected one GitHub markdown rendering call, actual %v", loader.renderMarkdownHTMLCalls)
	}
	if loader.authTokenCalls != 1 {
		t.Fatalf("expected one auth token lookup, actual %d", loader.authTokenCalls)
	}
	if len(transport.requests) != 1 {
		t.Fatalf("expected one image download request, actual %d", len(transport.requests))
	}
	if actual := transport.requests[0].Header.Get("Authorization"); actual != "token ghp_secret-token" {
		t.Fatalf("expected Authorization header %q, actual %q", "token ghp_secret-token", actual)
	}
	if _, ok := subject.detailImageStore.ImageBySource("https://raw.githubusercontent.com/acme/widgets/main/docs/diagram.png"); !ok {
		t.Fatal("expected the resolved private image to be stored")
	}
	if len(manager.images) != 1 {
		t.Fatalf("expected one synced image placement after the private image load, actual %d", len(manager.images))
	}
}

type recordingAsyncRunner struct {
	runs []func()
}

func (runner *recordingAsyncRunner) Go(run func()) {
	runner.runs = append(runner.runs, run)
}

type capturingDetailImageManager struct {
	images []detailImagePlacement
}

func (manager *capturingDetailImageManager) Sync(images []detailImagePlacement) {
	manager.images = append([]detailImagePlacement(nil), images...)
}

type stubImageRoundTripper struct {
	statusCode int
	body       []byte
	requests   []*http.Request
}

func (transport *stubImageRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.requests = append(transport.requests, request.Clone(request.Context()))
	statusCode := transport.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(append([]byte(nil), transport.body...))),
		Header:     http.Header{},
		Request:    request,
	}, nil
}

func given_pngImageBytes(t *testing.T, width int, height int) []byte {
	t.Helper()

	actualImage := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			actualImage.Set(x, y, color.RGBA{R: 0x44, G: 0x88, B: 0xCC, A: 0xFF})
		}
	}

	var buffer bytes.Buffer
	actualErr := png.Encode(&buffer, actualImage)
	then_noError(t, actualErr)
	return buffer.Bytes()
}
