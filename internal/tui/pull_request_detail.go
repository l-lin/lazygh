package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	defaultDetailWrapWidth      = 80
	minimumMarkdownRenderWidth  = 20
	markdownRenderFailurePrefix = "Markdown rendering failed. Showing source."
	maximumBranchLabelWidth     = 28

	detailDescriptionIcon           = "󰈙"
	detailCommentsIcon              = " "
	detailRepositoryIcon            = ""
	detailBranchIcon                = ""
	detailStatusIcon                = ""
	detailChecksIcon                = "󰄬"
	detailInlineCommentLocationIcon = "󰈔"
)

type DetailTab int

const (
	DescriptionDetailTab DetailTab = iota
	CommentsDetailTab
)

type MarkdownRenderer interface {
	Render(markdown string, width int) (string, error)
}

type glamourMarkdownRenderer struct{}

type pullRequestDetailResult struct {
	detail githubcli.PullRequestDetail
	err    error
}

func (tab DetailTab) Label() string {
	switch tab {
	case CommentsDetailTab:
		return fmt.Sprintf("%s Comments", detailCommentsIcon)
	default:
		return fmt.Sprintf("%s Description", detailDescriptionIcon)
	}
}

func (glamourMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(prettyMarkdownStyle()),
		glamour.WithWordWrap(effectiveMarkdownWidth(width)),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(strings.TrimSpace(markdown))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(rendered), nil
}
