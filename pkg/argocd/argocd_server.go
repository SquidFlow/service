package argocd

import (
	"context"

	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	sessionpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/cli"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/log"
)

// TODO: in-cluster mode
func GetArgoServerClient() argocdclient.Client {
	opts := argocdclient.ClientOptions{
		ServerAddr: viper.GetString("argocd.server_address"),
		PlainText:  false,
		Insecure:   true,
		GRPCWeb:    true,
	}

	opts.AuthToken = passwordLogin(
		context.Background(),
		argocdclient.NewClientOrDie(&opts),
		viper.GetString("argocd.username"),
		viper.GetString("argocd.password"),
	)

	return argocdclient.NewClientOrDie(&opts)
}

// passwordLogin performs the login and returns the token
func passwordLogin(ctx context.Context, acdClient argocdclient.Client, username, password string) string {
	username, password = cli.PromptCredentials(username, password)
	sessConn, sessionIf := acdClient.NewSessionClientOrDie()
	defer io.Close(sessConn)
	sessionRequest := sessionpkg.SessionCreateRequest{
		Username: username,
		Password: password,
	}
	createdSession, err := sessionIf.Create(ctx, &sessionRequest)
	if err != nil {
		log.G().Fatalf("Failed to login: %v", err)
	}
	log.G().Infof("connected to argocd server: %s with user: %s successfully", viper.GetString("argocd.server_address"), username)
	return createdSession.Token
}
