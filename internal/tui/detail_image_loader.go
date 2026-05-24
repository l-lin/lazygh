package tui

import (
	"crypto/sha1"
	"fmt"
	urlpkg "net/url"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type detailImageHTMLSource struct {
	key               string
	repository        string
	markdown          string
	renderedHTML      string
	applyRenderedHTML func(*Program, string)
}

func (program *Program) maybeLoadCurrentDetailImageHTML(gui *gocui.Gui) {
	program.executeCmds(gui, program.imageLoadCoordinator.planCurrentDetailImageHTMLLoads(program, gui))
}

func (program *Program) loadCurrentDetailImageHTML(gui *gocui.Gui, source detailImageHTMLSource) {
	renderedHTML, err := program.markdownHTMLRenderer.RenderMarkdownHTML(source.repository, source.markdown)

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.detailImageHTMLLoadInFlight, source.key)
		if err != nil || strings.TrimSpace(renderedHTML) == "" {
			program.detailImageHTMLLoadFailed[source.key] = true
			return nil
		}

		source.applyRenderedHTML(program, renderedHTML)
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		return program.afterStateChange(gui)
	})
}

func (program *Program) maybeLoadCurrentDetailImages(gui *gocui.Gui) {
	program.executeCmds(gui, program.imageLoadCoordinator.planCurrentDetailImageLoads(program, gui))
}

func (program *Program) loadCurrentDetailImage(gui *gocui.Gui, imageURL string) {
	githubToken := ""
	if isGitHubImageSource(imageURL) {
		githubToken = program.detailImageAuthToken()
	}
	loadedImage, err := loadDetailImage(imageURL, program.imageHTTPClient, githubToken)

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.detailImageLoadInFlight, imageURL)
		if err != nil {
			program.detailImageLoadFailed[imageURL] = true
			return nil
		}

		program.detailImageStore.Store(imageURL, loadedImage)
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		return program.afterStateChange(gui)
	})
}

func (program *Program) detailImageAuthToken() string {
	if !program.hasAuthTokenProvider() {
		return ""
	}

	program.detailImageAuthTokenMu.Lock()
	defer program.detailImageAuthTokenMu.Unlock()

	if program.githubAuthTokenLoaded {
		return program.githubAuthToken
	}

	actual, err := program.authTokenProvider.GetAuthToken()
	program.githubAuthTokenLoaded = true
	if err == nil {
		program.githubAuthToken = strings.TrimSpace(actual)
	}
	return program.githubAuthToken
}

func (program *Program) currentDetailImageHTMLSources() []detailImageHTMLSource {
	if program.reviewModeActive() {
		return program.currentReviewSessionImageHTMLSources()
	}
	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
			return program.currentPullRequestImageHTMLSources(summary, result.detail)
		}
	}
	if program.model.currentSideFocus() != FocusNotificationsView {
		return nil
	}
	if notification, ok := program.model.SelectedNotification(); ok {
		if repository, number, ok := notification.IssueIdentity(); ok {
			return program.currentIssueImageHTMLSources(repository, number)
		}
		if repository, id, ok := notification.ReleaseIdentity(); ok {
			return program.currentReleaseImageHTMLSources(repository, id)
		}
	}
	return nil
}

func (program *Program) currentReviewSessionImageHTMLSources() []detailImageHTMLSource {
	if program.reviewSessionShowsDescription() {
		summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
		if !ok {
			return nil
		}
		return []detailImageHTMLSource{program.pullRequestDescriptionImageHTMLSource(summary, detail)}
	}
	if program.reviewSessionShowsStoryChapter() {
		chapter, ok := program.selectedReviewSessionStoryChapter()
		if !ok {
			return nil
		}
		markdown := strings.TrimSpace(strings.Join(filterEmptyStrings([]string{"# " + strings.TrimSpace(chapter.Title), strings.TrimSpace(chapter.Narrative)}), "\n\n"))
		return []detailImageHTMLSource{{
			key:          detailImageSourceKey(fmt.Sprintf("story:%s:%s", pullRequestDetailKey(program.reviewSession.summary.Repository, program.reviewSession.summary.Number), firstNonEmpty(strings.TrimSpace(chapter.ID), strings.TrimSpace(chapter.Title))), markdown),
			repository:   pullRequestRepositoryName(program.reviewSession.summary.Repository),
			markdown:     markdown,
			renderedHTML: "",
		}}
	}

	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return nil
	}
	return program.reviewDiffFileImageHTMLSources(program.reviewSession.summary, selectedFile)
}

func (program *Program) currentPullRequestImageHTMLSources(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) []detailImageHTMLSource {
	switch program.activeDetailTab {
	case CommentsDetailTab:
		return program.pullRequestCommentsImageHTMLSources(summary, detail)
	case CommitsDetailTab:
		return program.pullRequestCommitImageHTMLSources(summary, detail)
	case ChangesDetailTab:
		if result, ok := program.pullRequestDiffForSummary(summary); ok && result.err == nil {
			return program.pullRequestDiffImageHTMLSources(summary, result.data.Files)
		}
		return nil
	default:
		return []detailImageHTMLSource{program.pullRequestDescriptionImageHTMLSource(summary, detail)}
	}
}

func (program *Program) pullRequestDescriptionImageHTMLSource(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) detailImageHTMLSource {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	markdown := detailBody(detail, summary)
	revision := detailImageMarkdownRevision(markdown)
	return detailImageHTMLSource{
		key:          detailImageSourceKey(pullRequestKey+":description", markdown),
		repository:   pullRequestRepositoryName(summary.Repository),
		markdown:     markdown,
		renderedHTML: detail.BodyHTML,
		applyRenderedHTML: func(program *Program, renderedHTML string) {
			cachedResult, ok := program.pullRequestDetailCache[pullRequestKey]
			if !ok || detailImageMarkdownRevision(detailBody(cachedResult.detail, summary)) != revision {
				return
			}
			cachedResult.detail.BodyHTML = strings.TrimSpace(renderedHTML)
			program.pullRequestDetailCache[pullRequestKey] = cachedResult
		},
	}
}

func (program *Program) pullRequestCommentsImageHTMLSources(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) []detailImageHTMLSource {
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
			applyRenderedHTML: func(program *Program, renderedHTML string) {
				cachedResult, ok := program.pullRequestDetailCache[pullRequestKey]
				if !ok || commentIndex >= len(cachedResult.detail.Comments) || detailImageMarkdownRevision(cachedResult.detail.Comments[commentIndex].Body) != revision {
					return
				}
				cachedResult.detail.Comments[commentIndex].BodyHTML = strings.TrimSpace(renderedHTML)
				program.pullRequestDetailCache[pullRequestKey] = cachedResult
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
					applyRenderedHTML: func(program *Program, renderedHTML string) {
						cachedResult, ok := program.pullRequestDetailCache[pullRequestKey]
						if !ok || threadPosition >= len(cachedResult.detail.InlineCommentThreads) || commentPosition >= len(cachedResult.detail.InlineCommentThreads[threadPosition].Comments) || detailImageMarkdownRevision(cachedResult.detail.InlineCommentThreads[threadPosition].Comments[commentPosition].Body) != revision {
							return
						}
						cachedResult.detail.InlineCommentThreads[threadPosition].Comments[commentPosition].BodyHTML = strings.TrimSpace(renderedHTML)
						program.pullRequestDetailCache[pullRequestKey] = cachedResult
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
			applyRenderedHTML: func(program *Program, renderedHTML string) {
				cachedResult, ok := program.pullRequestDetailCache[pullRequestKey]
				if !ok || inlineIndex >= len(cachedResult.detail.InlineComments) || detailImageMarkdownRevision(cachedResult.detail.InlineComments[inlineIndex].Body) != revision {
					return
				}
				cachedResult.detail.InlineComments[inlineIndex].BodyHTML = strings.TrimSpace(renderedHTML)
				program.pullRequestDetailCache[pullRequestKey] = cachedResult
			},
		})
	}
	return sources
}

func (program *Program) pullRequestCommitImageHTMLSources(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) []detailImageHTMLSource {
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
			applyRenderedHTML: func(program *Program, renderedHTML string) {
				cachedResult, ok := program.pullRequestDetailCache[pullRequestKey]
				if !ok || commitIndex >= len(cachedResult.detail.Commits) || detailImageMarkdownRevision(cachedResult.detail.Commits[commitIndex].MessageBody) != revision {
					return
				}
				cachedResult.detail.Commits[commitIndex].MessageBodyHTML = strings.TrimSpace(renderedHTML)
				program.pullRequestDetailCache[pullRequestKey] = cachedResult
			},
		})
	}
	return sources
}

func (program *Program) pullRequestDiffImageHTMLSources(summary any, files []reviewDiffFile) []detailImageHTMLSource {
	sources := make([]detailImageHTMLSource, 0)
	for fileIndex, file := range files {
		sources = append(sources, program.reviewDiffFileImageHTMLSourcesWithIndex(summary, file, fileIndex)...)
	}
	return sources
}

func (program *Program) reviewDiffFileImageHTMLSources(summary any, file reviewDiffFile) []detailImageHTMLSource {
	return program.reviewDiffFileImageHTMLSourcesWithIndex(summary, file, -1)
}

func (program *Program) reviewDiffFileImageHTMLSourcesWithIndex(summary any, file reviewDiffFile, fileIndexHint int) []detailImageHTMLSource {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return nil
	}
	diffKey := pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)
	repository := pullRequestRepositoryName(summaryValue.Repository)
	sources := make([]detailImageHTMLSource, 0)
	for threadIndex, thread := range file.Threads {
		for commentIndex, comment := range thread.Comments {
			resolvedFileIndex := fileIndexHint
			if resolvedFileIndex < 0 {
				resolvedFileIndex = program.pullRequestDiffFileIndex(summary, file.Path)
			}
			threadPosition := threadIndex
			commentPosition := commentIndex
			revision := detailImageMarkdownRevision(comment.Body)
			sources = append(sources, detailImageHTMLSource{
				key:          detailImageSourceKey(fmt.Sprintf("%s:diff:%s:%s:%d:%d", diffKey, strings.TrimSpace(file.Path), strings.TrimSpace(thread.ID), threadPosition, commentPosition), comment.Body),
				repository:   repository,
				markdown:     comment.Body,
				renderedHTML: comment.BodyHTML,
				applyRenderedHTML: func(program *Program, renderedHTML string) {
					cachedResult, ok := program.pullRequestDiffCache[diffKey]
					if !ok || resolvedFileIndex < 0 || resolvedFileIndex >= len(cachedResult.data.Files) || threadPosition >= len(cachedResult.data.Files[resolvedFileIndex].Threads) || commentPosition >= len(cachedResult.data.Files[resolvedFileIndex].Threads[threadPosition].Comments) || detailImageMarkdownRevision(cachedResult.data.Files[resolvedFileIndex].Threads[threadPosition].Comments[commentPosition].Body) != revision {
						return
					}
					cachedResult.data.Files[resolvedFileIndex].Threads[threadPosition].Comments[commentPosition].BodyHTML = strings.TrimSpace(renderedHTML)
					program.pullRequestDiffCache[diffKey] = cachedResult
				},
			})
		}
	}
	return sources
}

func (program *Program) pullRequestDiffFileIndex(summary any, filePath string) int {
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok {
		return -1
	}
	for index, file := range result.data.Files {
		if strings.TrimSpace(file.Path) == strings.TrimSpace(filePath) {
			return index
		}
	}
	return -1
}

func (program *Program) currentIssueImageHTMLSources(repository string, number int) []detailImageHTMLSource {
	key := notificationDetailKey(repository, number)
	result, ok := program.issueDetailCache[key]
	if !ok || result.err != nil {
		return nil
	}
	markdown := result.detail.Body
	revision := detailImageMarkdownRevision(markdown)
	return []detailImageHTMLSource{{
		key:          detailImageSourceKey(key+":issue", markdown),
		repository:   repository,
		markdown:     markdown,
		renderedHTML: result.detail.BodyHTML,
		applyRenderedHTML: func(program *Program, renderedHTML string) {
			cachedResult, ok := program.issueDetailCache[key]
			if !ok || detailImageMarkdownRevision(cachedResult.detail.Body) != revision {
				return
			}
			cachedResult.detail.BodyHTML = strings.TrimSpace(renderedHTML)
			program.issueDetailCache[key] = cachedResult
		},
	}}
}

func (program *Program) currentReleaseImageHTMLSources(repository string, id int) []detailImageHTMLSource {
	key := notificationDetailKey(repository, id)
	result, ok := program.releaseDetailCache[key]
	if !ok || result.err != nil {
		return nil
	}
	markdown := result.detail.Body
	revision := detailImageMarkdownRevision(markdown)
	return []detailImageHTMLSource{{
		key:          detailImageSourceKey(key+":release", markdown),
		repository:   repository,
		markdown:     markdown,
		renderedHTML: result.detail.BodyHTML,
		applyRenderedHTML: func(program *Program, renderedHTML string) {
			cachedResult, ok := program.releaseDetailCache[key]
			if !ok || detailImageMarkdownRevision(cachedResult.detail.Body) != revision {
				return
			}
			cachedResult.detail.BodyHTML = strings.TrimSpace(renderedHTML)
			program.releaseDetailCache[key] = cachedResult
		},
	}}
}

func (source detailImageHTMLSource) canLoadRenderedHTML() bool {
	return source.applyRenderedHTML != nil && strings.TrimSpace(source.repository) != "" && strings.TrimSpace(source.markdown) != "" && strings.TrimSpace(source.renderedHTML) == "" && needsRenderedMarkdownHTML(source.markdown)
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
