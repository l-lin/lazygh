package tui

import (
	"path"
	"strings"
)

const (
	iconPullRequest        = ""
	iconPullRequestDraft   = ""
	iconDescription        = ""
	iconComment            = ""
	iconCommit             = ""
	iconChanges            = ""
	iconUser               = ""
	iconReviewRequest      = ""
	iconLabel              = "󰓼"
	iconChecks             = "󰄬"
	iconApproval           = ""
	iconFile               = ""
	iconDirectory          = ""
	iconCommentCount       = ""
	iconTeamOwnership      = "󱄭"
	iconReview             = ""
	iconCopy               = "󰆏"
	iconOpenBrowser        = ""
	iconLink               = ""
	iconRefresh            = ""
	iconReviewApprove      = "󰆀"
	iconReviewComment      = "󰆂"
	iconRequestChanges     = "󰅾"
	iconPullRequestComment = "󰆆"
	iconEdit               = ""
	iconDelete             = ""
	iconBuild              = ""
	iconAddReaction        = "󰞅"
	iconTheme              = "󰸌"
	iconReviewURL          = ""
	iconMetadata           = "󰋽"
	iconChapter            = "󰭤"
	iconStatusSuccess      = ""
	iconStatusFailure      = ""
	iconStatusPending      = "󰦖"
	iconStatusMuted        = "•"
	iconChevronExpanded    = ""
	iconChevronCollapsed   = ""
	iconWarning            = "󰅚"
	iconUnavailable        = "󰌑"
	iconMarkdownLink       = "󰌹"
	iconMarkdownImage      = ""

	iconNotificationPullRequest = iconPullRequest
	iconNotificationIssue       = ""
	iconNotificationRelease     = ""
	iconNotificationUnread      = "●"
	iconNotificationRead        = "○"

	iconFileGo         = ""
	iconFileRuby       = ""
	iconFileMarkdown   = ""
	iconFileMakefile   = ""
	iconFileDocker     = "󰡨"
	iconFileYAML       = ""
	iconFileJSON       = ""
	iconFileTOML       = ""
	iconFileXML        = "󰗀"
	iconFileHTML       = ""
	iconFileCSS        = ""
	iconFileJavaScript = ""
	iconFileTypeScript = ""
	iconFileTSX        = ""
	iconFileShell      = ""
	iconFileSQL        = ""
	iconFileProto      = "󱘖"
	iconFileImage      = ""
	iconFileSVG        = "󰜡"
	iconFileKotlin     = "󱈙"
	iconFileJava       = ""
	iconFileLocation   = iconFile
)

const (
	detailDescriptionIcon           = iconDescription
	detailCommentsIcon              = iconComment + " "
	detailCommitsIcon               = iconCommit
	detailChangesIcon               = iconChanges
	pullRequestIcon                 = iconPullRequest
	draftPullRequestIcon            = iconPullRequestDraft
	detailRepositoryIcon            = iconPullRequest
	detailAuthorIcon                = iconUser
	detailReviewRequestsIcon        = iconReviewRequest
	detailLabelIcon                 = iconLabel
	detailStatusIcon                = iconPullRequest
	detailChecksIcon                = iconChecks
	detailApprovalIcon              = iconApproval
	detailInlineCommentLocationIcon = iconFileLocation

	reviewDiffHeaderPathIcon    = iconFileLocation
	reviewDiffDirectoryIcon     = iconDirectory
	reviewDiffDefaultFileIcon   = iconFile
	reviewDiffTeamOwnershipIcon = iconTeamOwnership

	reviewModeMetadataIcon = iconMetadata
	reviewModeChapterIcon  = iconChapter

	browserDetailExpandedChevron  = iconChevronExpanded
	browserDetailCollapsedChevron = iconChevronCollapsed

	pullRequestOverviewSuccessIcon         = iconStatusSuccess
	pullRequestOverviewFailureIcon         = iconStatusFailure
	pullRequestOverviewPendingIcon         = iconStatusPending
	pullRequestOverviewMutedIcon           = iconStatusMuted
	pullRequestOverviewReRequestReviewIcon = iconReviewRequest

	actionsPopupStartReviewIcon              = iconReview
	actionsPopupCancelPendingReviewIcon      = iconDelete
	actionsPopupYankPullRequestURLIcon       = iconCopy
	actionsPopupOpenPullRequestBrowserIcon   = iconOpenBrowser
	actionsPopupOpenLinkIcon                 = iconLink
	actionsPopupRefreshPullRequestIcon       = iconRefresh
	actionsPopupReviewApproveIcon            = iconReviewApprove
	actionsPopupReviewCommentIcon            = iconReviewComment
	actionsPopupReviewRequestChangesIcon     = iconRequestChanges
	actionsPopupReRequestReviewIcon          = iconReviewRequest
	actionsPopupCommentOnPullRequestIcon     = iconPullRequestComment
	actionsPopupEditPullRequestIcon          = iconEdit
	actionsPopupDeleteInlineCommentIcon      = iconDelete
	actionsPopupMarkNotificationReadIcon     = iconStatusSuccess
	actionsPopupMarkNotificationDoneIcon     = iconDelete
	actionsPopupMarkAllNotificationsReadIcon = iconStatusSuccess
	actionsPopupMarkAllNotificationsDoneIcon = iconDelete
	actionsPopupOpenNotificationBrowserIcon  = iconOpenBrowser
	actionsPopupResolveInlineCommentIcon     = iconChecks
	actionsPopupBuildActionIcon              = iconBuild
	actionsPopupBuildRunIcon                 = iconBuild
	actionsPopupBuildRunLogsIcon             = iconBuild
	actionsPopupAddReactionIcon              = iconAddReaction
	actionsPopupRemoveReactionIcon           = iconDelete
	actionsPopupReviewPullRequestURLIcon     = iconReviewURL
	actionsPopupChangeThemeIcon              = iconTheme
	actionsPopupReviewStoryIcon              = iconReview

	reviewDiffTreeCommentCountIcon = iconCommentCount
)

var iconFileByName = map[string]string{
	"go.mod":     iconFileGo,
	"go.sum":     iconFileGo,
	"Gemfile":    iconFileRuby,
	"Rakefile":   iconFileRuby,
	"README":     iconFileMarkdown,
	"README.md":  iconFileMarkdown,
	"Makefile":   iconFileMakefile,
	"Dockerfile": iconFileDocker,
}

var iconFileByExtension = map[string]string{
	".go":     iconFileGo,
	".rb":     iconFileRuby,
	".md":     iconFileMarkdown,
	".yaml":   iconFileYAML,
	".yml":    iconFileYAML,
	".json":   iconFileJSON,
	".toml":   iconFileTOML,
	".xml":    iconFileXML,
	".html":   iconFileHTML,
	".css":    iconFileCSS,
	".js":     iconFileJavaScript,
	".ts":     iconFileTypeScript,
	".tsx":    iconFileTSX,
	".jsx":    iconFileTSX,
	".sh":     iconFileShell,
	".bash":   iconFileShell,
	".zsh":    iconFileShell,
	".sql":    iconFileSQL,
	".proto":  iconFileProto,
	".png":    iconFileImage,
	".jpg":    iconFileImage,
	".jpeg":   iconFileImage,
	".gif":    iconFileImage,
	".svg":    iconFileSVG,
	".kt":     iconFileKotlin,
	".java":   iconFileJava,
	".docker": iconFileDocker,
}

func iconFileForPath(filePath string) string {
	trimmedPath := strings.TrimSpace(strings.TrimSuffix(filePath, "/"))
	if trimmedPath == "" {
		return iconFile
	}

	baseName := path.Base(trimmedPath)
	if icon, ok := iconFileByName[baseName]; ok {
		return icon
	}
	if icon, ok := iconFileByExtension[strings.ToLower(path.Ext(baseName))]; ok {
		return icon
	}
	return iconFile
}
