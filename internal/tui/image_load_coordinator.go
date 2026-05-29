package tui

import "strings"

func (coordinator *imageLoadCoordinator) markDetailImageHTMLLoadPlanned(sourceKey string) {
	trimmedKey := strings.TrimSpace(sourceKey)
	if coordinator == nil || trimmedKey == "" {
		return
	}
	if coordinator.detailImageHTMLLoadInFlight == nil {
		coordinator.detailImageHTMLLoadInFlight = map[string]bool{}
	}
	coordinator.detailImageHTMLLoadInFlight[trimmedKey] = true
}

func (coordinator *imageLoadCoordinator) recordDetailImageHTMLLoadFinished(sourceKey string, failed bool) {
	trimmedKey := strings.TrimSpace(sourceKey)
	if coordinator == nil || trimmedKey == "" {
		return
	}
	if coordinator.detailImageHTMLLoadInFlight != nil {
		delete(coordinator.detailImageHTMLLoadInFlight, trimmedKey)
	}
	if failed {
		if coordinator.detailImageHTMLLoadFailed == nil {
			coordinator.detailImageHTMLLoadFailed = map[string]bool{}
		}
		coordinator.detailImageHTMLLoadFailed[trimmedKey] = true
		return
	}
	if coordinator.detailImageHTMLLoadFailed != nil {
		delete(coordinator.detailImageHTMLLoadFailed, trimmedKey)
	}
}

func (coordinator *imageLoadCoordinator) markDetailImageLoadPlanned(imageURL string) {
	trimmedURL := strings.TrimSpace(imageURL)
	if coordinator == nil || trimmedURL == "" {
		return
	}
	if coordinator.detailImageLoadInFlight == nil {
		coordinator.detailImageLoadInFlight = map[string]bool{}
	}
	coordinator.detailImageLoadInFlight[trimmedURL] = true
}

func (coordinator *imageLoadCoordinator) recordDetailImageLoadFinished(imageURL string, failed bool) {
	trimmedURL := strings.TrimSpace(imageURL)
	if coordinator == nil || trimmedURL == "" {
		return
	}
	if coordinator.detailImageLoadInFlight != nil {
		delete(coordinator.detailImageLoadInFlight, trimmedURL)
	}
	if failed {
		if coordinator.detailImageLoadFailed == nil {
			coordinator.detailImageLoadFailed = map[string]bool{}
		}
		coordinator.detailImageLoadFailed[trimmedURL] = true
		return
	}
	if coordinator.detailImageLoadFailed != nil {
		delete(coordinator.detailImageLoadFailed, trimmedURL)
	}
}

func (coordinator *imageLoadCoordinator) cachedGitHubAuthToken() (string, bool) {
	if coordinator == nil || !coordinator.githubAuthTokenLoaded {
		return "", false
	}
	return coordinator.githubAuthToken, true
}

func (coordinator *imageLoadCoordinator) cacheGitHubAuthToken(token string) {
	if coordinator == nil {
		return
	}
	coordinator.githubAuthTokenLoaded = true
	coordinator.githubAuthToken = strings.TrimSpace(token)
}
