# H4 Service Helm Chart

This Helm Chart is designed to deploy H4 Service with both frontend and backend components.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Global Configuration](#global-configuration)
  - [Backend Configuration](#backend-configuration)
  - [Frontend Configuration](#frontend-configuration)
  - [Ingress Configuration](#ingress-configuration)
  - [ALB Configuration](#alb-specific-configuration)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Ingress controller (nginx-ingress or Alibaba Cloud ALB Ingress)

## Features

- Frontend and Backend service deployment
- Configurable resource limits
- Support for both Nginx Ingress and Alibaba Cloud ALB
- Configurable autoscaling
- Built-in configuration management via ConfigMap

## Quick Start

### Installation

```shell
helm install h4 helm/chart  \
      --set argocd.password=****** \
      --set applicationRepo.accessToken=******
```

### Uninstallation

```shell
helm uninstall h4
```

### Preview Template

```shell
helm template helm/chart  \
      --set argocd.password=****** \
      --set applicationRepo.accessToken=******
```

## Configuration

### Global Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `image.repository` | Image repository | `docker.io/wangguohao/h4-service` | Yes |
| `image.tag` | Image tag | `e8b887b` | Yes |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` | No |

### Backend Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `backend.replicaCount` | Number of backend replicas | `1` | No |
| `backend.service.type` | Backend service type | `ClusterIP` | No |
| `backend.service.port` | Backend service port | `38080` | Yes |
| `backend.resources.requests.cpu` | CPU requests | `100m` | No |
| `backend.resources.requests.memory` | Memory requests | `128Mi` | No |
| `backend.resources.limits.cpu` | CPU limits | `500m` | No |
| `backend.resources.limits.memory` | Memory limits | `512Mi` | No |
| `backend.env` | Environment variables | `[]` | No |

### Frontend Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `frontend.replicaCount` | Number of frontend replicas | `1` | No |
| `frontend.service.type` | Frontend service type | `ClusterIP` | No |
| `frontend.service.port` | Frontend service port | `80` | Yes |
| `frontend.resources.requests.cpu` | CPU requests | `100m` | No |
| `frontend.resources.requests.memory` | Memory requests | `128Mi` | No |
| `frontend.resources.limits.cpu` | CPU limits | `500m` | No |
| `frontend.resources.limits.memory` | Memory limits | `512Mi` | No |
| `frontend.env` | Environment variables | `[]` | No |

### Ingress Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `ingress.enabled` | Enable ingress | `true` | No |
| `ingress.type` | Ingress type (`nginx` or `alb`) | `nginx` | Yes when ingress enabled |
| `ingress.className` | Ingress class name | `nginx` | Yes when ingress enabled |
| `ingress.annotations` | Ingress annotations | `{}` | No |

### ALB Specific Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `ingress.alb.enabled` | Enable ALB | `false` | No |
| `ingress.alb.annotations` | ALB specific annotations | See values.yaml | No |

## Troubleshooting

### Common Issues

1. **Pod fails to start**
   - Check the pod logs using `kubectl logs <pod-name>`
   - Verify the image pull policy and credentials

2. **Ingress not working**
   - Ensure the ingress controller is properly installed
   - Verify the ingress class name matches your cluster setup

### Getting Support

For issues and feature requests, please create an issue in the project repository.
