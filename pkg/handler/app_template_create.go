package handler

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
)

// CreateTemplateRequest represents the request for creating a new application template
type CreateTemplateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Path        string            `json:"path" binding:"required"`
	Owner       string            `json:"owner" binding:"required"`
	Source      ApplicationSource `json:"source" binding:"required"`
	Description string            `json:"description,omitempty"`
	AppType     string            `json:"appType" binding:"required,oneof=kustomize helm"`
}

type CreateTemplateResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

// CreateApplicationTemplate handles the creation of a new application template
func CreateApplicationTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, CreateTemplateResponse{
			Success: false,
			ID:      0,
			Name:    req.Name,
		})
		return
	}

	// Generate template ID
	templateID := generateTemplateID()

	// Create new template with minimal required fields
	template := ApplicationTemplate{
		ID:          templateID,
		Name:        req.Name,
		Path:        req.Path,
		Owner:       req.Owner,
		AppType:     req.AppType,
		Source:      req.Source,
		Description: req.Description,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	log.WithFields(log.Fields{
		"template": template,
	}).Info("create template")

	// Clone gitops repo
	var opt = AppCreateOptions{
		CloneOpts: &git.CloneOptions{
			Repo:     viper.GetString("application_repo.remote_url"),
			FS:       fs.Create(memfs.New()),
			Provider: "github",
			Auth: git.Auth{
				Password: viper.GetString("application_repo.access_token"),
			},
			CloneForWrite: false,
		},
		AppsCloneOpts: &git.CloneOptions{
			CloneForWrite: false,
		},
		createOpts: &application.CreateOptions{
			AppName:          req.Name,
			AppType:          application.AppTypeKustomize,
			AppSpecifier:     req.Name,
			InstallationMode: application.InstallationModeNormal,
			DestServer:       "https://kubernetes.default.svc",
			Labels:           nil,
			Annotations:      nil,
			Include:          "",
			Exclude:          "",
		},
		ProjectName: req.Name,
		Timeout:     0,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	if err := RunCreateTemplate(opt); err != nil {
		c.JSON(500, CreateTemplateResponse{
			Success: false,
			ID:      templateID,
			Name:    req.Name,
		})
		return
	}
	c.JSON(http.StatusCreated, CreateTemplateResponse{
		Success: true,
		ID:      templateID,
		Name:    req.Name,
	})
}

// TODO: not decided where to store the template files
func RunCreateTemplate(opt AppCreateOptions) error {
	log.WithFields(log.Fields{
		"app-url":      opt.AppsCloneOpts.URL(),
		"app-revision": opt.AppsCloneOpts.Revision(),
		"app-path":     opt.AppsCloneOpts.Path(),
	}).Debug("starting with options: create template")

	return nil
}

// detectEnvironments analyzes the path to identify environments
func detectEnvironments(path string) []string {
	environments := make([]string, 0)
	envPatterns := map[string]*regexp.Regexp{
		"SIT": regexp.MustCompile(`(?i)(^|\W)sit($|\W)`),
		"UAT": regexp.MustCompile(`(?i)(^|\W)(uat|staging)($|\W)`),
		"PRD": regexp.MustCompile(`(?i)(^|\W)(prd|prod|production)($|\W)`),
	}

	for env, pattern := range envPatterns {
		if pattern.MatchString(path) {
			environments = append(environments, env)
		}
	}

	return environments
}

// storeTemplate saves the template to persistent storage
func storeTemplate(template *ApplicationTemplate) error {
	// TODO: Implement database storage
	return nil
}

// updateTemplate updates an existing template in storage
func updateTemplate(template *ApplicationTemplate) error {
	// TODO: Implement database update
	return nil
}

// generateTemplateID generates a unique template ID
func generateTemplateID() int {
	// TODO: Implement proper ID generation
	return 1
}

// validateTemplateResources performs detailed validation of template resources
func validateTemplateResources(template *ApplicationTemplate) error {
	// TODO: Implement detailed template validation
	return nil
}
