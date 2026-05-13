package tui

import (
	"strings"

	"charm.land/glamour/v2"
	glamouransi "charm.land/glamour/v2/ansi"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	defaultDetailWrapWidth      = 80
	minimumMarkdownRenderWidth  = 20
	disabledMarkdownWordWrap    = 0
	markdownRenderFailurePrefix = "Markdown rendering failed. Showing source."
	maximumBranchLabelWidth     = 28
)

type DetailTab int

const (
	DescriptionDetailTab DetailTab = iota
	CommentsDetailTab
	CommitsDetailTab
	ChangesDetailTab
)

type MarkdownRenderer interface {
	Render(markdown string, width int) (string, error)
}

type glamourMarkdownRenderer struct {
	imageStore       detailImageStore
	imageProtocol    detailImageProtocol
	terminalCellSize terminalCellSizeProvider
}

type pullRequestDetailResult struct {
	detail          githubdomain.PullRequestDetail
	err             error
	sourceUpdatedAt string
	needsRefresh    bool
}

func (tab DetailTab) Label() string {
	switch tab {
	case CommentsDetailTab:
		return strings.TrimSpace(detailCommentsIcon) + " Comments"
	case CommitsDetailTab:
		return detailCommitsIcon + " Commits"
	case ChangesDetailTab:
		return detailChangesIcon + " Changes"
	default:
		return detailDescriptionIcon + " Description"
	}
}

func (renderer glamourMarkdownRenderer) Render(markdown string, width int) (string, error) {
	return renderMarkdownWithImageMarkers(markdown, width, renderer.actualImageStore(), renderer.actualImageProtocol(), renderer.actualTerminalCellSize())
}

func renderMarkdownWithGlamour(markdown string) (string, error) {
	return renderMarkdownWithGlamourStyle(markdown, disabledMarkdownWordWrap, prettyMarkdownStyle())
}

func renderMarkdownWithGlamourStyle(markdown string, width int, style glamouransi.StyleConfig) (string, error) {
	registerMarkdownChromaStyle(style)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(markdownWordWrap(width)),
		glamour.WithPreservedNewLines(),
		glamour.WithChromaFormatter("terminal16m"),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(strings.TrimSpace(markdown))
	if err != nil {
		return "", err
	}

	return trimStyledBlankLines(rendered), nil
}

func markdownWordWrap(width int) int {
	if width <= disabledMarkdownWordWrap {
		return disabledMarkdownWordWrap
	}
	return maxInt(width, minimumMarkdownRenderWidth)
}
