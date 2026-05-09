package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdownWithGlamour_GivenAMarkdownImage_WhenRendering_ThenItFallsBackToAltTextAndURL(t *testing.T) {
	actual, actualErr := renderMarkdownWithGlamour("![Architecture](https://example.com/diagram.png)")

	then_noError(t, actualErr)
	if !strings.Contains(actual, "Architecture") {
		t.Fatalf("expected the glamour output to include the image alt text, actual %q", actual)
	}
	if !strings.Contains(actual, "https://example.com/diagram.png") {
		t.Fatalf("expected the glamour output to include the image URL, actual %q", actual)
	}
	if strings.Contains(actual, "\x1b_G") {
		t.Fatalf("expected glamour to avoid emitting kitty graphics escape codes, actual %q", actual)
	}
}

func TestRenderMarkdownWithGlamour_GivenAnHTMLImageTag_WhenRendering_ThenItDropsTheTagEntirely(t *testing.T) {
	actual, actualErr := renderMarkdownWithGlamour("<img src=\"https://example.com/diagram.png\" alt=\"Architecture\">")

	then_noError(t, actualErr)
	if strings.TrimSpace(actual) != "" {
		t.Fatalf("expected glamour to drop the HTML image tag, actual %q", actual)
	}
}
