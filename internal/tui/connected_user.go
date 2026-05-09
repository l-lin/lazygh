package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	connectedUserLoadingTitle          = "Loading connected user..."
	connectedUserLoadingDetail         = "Running `gh api user` to load the authenticated GitHub profile."
	connectedUserEmptyTitle            = "No connected user"
	connectedUserEmptyDetail           = "GitHub CLI returned no connected user data.\n\nRun `gh auth status` to verify the active account."
	connectedUserUnauthenticatedTitle  = "GitHub authentication required"
	connectedUserUnauthenticatedDetail = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	connectedUserUnavailableTitle      = "`gh` not found"
	connectedUserUnavailableDetail     = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	connectedUserGenericErrorTitle     = "Could not load connected user"
	connectedUserGenericErrorPrefix    = "Failed to run `gh api user`."
)

func connectedUserLoadingItem() Item {
	return Item{
		Title:  connectedUserLoadingTitle,
		Detail: connectedUserLoadingDetail,
	}
}

func connectedUserItem(user githubcli.ConnectedUser) Item {
	if user.Login == "" {
		return connectedUserEmptyItem()
	}

	lines := []string{
		fmt.Sprintf("Login: %s", formatLogin(user.Login)),
		fmt.Sprintf("Name: %s", valueOrDash(user.Name)),
		fmt.Sprintf("Bio: %s", valueOrDash(user.Bio)),
		fmt.Sprintf("Company: %s", valueOrDash(user.Company)),
		fmt.Sprintf("Location: %s", valueOrDash(user.Location)),
		fmt.Sprintf("Public repos: %d", user.PublicRepos),
		fmt.Sprintf("Followers: %d", user.Followers),
		fmt.Sprintf("Profile: %s", valueOrDash(user.URL)),
	}

	return Item{
		Title:  formatLogin(user.Login),
		Detail: strings.Join(lines, "\n"),
	}
}

func connectedUserErrorItem(err error) Item {
	switch {
	case errors.Is(err, githubcli.ErrEmptyConnectedUser):
		return connectedUserEmptyItem()
	case errors.Is(err, githubcli.ErrUnauthenticated):
		return Item{Title: connectedUserUnauthenticatedTitle, Detail: connectedUserUnauthenticatedDetail}
	case errors.Is(err, githubcli.ErrUnavailable):
		return Item{Title: connectedUserUnavailableTitle, Detail: connectedUserUnavailableDetail}
	default:
		return Item{Title: connectedUserGenericErrorTitle, Detail: formatConnectedUserErrorDetail(err)}
	}
}

func connectedUserStateItem(user githubcli.ConnectedUser, err error) Item {
	if err != nil {
		return connectedUserErrorItem(err)
	}

	return connectedUserItem(user)
}

func connectedUserEmptyItem() Item {
	return Item{Title: connectedUserEmptyTitle, Detail: connectedUserEmptyDetail}
}

func formatConnectedUserErrorDetail(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return connectedUserGenericErrorPrefix
	}

	return fmt.Sprintf("%s\n\n%s", connectedUserGenericErrorPrefix, message)
}

func formatLogin(login string) string {
	trimmedLogin := strings.TrimSpace(login)
	if trimmedLogin == "" {
		return "-"
	}

	return "@" + trimmedLogin
}

func valueOrDash(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "-"
	}

	return trimmedValue
}
