package tui

import (
	"strings"

	"charm.land/glamour/v2"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	defaultDetailWrapWidth      = 80
	minimumMarkdownRenderWidth  = 20
	disabledMarkdownWordWrap    = 0
	markdownRenderFailurePrefix = "Markdown rendering failed. Showing source."
	maximumBranchLabelWidth     = 28

	detailDescriptionIcon           = ""
	detailCommentsIcon              = " "
	detailCommitsIcon               = ""
	detailChangesIcon               = ""
	pullRequestIcon                 = ""
	draftPullRequestIcon            = ""
	detailRepositoryIcon            = pullRequestIcon
	detailAuthorIcon                = ""
	detailAssigneesIcon             = "󰀄"
	detailReviewRequestsIcon        = "󰀆"
	detailLabelIcon                 = "󰓼"
	detailStatusIcon                = pullRequestIcon
	detailChecksIcon                = "󰄬"
	detailApprovalIcon              = ""
	detailInlineCommentLocationIcon = ""
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

type glamourMarkdownRenderer struct{}

type pullRequestDetailResult struct {
	detail          githubcli.PullRequestDetail
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

func (glamourMarkdownRenderer) Render(markdown string, width int) (string, error) {
	markdownStyle := prettyMarkdownStyle()
	registerMarkdownChromaStyle(markdownStyle)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle),
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

	return strings.TrimSpace(rendered), nil
}

func markdownWordWrap(width int) int {
	if width <= disabledMarkdownWordWrap {
		return disabledMarkdownWordWrap
	}
	return maxInt(width, minimumMarkdownRenderWidth)
}
