package githubcli

import "strings"

type ReviewThreadAssembler struct {
	LoadThreadsPage        func(owner string, name string, number int, cursor string) (pullRequestReviewThreadsPage, error)
	LoadThreadCommentsPage func(threadID string, cursor string) (pullRequestReviewThreadCommentsPage, error)
}

func newReviewThreadAssembler(service *PullRequestDetailService) ReviewThreadAssembler {
	return ReviewThreadAssembler{
		LoadThreadsPage:        service.pullRequestReviewThreadsPage,
		LoadThreadCommentsPage: service.pullRequestReviewThreadCommentsPage,
	}
}

func (assembler ReviewThreadAssembler) List(repository string, number int) ([]PullRequestReviewThread, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return nil, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return nil, err
	}
	if assembler.LoadThreadsPage == nil {
		return nil, ErrInvalidPullRequestReviewThreadsResponse
	}

	threads := make([]PullRequestReviewThread, 0)
	cursor := ""
	for {
		page, err := assembler.LoadThreadsPage(owner, name, number, cursor)
		if err != nil {
			return nil, err
		}
		for _, threadNode := range page.Threads {
			thread := threadNode.Thread
			if threadNode.CommentsHasNextPage {
				if strings.TrimSpace(thread.ID) == "" || strings.TrimSpace(threadNode.CommentsEndCursor) == "" || assembler.LoadThreadCommentsPage == nil {
					return nil, ErrInvalidPullRequestReviewThreadsResponse
				}
				remainingComments, err := assembler.loadCommentsAfter(thread.ID, threadNode.CommentsEndCursor)
				if err != nil {
					return nil, err
				}
				thread.Comments = append(thread.Comments, remainingComments...)
			}
			threads = append(threads, thread.normalized())
		}
		if !page.HasNextPage {
			return threads, nil
		}
		if strings.TrimSpace(page.EndCursor) == "" {
			return nil, ErrInvalidPullRequestReviewThreadsResponse
		}
		cursor = page.EndCursor
	}
}

func (assembler ReviewThreadAssembler) loadCommentsAfter(threadID string, cursor string) ([]PullRequestComment, error) {
	comments := make([]PullRequestComment, 0)
	nextCursor := strings.TrimSpace(cursor)
	for nextCursor != "" {
		page, err := assembler.LoadThreadCommentsPage(threadID, nextCursor)
		if err != nil {
			return nil, err
		}
		comments = append(comments, page.Comments...)
		if !page.HasNextPage {
			return comments, nil
		}
		if strings.TrimSpace(page.EndCursor) == "" {
			return nil, ErrInvalidPullRequestReviewThreadsResponse
		}
		nextCursor = page.EndCursor
	}
	return comments, nil
}
