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


## recover from gitops repo

bootstrap everything from repo

```shell
➜  service git:(dev/sn0rt) ✗ minikube delete
🔥  Deleting "minikube" in docker ...
🔥  Deleting container "minikube" ...
🔥  Removing /Users/guohao/.minikube/machines/minikube ...
💀  Removed all traces of the "minikube" cluster.
```

```shell
➜  service git:(dev/sn0rt) ✗ minikube start
😄  minikube v1.33.1 on Darwin 14.5 (arm64)
🎉  minikube 1.34.0 is available! Download it: https://github.com/kubernetes/minikube/releases/tag/v1.34.0
💡  To disable this notice, run: 'minikube config set WantUpdateNotification false'

✨  Automatically selected the docker driver
📌  Using Docker Desktop driver with root privileges
👍  Starting "minikube" primary control-plane node in "minikube" cluster
🚜  Pulling base image v0.0.44 ...
🔥  Creating docker container (CPUs=2, Memory=3885MB) ...
🐳  Preparing Kubernetes v1.30.0 on Docker 26.1.1 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring bridge CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass

❗  /usr/local/bin/kubectl is version 1.25.4, which may have incompatibilities with Kubernetes 1.30.0.
    ▪ Want kubectl v1.30.0? Try 'minikube kubectl -- get pods -A'
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```

```shell
➜  service git:(dev/sn0rt) ✗ ./output/supervisor bootstrap --git-token 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 --repo https://github.com/h4-poc/application-repo.git --recover
DEBU[0000] start clone options
WARN[0000] detected local bootstrap manifests, using 'normal' installation mode
INFO[0001] cloning repo: https://github.com/h4-poc/application-repo.git
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Compressing objects: 100% (14/14), done.
Total 18 (delta 2), reused 16 (delta 2), pack-reused 0 (from 0)
INFO[0003] using revision: "", installation path: ""
INFO[0003] using context: "minikube", namespace: "argocd"
INFO[0003] applying bootstrap manifests to cluster...
namespace/argocd created
customresourcedefinition.apiextensions.k8s.io/applications.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/applicationsets.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/appprojects.argoproj.io created
serviceaccount/argocd-application-controller created
serviceaccount/argocd-applicationset-controller created
serviceaccount/argocd-dex-server created
serviceaccount/argocd-notifications-controller created
serviceaccount/argocd-redis created
serviceaccount/argocd-repo-server created
serviceaccount/argocd-server created
role.rbac.authorization.k8s.io/argocd-application-controller created
role.rbac.authorization.k8s.io/argocd-applicationset-controller created
role.rbac.authorization.k8s.io/argocd-dex-server created
role.rbac.authorization.k8s.io/argocd-notifications-controller created
role.rbac.authorization.k8s.io/argocd-redis created
role.rbac.authorization.k8s.io/argocd-server created
clusterrole.rbac.authorization.k8s.io/argocd-application-controller created
clusterrole.rbac.authorization.k8s.io/argocd-applicationset-controller created
clusterrole.rbac.authorization.k8s.io/argocd-server created
rolebinding.rbac.authorization.k8s.io/argocd-application-controller created
rolebinding.rbac.authorization.k8s.io/argocd-applicationset-controller created
rolebinding.rbac.authorization.k8s.io/argocd-dex-server created
rolebinding.rbac.authorization.k8s.io/argocd-notifications-controller created
rolebinding.rbac.authorization.k8s.io/argocd-redis created
rolebinding.rbac.authorization.k8s.io/argocd-server created
clusterrolebinding.rbac.authorization.k8s.io/argocd-application-controller created
clusterrolebinding.rbac.authorization.k8s.io/argocd-applicationset-controller created
clusterrolebinding.rbac.authorization.k8s.io/argocd-server created
configmap/argocd-cm created
configmap/argocd-cmd-params-cm created
configmap/argocd-gpg-keys-cm created
configmap/argocd-notifications-cm created
configmap/argocd-rbac-cm created
configmap/argocd-ssh-known-hosts-cm created
configmap/argocd-tls-certs-cm created
secret/argocd-notifications-secret created
secret/argocd-secret created
service/argocd-applicationset-controller created
service/argocd-dex-server created
service/argocd-metrics created
service/argocd-notifications-controller-metrics created
service/argocd-redis created
service/argocd-repo-server created
service/argocd-server created
service/argocd-server-metrics created
deployment.apps/argocd-applicationset-controller created
deployment.apps/argocd-dex-server created
deployment.apps/argocd-notifications-controller created
deployment.apps/argocd-redis created
deployment.apps/argocd-repo-server created
deployment.apps/argocd-server created
statefulset.apps/argocd-application-controller created
networkpolicy.networking.k8s.io/argocd-application-controller-network-policy created
networkpolicy.networking.k8s.io/argocd-applicationset-controller-network-policy created
networkpolicy.networking.k8s.io/argocd-dex-server-network-policy created
networkpolicy.networking.k8s.io/argocd-notifications-controller-network-policy created
networkpolicy.networking.k8s.io/argocd-redis-network-policy created
networkpolicy.networking.k8s.io/argocd-repo-server-network-policy created
networkpolicy.networking.k8s.io/argocd-server-network-policy created
secret/h4-secret created

INFO[0099] applying argo-cd bootstrap application
W1027 18:27:29.172436   11510 warnings.go:70] metadata.finalizers: "resources-finalizer.argocd.argoproj.io": prefer a domain-qualified finalizer name to avoid accidental conflicts with other finalizer writers
application.argoproj.io/h4-bootstrap created
INFO[0099] running argocd login to initialize argocd config
'admin:login' logged in successfully
Context 'autopilot' updated
INFO[0100]
INFO[0100] argocd initialized. password: 6DFqWrCRKJuN0Ugv
INFO[0100] run:

    kubectl port-forward -n argocd svc/argocd-server 8080:80
```