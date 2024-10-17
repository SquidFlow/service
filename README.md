# H4 Platform Service

H4 Platform is a Kubernetes-based computing platform supporting `multi-tenancy`, `multi-cloud`, and `big data component integration`.

## Key Features

- Kubernetes-based, leveraging cloud-native technologies
- GitOps-style deployments with ArgoCD
- Custom H4 components for business-specific needs (e.g., approval, billing)
- User-friendly web UI for easy management
- RESTful API backend for core functionalities

## Architecture

H4 Platform Architecture Diagram:

```mermaid
graph TD
User((user)) --> |user| H4UI[H4 Web UI]
H4UI --> |API request| H4Backend[H4 Backend Service]
H4Backend --> |create/update PR| GitHubRepo[GitHub Application Repo]
GitHubRepo --> |watch update| ArgoCD[ArgoCD]
ArgoCD --> |deploy| K8s[Kubernetes Cluster]
H4Backend --> |validate| TargetRepo[Target Repo]
TargetRepo --> |source support| Helm[Helm Charts]
subgraph H4 Platform
H4UI
H4Backend
GitHubRepo
ArgoCD
K8s
end
style H4 Platform fill:#f0f0f0,stroke:#333,stroke-width:2px
style User fill:#85C1E9,stroke:#333,stroke-width:2px
style TargetRepo fill:#82E0AA,stroke:#333,stroke-width:2px
style Helm fill:#F8C471,stroke:#333,stroke-width:2px
```


## Components

1. **H4 Web UI**: User-friendly interface for platform management
2. **H4 Backend Service**: Core component providing RESTful APIs
3. **GitHub Application Repo**: Central repository for deployment configurations
4. **ArgoCD**: Manages GitOps-style deployments
5. **Kubernetes Cluster**: Underlying infrastructure for running applications

## Command-Line Tools

H4 Platform provides two main command-line tools:

1. **supervisor CLI** (`supervisor`):
   - Purpose: Initialize and manage application deployments
   - Key functions: Platform initialization, project management, status checks
   - Location: `cmd/supervisor/supervisor.go`

2. **Server CLI** (`server`):
   - Purpose: Run the API server for the H4 Platform
   - Key functions: Start API server, display version information
   - Location: `cmd/service/service.go`

## Getting Started

To build the command-line tools, use the following make commands:

## build

```shell
$ make build-service # Builds the server CLI
$ make build-supervisor # Builds the supervisor CLI
$ make build # Builds both CLIs
```