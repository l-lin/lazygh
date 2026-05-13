package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	transportgithub "github.com/l-lin/lazygh/internal/githubcli"
)

func toDomainConnectedUser(user any) (githubdomain.ConnectedUser, bool) {
	switch actual := user.(type) {
	case githubdomain.ConnectedUser:
		return actual, true
	case *githubdomain.ConnectedUser:
		if actual == nil {
			return githubdomain.ConnectedUser{}, false
		}
		return *actual, true
	case transportgithub.ConnectedUser:
		return transportgithub.ToDomainConnectedUser(actual), true
	case *transportgithub.ConnectedUser:
		if actual == nil {
			return githubdomain.ConnectedUser{}, false
		}
		return transportgithub.ToDomainConnectedUser(*actual), true
	default:
		return githubdomain.ConnectedUser{}, false
	}
}

func toDomainNotification(notification any) (githubdomain.Notification, bool) {
	switch actual := notification.(type) {
	case githubdomain.Notification:
		return actual, true
	case *githubdomain.Notification:
		if actual == nil {
			return githubdomain.Notification{}, false
		}
		return *actual, true
	case transportgithub.Notification:
		return transportgithub.ToDomainNotification(actual), true
	case *transportgithub.Notification:
		if actual == nil {
			return githubdomain.Notification{}, false
		}
		return transportgithub.ToDomainNotification(*actual), true
	default:
		return githubdomain.Notification{}, false
	}
}

func toDomainPullRequestSummary(summary any) (githubdomain.PullRequest, bool) {
	switch actual := summary.(type) {
	case githubdomain.PullRequest:
		return actual, true
	case *githubdomain.PullRequest:
		if actual == nil {
			return githubdomain.PullRequest{}, false
		}
		return *actual, true
	case transportgithub.PullRequest:
		return transportgithub.ToDomainPullRequestSummary(actual), true
	case *transportgithub.PullRequest:
		if actual == nil {
			return githubdomain.PullRequest{}, false
		}
		return transportgithub.ToDomainPullRequestSummary(*actual), true
	default:
		return githubdomain.PullRequest{}, false
	}
}

func toDomainPullRequestDetail(detail any) (githubdomain.PullRequestDetail, bool) {
	switch actual := detail.(type) {
	case githubdomain.PullRequestDetail:
		return actual, true
	case *githubdomain.PullRequestDetail:
		if actual == nil {
			return githubdomain.PullRequestDetail{}, false
		}
		return *actual, true
	case transportgithub.PullRequestDetail:
		return transportgithub.ToDomainPullRequestDetail(actual), true
	case *transportgithub.PullRequestDetail:
		if actual == nil {
			return githubdomain.PullRequestDetail{}, false
		}
		return transportgithub.ToDomainPullRequestDetail(*actual), true
	default:
		return githubdomain.PullRequestDetail{}, false
	}
}

func toDomainRepository(repository any) (githubdomain.Repository, bool) {
	switch actual := repository.(type) {
	case githubdomain.Repository:
		return actual, true
	case *githubdomain.Repository:
		if actual == nil {
			return githubdomain.Repository{}, false
		}
		return *actual, true
	case transportgithub.Repository:
		return transportgithub.ToDomainRepository(actual), true
	case *transportgithub.Repository:
		if actual == nil {
			return githubdomain.Repository{}, false
		}
		return transportgithub.ToDomainRepository(*actual), true
	default:
		return githubdomain.Repository{}, false
	}
}

func toDomainPullRequestInlineComment(comment any) (githubdomain.PullRequestInlineComment, bool) {
	switch actual := comment.(type) {
	case githubdomain.PullRequestInlineComment:
		return actual, true
	case *githubdomain.PullRequestInlineComment:
		if actual == nil {
			return githubdomain.PullRequestInlineComment{}, false
		}
		return *actual, true
	case transportgithub.PullRequestInlineComment:
		return domainPullRequestInlineCommentFromTransport(actual), true
	case *transportgithub.PullRequestInlineComment:
		if actual == nil {
			return githubdomain.PullRequestInlineComment{}, false
		}
		return domainPullRequestInlineCommentFromTransport(*actual), true
	default:
		return githubdomain.PullRequestInlineComment{}, false
	}
}

func domainPullRequestInlineCommentFromTransport(comment transportgithub.PullRequestInlineComment) githubdomain.PullRequestInlineComment {
	actual := githubdomain.PullRequestInlineComment{
		ID:                comment.ID,
		Body:              comment.Body,
		BodyHTML:          comment.BodyHTML,
		CreatedAt:         comment.CreatedAt,
		URL:               comment.URL,
		Path:              comment.Path,
		DiffHunk:          comment.DiffHunk,
		Line:              comment.Line,
		OriginalLine:      comment.OriginalLine,
		StartLine:         comment.StartLine,
		OriginalStartLine: comment.OriginalStartLine,
		Side:              comment.Side,
		StartSide:         comment.StartSide,
		SubjectType:       comment.SubjectType,
		ReactionGroups:    toDomainReactionGroups(comment.ReactionGroups),
	}
	if comment.Author != nil {
		actual.Author = &githubdomain.PullRequestCommentAuthor{Login: comment.Author.Login}
	}
	return actual
}

func toDomainPullRequestDiff(diff any) (githubdomain.PullRequestDiff, bool) {
	switch actual := diff.(type) {
	case githubdomain.PullRequestDiff:
		return actual, true
	case *githubdomain.PullRequestDiff:
		if actual == nil {
			return githubdomain.PullRequestDiff{}, false
		}
		return *actual, true
	case transportgithub.PullRequestDiff:
		return transportgithub.ToDomainPullRequestDiff(actual), true
	case *transportgithub.PullRequestDiff:
		if actual == nil {
			return githubdomain.PullRequestDiff{}, false
		}
		return transportgithub.ToDomainPullRequestDiff(*actual), true
	default:
		return githubdomain.PullRequestDiff{}, false
	}
}

func toDomainPullRequestComment(comment any) (githubdomain.PullRequestComment, bool) {
	switch actual := comment.(type) {
	case githubdomain.PullRequestComment:
		return actual, true
	case *githubdomain.PullRequestComment:
		if actual == nil {
			return githubdomain.PullRequestComment{}, false
		}
		return *actual, true
	case transportgithub.PullRequestComment:
		return domainPullRequestCommentFromTransport(actual), true
	case *transportgithub.PullRequestComment:
		if actual == nil {
			return githubdomain.PullRequestComment{}, false
		}
		return domainPullRequestCommentFromTransport(*actual), true
	default:
		return githubdomain.PullRequestComment{}, false
	}
}

func domainPullRequestCommentFromTransport(comment transportgithub.PullRequestComment) githubdomain.PullRequestComment {
	actual := githubdomain.PullRequestComment{
		ID:              comment.ID,
		Body:            comment.Body,
		BodyHTML:        comment.BodyHTML,
		CreatedAt:       comment.CreatedAt,
		URL:             comment.URL,
		DiffHunk:        comment.DiffHunk,
		State:           comment.State,
		ViewerDidAuthor: comment.ViewerDidAuthor,
		ReactionGroups:  toDomainReactionGroups(comment.ReactionGroups),
	}
	if comment.Author != nil {
		actual.Author = &githubdomain.PullRequestCommentAuthor{Login: comment.Author.Login}
	}
	return actual
}

func toDomainReviewThreadTarget(target any) (githubdomain.PullRequestReviewThreadTarget, bool) {
	switch actual := target.(type) {
	case githubdomain.PullRequestReviewThreadTarget:
		return actual, true
	case *githubdomain.PullRequestReviewThreadTarget:
		if actual == nil {
			return githubdomain.PullRequestReviewThreadTarget{}, false
		}
		return *actual, true
	case transportgithub.PullRequestReviewThreadTarget:
		return githubdomain.PullRequestReviewThreadTarget{
			Path:        actual.Path,
			Line:        actual.Line,
			Side:        actual.Side,
			StartLine:   actual.StartLine,
			StartSide:   actual.StartSide,
			SubjectType: actual.SubjectType,
		}, true
	case *transportgithub.PullRequestReviewThreadTarget:
		if actual == nil {
			return githubdomain.PullRequestReviewThreadTarget{}, false
		}
		return githubdomain.PullRequestReviewThreadTarget{
			Path:        actual.Path,
			Line:        actual.Line,
			Side:        actual.Side,
			StartLine:   actual.StartLine,
			StartSide:   actual.StartSide,
			SubjectType: actual.SubjectType,
		}, true
	default:
		return githubdomain.PullRequestReviewThreadTarget{}, false
	}
}

func toDomainPullRequestStatusCheck(check any) (githubdomain.PullRequestStatusCheck, bool) {
	switch actual := check.(type) {
	case githubdomain.PullRequestStatusCheck:
		return actual, true
	case *githubdomain.PullRequestStatusCheck:
		if actual == nil {
			return githubdomain.PullRequestStatusCheck{}, false
		}
		return *actual, true
	case transportgithub.PullRequestStatusCheck:
		return transportgithub.ToDomainPullRequestStatusCheck(actual), true
	case *transportgithub.PullRequestStatusCheck:
		if actual == nil {
			return githubdomain.PullRequestStatusCheck{}, false
		}
		return transportgithub.ToDomainPullRequestStatusCheck(*actual), true
	default:
		return githubdomain.PullRequestStatusCheck{}, false
	}
}

func toDomainPullRequestReviewThread(thread any) (githubdomain.PullRequestReviewThread, bool) {
	switch actual := thread.(type) {
	case githubdomain.PullRequestReviewThread:
		return actual, true
	case *githubdomain.PullRequestReviewThread:
		if actual == nil {
			return githubdomain.PullRequestReviewThread{}, false
		}
		return *actual, true
	case transportgithub.PullRequestReviewThread:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{InlineCommentThreads: []transportgithub.PullRequestReviewThread{actual}}).InlineCommentThreads[0], true
	case *transportgithub.PullRequestReviewThread:
		if actual == nil {
			return githubdomain.PullRequestReviewThread{}, false
		}
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{InlineCommentThreads: []transportgithub.PullRequestReviewThread{*actual}}).InlineCommentThreads[0], true
	default:
		return githubdomain.PullRequestReviewThread{}, false
	}
}

func toDomainPullRequestComments(comments any) []githubdomain.PullRequestComment {
	switch actual := comments.(type) {
	case []githubdomain.PullRequestComment:
		return append([]githubdomain.PullRequestComment(nil), actual...)
	case []transportgithub.PullRequestComment:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{Comments: actual}).Comments
	default:
		return nil
	}
}

func toDomainPullRequestReviewThreads(threads any) []githubdomain.PullRequestReviewThread {
	switch actual := threads.(type) {
	case []githubdomain.PullRequestReviewThread:
		return append([]githubdomain.PullRequestReviewThread(nil), actual...)
	case []transportgithub.PullRequestReviewThread:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{InlineCommentThreads: actual}).InlineCommentThreads
	default:
		return nil
	}
}

func toDomainPullRequestInlineComments(comments any) []githubdomain.PullRequestInlineComment {
	switch actual := comments.(type) {
	case []githubdomain.PullRequestInlineComment:
		return append([]githubdomain.PullRequestInlineComment(nil), actual...)
	case []transportgithub.PullRequestInlineComment:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{InlineComments: actual}).InlineComments
	default:
		return nil
	}
}

func toDomainPullRequestReviews(reviews any) []githubdomain.PullRequestReview {
	switch actual := reviews.(type) {
	case []githubdomain.PullRequestReview:
		return append([]githubdomain.PullRequestReview(nil), actual...)
	case []transportgithub.PullRequestReview:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{Reviews: actual}).Reviews
	default:
		return nil
	}
}

func toDomainPullRequestCommits(commits any) []githubdomain.PullRequestCommit {
	switch actual := commits.(type) {
	case []githubdomain.PullRequestCommit:
		return append([]githubdomain.PullRequestCommit(nil), actual...)
	case []transportgithub.PullRequestCommit:
		return transportgithub.ToDomainPullRequestDetail(transportgithub.PullRequestDetail{Commits: actual}).Commits
	default:
		return nil
	}
}

func toDomainReactionGroup(group any) (githubdomain.ReactionGroup, bool) {
	switch actual := group.(type) {
	case githubdomain.ReactionGroup:
		return actual, true
	case *githubdomain.ReactionGroup:
		if actual == nil {
			return githubdomain.ReactionGroup{}, false
		}
		return *actual, true
	case transportgithub.ReactionGroup:
		return githubdomain.ReactionGroup{Content: githubdomain.NormalizeReactionContent(string(actual.Content)), TotalCount: actual.TotalCount, ViewerHasReacted: actual.ViewerHasReacted}, true
	case *transportgithub.ReactionGroup:
		if actual == nil {
			return githubdomain.ReactionGroup{}, false
		}
		return githubdomain.ReactionGroup{Content: githubdomain.NormalizeReactionContent(string(actual.Content)), TotalCount: actual.TotalCount, ViewerHasReacted: actual.ViewerHasReacted}, true
	default:
		return githubdomain.ReactionGroup{}, false
	}
}

func toDomainReactionGroups(groups any) []githubdomain.ReactionGroup {
	switch actual := groups.(type) {
	case []githubdomain.ReactionGroup:
		return append([]githubdomain.ReactionGroup(nil), actual...)
	case []transportgithub.ReactionGroup:
		converted := make([]githubdomain.ReactionGroup, 0, len(actual))
		for _, group := range actual {
			converted = append(converted, githubdomain.ReactionGroup{
				Content:          githubdomain.NormalizeReactionContent(string(group.Content)),
				TotalCount:       group.TotalCount,
				ViewerHasReacted: group.ViewerHasReacted,
			})
		}
		return converted
	default:
		return nil
	}
}
