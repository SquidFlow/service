# Application Repository Layout Convention

This document describes the supported repository layouts for applications deployed through our platform. The platform supports two main deployment methods: Kustomize and Helm.

## Kustomize Application Layouts

this section describes the supported Kustomize layouts.

### 1. Simple Single Environment Layout

if find `kustomization.yaml` in root directory, it will be treated as a simple single environment layout.

the layout of the repository should be like this:
```shell
/
├── deployment.yaml        # Deployment configuration
├── service.yaml          # Service configuration
└── kustomization.yaml    # Kustomization configuration
```

### 2. Standard Multi-Environment Layout

if find `base/` and `overlays/` directories, it will be treated as a standard multi-environment layout.
and the overlays directory will be used as the environment-specific overlays.

the layout of the repository should be like this:

```shell
/
├── base/                    # Base configuration directory
│   ├── deployment.yaml      # Base deployment configuration
│   ├── service.yaml        # Base service configuration
│   └── kustomization.yaml  # Base kustomization file
└── overlays/               # Environment-specific overlays
    ├── dev/                # Development environment
    │   ├── kustomization.yaml
    │   └── patch.yaml      # Dev-specific patches
    ├── staging/           # Staging environment
    │   ├── kustomization.yaml
    │   └── patch.yaml     # Staging-specific patches
    └── prod/              # Production environment
        ├── kustomization.yaml
        └── patch.yaml     # Production-specific patches
```



## Helm Application Layouts

this section describes the supported Helm chart layouts.

### 1. Standard Helm Chart Layout

If find `Chart.yaml` and `values.yaml` with root directory, it will be treated as a standard Helm chart.

the layout of the repository should be like this:
```shell
/
├── Chart.yaml            # Chart metadata
├── values.yaml          # Default values
├── templates/           # Template directory
│   ├── deployment.yaml  # Deployment template
│   ├── service.yaml    # Service template
│   └── ingress.yaml    # Ingress template
└── environments/        # Environment-specific values
    ├── dev/            # Development environment
    │   └── values.yaml # Dev-specific values
    ├── staging/        # Staging environment
    │   └── values.yaml # Staging-specific values
    └── prod/           # Production environment
        └── values.yaml # Production-specific values
```

### 2. Versioned Helm Chart Layout

if find `manifests/` and `environments/` under `path` directory, it will be treated as a versioned Helm chart.

the user need to specify the version of the Helm chart to deploy with `helm_manifest_path` field in the `application_specifier` section.

```json
    "application_source": {
        "repo":"git@github.com:SquidFlow/helm-example-app.git",
        "target_revision": "main",
        "path":"/",
        "application_specifier": {
            "helm_manifest_path": "manifests/4.0.0"
        }
    },
```

the layout of the repository should be like this:

```shell
/
├── manifests/          # Versioned manifests directory
│   ├── 1.0.0/          # Version 1.0.0
│   │   └── values.yaml # Version 1.0.0 values
│   └── 2.0.0/          # Version 2.0.0
│       └── values.yaml # Version 2.0.0 values
└── environments/       # Environment-specific values
    ├── dev/            # Development environment
    │   └── values.yaml # Dev-specific values
    ├── staging/        # Staging environment
    │   └── values.yaml # Staging-specific values
    └── prod/           # Production environment
        └── values.yaml # Production-specific values
```

## Validation

You can validate your repository layout using the following API:

```http
PUT /api/v1/deploy/argocdapplications/validate
Content-Type: application/json

{
    "repo": "git@github.com:org/repo.git",
    "target_revision": "main",
    "path": "/",
    "application_specifier": {
        "helm_manifest_path": "manifests/4.0.0"  // Optional, for Helm charts
    }
}
```

## Best Practices

1. Directory Naming
   - Use lowercase letters and hyphens
   - Avoid special characters and spaces
   - Use clear, descriptive names

2. File Organization
   - Keep related configurations together
   - Place common configurations in base/chart root
   - Environment-specific configs in respective directories

3. Version Control
   - All configuration files must be version controlled
   - Use semantic versioning for releases
   - Document significant changes

4. Security
   - Don't store sensitive data in configuration files
   - Use Secrets or ConfigMaps for sensitive data
   - Implement proper access controls

5. Environment Isolation
   - Keep environment configurations separate
   - Use clear environment naming
   - Document environment-specific requirements

## Requirements

1. Kustomize Applications
   - Must have kustomization.yaml in each environment
   - Base directory must contain core resources
   - Overlays should only contain patches

2. Helm Applications
   - Must have valid Chart.yaml
   - Must have values.yaml for each environment
   - Version directories must be complete

## how to debug the application layout

of course, you can use the `helm` or `kustomize` command to validate the application layout without deploying it.

### kustomize

for kustomize, you can use the following command to validate the application layout with single environment:

```shell
kustomize build 'https://github.com/SquidFlow/kustomize-example-app/overlays/sit'
```

the output should be like this:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: sit-simple-service
  namespace: sit
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: trivial-go-web-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sit-simple-deployment
  namespace: sit
spec:
  replicas: 1
  selector:
    matchLabels:
      app: trivial-go-web-app
  template:
    metadata:
      labels:
        app: trivial-go-web-app
    spec:
      containers:
      - image: docker.io/kostiscodefresh/simple-web-app:3d9b390
        name: webserver-simple
        ports:
        - containerPort: 8080
```

### helm

for helm, you can use the following command to validate the application layout with single environment:

for multi-environment, you can use the following command (need download the helm chart first):

```shell
helm template jupyterhub manifests/4.0.0 \
    --values environments/dev/values.yaml \
```
