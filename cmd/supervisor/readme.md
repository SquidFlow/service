# Bootstrap CLI

Bootstrap is a command-line tool for initializing and managing the H4 Platform environment.

## Quick Start

### bootstrap the gitops repo

```shell
~/w/g/s/g/h/s/output *main> ./supervisor bootstrap --git-token github_pat_******** --repo https://github.com/h4-poc/gitops2.git --app github.com/h4-poc/service/manifests/base
DEBU[0000] start clone options
INFO[0000] building bootstrap manifests                  app=github.com/h4-poc/service/manifests/base bootstrapAppsLabels="map[]" labels="map[]" namespace=argocd namespaceLabels="map[]" path= repo="https://github.com/h4-poc/gitops2.git" revision=
INFO[0010] cloning repo: https://github.com/h4-poc/gitops2.git
INFO[0014] repository 'https://github.com/h4-poc/gitops2.git' was not found, trying to create it...
INFO[0015] empty repository, initializing a new one with specified remote
INFO[0015] using revision: "", installation path: ""
INFO[0015] using context: "minikube", namespace: "argocd"
INFO[0015] applying bootstrap manifests to cluster...
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

INFO[0081] pushing bootstrap manifests to repo
Resolving deltas: 100% (1/1), done.
INFO[0083] applying argo-cd bootstrap application
W1106 18:58:13.108766   72537 warnings.go:70] metadata.finalizers: "resources-finalizer.argocd.argoproj.io": prefer a domain-qualified finalizer name to avoid accidental conflicts with other finalizer writers
application.argoproj.io/h4-bootstrap created
INFO[0083]
INFO[0083] argocd initialized. password: 16oADdAuGFslTfBz
INFO[0083] run:

    kubectl port-forward -n argocd svc/argocd-server 8080:80
```

for now the bootstrap include `argocd`

```shell
~/w/g/s/g/h/service *main> kubectl get pod --all-namespaces
NAMESPACE     NAME                                               READY   STATUS    RESTARTS   AGE
argocd        argocd-application-controller-0                    1/1     Running   0          5m13s
argocd        argocd-applicationset-controller-7d7c89f5f-vfsvp   1/1     Running   0          5m14s
argocd        argocd-dex-server-77b75c4cff-l8bgm                 1/1     Running   0          5m14s
argocd        argocd-notifications-controller-64775dbfc4-swbhh   1/1     Running   0          5m14s
argocd        argocd-redis-7d85c5d7b8-fhjl7                      1/1     Running   0          5m14s
argocd        argocd-repo-server-75bf446df7-ztgvs                1/1     Running   0          5m13s
argocd        argocd-server-7bb58d96d7-swh6j                     1/1     Running   0          5m13s
kube-system   coredns-7db6d8ff4d-7vtzd                           1/1     Running   0          6m20s
kube-system   etcd-minikube                                      1/1     Running   0          6m34s
kube-system   kube-apiserver-minikube                            1/1     Running   0          6m34s
kube-system   kube-controller-manager-minikube                   1/1     Running   0          6m34s
kube-system   kube-proxy-jgl8j                                   1/1     Running   0          6m20s
kube-system   kube-scheduler-minikube                            1/1     Running   0          6m34s
kube-system   storage-provisioner                                1/1     Running   0          6m33s
```

### create a argocd project

```shell
~/w/g/s/g/h/s/output *main> ./supervisor project --git-token github_pat_******** --repo https://github.com/h4-poc/gitops2.git create testing
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/gitops2.git
Enumerating objects: 17, done.
Counting objects: 100% (17/17), done.
Compressing objects: 100% (12/12), done.
Total 17 (delta 1), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0002] using revision: "", installation path: "/"
INFO[0002] pushing new project manifest to repo
INFO[0003] project created: 'testing'

~/w/g/s/g/h/s/output *main> kubectl get appprojects.argoproj.io --all-namespaces
NAMESPACE   NAME      AGE
argocd      default   20m
argocd      testing   32s
```

### list the project

```shell
~/w/g/s/g/h/s/output *main> ./supervisor project --git-token github_pat_****** --repo https://github.com/h4-poc/gitops2.git  list
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/gitops2.git
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Compressing objects: 100% (14/14), done.
Total 18 (delta 2), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0001] using revision: "", installation path: "/"
NAME     NAMESPACE  DEFAULT CLUSTER
testing  argocd     https://kubernetes.default.svc
```

### delete project

```shell
~/w/g/s/g/h/s/output *main> ./supervisor project --git-token github_pat_****** --repo https://github.com/h4-poc/gitops2.git  delete testing
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/gitops2.git
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Compressing objects: 100% (14/14), done.
Total 18 (delta 2), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0002] using revision: "", installation path: "/"
INFO[0002] committing changes to gitops repo...

~/w/g/s/g/h/s/output *main> ./supervisor project --git-token ****** --repo https://github.com/h4-poc/gitops2.git  list
DEBU[0000] start clone options
INFO[0000] cloning git repository: https://github.com/h4-poc/gitops2.git
Enumerating objects: 17, done.
Counting objects: 100% (17/17), done.
Compressing objects: 100% (12/12), done.
Total 17 (delta 1), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0001] using revision: "", installation path: "/"
NAME  NAMESPACE  DEFAULT CLUSTER
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
~/w/g/s/g/h/service *main> kubectl get pod --all-namespaces
NAMESPACE     NAME                               READY   STATUS    RESTARTS   AGE
kube-system   coredns-7db6d8ff4d-jbcj7           1/1     Running   0          73s
kube-system   etcd-minikube                      1/1     Running   0          89s
kube-system   kube-apiserver-minikube            1/1     Running   0          88s
kube-system   kube-controller-manager-minikube   1/1     Running   0          88s
kube-system   kube-proxy-lnrn6                   1/1     Running   0          73s
kube-system   kube-scheduler-minikube            1/1     Running   0          88s
kube-system   storage-provisioner                1/1     Running   0          87s

~/w/g/s/g/h/s/output main> ./supervisor bootstrap  --git-token ****** --repo https://github.com/h4-poc/gitops2.git --recover
DEBU[0000] start clone options
INFO[0000] starting with options:                        app=github.com/h4-poc/service/manifests/base kube-context=minikube namespace=argocd repo-url="https://github.com/h4-poc/gitops2.git" revision=
INFO[0027] cloning repo: https://github.com/h4-poc/gitops2.git
Enumerating objects: 17, done.
Counting objects: 100% (17/17), done.
Compressing objects: 100% (12/12), done.
Total 17 (delta 1), reused 17 (delta 1), pack-reused 0 (from 0)
INFO[0030] using revision: "", installation path: ""
INFO[0030] using context: "minikube", namespace: "argocd"
INFO[0030] applying bootstrap manifests to cluster...
namespace/argocd configured
customresourcedefinition.apiextensions.k8s.io/applications.argoproj.io unchanged
customresourcedefinition.apiextensions.k8s.io/applicationsets.argoproj.io unchanged
customresourcedefinition.apiextensions.k8s.io/appprojects.argoproj.io unchanged
serviceaccount/argocd-application-controller configured
serviceaccount/argocd-applicationset-controller configured
serviceaccount/argocd-dex-server configured
serviceaccount/argocd-notifications-controller configured
serviceaccount/argocd-redis configured
serviceaccount/argocd-repo-server configured
serviceaccount/argocd-server configured
role.rbac.authorization.k8s.io/argocd-application-controller configured
role.rbac.authorization.k8s.io/argocd-applicationset-controller configured
role.rbac.authorization.k8s.io/argocd-dex-server configured
role.rbac.authorization.k8s.io/argocd-notifications-controller configured
role.rbac.authorization.k8s.io/argocd-redis configured
role.rbac.authorization.k8s.io/argocd-server configured
clusterrole.rbac.authorization.k8s.io/argocd-application-controller configured
clusterrole.rbac.authorization.k8s.io/argocd-applicationset-controller configured
clusterrole.rbac.authorization.k8s.io/argocd-server configured
rolebinding.rbac.authorization.k8s.io/argocd-application-controller configured
rolebinding.rbac.authorization.k8s.io/argocd-applicationset-controller configured
rolebinding.rbac.authorization.k8s.io/argocd-dex-server configured
rolebinding.rbac.authorization.k8s.io/argocd-notifications-controller configured
rolebinding.rbac.authorization.k8s.io/argocd-redis configured
rolebinding.rbac.authorization.k8s.io/argocd-server configured
clusterrolebinding.rbac.authorization.k8s.io/argocd-application-controller configured
clusterrolebinding.rbac.authorization.k8s.io/argocd-applicationset-controller configured
clusterrolebinding.rbac.authorization.k8s.io/argocd-server configured
configmap/argocd-cm configured
configmap/argocd-cmd-params-cm configured
configmap/argocd-gpg-keys-cm configured
configmap/argocd-notifications-cm configured
configmap/argocd-rbac-cm configured
configmap/argocd-ssh-known-hosts-cm configured
configmap/argocd-tls-certs-cm configured
secret/argocd-notifications-secret configured
secret/argocd-secret configured
service/argocd-applicationset-controller configured
service/argocd-dex-server configured
service/argocd-metrics configured
service/argocd-notifications-controller-metrics configured
service/argocd-redis configured
service/argocd-repo-server configured
service/argocd-server configured
service/argocd-server-metrics configured
deployment.apps/argocd-applicationset-controller configured
deployment.apps/argocd-dex-server configured
deployment.apps/argocd-notifications-controller configured
deployment.apps/argocd-redis configured
deployment.apps/argocd-repo-server configured
deployment.apps/argocd-server configured
statefulset.apps/argocd-application-controller configured
networkpolicy.networking.k8s.io/argocd-application-controller-network-policy configured
networkpolicy.networking.k8s.io/argocd-applicationset-controller-network-policy configured
networkpolicy.networking.k8s.io/argocd-dex-server-network-policy configured
networkpolicy.networking.k8s.io/argocd-notifications-controller-network-policy configured
networkpolicy.networking.k8s.io/argocd-redis-network-policy configured
networkpolicy.networking.k8s.io/argocd-repo-server-network-policy configured
networkpolicy.networking.k8s.io/argocd-server-network-policy configured
secret/h4-secret configured

INFO[0033] applying argo-cd bootstrap application
application.argoproj.io/h4-bootstrap configured
INFO[0033]
INFO[0033] argocd initialized. password: hIXr0K3Req4bcRkw
INFO[0033] run:

    kubectl port-forward -n argocd svc/argocd-server 8080:80
```