package application

import (
	"fmt"
	"strings"
)

type ApplicationSourceOption struct {
	Repo           string
	Path           string
	TargetRevision string
}

// BuildKustomizeResourceRef builds a kustomize resource reference from an ApplicationSourceOption
func BuildKustomizeResourceRef(source ApplicationSourceOption) string {
	// remove possible .git suffix
	repoURL := strings.TrimSuffix(source.Repo, ".git")

	// if git@ format, convert to https:// format
	if strings.HasPrefix(repoURL, "git@") {
		repoURL = strings.Replace(repoURL, "git@", "", 1)
		repoURL = strings.Replace(repoURL, ":", "/", 1)
	}

	// remove https:// prefix if exists
	repoURL = strings.TrimPrefix(repoURL, "https://")

	// build path part
	pathPart := ""
	if source.Path != "" {
		pathPart = "/" + source.Path
	}

	// build reference
	ref := source.TargetRevision
	if ref == "" {
		ref = "main" // default to main branch
	}

	// return format: repository/path?ref=revision
	return fmt.Sprintf("%s%s?ref=%s", repoURL, pathPart, ref)
}
