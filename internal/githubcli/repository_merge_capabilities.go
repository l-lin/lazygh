package githubcli

import "errors"

const repositoryMergeCapabilitiesQuery = `query($owner:String!,$name:String!){repository(owner:$owner,name:$name){autoMergeAllowed}}`

var ErrInvalidRepositoryMergeCapabilitiesResponse = errors.New("invalid repository merge capabilities response")

type repositoryMergeCapabilities struct {
	AutoMergeAllowed bool
}

func (client *PullRequestMutationService) loadRepositoryMergeCapabilities(repository string) (repositoryMergeCapabilities, error) {
	owner, name, err := splitRepositoryOwnerAndName(repository)
	if err != nil {
		return repositoryMergeCapabilities{}, err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: repositoryMergeCapabilitiesQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", owner), typedGraphQLVariable("name", name)}})
	if err != nil {
		return repositoryMergeCapabilities{}, err
	}

	return parseRepositoryMergeCapabilities(result.Stdout)
}

func parseRepositoryMergeCapabilities(stdout []byte) (repositoryMergeCapabilities, error) {
	var response struct {
		Repository *struct {
			AutoMergeAllowed bool `json:"autoMergeAllowed"`
		} `json:"repository"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidRepositoryMergeCapabilitiesResponse); err != nil {
		return repositoryMergeCapabilities{}, err
	}
	if response.Repository == nil {
		return repositoryMergeCapabilities{}, ErrInvalidRepositoryMergeCapabilitiesResponse
	}

	return repositoryMergeCapabilities{AutoMergeAllowed: response.Repository.AutoMergeAllowed}, nil
}
