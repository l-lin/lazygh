package githubcli

import (
	"bufio"
	"fmt"
	"strings"

	githubcodeowners "github.com/hmarr/codeowners"
)

var (
	ErrInvalidPullRequestBaseRefNameResponse = fmt.Errorf("invalid pull request base ref response")
	ErrInvalidPullRequestCodeownersResponse  = fmt.Errorf("invalid pull request CODEOWNERS response")
	ErrMissingPullRequestBaseRefName         = fmt.Errorf("missing pull request base ref name")
)

const pullRequestBaseRefNameQuery = `query($owner:String!,$name:String!,$number:Int!){repository(owner:$owner,name:$name){pullRequest(number:$number){baseRefName}}}`

const pullRequestCodeownersBlobQuery = `query($owner:String!,$name:String!,$dotgithubExpression:String!,$rootExpression:String!,$docsExpression:String!){repository(owner:$owner,name:$name){dotgithub:object(expression:$dotgithubExpression){... on Blob{text}} root:object(expression:$rootExpression){... on Blob{text}} docs:object(expression:$docsExpression){... on Blob{text}}}}`

type pullRequestBaseRefNameResponse struct {
	Data *struct {
		Repository *struct {
			PullRequest *struct {
				BaseRefName string `json:"baseRefName"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

type pullRequestCodeownersBlobResponse struct {
	Data *struct {
		Repository *struct {
			DotGitHub *struct {
				Text string `json:"text"`
			} `json:"dotgithub"`
			Root *struct {
				Text string `json:"text"`
			} `json:"root"`
			Docs *struct {
				Text string `json:"text"`
			} `json:"docs"`
		} `json:"repository"`
	} `json:"data"`
}

func (client *PullRequestDetailService) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return nil, err
	}

	normalizedPaths := normalizePullRequestFilePaths(filePaths)
	if len(normalizedPaths) == 0 {
		return nil, nil
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return nil, err
	}

	baseRefName, err := client.pullRequestBaseRefName(owner, name, number)
	if err != nil {
		return nil, err
	}

	codeownersText, found, err := client.pullRequestCodeownersBlob(owner, name, baseRefName)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	return matchPullRequestCodeownersTeamOwners(codeownersText, normalizedPaths), nil
}

func (client *PullRequestDetailService) pullRequestBaseRefName(owner string, name string, number int) (string, error) {
	result, err := client.queryGraphQL(GraphQLRequest{Query: pullRequestBaseRefNameQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", strings.TrimSpace(owner)), typedGraphQLVariable("name", strings.TrimSpace(name)), typedGraphQLVariable("number", number)}})
	if err != nil {
		return "", err
	}

	var response pullRequestBaseRefNameResponse
	if err := client.transport.decoder.DecodeJSON(result.Stdout, &response); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPullRequestBaseRefNameResponse, err)
	}
	if response.Data == nil || response.Data.Repository == nil || response.Data.Repository.PullRequest == nil {
		return "", ErrInvalidPullRequestBaseRefNameResponse
	}

	return validateNonEmptyPullRequestField(response.Data.Repository.PullRequest.BaseRefName, ErrMissingPullRequestBaseRefName)
}

func (client *PullRequestDetailService) pullRequestCodeownersBlob(owner string, name string, baseRefName string) (string, bool, error) {
	trimmedBaseRefName := strings.TrimSpace(baseRefName)
	if trimmedBaseRefName == "" {
		return "", false, ErrMissingPullRequestBaseRefName
	}

	refPrefix := "refs/heads/" + trimmedBaseRefName + ":"
	result, err := client.queryGraphQL(GraphQLRequest{Query: pullRequestCodeownersBlobQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", strings.TrimSpace(owner)), typedGraphQLVariable("name", strings.TrimSpace(name)), typedGraphQLVariable("dotgithubExpression", refPrefix+".github/CODEOWNERS"), typedGraphQLVariable("rootExpression", refPrefix+"CODEOWNERS"), typedGraphQLVariable("docsExpression", refPrefix+"docs/CODEOWNERS")}})
	if err != nil {
		return "", false, err
	}

	var response pullRequestCodeownersBlobResponse
	if err := client.transport.decoder.DecodeJSON(result.Stdout, &response); err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrInvalidPullRequestCodeownersResponse, err)
	}
	if response.Data == nil || response.Data.Repository == nil {
		return "", false, ErrInvalidPullRequestCodeownersResponse
	}

	if response.Data.Repository.DotGitHub != nil {
		return normalizePullRequestDiffText(response.Data.Repository.DotGitHub.Text), true, nil
	}
	if response.Data.Repository.Root != nil {
		return normalizePullRequestDiffText(response.Data.Repository.Root.Text), true, nil
	}
	if response.Data.Repository.Docs != nil {
		return normalizePullRequestDiffText(response.Data.Repository.Docs.Text), true, nil
	}

	return "", false, nil
}

func normalizePullRequestFilePaths(filePaths []string) []string {
	if len(filePaths) == 0 {
		return nil
	}

	normalizedPaths := make([]string, 0, len(filePaths))
	seenPaths := map[string]bool{}
	for _, filePath := range filePaths {
		trimmedFilePath := strings.TrimSpace(filePath)
		if trimmedFilePath == "" || seenPaths[trimmedFilePath] {
			continue
		}
		seenPaths[trimmedFilePath] = true
		normalizedPaths = append(normalizedPaths, trimmedFilePath)
	}
	if len(normalizedPaths) == 0 {
		return nil
	}
	return normalizedPaths
}

func matchPullRequestCodeownersTeamOwners(codeownersText string, filePaths []string) map[string][]string {
	ruleset := parsePullRequestCodeownersRuleset(codeownersText)
	if len(ruleset) == 0 {
		return nil
	}

	teamOwnersByPath := map[string][]string{}
	for _, filePath := range normalizePullRequestFilePaths(filePaths) {
		rule, err := ruleset.Match(filePath)
		if err != nil || rule == nil {
			continue
		}

		teamOwners := codeownersRuleTeamOwnerLabels(rule.Owners)
		if len(teamOwners) == 0 {
			continue
		}
		teamOwnersByPath[filePath] = teamOwners
	}
	if len(teamOwnersByPath) == 0 {
		return nil
	}
	return teamOwnersByPath
}

func parsePullRequestCodeownersRuleset(codeownersText string) githubcodeowners.Ruleset {
	scanner := bufio.NewScanner(strings.NewReader(normalizePullRequestDiffText(codeownersText)))
	rules := make(githubcodeowners.Ruleset, 0)
	for scanner.Scan() {
		parsedRules, err := githubcodeowners.ParseFile(strings.NewReader(scanner.Text()))
		if err != nil || len(parsedRules) == 0 {
			continue
		}
		rules = append(rules, parsedRules...)
	}
	return rules
}

func codeownersRuleTeamOwnerLabels(owners []githubcodeowners.Owner) []string {
	fullValues := make([]string, 0, len(owners))
	seenValues := map[string]bool{}
	labelCounts := map[string]int{}
	for _, owner := range owners {
		if owner.Type != githubcodeowners.TeamOwner {
			continue
		}
		trimmedValue := strings.TrimSpace(owner.Value)
		if trimmedValue == "" || seenValues[trimmedValue] {
			continue
		}
		seenValues[trimmedValue] = true
		fullValues = append(fullValues, trimmedValue)
		labelCounts[codeownersTeamOwnerLabel(trimmedValue)]++
	}
	if len(fullValues) == 0 {
		return nil
	}

	labels := make([]string, 0, len(fullValues))
	for _, fullValue := range fullValues {
		label := codeownersTeamOwnerLabel(fullValue)
		if labelCounts[label] > 1 {
			label = fullValue
		}
		labels = append(labels, label)
	}
	return labels
}

func codeownersTeamOwnerLabel(teamOwner string) string {
	trimmedTeamOwner := strings.TrimSpace(teamOwner)
	if trimmedTeamOwner == "" {
		return ""
	}
	_, slug, ok := strings.Cut(trimmedTeamOwner, "/")
	if ok && strings.TrimSpace(slug) != "" {
		return strings.TrimSpace(slug)
	}
	return trimmedTeamOwner
}
