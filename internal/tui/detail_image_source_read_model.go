package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type detailImageSourceReadModel struct {
	activeTab               DetailTab
	reviewModeActive        bool
	reviewShowsDescription  bool
	reviewShowsStoryChapter bool
	summary                 githubdomain.PullRequest
	summaryKnown            bool
	detail                  githubdomain.PullRequestDetail
	detailKnown             bool
	diffFiles               []reviewDiffFile
	diffFilesKnown          bool
	reviewStoryChapter      reviewStoryChapter
	reviewStoryChapterKnown bool
	reviewDiffFile          reviewDiffFile
	reviewDiffFileKnown     bool
	reviewDiffFileIndex     int
	issueRepository         string
	issueNumber             int
	issueKnown              bool
	issueDetail             githubdomain.IssueDetail
	issueDetailKnown        bool
	releaseRepository       string
	releaseID               int
	releaseKnown            bool
	releaseDetail           githubdomain.ReleaseDetail
	releaseDetailKnown      bool
}

func (model detailImageSourceReadModel) sources() []detailImageHTMLSource {
	if model.reviewModeActive {
		return model.reviewSources()
	}
	if model.summaryKnown && model.detailKnown {
		switch model.activeTab {
		case CommentsDetailTab:
			return pullRequestCommentsImageHTMLSources(model.summary, model.detail)
		case CommitsDetailTab:
			return pullRequestCommitImageHTMLSources(model.summary, model.detail)
		case ChangesDetailTab:
			if !model.diffFilesKnown {
				return nil
			}
			return pullRequestDiffImageHTMLSources(model.summary, model.diffFiles)
		default:
			return []detailImageHTMLSource{pullRequestDescriptionImageHTMLSource(model.summary, model.detail)}
		}
	}
	if model.issueKnown && model.issueDetailKnown {
		return issueImageHTMLSources(model.issueRepository, model.issueNumber, model.issueDetail)
	}
	if model.releaseKnown && model.releaseDetailKnown {
		return releaseImageHTMLSources(model.releaseRepository, model.releaseID, model.releaseDetail)
	}
	return nil
}

func (model detailImageSourceReadModel) reviewSources() []detailImageHTMLSource {
	if !model.summaryKnown {
		return nil
	}
	if model.reviewShowsDescription {
		if !model.detailKnown {
			return nil
		}
		return []detailImageHTMLSource{pullRequestDescriptionImageHTMLSource(model.summary, model.detail)}
	}
	if model.reviewShowsStoryChapter {
		if !model.reviewStoryChapterKnown {
			return nil
		}
		return []detailImageHTMLSource{storyChapterImageHTMLSource(model.summary, model.reviewStoryChapter)}
	}
	if !model.reviewDiffFileKnown {
		return nil
	}
	return reviewDiffFileImageHTMLSourcesWithIndex(model.summary, model.reviewDiffFile, model.reviewDiffFileIndex)
}
