package tui

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestRenderPullRequestDetailHeader_GivenRichMetadata_WhenFormatting_ThenItShowsTheOverviewReferenceLifecycleAndStatusLines(t *testing.T) {
	summary := githubcli.PullRequest{
		Title:      "Fallback title",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	}
	detail := githubcli.PullRequestDetail{
		Title:     "Add a real detail pane",
		Number:    42,
		Author:    &githubcli.PullRequestAuthor{Login: "octocat"},
		State:     "OPEN",
		CreatedAt: "2026-04-18T10:00:00Z",
		UpdatedAt: "2026-04-18T12:30:00Z",
		Labels:    []githubcli.PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		Assignees: []githubcli.PullRequestAuthor{{Login: "assignee-one"}, {Login: "assignee-two"}},
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
		BaseRefName: "main",
		HeadRefName: "feature/detail",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "Looks good",
			CreatedAt: "2026-04-18T13:00:00Z",
		}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{
			{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
		},
		Additions:    12,
		Deletions:    3,
		ChangedFiles: 5,
	}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	actualText := string(actualDocument.text)

	for _, expected := range []string{
		"acme/widgets#42 Add a real detail pane",
		"Created by",
		"@octocat",
		"the 2026-04-18 10:00 UTC",
		"(last updated at 2026-04-18 12:30 UTC)",
		"Assigned to",
		"@assignee-one",
		"@assignee-two",
		detailStatusIcon + " OPEN",
		"main ← feature/detail",
		detailChecksIcon + " 1 passing, 1 failing",
		"+12",
		"-3",
		detailLabelIcon + " bug",
		detailLabelIcon + " backend",
		detailReviewRequestsIcon + " @reviewer-requested",
		detailReviewRequestsIcon + " @acme/platform",
	} {
		if !strings.Contains(actualText, expected) {
			t.Fatalf("expected header to contain %q, actual %q", expected, actualText)
		}
	}
	for _, unexpected := range []string{"Created:", "Updated:", detailCommentsIcon + " 1 comment"} {
		if strings.Contains(actualText, unexpected) {
			t.Fatalf("expected header to omit %q, actual %q", unexpected, actualText)
		}
	}
	if strings.Contains(actualText, "  ·  ") {
		t.Fatalf("expected header metadata to avoid dot separators, actual %q", actualText)
	}
}

func TestRenderPullRequestDetailHeader_GivenRichMetadata_WhenFormatting_ThenItStylesTheReferenceLifecycleBadgesAndLabels(t *testing.T) {
	summary := githubcli.PullRequest{Title: "Fallback title", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Title:       "Add a real detail pane",
		Number:      42,
		Author:      &githubcli.PullRequestAuthor{Login: "octocat"},
		CreatedAt:   "2026-04-18T10:00:00Z",
		UpdatedAt:   "2026-04-18T12:30:00Z",
		Labels:      []githubcli.PullRequestLabel{{Name: "bug"}},
		Assignees:   []githubcli.PullRequestAuthor{{Login: "assignee-one"}},
		BaseRefName: "main",
		HeadRefName: "feature/detail",
		State:       "OPEN",
	}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	titleLineIndex, titleLine := given_detailDocumentLineContaining(t, actualDocument, "acme/widgets#42")
	referenceIndex := given_runeIndexInString(t, titleLine, "acme/widgets#42")
	titleIndex := given_runeIndexInString(t, titleLine, "Add a real detail pane")
	lifecycleLineIndex, lifecycleLine := given_detailDocumentLineContaining(t, actualDocument, "@octocat")
	createdByIndex := given_runeIndexInString(t, lifecycleLine, "Created by")
	authorIndex := given_runeIndexInString(t, lifecycleLine, "@octocat")
	dateIndex := given_runeIndexInString(t, lifecycleLine, "2026-04-18 10:00 UTC")
	labelsLineIndex, labelsLine := given_detailDocumentLineContaining(t, actualDocument, detailLabelIcon+" bug")
	labelIndex := given_runeIndexInString(t, labelsLine, detailLabelIcon)

	if actualStylePrefix := actualDocument.lineStylePrefixes[titleLineIndex][referenceIndex]; actualStylePrefix != foregroundColorEscape(theme.PullRequestReferenceHex) {
		t.Fatalf("expected reference prefix %q, actual %q", foregroundColorEscape(theme.PullRequestReferenceHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[titleLineIndex][titleIndex]; actualStylePrefix != foregroundColorEscape(theme.PullRequestTitleHex) {
		t.Fatalf("expected title prefix %q, actual %q", foregroundColorEscape(theme.PullRequestTitleHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[lifecycleLineIndex][createdByIndex]; actualStylePrefix != foregroundColorEscape(theme.PullRequestTitleHex) {
		t.Fatalf("expected lifecycle prefix %q, actual %q", foregroundColorEscape(theme.PullRequestTitleHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[lifecycleLineIndex][authorIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.CommentAuthorBadgeHex)) || !strings.Contains(actualStylePrefix, backgroundColorEscape(theme.CommentAuthorBadgeBackgroundHex)) {
		t.Fatalf("expected author badge prefix to contain %q and %q, actual %q", foregroundColorEscape(theme.CommentAuthorBadgeHex), backgroundColorEscape(theme.CommentAuthorBadgeBackgroundHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[lifecycleLineIndex][dateIndex]; actualStylePrefix != foregroundColorEscape(theme.PullRequestReferenceHex) {
		t.Fatalf("expected lifecycle timestamp prefix %q, actual %q", foregroundColorEscape(theme.PullRequestReferenceHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[labelsLineIndex][labelIndex]; actualStylePrefix != foregroundColorEscape(theme.PullRequestReferenceHex) {
		t.Fatalf("expected labels prefix %q, actual %q", foregroundColorEscape(theme.PullRequestReferenceHex), actualStylePrefix)
	}
}

func TestPullRequestDetailCommentCount_GivenInlineCommentThreadsAndRestInlineComments_WhenCounting_ThenItCountsThreadCommentsWithoutDoubleCounting(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		Number: 42,
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "General feedback",
			CreatedAt: "2026-04-18T13:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{Body: "First inline"}, {Body: "Second inline"}},
		InlineCommentThreads: []githubcli.PullRequestReviewThread{{
			ID:         "thread-1",
			IsResolved: true,
			Comments:   []githubcli.PullRequestComment{{Body: "First inline"}, {Body: "Second inline"}},
		}},
	}

	actual := pullRequestDetailCommentCount(detail)

	if actual != 3 {
		t.Fatalf("expected comment count %d, actual %d", 3, actual)
	}
}

func TestRenderPullRequestDetailHeader_GivenChurnCounts_WhenFormatting_ThenItUsesTheDiffPalette(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Number: 42, Additions: 12, Deletions: 3, ChangedFiles: 5}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	lineIndex, line := given_detailDocumentLineContaining(t, actualDocument, "+12")
	additionIndex := given_runeIndexInString(t, line, "+12")
	deletionIndex := given_runeIndexInString(t, line, "-3")

	if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][additionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionHex) {
		t.Fatalf("expected additions prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][deletionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionHex) {
		t.Fatalf("expected deletions prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionHex), actualStylePrefix)
	}
}

func TestRenderPullRequestDetailHeader_GivenSummaryOnlyUpdatedTimestamp_WhenFormatting_ThenItOmitsMissingCreatedTimeAndChurn(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-04-18T12:30:00Z"}
	detail := githubcli.PullRequestDetail{Number: 42}

	actual := renderPullRequestDetailHeader(summary, detail)

	if strings.Contains(actual, "Created:") || strings.Contains(actual, "Created by") {
		t.Fatalf("expected the header to omit the missing created metadata, actual %q", actual)
	}
	if !strings.Contains(actual, "Last updated at 2026-04-18 12:30 UTC") {
		t.Fatalf("expected the header to contain the updated timestamp, actual %q", actual)
	}
	if strings.Contains(actual, "  +0") || strings.Contains(actual, "  -0") {
		t.Fatalf("expected the header to omit placeholder churn counts, actual %q", actual)
	}
}

func TestRenderPullRequestDetailContentWithSeparator_GivenHeaderAndBody_WhenFormatting_ThenItPlacesAHorizontalRuleBetweenMetadataAndContent(t *testing.T) {
	actualDocument := newDetailDocument(renderPullRequestDetailContentWithSeparator("repo\ntitle\nmeta", "Body", 12), 12)
	actual := strings.Join([]string{string(actualDocument.lines[0]), string(actualDocument.lines[1]), string(actualDocument.lines[2]), string(actualDocument.lines[3]), string(actualDocument.lines[4])}, "\n")

	expected := strings.Join([]string{"repo", "title", "meta", "────────────", "Body"}, "\n")
	if actual != expected {
		t.Fatalf("expected detail content %q, actual %q", expected, actual)
	}
}

func TestRenderPullRequestDetailContentWithSeparator_GivenHeaderAndBody_WhenFormatting_ThenItStylesTheSeparatorLikeABorder(t *testing.T) {
	actualDocument := newDetailDocument(renderPullRequestDetailContentWithSeparator("repo", "Body", 12), 12)
	separatorLineIndex, separatorLine := given_detailDocumentLineContaining(t, actualDocument, "────────────")

	if separatorLine != "────────────" {
		t.Fatalf("expected separator line %q, actual %q", "────────────", separatorLine)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[separatorLineIndex][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected separator border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}
}

func TestRenderPullRequestDescription_GivenMarkdownBody_WhenFormatting_ThenItUsesTheMarkdownRendererAndWrapWidth(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered markdown body"}
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Title: "Add a real detail pane", Number: 42, Body: "## Summary\n\n- render markdown"}

	actual := renderPullRequestDescription(summary, detail, renderer, 48)

	if actual != "Rendered markdown body" {
		t.Fatalf("expected rendered description %q, actual %q", "Rendered markdown body", actual)
	}
	if renderer.lastWidth != 48 {
		t.Fatalf("expected width %d, actual %d", 48, renderer.lastWidth)
	}
	if renderer.lastMarkdown != "## Summary\n\n- render markdown" {
		t.Fatalf("expected markdown %q, actual %q", "## Summary\n\n- render markdown", renderer.lastMarkdown)
	}
}

func TestRenderPullRequestCommentsTab_GivenComments_WhenFormatting_ThenItKeepsUsernamesClearlyVisible(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"**Ship it**":   "Rendered comment one",
		"Needs changes": "Rendered comment two",
	}}
	comments := []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}, {Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, CreatedAt: "2026-04-18T14:15:00Z", Body: "Needs changes"}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 60)

	for _, expected := range []string{detailCommentsIcon + " @reviewer-one", "2026-04-18 13:00 UTC", "Rendered comment one", detailCommentsIcon + " @reviewer-two", "2026-04-18 14:15 UTC", "Rendered comment two"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected comments tab to contain %q, actual %q", expected, actual)
		}
	}
}

func TestRenderPullRequestCommitsTab_GivenCommits_WhenFormatting_ThenItShowsTheShortShaHeadlineAuthorsAndTimestamps(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"this commit adds gh pr back": "Rendered commit body"}}
	commits := []githubcli.PullRequestCommit{{
		OID:             "e9a3253762e768badaa1d4a5b3d267416d1e42f4",
		MessageHeadline: "reintroduce interactive gh pr",
		MessageBody:     "this commit adds gh pr back",
		AuthoredDate:    "2019-10-04T15:23:39Z",
		CommittedDate:   "2019-10-04T15:57:48Z",
		Authors: []githubcli.PullRequestCommitAuthor{{
			Name:  "nate smith",
			Login: "vilmibm",
			Email: "vilmibm@github.com",
		}},
	}}

	actual := renderPullRequestCommitsTab(commits, renderer, 72)

	for _, expected := range []string{"e9a3253", "reintroduce interactive gh pr", "Authors: nate smith", "Authored 2019-10-04 15:23 UTC", "Committed 2019-10-04 15:57 UTC", "Rendered commit body"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected commits tab to contain %q, actual %q", expected, actual)
		}
	}
	if renderer.lastWidth != commentBoxInnerWidth(72) {
		t.Fatalf("expected commit render width %d, actual %d", commentBoxInnerWidth(72), renderer.lastWidth)
	}
}

func TestRenderPullRequestCommentsTab_GivenComments_WhenFormatting_ThenItRendersEachCommentInsideAGreyRoundedBoxWithTheAuthorAndDateOnTheSameLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered comment one"}
	comments := []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 60)
	actualDocument := newDetailDocument(actual, 60)

	if actualTopBorder := string(actualDocument.lines[0]); !strings.HasPrefix(actualTopBorder, "╭") || !strings.HasSuffix(actualTopBorder, "╮") {
		t.Fatalf("expected a rounded top border, actual %q", actualTopBorder)
	}
	metadataLine := string(actualDocument.lines[1])
	if !strings.HasPrefix(metadataLine, "│ ") || !strings.HasSuffix(metadataLine, " │") {
		t.Fatalf("expected the metadata line to stay inside the rounded box, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, detailCommentsIcon+" @reviewer-one") {
		t.Fatalf("expected the metadata line to contain the comment author badge, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, "2026-04-18 13:00 UTC") {
		t.Fatalf("expected the metadata line to contain the comment timestamp, actual %q", metadataLine)
	}
	if strings.Contains(metadataLine, "·") {
		t.Fatalf("expected the metadata line to avoid dot separators, actual %q", metadataLine)
	}
	if actualBodyLine := string(actualDocument.lines[2]); !strings.HasPrefix(actualBodyLine, "│ Rendered comment one") {
		t.Fatalf("expected boxed comment body, actual %q", actualBodyLine)
	}
	if actualBottomBorder := string(actualDocument.lines[3]); !strings.HasPrefix(actualBottomBorder, "╰") || !strings.HasSuffix(actualBottomBorder, "╯") {
		t.Fatalf("expected a rounded bottom border, actual %q", actualBottomBorder)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[0][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected the comment border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}
}

func TestRenderPullRequestDetailHeader_GivenReactionGroups_WhenFormatting_ThenItShowsAReactionsLine(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42}
	detail := githubcli.PullRequestDetail{
		Title:          "First PR",
		Number:         42,
		State:          "OPEN",
		ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 2, ViewerHasReacted: true}, {Content: githubcli.ReactionContentEyes, TotalCount: 1}},
	}

	actual := renderPullRequestDetailHeader(summary, detail)
	actualDocument := newDetailDocument(actual, 120)
	_, reactionLine := given_detailDocumentLineContaining(t, actualDocument, "👍")

	if !strings.Contains(reactionLine, "👍 2") {
		t.Fatalf("expected the reactions line to contain %q, actual %q", "👍 2", reactionLine)
	}
	if !strings.Contains(reactionLine, "👀 1") {
		t.Fatalf("expected the reactions line to contain %q, actual %q", "👀 1", reactionLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenCommentReactionGroups_WhenFormatting_ThenItShowsThemOnTheMetadataLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered comment one"}
	comments := []githubcli.PullRequestComment{{
		Author:         &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
		CreatedAt:      "2026-04-18T13:00:00Z",
		Body:           "**Ship it**",
		ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 2}, {Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}},
	}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 80)
	actualDocument := newDetailDocument(actual, 80)
	_, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-one")

	if !strings.Contains(metadataLine, "👍 2") {
		t.Fatalf("expected the metadata line to contain %q, actual %q", "👍 2", metadataLine)
	}
	if !strings.Contains(metadataLine, "❤️ 1") {
		t.Fatalf("expected the metadata line to contain %q, actual %q", "❤️ 1", metadataLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineThreadCommentReactionGroups_WhenFormatting_ThenItShowsThemOnTheMetadataLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineThreads := []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     43,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:         &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			CreatedAt:      "2026-04-18T14:15:00Z",
			Body:           "Needs more spacing",
			ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentRocket, TotalCount: 1}},
			DiffHunk:       "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
		}},
	}}

	actual := renderPullRequestCommentsTab(nil, inlineThreads, nil, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	_, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-inline")

	if !strings.Contains(metadataLine, "🚀 1") {
		t.Fatalf("expected the metadata line to contain %q, actual %q", "🚀 1", metadataLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItShowsAFileIconLineRangeAndColoredChangeCounts(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	locationLineIndex, locationLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:43")

	expectedVisibleLine := detailInlineCommentLocationIcon + " internal/tui/render.go:43  +1  -1"
	if locationLine != expectedVisibleLine {
		t.Fatalf("expected inline comment location %q, actual %q", expectedVisibleLine, locationLine)
	}
	additionIndex := given_runeIndexInString(t, locationLine, "+1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][additionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionHex) {
		t.Fatalf("expected inline addition count prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
	}
	deletionIndex := given_runeIndexInString(t, locationLine, "-1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][deletionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionHex) {
		t.Fatalf("expected inline deletion count prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionHex), actualStylePrefix)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItRendersABatLikeDiffPreview(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)

	_, actualHunkHeader := given_detailDocumentLineContaining(t, actualDocument, "@@ -42,2 +42,2 @@")
	if actualHunkHeader != "@@ -42,2 +42,2 @@" {
		t.Fatalf("expected diff hunk header %q, actual %q", "@@ -42,2 +42,2 @@", actualHunkHeader)
	}
	_, actualContextLine := given_detailDocumentLineContaining(t, actualDocument, "\"deny\": []")
	if actualContextLine != "42 : 42 │ \"deny\": []" {
		t.Fatalf("expected diff context line %q, actual %q", "42 : 42 │ \"deny\": []", actualContextLine)
	}
	_, actualDeletionLine := given_detailDocumentLineContaining(t, actualDocument, "\"opusplan\"")
	if actualDeletionLine != "43 :    │ \"model\": \"opusplan\"," {
		t.Fatalf("expected diff deletion line %q, actual %q", "43 :    │ \"model\": \"opusplan\",", actualDeletionLine)
	}
	_, actualAdditionLine := given_detailDocumentLineContaining(t, actualDocument, "\"opus\"")
	if actualAdditionLine != "   : 43 │ \"model\": \"opus\"," {
		t.Fatalf("expected diff addition line %q, actual %q", "   : 43 │ \"model\": \"opus\",", actualAdditionLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItRendersTheAuthorAndDateOnTheSameLineInsideTheCommentBox(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	topBorderLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "╭")
	metadataLineIndex, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-inline")
	bodyLineIndex, bodyLine := given_detailDocumentLineContaining(t, actualDocument, "Rendered inline comment")

	if metadataLineIndex != topBorderLineIndex+1 {
		t.Fatalf("expected the metadata line to render inside the box immediately after the top border, actual %q", metadataLine)
	}
	if bodyLineIndex != metadataLineIndex+1 {
		t.Fatalf("expected the body line to render after the metadata line, actual %q", bodyLine)
	}
	if !strings.Contains(metadataLine, detailCommentsIcon+" @reviewer-inline") {
		t.Fatalf("expected the metadata line to contain the author badge, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, "2026-04-18 14:15 UTC") {
		t.Fatalf("expected the metadata line to keep the timestamp on the same line, actual %q", metadataLine)
	}
	if strings.Contains(metadataLine, "·") {
		t.Fatalf("expected the inline comment metadata line to avoid dot separators, actual %q", metadataLine)
	}
	if !strings.HasPrefix(metadataLine, "│ ") || !strings.HasSuffix(metadataLine, " │") {
		t.Fatalf("expected the metadata line to render inside the rounded box, actual %q", metadataLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenResolvedInlineCommentThreads_WhenFormatting_ThenItShowsTheResolvedStateAsAPillOnTheHeaderLineAndRendersRepliesTogether(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Needs more spacing":     "Rendered inline comment",
		"Fixed in the next push": "Rendered reply",
	}}
	inlineThreads := []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsResolved: true,
		Path:       "internal/tui/render.go",
		Line:       43,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				CreatedAt: "2026-04-18T14:15:00Z",
				Body:      "Needs more spacing",
				DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				CreatedAt: "2026-04-18T14:45:00Z",
				Body:      "Fixed in the next push",
			},
		},
	}}

	actual := renderPullRequestCommentsTab(nil, inlineThreads, nil, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	headerLineIndex, headerLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:43")
	resolvedIndex := given_runeIndexInString(t, headerLine, "Resolved")

	if !strings.Contains(headerLine, " internal/tui/render.go:43") {
		t.Fatalf("expected the header line to show the location, actual %q", headerLine)
	}
	if strings.Contains(headerLine, "R43") || strings.Contains(headerLine, "L43") {
		t.Fatalf("expected the header line to drop the side anchor, actual %q", headerLine)
	}
	if !strings.Contains(headerLine, "Resolved") {
		t.Fatalf("expected the header line to show the resolved state, actual %q", headerLine)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[headerLineIndex][resolvedIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.DiffAdditionHex)) || !strings.Contains(actualStylePrefix, backgroundColorEscape(theme.DiffAdditionBackgroundHex)) {
		t.Fatalf("expected the resolved pill prefix to contain %q and %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), backgroundColorEscape(theme.DiffAdditionBackgroundHex), actualStylePrefix)
	}
	for _, expected := range []string{"@@ -42,2 +42,2 @@", "42 : 42 │ \"deny\": []", "43 :    │ \"model\": \"opusplan\",", "   : 43 │ \"model\": \"opus\","} {
		if _, actualLine := given_detailDocumentLineContaining(t, actualDocument, expected); actualLine != expected {
			t.Fatalf("expected inline thread rendering to keep the visible diff preview line %q, actual %q", expected, actualLine)
		}
	}
	if _, replyLine := given_detailDocumentLineContaining(t, actualDocument, "Rendered reply"); !strings.Contains(replyLine, "Rendered reply") {
		t.Fatalf("expected the reply to render in the same inline thread section, actual %q", replyLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenPendingOutdatedInlineCommentThreads_WhenFormatting_ThenItShowsThreadStatesOnTheHeaderAndPendingOnTheComment(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineThreads := []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsOutdated: true,
		Path:       "internal/tui/render.go",
		Line:       43,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			CreatedAt: "2026-04-18T14:15:00Z",
			Body:      "Needs more spacing",
			State:     "PENDING",
			DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
		}},
	}}

	actualDocument := newDetailDocument(renderPullRequestCommentsTab(nil, inlineThreads, nil, renderer, 120), 120)
	_, headerLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:43")
	metadataLineIndex, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-inline")

	for _, expected := range []string{"Unresolved", "Outdated"} {
		if !strings.Contains(headerLine, expected) {
			t.Fatalf("expected the header line to contain %q, actual %q", expected, headerLine)
		}
	}
	if strings.Contains(headerLine, "Pending") {
		t.Fatalf("expected the header line to move the pending state onto the comment, actual %q", headerLine)
	}
	if strings.Contains(headerLine, "R43") || strings.Contains(headerLine, "L43") {
		t.Fatalf("expected the header line to drop the side anchor, actual %q", headerLine)
	}
	pendingIndex := given_runeIndexInString(t, metadataLine, "Pending")
	if actualStylePrefix := actualDocument.lineStylePrefixes[metadataLineIndex][pendingIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.PendingHex)) || !strings.Contains(actualStylePrefix, backgroundColorEscape(theme.PendingBackgroundHex)) {
		t.Fatalf("expected the pending comment pill prefix to contain %q and %q, actual %q", foregroundColorEscape(theme.PendingHex), backgroundColorEscape(theme.PendingBackgroundHex), actualStylePrefix)
	}
}

func TestRenderPullRequestCommentsTab_GivenMultiLineInlineCommentThreads_WhenFormatting_ThenItShowsAtLeastFiveDiffLinesWithLeadingContext(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineThreads := []githubcli.PullRequestReviewThread{{
		ID:        "thread-1",
		Path:      "internal/tui/render.go",
		Line:      71,
		StartLine: 70,
		DiffSide:  "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			CreatedAt: "2026-04-18T14:15:00Z",
			Body:      "Needs more spacing",
			DiffHunk:  "@@ -66,6 +66,7 @@\n one\n two\n three\n four\n-five old\n-six old\n+five new\n+six new\n seven",
		}},
	}}

	actual := renderPullRequestCommentsTab(nil, inlineThreads, nil, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)

	for _, expected := range []string{"@@ -66,6 +66,7 @@", "67 : 67 │ two", "68 : 68 │ three", "69 : 69 │ four", "70 :    │ five old", "71 :    │ six old", "   : 70 │ five new", "   : 71 │ six new"} {
		if _, actualLine := given_detailDocumentLineContaining(t, actualDocument, expected); actualLine != expected {
			t.Fatalf("expected the diff preview to contain the visible line %q, actual %q", expected, actualLine)
		}
	}
	if strings.Contains(actual, "66 : 66 │ one") || strings.Contains(actual, "72 : 72 │ seven") {
		t.Fatalf("expected the diff preview to crop to three leading context lines plus the selected lines, actual %q", actual)
	}
}

func TestRenderRoundedCommentBox_GivenMultiLineStyledText_WhenFormatting_ThenItReappliesTheVisibleStyleAtEachContentLineStart(t *testing.T) {
	styledBody := foregroundColorEscape(theme.MarkdownHeadingHex) + "Styled line one\nline two" + ansiReset

	actual := renderRoundedCommentBox(styledBody, 32)
	actualDocument := newDetailDocument(actual, 32)
	if actualStylePrefix := actualDocument.lineStylePrefixes[1][2]; actualStylePrefix != foregroundColorEscape(theme.MarkdownHeadingHex) {
		t.Fatalf("expected the first content line style prefix %q, actual %q", foregroundColorEscape(theme.MarkdownHeadingHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[2][2]; actualStylePrefix != foregroundColorEscape(theme.MarkdownHeadingHex) {
		t.Fatalf("expected the second content line style prefix %q, actual %q", foregroundColorEscape(theme.MarkdownHeadingHex), actualStylePrefix)
	}
}

func TestGlamourMarkdownRenderer_GivenHeadingMarkdown_WhenRendering_ThenItKeepsHeadingStyledAndDoesNotAddDocumentIndent(t *testing.T) {
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("## Why\n\nParagraph body", 40)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 40)
	if actualHeading := string(actualDocument.lines[0]); actualHeading != "## Why" {
		t.Fatalf("expected visible heading %q, actual %q", "## Why", actualHeading)
	}
	if actualParagraph := string(actualDocument.lines[2]); actualParagraph != "Paragraph body" {
		t.Fatalf("expected visible paragraph %q, actual %q", "Paragraph body", actualParagraph)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[0][0]; actualStylePrefix == "" {
		t.Fatal("expected the heading to keep a style prefix")
	}
}

func TestRenderPullRequestDescription_GivenMarkdownRendererFailure_WhenFormatting_ThenItFallsBackToRawMarkdown(t *testing.T) {
	renderer := &fakeMarkdownRenderer{err: errors.New("boom")}
	summary := githubcli.PullRequest{Number: 9, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Title: "Fallback body", Number: 9, Body: "## Summary\n\n- keep the source"}

	actual := renderPullRequestDescription(summary, detail, renderer, 40)

	for _, expected := range []string{"Markdown rendering failed", "## Summary", "- keep the source"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected description to contain %q, actual %q", expected, actual)
		}
	}
}

func TestDetailStatus_GivenDraftMetadata_WhenFormatting_ThenItPrefersDRAFT(t *testing.T) {
	summary := githubcli.PullRequest{State: "OPEN"}
	detail := githubcli.PullRequestDetail{State: "MERGED", IsDraft: true}

	actual := detailStatus(detail, summary)

	if actual != "DRAFT" {
		t.Fatalf("expected status %q, actual %q", "DRAFT", actual)
	}
}

func TestRenderPullRequestDetailHeader_GivenPullRequestStatuses_WhenFormatting_ThenItUsesStateSpecificStatusBadgeBackgrounds(t *testing.T) {
	testCases := []struct {
		name                  string
		summary               githubcli.PullRequest
		detail                githubcli.PullRequestDetail
		expectedStatus        string
		expectedBackgroundHex string
	}{
		{name: "open", summary: githubcli.PullRequest{State: "OPEN"}, detail: githubcli.PullRequestDetail{State: "OPEN"}, expectedStatus: "OPEN", expectedBackgroundHex: theme.PullRequestStatusOpenBackgroundHex},
		{name: "draft", summary: githubcli.PullRequest{State: "OPEN"}, detail: githubcli.PullRequestDetail{State: "OPEN", IsDraft: true}, expectedStatus: "DRAFT", expectedBackgroundHex: theme.PullRequestStatusDraftBackgroundHex},
		{name: "closed", summary: githubcli.PullRequest{State: "CLOSED"}, detail: githubcli.PullRequestDetail{State: "CLOSED"}, expectedStatus: "CLOSED", expectedBackgroundHex: theme.PullRequestStatusClosedBackgroundHex},
		{name: "merged", summary: githubcli.PullRequest{State: "MERGED"}, detail: githubcli.PullRequestDetail{State: "MERGED"}, expectedStatus: "MERGED", expectedBackgroundHex: theme.PullRequestStatusMergedBackgroundHex},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			header := renderPullRequestDetailHeader(githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, State: testCase.summary.State, IsDraft: testCase.summary.IsDraft}, githubcli.PullRequestDetail{Number: 42, State: testCase.detail.State, IsDraft: testCase.detail.IsDraft})
			actualDocument := newDetailDocument(header, 120)
			lineIndex, line := given_detailDocumentLineContaining(t, actualDocument, testCase.expectedStatus)
			statusIndex := given_runeIndexInString(t, line, testCase.expectedStatus)

			if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][statusIndex]; !strings.Contains(actualStylePrefix, backgroundColorEscape(testCase.expectedBackgroundHex)) {
				t.Fatalf("expected status badge prefix to contain background %q, actual %q", backgroundColorEscape(testCase.expectedBackgroundHex), actualStylePrefix)
			}
		})
	}
}

func TestRenderPullRequestDetailHeader_GivenApprovalReviews_WhenFormatting_ThenItShowsTheCurrentApproversWithGreenCheckIcons(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Number: 42,
		Reviews: []githubcli.PullRequestReview{
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, State: "COMMENTED", SubmittedAt: "2026-04-21T11:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, State: "APPROVED", SubmittedAt: "2026-04-21T12:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-three"}, State: "APPROVED", SubmittedAt: "2026-04-21T09:00:00Z"},
		},
	}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	lineIndex, approvalsLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-two")
	firstIconIndex := given_runeIndexInString(t, approvalsLine, detailApprovalIcon+" @reviewer-two")
	secondIconIndex := given_runeIndexInString(t, approvalsLine, detailApprovalIcon+" @reviewer-three")

	for _, expected := range []string{detailApprovalIcon + " @reviewer-two", detailApprovalIcon + " @reviewer-three"} {
		if !strings.Contains(approvalsLine, expected) {
			t.Fatalf("expected approvals line to contain %q, actual %q", expected, approvalsLine)
		}
	}
	if strings.Contains(approvalsLine, "@reviewer-one") {
		t.Fatalf("expected approvals line to hide reviewers whose latest review is not an approval, actual %q", approvalsLine)
	}
	for _, iconIndex := range []int{firstIconIndex, secondIconIndex} {
		if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][iconIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.DiffAdditionHex)) {
			t.Fatalf("expected approval icon prefix to contain the addition foreground %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
		}
	}
}

func TestRenderPullRequestDetailHeader_GivenALongBranchName_WhenFormatting_ThenItKeepsTheFullBranchName(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Number: 42, State: "OPEN", BaseRefName: "main", HeadRefName: "P3C-7048/refactor_shared_modules"}

	actual := renderPullRequestDetailHeader(summary, detail)

	if !strings.Contains(actual, "main ← P3C-7048/refactor_shared_modules") {
		t.Fatalf("expected the header to keep the full branch name, actual %q", actual)
	}
	if strings.Contains(actual, "…") {
		t.Fatalf("expected the header to avoid truncating the branch name, actual %q", actual)
	}
}

type fakeMarkdownRenderer struct {
	output       string
	outputs      map[string]string
	err          error
	lastMarkdown string
	lastWidth    int
	callCount    int
}

func (renderer *fakeMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer.lastMarkdown = markdown
	renderer.lastWidth = width
	renderer.callCount++
	if renderer.err != nil {
		return "", renderer.err
	}
	if renderer.outputs != nil {
		if output, ok := renderer.outputs[markdown]; ok {
			return output, nil
		}
	}
	return renderer.output, nil
}

func given_detailDocumentLineContaining(t *testing.T, document detailDocument, segment string) (int, string) {
	t.Helper()

	for lineIndex, line := range document.lines {
		visibleLine := string(line)
		if strings.Contains(visibleLine, segment) {
			return lineIndex, visibleLine
		}
	}

	t.Fatalf("expected detail document to contain %q, actual %q", segment, document.text)
	return -1, ""
}

func given_runeIndexInString(t *testing.T, text string, segment string) int {
	t.Helper()

	byteIndex := strings.Index(text, segment)
	if byteIndex < 0 {
		t.Fatalf("expected %q to contain %q", text, segment)
	}
	return utf8.RuneCountInString(text[:byteIndex])
}
