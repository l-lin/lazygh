package githubcli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const notificationsBulkReadAPIPath = "/notifications"

var (
	ErrMissingNotificationThreadID     = errors.New("missing notification thread id")
	ErrNotificationEndpointAuthRefused = errors.New("notification endpoint authentication refused")
)

type NotificationBulkReadResult struct {
	Accepted bool
}

func (client *Client) MarkNotificationRead(threadID string) error {
	threadAPIPath, err := notificationThreadAPIPath(threadID)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh api notification thread read", "api", threadAPIPath, "--method", "PATCH"); err != nil {
		return normalizeNotificationEndpointError(err)
	}
	return nil
}

func (client *Client) MarkNotificationDone(threadID string) error {
	threadAPIPath, err := notificationThreadAPIPath(threadID)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh api notification thread done", "api", threadAPIPath, "--method", "DELETE"); err != nil {
		return normalizeNotificationEndpointError(err)
	}
	return nil
}

func (client *Client) MarkAllNotificationsRead() (NotificationBulkReadResult, error) {
	result, err := client.runGH("gh api notifications read", "api", notificationsBulkReadAPIPath, "--method", "PUT", "--include")
	if err != nil {
		return NotificationBulkReadResult{}, normalizeNotificationEndpointError(err)
	}

	return NotificationBulkReadResult{Accepted: notificationIncludedHTTPStatus(result.Stdout) == 202}, nil
}

func (client *Client) MarkAllNotificationsDone(notifications []Notification) (int, error) {
	markedCount := 0
	for _, notification := range notifications {
		trimmedThreadID := strings.TrimSpace(notification.ID)
		if trimmedThreadID == "" {
			continue
		}
		if err := client.MarkNotificationDone(trimmedThreadID); err != nil {
			return markedCount, err
		}
		markedCount++
	}
	return markedCount, nil
}

func notificationThreadAPIPath(threadID string) (string, error) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return "", ErrMissingNotificationThreadID
	}
	return "/notifications/threads/" + trimmedThreadID, nil
}

func notificationIncludedHTTPStatus(stdout []byte) int {
	for _, line := range strings.Split(strings.ReplaceAll(string(stdout), "\r\n", "\n"), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		parts := strings.Fields(trimmedLine)
		if len(parts) < 2 {
			return 0
		}
		actual, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0
		}
		return actual
	}
	return 0
}

func normalizeNotificationEndpointError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrUnauthenticated) || errors.Is(err, ErrUnavailable) {
		return err
	}

	message := strings.TrimSpace(err.Error())
	lowerMessage := strings.ToLower(message)
	switch {
	case strings.Contains(lowerMessage, "resource not accessible by personal access token"),
		strings.Contains(lowerMessage, "resource not accessible by integration"),
		strings.Contains(lowerMessage, "only support authentication using a personal access token (classic)"),
		strings.Contains(lowerMessage, "personal access token (classic)"):
		return fmt.Errorf(
			"%w: GitHub notification endpoints reject the current credential type. Re-authenticate `gh` with a personal access token (classic) that has the `notifications` scope and `repo` access for private repositories",
			ErrNotificationEndpointAuthRefused,
		)
	case strings.Contains(lowerMessage, "notifications scope"),
		strings.Contains(lowerMessage, "requires one of the following scopes"),
		strings.Contains(lowerMessage, "missing required scope"),
		strings.Contains(lowerMessage, "insufficient scopes"):
		return fmt.Errorf(
			"%w: GitHub notification endpoints need the `notifications` scope and `repo` access for private repositories. Re-authenticate `gh` and grant those scopes",
			ErrNotificationEndpointAuthRefused,
		)
	default:
		return err
	}
}
