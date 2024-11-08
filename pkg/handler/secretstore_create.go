package handler

import (
	"context"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
)

// SecretStoreRequest 定义创建 SecretStore 的请求结构
type SecretStoreRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Provider  string            `json:"provider" binding:"required"` // aws/vault/azure 等
	Auth      map[string]string `json:"auth" binding:"required"`     // 认证信息
}

func CreateSecretStore(c *gin.Context) {
	var req SecretStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	var gitOpsFs = memfs.New()
	var opt = AppCreateOptions{
		CloneOpts: &git.CloneOptions{
			Repo:     viper.GetString("application_repo.remote_url"),
			FS:       fs.Create(gitOpsFs),
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
		ProjectName: req.Namespace,
		Timeout:     0,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	if err := createSecretStoreYAML(context.Background(), opt); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create application: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "SecretStore created successfully",
	})
}

// createSecretStoreYAML generates the YAML representation of a SecretStore
func createSecretStoreYAML(ctx context.Context, opt AppCreateOptions) error {

	// 创建 SecretStore CR
	secretStore := &esv1beta1.SecretStore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "external-secrets.io/v1beta1",
			Kind:       "SecretStore",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opt.createOpts.AppName,
			Namespace: opt.ProjectName,
		},
		Spec: esv1beta1.SecretStoreSpec{},
	}

	log.WithFields(log.Fields{
		"secret-store": secretStore,
	}).Info("secret store")

	return nil
}
