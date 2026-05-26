package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestUpdate_GivenMsgCurrentDetailImageHTMLLoadedForPullRequestDescriptionUsingSummaryFallback_WhenApplying_ThenItUpdatesTheCachedHTML(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	given_markdown := "![Architecture](./docs/diagram.png)"
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{Body: "", BodyHTML: ""}}

	Update(subject, MsgCurrentDetailImageHTMLLoaded{Source: detailImageHTMLSource{
		key:          detailImageSourceKey("acme/widgets#42:description", given_markdown),
		repository:   "acme/widgets",
		markdown:     given_markdown,
		renderedHTML: "",
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindPullRequestDescription,
			cacheKey:         "acme/widgets#42",
			markdownRevision: detailImageMarkdownRevision(given_markdown),
			fallbackMarkdown: given_markdown,
		},
	}, RenderedHTML: "<p>resolved</p>"})

	actual := subject.pullRequestDetailCache["acme/widgets#42"].detail.BodyHTML
	expected := "<p>resolved</p>"
	if actual != expected {
		t.Fatalf("expected cached body HTML %q, actual %q", expected, actual)
	}
}

func TestUpdate_GivenMsgCurrentDetailImageHTMLLoadedForPullRequestDiffThreadComment_WhenApplying_ThenItUpdatesTheNestedCachedHTML(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	given_markdown := "![Architecture](./docs/diagram.png)"
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "docs/diagram.md", Threads: []reviewDiffThread{{Comments: []githubdomain.PullRequestComment{{Body: given_markdown}}}}}}}}

	Update(subject, MsgCurrentDetailImageHTMLLoaded{Source: detailImageHTMLSource{
		key:          detailImageSourceKey("acme/widgets#42:diff:docs/diagram.md:thread-1:0:0", given_markdown),
		repository:   "acme/widgets",
		markdown:     given_markdown,
		renderedHTML: "",
		applyTarget: detailImageHTMLApplyTarget{
			kind:             detailImageHTMLApplyKindPullRequestDiffThreadComment,
			cacheKey:         "acme/widgets#42",
			markdownRevision: detailImageMarkdownRevision(given_markdown),
			fileIndex:        0,
			threadIndex:      0,
			commentIndex:     0,
		},
	}, RenderedHTML: "<p>resolved</p>"})

	actual := subject.pullRequestDiffCache["acme/widgets#42"].data.Files[0].Threads[0].Comments[0].BodyHTML
	expected := "<p>resolved</p>"
	if actual != expected {
		t.Fatalf("expected cached diff comment HTML %q, actual %q", expected, actual)
	}
}
