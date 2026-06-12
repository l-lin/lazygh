package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func clonePullRequestDetail(detail githubdomain.PullRequestDetail) githubdomain.PullRequestDetail {
	detail.Author = clonePullRequestAuthor(detail.Author)
	detail.Labels = append([]githubdomain.PullRequestLabel(nil), detail.Labels...)
	detail.Assignees = clonePullRequestAuthors(detail.Assignees)
	detail.ReviewRequests = append([]githubdomain.PullRequestReviewRequest(nil), detail.ReviewRequests...)
	detail.AutoMergeRequest = clonePullRequestAutoMergeRequest(detail.AutoMergeRequest)
	detail.MergeQueueEntry = clonePullRequestMergeQueueEntry(detail.MergeQueueEntry)
	detail.ReactionGroups = append([]githubdomain.ReactionGroup(nil), detail.ReactionGroups...)
	detail.Comments = clonePullRequestComments(detail.Comments)
	detail.Commits = clonePullRequestCommits(detail.Commits)
	detail.Reviews = clonePullRequestReviews(detail.Reviews)
	detail.InlineComments = clonePullRequestInlineComments(detail.InlineComments)
	detail.InlineCommentThreads = clonePullRequestReviewThreads(detail.InlineCommentThreads)
	detail.StatusCheckRollup = append([]githubdomain.PullRequestStatusCheck(nil), detail.StatusCheckRollup...)
	return detail
}

func clonePullRequestAutoMergeRequest(request *githubdomain.PullRequestAutoMergeRequest) *githubdomain.PullRequestAutoMergeRequest {
	if request == nil {
		return nil
	}
	copy := *request
	return &copy
}

func clonePullRequestMergeQueueEntry(entry *githubdomain.PullRequestMergeQueueEntry) *githubdomain.PullRequestMergeQueueEntry {
	if entry == nil {
		return nil
	}
	copy := *entry
	return &copy
}

func clonePullRequestAuthor(author *githubdomain.PullRequestAuthor) *githubdomain.PullRequestAuthor {
	if author == nil {
		return nil
	}
	copy := *author
	return &copy
}

func clonePullRequestAuthors(authors []githubdomain.PullRequestAuthor) []githubdomain.PullRequestAuthor {
	return append([]githubdomain.PullRequestAuthor(nil), authors...)
}

func clonePullRequestCommentAuthor(author *githubdomain.PullRequestCommentAuthor) *githubdomain.PullRequestCommentAuthor {
	if author == nil {
		return nil
	}
	copy := *author
	return &copy
}

func clonePullRequestComments(comments []githubdomain.PullRequestComment) []githubdomain.PullRequestComment {
	cloned := make([]githubdomain.PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		comment.Author = clonePullRequestCommentAuthor(comment.Author)
		comment.ReactionGroups = append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...)
		cloned = append(cloned, comment)
	}
	return cloned
}

func clonePullRequestCommits(commits []githubdomain.PullRequestCommit) []githubdomain.PullRequestCommit {
	cloned := make([]githubdomain.PullRequestCommit, 0, len(commits))
	for _, commit := range commits {
		commit.Authors = append([]githubdomain.PullRequestCommitAuthor(nil), commit.Authors...)
		cloned = append(cloned, commit)
	}
	return cloned
}

func clonePullRequestReviews(reviews []githubdomain.PullRequestReview) []githubdomain.PullRequestReview {
	cloned := make([]githubdomain.PullRequestReview, 0, len(reviews))
	for _, review := range reviews {
		review.Author = clonePullRequestCommentAuthor(review.Author)
		cloned = append(cloned, review)
	}
	return cloned
}

func clonePullRequestInlineComments(comments []githubdomain.PullRequestInlineComment) []githubdomain.PullRequestInlineComment {
	cloned := make([]githubdomain.PullRequestInlineComment, 0, len(comments))
	for _, comment := range comments {
		comment.Author = clonePullRequestCommentAuthor(comment.Author)
		comment.ReactionGroups = append([]githubdomain.ReactionGroup(nil), comment.ReactionGroups...)
		cloned = append(cloned, comment)
	}
	return cloned
}

func clonePullRequestReviewThreads(threads []githubdomain.PullRequestReviewThread) []githubdomain.PullRequestReviewThread {
	cloned := make([]githubdomain.PullRequestReviewThread, 0, len(threads))
	for _, thread := range threads {
		thread.Comments = clonePullRequestComments(thread.Comments)
		cloned = append(cloned, thread)
	}
	return cloned
}
