# Bootstrap CLI

Bootstrap is a command-line tool for initializing and managing the H4 Platform environment.

## Quick Start

### bootstrap the gitops repo

```shell
➜  service git:(main) ✗ ./output/supervisor bootstrap --git-token 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 --repo https://github.com/h4-poc/application-repo.git
DEBU[0000] start clone options
WARN[0000] detected local bootstrap manifests, using 'normal' installation mode
INFO[0002] cloning repo: https://github.com/h4-poc/application-repo.git
INFO[0004] empty repository, initializing a new one with specified remote
INFO[0005] using revision: "", installation path: ""
INFO[0005] using context: "minikube", namespace: "argocd"
INFO[0005] applying bootstrap manifests to cluster...
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
serviceaccount/vault-agent-injector created
serviceaccount/vault created
clusterrole.rbac.authorization.k8s.io/vault-agent-injector-clusterrole created
clusterrolebinding.rbac.authorization.k8s.io/vault-agent-injector-binding created
clusterrolebinding.rbac.authorization.k8s.io/vault-server-binding created
service/vault-agent-injector-svc created
service/vault-internal created
service/vault created
deployment.apps/vault-agent-injector created
statefulset.apps/vault created
mutatingwebhookconfiguration.admissionregistration.k8s.io/vault-agent-injector-cfg created
pod/vault-server-test created
serviceaccount/external-secrets-cert-controller created
serviceaccount/external-secrets created
serviceaccount/external-secrets-webhook created
secret/external-secrets-webhook created
customresourcedefinition.apiextensions.k8s.io/acraccesstokens.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/clusterexternalsecrets.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/clustersecretstores.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/ecrauthorizationtokens.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/externalsecrets.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/fakes.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/gcraccesstokens.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/githubaccesstokens.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/passwords.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/pushsecrets.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/secretstores.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/vaultdynamicsecrets.generators.external-secrets.io created
customresourcedefinition.apiextensions.k8s.io/webhooks.generators.external-secrets.io created
clusterrole.rbac.authorization.k8s.io/external-secrets-cert-controller created
clusterrole.rbac.authorization.k8s.io/external-secrets-controller created
clusterrole.rbac.authorization.k8s.io/external-secrets-view created
clusterrole.rbac.authorization.k8s.io/external-secrets-edit created
clusterrole.rbac.authorization.k8s.io/external-secrets-servicebindings created
clusterrolebinding.rbac.authorization.k8s.io/external-secrets-cert-controller created
clusterrolebinding.rbac.authorization.k8s.io/external-secrets-controller created
role.rbac.authorization.k8s.io/external-secrets-leaderelection created
rolebinding.rbac.authorization.k8s.io/external-secrets-leaderelection created
service/external-secrets-webhook created
deployment.apps/external-secrets-cert-controller created
deployment.apps/external-secrets created
deployment.apps/external-secrets-webhook created
validatingwebhookconfiguration.admissionregistration.k8s.io/secretstore-validate created
validatingwebhookconfiguration.admissionregistration.k8s.io/externalsecret-validate created
secret/h4-secret created

INFO[0158] pushing bootstrap manifests to repo
Resolving deltas: 100% (1/1), done.
INFO[0159] applying argo-cd bootstrap application
W1027 22:33:50.126897   27132 warnings.go:70] metadata.finalizers: "resources-finalizer.argocd.argoproj.io": prefer a domain-qualified finalizer name to avoid accidental conflicts with other finalizer writers
application.argoproj.io/h4-bootstrap created
INFO[0159]
INFO[0159] argocd initialized. password: Xele0y9mhZyYjJI2
INFO[0159] run:

    kubectl port-forward -n argocd svc/argocd-server 8080:80

```

for now the bootstrap include `argocd`, `external-secret` and `vault`.

```shell
➜  service git:(main) ✗ kubectl get pod --all-namespaces
NAMESPACE     NAME                                               READY   STATUS            RESTARTS   AGE
argocd        argocd-application-controller-0                    1/1     Running           0          4m3s
argocd        argocd-applicationset-controller-75d8c9495-qnff4   1/1     Running           0          4m3s
argocd        argocd-dex-server-7c9b44b9f9-sxh9x                 0/1     PodInitializing   0          4m3s
argocd        argocd-notifications-controller-77f49c7745-b7ggm   1/1     Running           0          4m3s
argocd        argocd-redis-575c96bc4f-plnl4                      1/1     Running           0          4m3s
argocd        argocd-repo-server-7f44b474d7-wvwtn                0/1     PodInitializing   0          4m3s
argocd        argocd-server-5f4dd5d648-vxvgk                     1/1     Running           0          4m3s
argocd        external-secrets-5859d8dc69-cng5s                  1/1     Running           0          4m2s
argocd        external-secrets-cert-controller-7d675fdf6-b6cwq   0/1     Running           0          4m2s
argocd        external-secrets-webhook-6cc4bd4fd4-ggsgc          0/1     Running           0          4m2s
argocd        vault-0                                            1/1     Running           0          4m3s
argocd        vault-agent-injector-8667c5945-275zs               1/1     Running           0          4m3s
kube-system   coredns-7db6d8ff4d-ww68w                           1/1     Running           0          4m3s
kube-system   etcd-minikube                                      1/1     Running           0          4m23s
kube-system   kube-apiserver-minikube                            1/1     Running           0          4m23s
kube-system   kube-controller-manager-minikube                   1/1     Running           0          4m23s
kube-system   kube-proxy-vgj44                                   1/1     Running           0          4m9s
kube-system   kube-scheduler-minikube                            1/1     Running           0          4m23s
kube-system   storage-provisioner                                1/1     Running           0          4m22s
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