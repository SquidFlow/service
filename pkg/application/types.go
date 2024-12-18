package application

import (
	"errors"

	"github.com/squidflow/service/pkg/kube"
)

// AppCreateOptions represents options for creating an application
type AppCreateOptions struct {
	ProjectName     string
	KubeContextName string
	AppOpts         *CreateOptions
	KubeFactory     kube.Factory
	Labels          map[string]string
	Annotations     map[string]string
	Include         string
	Exclude         string
	DryRun          bool
}

// Errors
var (
	ErrEmptyAppSpecifier            = errors.New("empty app not allowed")
	ErrEmptyAppName                 = errors.New("app name can not be empty, please specify application name")
	ErrEmptyProjectName             = errors.New("project name can not be empty, please specificy project name with: --project")
	ErrAppAlreadyInstalledOnProject = errors.New("application already installed on project")
	ErrAppCollisionWithExistingBase = errors.New("an application with the same name and a different base already exists, consider choosing a different name")
	ErrUnknownAppType               = errors.New("unknown application type")
)
