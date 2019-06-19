package url

import (
	"path"
	"strings"
)

func NormalizeRepoUrl(repo string) (string, bool) {
	repo = strings.TrimSuffix(repo, "/")

	if strings.HasPrefix(repo, "http://") {
		repo = strings.TrimPrefix(repo, "http://")
	} else if strings.HasPrefix(repo, "https://") {
		repo = strings.TrimPrefix(repo, "https://")
	}

	// only support github for now
	if !strings.HasPrefix(repo, "github.com/") {
		return "", false
	}

	repo = path.Clean(repo)
	return "https://" + repo, true
}
