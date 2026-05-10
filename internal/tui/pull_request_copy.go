package tui

import "github.com/l-lin/lazygh/internal/githubcli"

func clonePullRequestDetail(detail githubcli.PullRequestDetail) githubcli.PullRequestDetail {
	detail.Author = clonePullRequestAuthor(detail.Author)
	detail.Labels = append([]githubcli.PullRequestLabel(nil), detail.Labels...)
	detail.Assignees = clonePullRequestAuthors(detail.Assignees)
	detail.ReviewRequests = append([]githubcli.PullRequestReviewRequest(nil), detail.ReviewRequests...)
	detail.ReactionGroups = append([]githubcli.ReactionGroup(nil), detail.ReactionGroups...)
	detail.Comments = clonePullRequestComments(detail.Comments)
	detail.Commits = clonePullRequestCommits(detail.Commits)
	detail.Reviews = clonePullRequestReviews(detail.Reviews)
	detail.InlineComments = clonePullRequestInlineComments(detail.InlineComments)
	detail.InlineCommentThreads = clonePullRequestReviewThreads(detail.InlineCommentThreads)
	detail.StatusCheckRollup = append([]githubcli.PullRequestStatusCheck(nil), detail.StatusCheckRollup...)
	return detail
}

func clonePullRequestAuthor(author *githubcli.PullRequestAuthor) *githubcli.PullRequestAuthor {
	if author == nil {
		return nil
	}
	copy := *author
	return &copy
}

func clonePullRequestAuthors(authors []githubcli.PullRequestAuthor) []githubcli.PullRequestAuthor {
	return append([]githubcli.PullRequestAuthor(nil), authors...)
}

func clonePullRequestCommentAuthor(author *githubcli.PullRequestCommentAuthor) *githubcli.PullRequestCommentAuthor {
	if author == nil {
		return nil
	}
	copy := *author
	return &copy
}

func clonePullRequestComments(comments []githubcli.PullRequestComment) []githubcli.PullRequestComment {
	cloned := make([]githubcli.PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		comment.Author = clonePullRequestCommentAuthor(comment.Author)
		comment.ReactionGroups = append([]githubcli.ReactionGroup(nil), comment.ReactionGroups...)
		cloned = append(cloned, comment)
	}
	return cloned
}

func clonePullRequestCommits(commits []githubcli.PullRequestCommit) []githubcli.PullRequestCommit {
	cloned := make([]githubcli.PullRequestCommit, 0, len(commits))
	for _, commit := range commits {
		commit.Authors = append([]githubcli.PullRequestCommitAuthor(nil), commit.Authors...)
		cloned = append(cloned, commit)
	}
	return cloned
}

func clonePullRequestReviews(reviews []githubcli.PullRequestReview) []githubcli.PullRequestReview {
	cloned := make([]githubcli.PullRequestReview, 0, len(reviews))
	for _, review := range reviews {
		review.Author = clonePullRequestCommentAuthor(review.Author)
		cloned = append(cloned, review)
	}
	return cloned
}

func clonePullRequestInlineComments(comments []githubcli.PullRequestInlineComment) []githubcli.PullRequestInlineComment {
	cloned := make([]githubcli.PullRequestInlineComment, 0, len(comments))
	for _, comment := range comments {
		comment.Author = clonePullRequestCommentAuthor(comment.Author)
		comment.ReactionGroups = append([]githubcli.ReactionGroup(nil), comment.ReactionGroups...)
		cloned = append(cloned, comment)
	}
	return cloned
}

func clonePullRequestReviewThreads(threads []githubcli.PullRequestReviewThread) []githubcli.PullRequestReviewThread {
	cloned := make([]githubcli.PullRequestReviewThread, 0, len(threads))
	for _, thread := range threads {
		thread.Comments = clonePullRequestComments(thread.Comments)
		cloned = append(cloned, thread)
	}
	return cloned
}
