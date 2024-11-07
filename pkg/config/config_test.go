package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	// Test case 1: Valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		validConfig := `
[log]
level = "debug"

[server]
address = "0.0.0.0"
port = 9090

[argocd]
server_address = "localhost:8080"
username = "admin"
password = "password123"

[application_repo]
provider = ["github"]
remote_url = "https://github.com/example/repo.git"
access_token = "token123"
`
		_, err = tmpfile.Write([]byte(validConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		config, err := ParseConfig(tmpfile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "debug", config.Log.Level)
		assert.Equal(t, 9090, config.Server.Port)
		assert.Equal(t, []string{"github"}, config.ApplicationRepo.Provider)
		assert.Equal(t, "https://github.com/example/repo.git", config.ApplicationRepo.RemoteURL)
		assert.Equal(t, "token123", config.ApplicationRepo.AccessToken)
	})

	// Test case 2: Invalid log level
	t.Run("Invalid Log Level", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidConfig := `
[log]
level = "invalid"

[server]
address = "0.0.0.0"
port = 8080

[argocd]
server_address = "localhost:8080"
username = "admin"
password = "password123"

[application_repo]
provider = ["github"]
remote_url = "https://github.com/example/repo.git"
access_token = "token123"
`
		_, err = tmpfile.Write([]byte(invalidConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		_, err = ParseConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config validation failed")
	})

	// Test case 3: Invalid server port
	t.Run("Invalid Server Port", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "config.*.toml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidConfig := `
[log]
level = "info"

[server]
address = "0.0.0.0"
port = 70000

[argocd]
server_address = "localhost:8080"
username = "admin"
password = "password123"

[application_repo]
provider = ["github"]
remote_url = "https://github.com/example/repo.git"
access_token = "token123"
`
		_, err = tmpfile.Write([]byte(invalidConfig))
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		_, err = ParseConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config validation failed")
	})
}
