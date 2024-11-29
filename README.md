# SquidFlow Platform

SquidFlow Platform is a modern GitOps-based platform that simplifies multi-cluster Kubernetes management and application deployment. It abstracts away the complexity of Kubernetes CRDs and infrastructure configurations, allowing users to focus on their applications rather than the underlying technical details.

## Overview

SquidFlow Platform uses Git repositories as the source of truth for declarative infrastructure and application definitions. Built on top of ArgoCD, it provides an enhanced user experience for managing multiple Kubernetes clusters across different environments. The platform automatically handles the generation and management of complex configurations, significantly reducing the cognitive load on users.

## Core Features

- **GitOps-Driven Architecture**
  - Uses Git repositories as the single source of truth
  - Leverages ArgoCD for automated deployment and synchronization
  - Automated PR creation and management

- **Multi-Cluster Management**
  - Centralized control plane for multiple Kubernetes clusters
  - Environment segregation (Dev, Staging, Production)
  - Cross-cluster resource management without manual intervention

- **Enhanced Security Integration**
  - Automated Vault configuration and secret management
  - External Secrets Operator integration without manual setup

## How It Simplifies Your Work

Traditional Kubernetes deployment requires:
- Deep understanding of CRDs
- Manual YAML configuration
- Complex secret management setup
- ArgoCD configuration expertise

With SquidFlow Platform:
- ✅ Click-to-deploy applications
- ✅ Automatic configuration generation
- ✅ Built-in best practices
- ✅ Simplified multi-cluster management
- ✅ Zero-touch secret management

## Architecture

SquidFlow Platform Architecture Diagram:

```mermaid
graph TD
User((user)) --> |user| SquidFlowUI[SquidFlow Web UI]
SquidFlowUI --> |API request| SquidFlowBackend[SquidFlow Backend Service]
SquidFlowBackend --> |create/update PR| GitHubRepo[GitHub Application Repo]
GitHubRepo --> |watch update| ArgoCD[ArgoCD]
ArgoCD --> |deploy| K8s[Kubernetes Cluster]
SquidFlowBackend --> |validate| TargetRepo[Target Repo]
TargetRepo --> |source support| Helm[Helm Charts]
subgraph SquidFlow Platform
SquidFlowUI
SquidFlowBackend
GitHubRepo
ArgoCD
K8s
end
style User fill:#85C1E9,stroke:#333,stroke-width:2px
style TargetRepo fill:#82E0AA,stroke:#333,stroke-width:2px
style Helm fill:#F8C471,stroke:#333,stroke-width:2px
```


## Components

1. **SquidFlow Web UI**: User-friendly interface for platform management
2. **SquidFlow Backend Service**: Core component providing RESTful APIs
3. **GitHub Application Repo**: Central repository for deployment configurations
4. **ArgoCD**: Manages GitOps-style deployments
5. **Kubernetes Cluster**: Underlying infrastructure for running applications

## Command-Line Tools

SquidFlow Platform provides two main command-line tools:

1. **supervisor CLI** (`supervisor`):
   - Purpose: Initialize and manage application deployments
   - Key functions: Platform initialization, project management, status checks
   - Location: `cmd/supervisor/supervisor.go`

2. **Server CLI** (`server`):
   - Purpose: Run the API server for the SquidFlow Platform
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