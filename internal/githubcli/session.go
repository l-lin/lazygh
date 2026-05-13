package githubcli

import (
	"fmt"
	"strings"
)

type ConnectedUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	URL         string `json:"html_url"`
}

func (service *SessionService) GetConnectedUser() (ConnectedUser, error) {
	result, err := service.doREST(RESTRequest{Path: "user"})
	if err != nil {
		return ConnectedUser{}, err
	}

	var user ConnectedUser
	if err := service.transport.decoder.DecodeJSON(result.Stdout, &user); err != nil {
		return ConnectedUser{}, fmt.Errorf("%w: %v", ErrInvalidConnectedUserResponse, err)
	}

	user = user.normalized()
	if user.Login == "" {
		return ConnectedUser{}, ErrEmptyConnectedUser
	}

	return user, nil
}

func (user ConnectedUser) normalized() ConnectedUser {
	user.Login = strings.TrimSpace(user.Login)
	user.Name = strings.TrimSpace(user.Name)
	user.Bio = strings.TrimSpace(user.Bio)
	user.Company = strings.TrimSpace(user.Company)
	user.Location = strings.TrimSpace(user.Location)
	user.URL = strings.TrimSpace(user.URL)
	return user
}
