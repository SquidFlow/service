package writer

import (
	"context"
	"fmt"

	billyUtils "github.com/go-git/go-billy/v5/util"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
)

// Note: all the project delete logic is based on meta repo
func (n *NativeRepoTarget) RunProjectDelete(ctx context.Context, projectName string) error {
	r, repofs, err := prepareRepo(ctx, n.metaRepoCloneOpts, projectName)
	if err != nil {
		return err
	}

	allApps, err := repofs.ReadDir(store.Default.AppsDir)
	if err != nil {
		return fmt.Errorf("failed to list all applications")
	}

	for _, app := range allApps {
		err = DeleteFromProject(repofs, app.Name(), projectName)
		if err != nil {
			return err
		}
	}

	err = repofs.Remove(repofs.Join(store.Default.ProjectsDir, projectName+".yaml"))
	if err != nil {
		return fmt.Errorf("failed to delete project '%s': %w", projectName, err)
	}

	log.G().WithFields(log.Fields{"project": projectName}).Info("deleting project")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: fmt.Sprintf("chore: deleted project '%s'", projectName)}); err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	return nil
}

func (n *NativeRepoTarget) RunProjectList(ctx context.Context) ([]types.TenantInfo, error) {
	n.metaRepoCloneOpts.Parse()

	_, repofs, err := prepareRepo(ctx, n.metaRepoCloneOpts, "")
	if err != nil {
		return nil, err
	}

	matches, err := billyUtils.Glob(repofs, repofs.Join(store.Default.ProjectsDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var tenants []types.TenantInfo
	for _, name := range matches {
		proj, appset, err := getProjectInfoFromFile(repofs, name)
		if err != nil {
			return nil, err
		}

		tenantInfo := types.TenantInfo{
			Name:           proj.Name,
			Namespace:      proj.Namespace,
			DefaultCluster: proj.Annotations[store.Default.DestServerAnnotation],
			GitOpsRepo:     appset.Spec.Generators[0].Git.RepoURL,
		}
		tenants = append(tenants, tenantInfo)
	}

	return tenants, nil
}

func (n *NativeRepoTarget) RunProjectGet(ctx context.Context, projectName string) (*types.TenantDetailInfo, error) {
	_, repofs, err := prepareRepo(ctx, n.metaRepoCloneOpts, projectName)
	if err != nil {
		return nil, err
	}

	projectFile := repofs.Join(store.Default.ProjectsDir, projectName+".yaml")
	if !repofs.ExistsOrDie(projectFile) {
		return nil, fmt.Errorf("project %s not found", projectName)
	}

	proj, appset, err := getProjectInfoFromFile(repofs, projectFile)
	if err != nil {
		return nil, err
	}

	detail := &types.TenantDetailInfo{
		Name:           proj.Name,
		Namespace:      proj.Namespace,
		Description:    proj.Annotations["description"],
		DefaultCluster: proj.Annotations[store.Default.DestServerAnnotation],
		CreatedBy:      proj.Annotations["created-by"],
		CreatedAt:      proj.CreationTimestamp.String(),
		GitOpsRepo:     appset.Spec.Generators[0].Git.RepoURL,
	}

	if len(proj.Spec.SourceRepos) > 0 {
		detail.SourceRepos = proj.Spec.SourceRepos
	}

	if len(proj.Spec.Destinations) > 0 {
		for _, dest := range proj.Spec.Destinations {
			detail.Destinations = append(detail.Destinations, types.ProjectDest{
				Server:    dest.Server,
				Namespace: dest.Namespace,
			})
		}
	}

	if len(proj.Spec.ClusterResourceWhitelist) > 0 {
		for _, res := range proj.Spec.ClusterResourceWhitelist {
			detail.ClusterResourceWhitelist = append(detail.ClusterResourceWhitelist, types.ProjectResource{
				Group: res.Group,
				Kind:  res.Kind,
			})
		}
	}

	if len(proj.Spec.NamespaceResourceWhitelist) > 0 {
		for _, res := range proj.Spec.NamespaceResourceWhitelist {
			detail.NamespaceResourceWhitelist = append(detail.NamespaceResourceWhitelist, types.ProjectResource{
				Group: res.Group,
				Kind:  res.Kind,
			})
		}
	}

	return detail, nil
}

func DeleteFromProject(repofs fs.FS, appName, projectName string) error {
	var dirToCheck string
	appDir := repofs.Join(store.Default.AppsDir, appName)
	appOverlay := repofs.Join(appDir, store.Default.OverlaysDir)
	if repofs.ExistsOrDie(appOverlay) {
		// kustApp
		dirToCheck = appOverlay
	} else {
		// dirApp
		dirToCheck = appDir
	}

	allProjects, err := repofs.ReadDir(dirToCheck)
	if err != nil {
		return fmt.Errorf("failed to check projects in '%s': %w", appName, err)
	}

	var found = false
	for _, project := range allProjects {
		if project.Name() == projectName {
			found = true
		}
	}
	if !found {
		return nil
	}

	var dirToRemove string
	if len(allProjects) == 1 {
		dirToRemove = appDir
		log.G().Infof("Deleting app '%s'", appName)
	} else {
		dirToRemove = repofs.Join(dirToCheck, projectName)
		log.G().Infof("Deleting app '%s' from project '%s'", appName, projectName)
	}

	err = billyUtils.RemoveAll(repofs, dirToRemove)
	if err != nil {
		return fmt.Errorf("failed to delete directory '%s': %w", dirToRemove, err)
	}

	return nil
}
