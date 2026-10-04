package githubcli

type pullRequestCodeownersEntry struct {
	ready chan struct{}
	text  string
	found bool
	err   error
}

func (client *PullRequestDetailService) cachedPullRequestCodeownersBlob(owner, name, revision string) (string, bool, error) {
	key := owner + "/" + name + "@" + revision
	client.codeownersMutex.Lock()
	if entry := client.codeowners[key]; entry != nil {
		client.codeownersMutex.Unlock()
		<-entry.ready
		return entry.text, entry.found, entry.err
	}
	if client.codeowners == nil {
		client.codeowners = make(map[string]*pullRequestCodeownersEntry)
	}
	// Bound session memory without evicting lookups that other reads are awaiting.
	if len(client.codeowners) >= 128 {
		for cachedKey, entry := range client.codeowners {
			select {
			case <-entry.ready:
				delete(client.codeowners, cachedKey)
			default:
			}
		}
	}
	entry := &pullRequestCodeownersEntry{ready: make(chan struct{})}
	client.codeowners[key] = entry
	client.codeownersMutex.Unlock()
	entry.text, entry.found, entry.err = client.pullRequestCodeownersBlob(owner, name, revision)
	client.codeownersMutex.Lock()
	if entry.err != nil {
		delete(client.codeowners, key)
	}
	close(entry.ready)
	client.codeownersMutex.Unlock()
	return entry.text, entry.found, entry.err
}
