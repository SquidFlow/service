package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	// Test case 1: Valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		validConfig := `
log_level = "debug"
[server]
address = "0.0.0.0"
port = 9090
[kubernetes]
mode = "kubeconfig"
[kubernetes.kubeconfig]
path = "/home/user/.kube/config"
[application_repo]
provider = ["github"]
local_path = "/path/to/repo"
remote_url = "https://github.com/example/repo.git"
access_token = "token123"
`
		_, err = tmpfile.Write([]byte(validConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		config, err := ParseConfig(tmpfile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, 9090, config.Server.Port)
		assert.Equal(t, KubeconfigMode, config.Kubernetes.Mode)
		assert.Equal(t, "/home/user/.kube/config", config.Kubernetes.KubeConfig.Path)
		assert.Equal(t, []string{"github"}, config.ApplicationRepo.Provider)
		assert.Equal(t, "/path/to/repo", config.ApplicationRepo.LocalPath)
		assert.Equal(t, "https://github.com/example/repo.git", config.ApplicationRepo.RemoteURL)
		assert.Equal(t, "token123", config.ApplicationRepo.AccessToken)
	})

	// Test case 2: Missing required fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidConfig := `
log_level = "info"
[server]
address = "0.0.0.0"
`
		_, err = tmpfile.Write([]byte(invalidConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		_, err = ParseConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server port is required")
	})

	// Test case 3: Invalid Kubernetes mode
	t.Run("Invalid Kubernetes Mode", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidConfig := `
log_level = "info"
[server]
port = 8080
[kubernetes]
mode = "invalid"
[application_repo]
provider = ["github"]
local_path = "/path/to/repo"
remote_url = "https://github.com/example/repo.git"
access_token = "token123"
`
		_, err = tmpfile.Write([]byte(invalidConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		_, err = ParseConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid kubernetes mode: must be either 'kubeconfig' or 'incluster'")
	})

	// Test case 4: Missing kubeconfig path when mode is kubeconfig
	t.Run("Missing Kubeconfig Path", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidConfig := `
log_level = "info"
[server]
port = 8080
[kubernetes]
mode = "kubeconfig"
[application_repo]
provider = ["github"]
local_path = "/path/to/repo"
remote_url = "https://github.com/example/repo.git"
`
		_, err = tmpfile.Write([]byte(invalidConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		_, err = ParseConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Equal(t, "https://github.com/example/repo.git", viper.GetString("application_repo.remote_url"))
		assert.Contains(t, err.Error(), "kubernetes kubeconfig path is required when mode is kubeconfig")
	})

}
