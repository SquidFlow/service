package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
)

// ValidationRequest represents the request structure for template validation
type ValidationRequest struct {
	Source         ApplicationSource `json:"source" binding:"required"`
	Path           string            `json:"path" binding:"required"`
	TargetRevision string            `json:"targetRevision" binding:"required"`
}

// ValidationResult represents the validation result for each environment
type ValidationResult struct {
	Environment []string `json:"environment"` // the repo support multiple environments
	IsValid     bool     `json:"isValid"`
	Message     []string `json:"message,omitempty"`
}

// ValidationResponse represents the response structure for template validation
type ValidationResponse struct {
	Success bool               `json:"success"`
	Error   string             `json:"error,omitempty"`
	Results []ValidationResult `json:"results"`
}

// ValidateApplicationTemplate handles the validation of application templates
func ValidateApplicationTemplate(c *gin.Context) {
	var req ValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ValidationResponse{
			Success: false,
			Results: nil,
		})
		return
	}

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
			AppName:          req.Path,
			AppType:          application.AppTypeKustomize,
			AppSpecifier:     req.Path,
			InstallationMode: application.InstallationModeNormal,
			DestServer:       "https://kubernetes.default.svc",
			Labels:           nil,
			Annotations:      nil,
			Include:          "",
			Exclude:          "",
		},
		ProjectName: req.Path,
		Timeout:     0,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	result, err := validateTemplateParams(opt)
	if err != nil {
		c.JSON(400, ValidationResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(200, ValidationResponse{
		Success: true,
		Results: result,
	})
}

func validateTemplateParams(opt AppCreateOptions) ([]ValidationResult, error) {
	var results []ValidationResult
	return results, nil
}
