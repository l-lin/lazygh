package githubcli

import (
	"encoding/json"
	"fmt"
	"strings"
)

const pullRequestReactionTargetsQuery = `query($owner:String!,$name:String!,$number:Int!,$cursor:String){repository(owner:$owner,name:$name){pullRequest(number:$number){id reactionGroups{content viewerHasReacted users{totalCount}} comments(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor author{login} body createdAt url reactionGroups{content viewerHasReacted users{totalCount}}}}}}}`
const pullRequestReviewCommentReactionGroupsQuery = `query($ids:[ID!]!){nodes(ids:$ids){... on PullRequestReviewComment{id reactionGroups{content viewerHasReacted users{totalCount}}}}}`

var (
	ErrInvalidPullRequestReactionTargetsResponse            = fmt.Errorf("invalid pull request reaction targets response")
	ErrInvalidPullRequestReviewCommentReactionGroupsPayload = fmt.Errorf("invalid pull request review comment reaction groups response")
)

type pullRequestReactionTargets struct {
	PullRequestID  string
	ReactionGroups []ReactionGroup
	Comments       []PullRequestComment
}

type pullRequestReactionTargetsPage struct {
	PullRequestID  string
	ReactionGroups []ReactionGroup
	Comments       []PullRequestComment
	HasNextPage    bool
	EndCursor      string
}

type pullRequestReactionTargetsResponse struct {
	Data struct {
		Repository *struct {
			PullRequest *struct {
				ID             string          `json:"id"`
				ReactionGroups []ReactionGroup `json:"reactionGroups"`
				Comments       struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []PullRequestComment `json:"nodes"`
				} `json:"comments"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type pullRequestReviewCommentReactionGroupsResponse struct {
	Data *struct {
		Nodes []*struct {
			ID             string          `json:"id"`
			ReactionGroups []ReactionGroup `json:"reactionGroups"`
		} `json:"nodes"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *PullRequestDetailService) listPullRequestReactionTargets(repository string, number int) (pullRequestReactionTargets, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return pullRequestReactionTargets{}, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return pullRequestReactionTargets{}, err
	}

	result := pullRequestReactionTargets{}
	cursor := ""
	for {
		page, err := client.loadPullRequestReactionTargetsPage(owner, name, number, cursor)
		if err != nil {
			return pullRequestReactionTargets{}, err
		}
		if result.PullRequestID == "" {
			result.PullRequestID = strings.TrimSpace(page.PullRequestID)
			result.ReactionGroups = append([]ReactionGroup(nil), page.ReactionGroups...)
		}
		result.Comments = append(result.Comments, page.Comments...)
		if !page.HasNextPage {
			return result, nil
		}
		if strings.TrimSpace(page.EndCursor) == "" {
			return pullRequestReactionTargets{}, ErrInvalidPullRequestReactionTargetsResponse
		}
		cursor = page.EndCursor
	}
}

func (client *PullRequestDetailService) loadPullRequestReactionTargetsPage(owner string, name string, number int, cursor string) (pullRequestReactionTargetsPage, error) {
	request := GraphQLRequest{Query: pullRequestReactionTargetsQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", strings.TrimSpace(owner)), typedGraphQLVariable("name", strings.TrimSpace(name)), typedGraphQLVariable("number", number)}}
	if strings.TrimSpace(cursor) != "" {
		request.Variables = append(request.Variables, typedGraphQLVariable("cursor", strings.TrimSpace(cursor)))
	}

	result, err := client.queryGraphQL(request)
	if err != nil {
		return pullRequestReactionTargetsPage{}, err
	}
	return parsePullRequestReactionTargetsPage(result.Stdout)
}

func parsePullRequestReactionTargetsPage(stdout []byte) (pullRequestReactionTargetsPage, error) {
	var response pullRequestReactionTargetsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return pullRequestReactionTargetsPage{}, fmt.Errorf("%w: %v", ErrInvalidPullRequestReactionTargetsResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return pullRequestReactionTargetsPage{}, fmt.Errorf("%w: %s", ErrInvalidPullRequestReactionTargetsResponse, message)
		}
	}
	if response.Data.Repository == nil || response.Data.Repository.PullRequest == nil {
		return pullRequestReactionTargetsPage{}, ErrInvalidPullRequestReactionTargetsResponse
	}

	pullRequest := response.Data.Repository.PullRequest
	page := pullRequestReactionTargetsPage{
		PullRequestID:  strings.TrimSpace(pullRequest.ID),
		ReactionGroups: normalizeReactionGroups(pullRequest.ReactionGroups),
		Comments:       normalizePullRequestComments(pullRequest.Comments.Nodes),
		HasNextPage:    pullRequest.Comments.PageInfo.HasNextPage,
		EndCursor:      strings.TrimSpace(pullRequest.Comments.PageInfo.EndCursor),
	}
	return page, nil
}

func (client *PullRequestDetailService) listPullRequestReviewCommentReactionGroups(ids []string) (map[string][]ReactionGroup, error) {
	trimmedIDs := uniqueReactionTargetIDs(ids)
	if len(trimmedIDs) == 0 {
		return nil, nil
	}

	groupsByID := map[string][]ReactionGroup{}
	for start := 0; start < len(trimmedIDs); start += 100 {
		end := start + 100
		if end > len(trimmedIDs) {
			end = len(trimmedIDs)
		}
		batchGroups, err := client.pullRequestReviewCommentReactionGroupsBatch(trimmedIDs[start:end])
		if err != nil {
			return nil, err
		}
		for id, groups := range batchGroups {
			groupsByID[id] = append([]ReactionGroup(nil), groups...)
		}
	}
	return groupsByID, nil
}

func (client *PullRequestDetailService) pullRequestReviewCommentReactionGroupsBatch(ids []string) (map[string][]ReactionGroup, error) {
	trimmedIDs := uniqueReactionTargetIDs(ids)
	if len(trimmedIDs) == 0 {
		return nil, nil
	}

	request := GraphQLRequest{Query: pullRequestReviewCommentReactionGroupsQuery}
	for _, id := range trimmedIDs {
		request.Variables = append(request.Variables, typedGraphQLVariable("ids[]", id))
	}

	result, err := client.queryGraphQL(request)
	if err != nil {
		return nil, err
	}
	return parsePullRequestReviewCommentReactionGroups(result.Stdout)
}

func parsePullRequestReviewCommentReactionGroups(stdout []byte) (map[string][]ReactionGroup, error) {
	var response pullRequestReviewCommentReactionGroupsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewCommentReactionGroupsPayload, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPullRequestReviewCommentReactionGroupsPayload, message)
		}
	}
	if response.Data == nil {
		return nil, ErrInvalidPullRequestReviewCommentReactionGroupsPayload
	}

	groupsByID := map[string][]ReactionGroup{}
	for _, node := range response.Data.Nodes {
		if node == nil {
			continue
		}
		trimmedID := strings.TrimSpace(node.ID)
		if trimmedID == "" {
			continue
		}
		groupsByID[trimmedID] = normalizeReactionGroups(node.ReactionGroups)
	}
	return groupsByID, nil
}

func uniqueReactionTargetIDs(ids []string) []string {
	uniqueIDs := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)
		if trimmedID == "" || seen[trimmedID] {
			continue
		}
		seen[trimmedID] = true
		uniqueIDs = append(uniqueIDs, trimmedID)
	}
	return uniqueIDs
}
