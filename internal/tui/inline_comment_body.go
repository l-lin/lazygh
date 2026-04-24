package tui

func renderInlineCommentBody(markdown string, renderer MarkdownRenderer, width int) string {
	return renderMarkdownWithFallback(markdown, renderer, width, "No comment body.")
}
