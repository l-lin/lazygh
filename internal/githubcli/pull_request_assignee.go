package githubcli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidAssignableUsersResponse = fmt.Errorf("invalid assignable users response")

const (
	assignableUsersPerPage         = 100
	assignableUserSearchResultSize = 20
)

const searchAssignableUsersQuery = `query($owner:String!,$name:String!,$first:Int!,$search:String){repository(owner:$owner,name:$name){assignableUsers(first:$first,query:$search){nodes{login name is_bot:isBot}}}}`

func (client *PullRequestMutationService) ListAssignableUsers(repository string) ([]PullRequestAuthor, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return nil, ErrMissingPullRequestIdentity
	}

	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/assignees?per_page=%d", trimmedRepository, assignableUsersPerPage), Paginate: true, Slurp: true})
	if err != nil {
		return nil, err
	}

	return parseAssignableUsers(result.Stdout)
}

func (client *PullRequestMutationService) SearchAssignableUsers(repository string, query string) ([]PullRequestAuthor, error) {
	trimmedRepository := strings.TrimSpace(repository)
	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return nil, err
	}

	variables := []GraphQLVariable{
		typedGraphQLVariable("owner", owner),
		typedGraphQLVariable("name", name),
		typedGraphQLVariable("first", assignableUserSearchResultSize),
	}
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery != "" {
		variables = append(variables, typedGraphQLVariable("search", trimmedQuery))
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: searchAssignableUsersQuery, Variables: variables, DisplayArgs: searchAssignableUsersDisplayArgs(owner, name, trimmedQuery)})
	if err != nil {
		return nil, err
	}

	return parseAssignableUserSearchResults(result.Stdout)
}

func (client *PullRequestMutationService) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	normalizedAddLogins := normalizePullRequestAssigneeLogins(addLogins)
	normalizedRemoveLogins := normalizePullRequestAssigneeLogins(removeLogins)
	if len(normalizedAddLogins) == 0 && len(normalizedRemoveLogins) == 0 {
		return nil
	}

	args := []string{"pr", "edit", strconv.Itoa(number), "-R", trimmedRepository}
	if len(normalizedAddLogins) > 0 {
		args = append(args, "--add-assignee", strings.Join(normalizedAddLogins, ","))
	}
	if len(normalizedRemoveLogins) > 0 {
		args = append(args, "--remove-assignee", strings.Join(normalizedRemoveLogins, ","))
	}

	if _, err := client.execute(rawCommand(args...)); err != nil {
		return err
	}

	return nil
}

func parseAssignableUsers(stdout []byte) ([]PullRequestAuthor, error) {
	var flatUsers []PullRequestAuthor
	if err := (paginator{}).DecodeSlurpedJSON(stdout, &flatUsers); err == nil {
		return normalizeAssignableUsers(flatUsers), nil
	}
	if err := (responseDecoder{}).DecodeJSON(stdout, &flatUsers); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAssignableUsersResponse, err)
	}

	return normalizeAssignableUsers(flatUsers), nil
}

func parseAssignableUserSearchResults(stdout []byte) ([]PullRequestAuthor, error) {
	var response struct {
		Repository *struct {
			AssignableUsers struct {
				Nodes []PullRequestAuthor `json:"nodes"`
			} `json:"assignableUsers"`
		} `json:"repository"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidAssignableUsersResponse); err != nil {
		return nil, err
	}
	if response.Repository == nil {
		return nil, nil
	}
	return normalizeAssignableUserSearchResults(response.Repository.AssignableUsers.Nodes), nil
}

func normalizeAssignableUsers(users []PullRequestAuthor) []PullRequestAuthor {
	normalizedUsers := make([]PullRequestAuthor, 0, len(users))
	seenLogins := map[string]bool{}
	for _, user := range users {
		normalizedUser := user.normalized()
		if normalizedUser.Login == "" || seenLogins[normalizedUser.Login] {
			continue
		}
		seenLogins[normalizedUser.Login] = true
		normalizedUsers = append(normalizedUsers, normalizedUser)
	}
	sort.SliceStable(normalizedUsers, func(i int, j int) bool {
		return normalizedUsers[i].Login < normalizedUsers[j].Login
	})
	return normalizedUsers
}

func normalizeAssignableUserSearchResults(users []PullRequestAuthor) []PullRequestAuthor {
	normalizedUsers := make([]PullRequestAuthor, 0, len(users))
	seenLogins := map[string]bool{}
	for _, user := range users {
		normalizedUser := user.normalized()
		if normalizedUser.Login == "" || seenLogins[normalizedUser.Login] {
			continue
		}
		seenLogins[normalizedUser.Login] = true
		normalizedUsers = append(normalizedUsers, normalizedUser)
	}
	return normalizedUsers
}

func FormatAssignableUsersCommand(repository string) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return formatCommand("api")
	}
	return formatCommand("api", fmt.Sprintf("repos/%s/assignees?per_page=%d", trimmedRepository, assignableUsersPerPage), "--paginate", "--slurp")
}

func FormatSearchAssignableUsersCommand(repository string, query string) string {
	trimmedRepository := strings.TrimSpace(repository)
	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return formatCommand("api", "graphql")
	}
	return formatCommandArguments(searchAssignableUsersDisplayArgs(owner, name, strings.TrimSpace(query)))
}

func searchAssignableUsersDisplayArgs(owner string, name string, query string) []string {
	args := []string{"api", "graphql", "-F", "owner=" + strings.TrimSpace(owner), "-F", "name=" + strings.TrimSpace(name), "-F", "first=" + strconv.Itoa(assignableUserSearchResultSize)}
	if strings.TrimSpace(query) != "" {
		args = append(args, "-F", "search="+strings.TrimSpace(query))
	}
	return args
}

func normalizePullRequestAssigneeLogins(logins []string) []string {
	normalizedLogins := make([]string, 0, len(logins))
	seenLogins := map[string]bool{}
	for _, login := range logins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" || seenLogins[trimmedLogin] {
			continue
		}
		seenLogins[trimmedLogin] = true
		normalizedLogins = append(normalizedLogins, trimmedLogin)
	}
	return normalizedLogins
}
