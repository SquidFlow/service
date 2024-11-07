package argocd

import (
	"context"

	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	sessionpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/cli"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/spf13/viper"
)

// TODO: in-cluster mode
func GetArgoServerClient() argocdclient.Client {
	opts := argocdclient.ClientOptions{
		ConfigPath:      "",
		ServerAddr:      viper.GetString("argocd.server_address"),
		PlainText:       false,
		GRPCWeb:         true,
		Insecure:        true,
		GRPCWebRootPath: "",
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
	errors.CheckError(err)
	return createdSession.Token
}
