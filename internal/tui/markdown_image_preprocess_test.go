package tui

import (
	"strings"
	"testing"
)

func TestPrepareMarkdownForImageRendering_GivenAnHTMLImageTag_WhenPreparing_ThenItConvertsItToMarkdownImageSyntax(t *testing.T) {
	actual := prepareMarkdownForImageRendering(`<img src="https://example.com/diagram.png" alt="Architecture">`, "")

	if actual != `![Architecture](https://example.com/diagram.png)` {
		t.Fatalf("expected markdown image syntax %q, actual %q", `![Architecture](https://example.com/diagram.png)`, actual)
	}
}

func TestPrepareMarkdownForImageRendering_GivenAParagraphWrappedHTMLImageTag_WhenPreparing_ThenItDropsTheWrapperAndConvertsItToMarkdownImageSyntax(t *testing.T) {
	actual := prepareMarkdownForImageRendering(`<p align="center">
  <img src="https://example.com/diagram.png" alt="Architecture">
</p>`, "")

	if strings.Contains(actual, "<p") {
		t.Fatalf("expected the wrapper paragraph to be removed, actual %q", actual)
	}
	if actual != `![Architecture](https://example.com/diagram.png)` {
		t.Fatalf("expected markdown image syntax %q, actual %q", `![Architecture](https://example.com/diagram.png)`, actual)
	}
}

func TestPrepareMarkdownForImageRendering_GivenRenderedHTMLImages_WhenPreparing_ThenItRewritesTheSourceURLsInOrder(t *testing.T) {
	markdown := strings.Join([]string{
		`![First](https://github.com/user-attachments/assets/raw-first)`,
		`<img src="https://github.com/user-attachments/assets/raw-second" alt="Second">`,
	}, "\n\n")
	renderedHTML := strings.Join([]string{
		`<p><img src="https://private-user-images.githubusercontent.com/signed-first"></p>`,
		`<p><img src="https://private-user-images.githubusercontent.com/signed-second"></p>`,
	}, "")

	actual := prepareMarkdownForImageRendering(markdown, renderedHTML)

	for _, unexpected := range []string{"raw-first", "raw-second", "<img"} {
		if strings.Contains(actual, unexpected) {
			t.Fatalf("expected prepared markdown to drop %q, actual %q", unexpected, actual)
		}
	}
	for _, expected := range []string{
		`![First](https://private-user-images.githubusercontent.com/signed-first)`,
		`![Second](https://private-user-images.githubusercontent.com/signed-second)`,
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected prepared markdown to contain %q, actual %q", expected, actual)
		}
	}
}
