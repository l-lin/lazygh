package tui

import (
	"crypto/sha1"
	"fmt"
	urlpkg "net/url"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type detailImageHTMLSource struct {
	key          string
	repository   string
	markdown     string
	renderedHTML string
	applyTarget  detailImageHTMLApplyTarget
}

func pullRequestDescriptionImageHTMLSource(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) detailImageHTMLSource {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	markdown := detailBody(detail, summary)
	revision := detailImageMarkdownRevision(markdown)
	return detailImageHTMLSource{
		key:          detailImageSourceKey(pullRequestKey+":description", markdown),
		repository:   pullRequestRepositoryName(summary.Repository),
		markdown:     markdown,
		renderedHTML: detail.BodyHTML,
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindPullRequestDescription,
			cacheKey:         pullRequestKey,
			markdownRevision: revision,
			fallbackMarkdown: summary.Body,
		},
	}
}

func storyChapterImageHTMLSource(summary githubdomain.PullRequest, chapter reviewStoryChapter) detailImageHTMLSource {
	markdown := strings.TrimSpace(strings.Join(filterEmptyStrings([]string{"# " + strings.TrimSpace(chapter.Title), strings.TrimSpace(chapter.Narrative)}), "\n\n"))
	return detailImageHTMLSource{
		key:          detailImageSourceKey(fmt.Sprintf("story:%s:%s", pullRequestDetailKey(summary.Repository, summary.Number), firstNonEmpty(strings.TrimSpace(chapter.ID), strings.TrimSpace(chapter.Title))), markdown),
		repository:   pullRequestRepositoryName(summary.Repository),
		markdown:     markdown,
		renderedHTML: "",
	}
}

func pullRequestCommentsImageHTMLSources(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) []detailImageHTMLSource {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	repository := pullRequestRepositoryName(summary.Repository)
	sources := make([]detailImageHTMLSource, 0, len(detail.Comments)+len(detail.InlineComments))
	for index, comment := range detail.Comments {
		commentIndex := index
		revision := detailImageMarkdownRevision(comment.Body)
		sources = append(sources, detailImageHTMLSource{
			key:          detailImageSourceKey(fmt.Sprintf("%s:comment:%s:%d", pullRequestKey, strings.TrimSpace(comment.ID), commentIndex), comment.Body),
			repository:   repository,
			markdown:     comment.Body,
			renderedHTML: comment.BodyHTML,
			applyTarget: detailImageHTMLApplyTarget{
				kind:             detailImageHTMLApplyKindPullRequestComment,
				cacheKey:         pullRequestKey,
				markdownRevision: revision,
				itemIndex:        commentIndex,
			},
		})
	}
	if len(detail.InlineCommentThreads) > 0 {
		for threadIndex, thread := range detail.InlineCommentThreads {
			for commentIndex, comment := range thread.Comments {
				threadPosition := threadIndex
				commentPosition := commentIndex
				revision := detailImageMarkdownRevision(comment.Body)
				sources = append(sources, detailImageHTMLSource{
					key:          detailImageSourceKey(fmt.Sprintf("%s:inline-thread:%s:%d:%d", pullRequestKey, strings.TrimSpace(thread.ID), threadPosition, commentPosition), comment.Body),
					repository:   repository,
					markdown:     comment.Body,
					renderedHTML: comment.BodyHTML,
					applyTarget: detailImageHTMLApplyTarget{
						kind:             detailImageHTMLApplyKindPullRequestInlineThreadComment,
						cacheKey:         pullRequestKey,
						markdownRevision: revision,
						threadIndex:      threadPosition,
						commentIndex:     commentPosition,
					},
				})
			}
		}
		return sources
	}
	for index, inlineComment := range detail.InlineComments {
		inlineIndex := index
		revision := detailImageMarkdownRevision(inlineComment.Body)
		sources = append(sources, detailImageHTMLSource{
			key:          detailImageSourceKey(fmt.Sprintf("%s:inline-comment:%s:%d", pullRequestKey, strings.TrimSpace(inlineComment.ID), inlineIndex), inlineComment.Body),
			repository:   repository,
			markdown:     inlineComment.Body,
			renderedHTML: inlineComment.BodyHTML,
			applyTarget: detailImageHTMLApplyTarget{
				kind:             detailImageHTMLApplyKindPullRequestInlineComment,
				cacheKey:         pullRequestKey,
				markdownRevision: revision,
				itemIndex:        inlineIndex,
			},
		})
	}
	return sources
}

func pullRequestCommitImageHTMLSources(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) []detailImageHTMLSource {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	repository := pullRequestRepositoryName(summary.Repository)
	sources := make([]detailImageHTMLSource, 0, len(detail.Commits))
	for index, commit := range detail.Commits {
		commitIndex := index
		revision := detailImageMarkdownRevision(commit.MessageBody)
		sources = append(sources, detailImageHTMLSource{
			key:          detailImageSourceKey(fmt.Sprintf("%s:commit:%s:%d", pullRequestKey, strings.TrimSpace(commit.OID), commitIndex), commit.MessageBody),
			repository:   repository,
			markdown:     commit.MessageBody,
			renderedHTML: commit.MessageBodyHTML,
			applyTarget: detailImageHTMLApplyTarget{
				kind:             detailImageHTMLApplyKindPullRequestCommit,
				cacheKey:         pullRequestKey,
				markdownRevision: revision,
				itemIndex:        commitIndex,
			},
		})
	}
	return sources
}

func pullRequestDiffImageHTMLSources(summary githubdomain.PullRequest, files []reviewDiffFile) []detailImageHTMLSource {
	sources := make([]detailImageHTMLSource, 0)
	for fileIndex, file := range files {
		sources = append(sources, reviewDiffFileImageHTMLSourcesWithIndex(summary, file, fileIndex)...)
	}
	return sources
}

func reviewDiffFileImageHTMLSourcesWithIndex(summary githubdomain.PullRequest, file reviewDiffFile, fileIndex int) []detailImageHTMLSource {
	diffKey := pullRequestDetailKey(summary.Repository, summary.Number)
	repository := pullRequestRepositoryName(summary.Repository)
	sources := make([]detailImageHTMLSource, 0)
	for threadIndex, thread := range file.Threads {
		for commentIndex, comment := range thread.Comments {
			threadPosition := threadIndex
			commentPosition := commentIndex
			revision := detailImageMarkdownRevision(comment.Body)
			sources = append(sources, detailImageHTMLSource{
				key:          detailImageSourceKey(fmt.Sprintf("%s:diff:%s:%s:%d:%d", diffKey, strings.TrimSpace(file.Path), strings.TrimSpace(thread.ID), threadPosition, commentPosition), comment.Body),
				repository:   repository,
				markdown:     comment.Body,
				renderedHTML: comment.BodyHTML,
				applyTarget: detailImageHTMLApplyTarget{
					kind:             detailImageHTMLApplyKindPullRequestDiffThreadComment,
					cacheKey:         diffKey,
					markdownRevision: revision,
					fileIndex:        fileIndex,
					threadIndex:      threadPosition,
					commentIndex:     commentPosition,
				},
			})
		}
	}
	return sources
}

func pullRequestDiffFileIndex(files []reviewDiffFile, filePath string) int {
	for index, file := range files {
		if strings.TrimSpace(file.Path) == strings.TrimSpace(filePath) {
			return index
		}
	}
	return -1
}

func issueImageHTMLSources(repository string, number int, detail githubdomain.IssueDetail) []detailImageHTMLSource {
	key := notificationDetailKey(repository, number)
	markdown := detail.Body
	revision := detailImageMarkdownRevision(markdown)
	return []detailImageHTMLSource{{
		key:          detailImageSourceKey(key+":issue", markdown),
		repository:   repository,
		markdown:     markdown,
		renderedHTML: detail.BodyHTML,
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindIssue,
			cacheKey:         key,
			markdownRevision: revision,
		},
	}}
}

func releaseImageHTMLSources(repository string, id int, detail githubdomain.ReleaseDetail) []detailImageHTMLSource {
	key := notificationDetailKey(repository, id)
	markdown := detail.Body
	revision := detailImageMarkdownRevision(markdown)
	return []detailImageHTMLSource{{
		key:          detailImageSourceKey(key+":release", markdown),
		repository:   repository,
		markdown:     markdown,
		renderedHTML: detail.BodyHTML,
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindRelease,
			cacheKey:         key,
			markdownRevision: revision,
		},
	}}
}

func (source detailImageHTMLSource) canLoadRenderedHTML() bool {
	return source.applyTarget.canApply() && strings.TrimSpace(source.repository) != "" && strings.TrimSpace(source.markdown) != "" && strings.TrimSpace(source.renderedHTML) == "" && needsRenderedMarkdownHTML(source.markdown)
}

func needsRenderedMarkdownHTML(markdown string) bool {
	trimmedMarkdown := strings.TrimSpace(markdown)
	if trimmedMarkdown == "" {
		return false
	}
	foundHTMLImage := false
	for _, fragment := range bareHTMLImagePattern.FindAllString(trimmedMarkdown, -1) {
		foundHTMLImage = true
		_, imageURL, ok := parseHTMLImageTag(fragment)
		if !ok || needsRenderedImageURL(strings.TrimSpace(imageURL)) {
			return true
		}
	}
	for _, occurrence := range collectMarkdownImageOccurrences(trimmedMarkdown) {
		if needsRenderedImageURL(strings.TrimSpace(occurrence.imageURL)) {
			return true
		}
	}
	_ = foundHTMLImage
	return false
}

func needsRenderedImageURL(imageURL string) bool {
	trimmedURL := strings.TrimSpace(imageURL)
	if trimmedURL == "" {
		return true
	}
	parsedURL, err := urlpkg.Parse(trimmedURL)
	if err != nil {
		return true
	}
	if parsedURL.Scheme == "data" || parsedURL.Scheme == "file" {
		return false
	}
	if parsedURL.Scheme == "" {
		return !strings.HasPrefix(trimmedURL, "/")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	return isGitHubImageRequest(parsedURL)
}

func detailImageSourceKey(prefix string, markdown string) string {
	return strings.TrimSpace(prefix) + ":" + detailImageMarkdownRevision(markdown)
}

func detailImageMarkdownRevision(markdown string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(markdown)))
	return fmt.Sprintf("%x", sum)
}
