# Bootstrap CLI

Bootstrap is a command-line tool for initializing and managing the H4 Platform environment.

## Quick Start

### bootstrap the gitops repo

```shell
➜  service git:(main) ✗ ./output/supervisor bootstrap --git-token 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 --repo https://github.com/h4-poc/application-repo.git
DEBU[0000] start clone options
WARN[0000] detected local bootstrap manifests, using 'normal' installation mode
INFO[0001] cloning repo: https://github.com/h4-poc/application-repo.git
INFO[0001] empty repository, initializing a new one with specified remote
INFO[0002] using revision: "", installation path: ""
INFO[0002] using context: "minikube", namespace: "argocd"
INFO[0002] applying bootstrap manifests to cluster...
namespace/argocd configured
...
secret/h4-secret configured

INFO[0005] pushing bootstrap manifests to repo
Resolving deltas: 100% (1/1), done.
INFO[0006] applying argo-cd bootstrap application
application.argoproj.io/h4-bootstrap configured
INFO[0006] running argocd login to initialize argocd config
INFO[0006]
INFO[0006] argocd initialized. password: HDnTdp0GCJEgLc5T
INFO[0006] run:

    kubectl port-forward -n argocd svc/argocd-server 8080:80


```

### create a argocd project

```shell
➜  service git:(main) ✗ ./output/supervisor project --git-token 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 --repo https://github.com/h4-poc/application-repo.git  create testing
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/application-repo.git
Enumerating objects: 17, done.
Counting objects: 100% (17/17), done.
Compressing objects: 100% (13/13), done.
Total 17 (delta 1), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0002] using revision: "", installation path: "/"
INFO[0002] pushing new project manifest to repo
INFO[0003] project created: 'testing'# create a new project

➜  service git:(main) ✗ ./output/supervisor --kubeconfig ~/.kube/config status
DEBU[0000] start clone options
COMPONENT   STATUS   DETAILS
Kubernetes  Healthy  Version: v1.30.0
ArgoCD      Healthy  Ready replicas: 1/1

```

### delete the project

```shell

➜  service git:(main) ✗ ./output/supervisor project --git-token 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 --repo https://github.com/h4-poc/application-repo.git  delete testing
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/application-repo.git
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Compressing objects: 100% (15/15), done.
Total 18 (delta 2), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0002] using revision: "", installation path: "/"
INFO[0002] committing changes to gitops repo...
```

### list the project

```shell
➜  service git:(main) ✗ ./output/supervisor project --git-token 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 --repo https://github.com/h4-poc/application-repo.git  create testing1
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/application-repo.git
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Compressing objects: 100% (15/15), done.
Total 18 (delta 2), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0002] using revision: "", installation path: "/"
INFO[0002] pushing new project manifest to repo
INFO[0003] project created: 'testing1'

➜  service git:(main) ✗ ./output/supervisor project --git-token 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 --repo https://github.com/h4-poc/application-repo.git  list
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/application-repo.git
Enumerating objects: 19, done.
Counting objects: 100% (19/19), done.
Compressing objects: 100% (16/16), done.
Total 19 (delta 3), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0001] using revision: "", installation path: "/"
NAME      NAMESPACE  DEFAULT CLUSTER
testing   argocd     https://kubernetes.default.svc
testing1  argocd     https://kubernetes.default.svc
```

For more detailed information, please refer to the main documentation.
