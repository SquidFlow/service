package reader

import (
	"fmt"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// this file is an abstraction layer for the git repository layout
// supports the following layouts:
// - helm
// - kustomize
// - helm with multiple environments
// - kustomize with multiple environments
// - some existing gitops patterns

func InferApplicationSource(repofs fs.FS, req types.ApplicationSourceRequest) (string, []string, error) {
	appSourceType, err := detectApplicationType(repofs, req)
	if err != nil {
		return "", nil, err
	}

	switch appSourceType {
	case AppTypeHelm, AppTypeKustomize:
		return appSourceType, []string{"default"}, nil

	case AppTypeHelmMultiEnv, AppTypeKustomizeMultiEnv:
		detector := NewEnvironmentDetector(appSourceType, repofs, req.Path)
		environments, err := detector.DetectEnvironments()
		if err != nil {
			return appSourceType, nil, err
		}
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
			"app_source_type": appSourceType,
			"environments":    environments,
		}).Debug("detected application environments")
		return appSourceType, environments, nil

	default:
		return "", nil, fmt.Errorf("unknown application source type: %s", appSourceType)
	}
}
