package repo

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetOrCreateLocalAppRepo() error {
	scmProvider := viper.GetStringSlice("application_repo.provider")
	scmLocalPath := viper.GetString("application_repo.local_path")
	scmRemoteURL := viper.GetString("application_repo.remote_url")
	scmAccessToken := viper.GetString("application_repo.access_token")
	log.WithFields(log.Fields{
		"provider":     scmProvider,
		"local_path":   scmLocalPath,
		"remote_url":   scmRemoteURL,
		"access_token": scmAccessToken,
	}).Info("scm info")
	return nil
}
