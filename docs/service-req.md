# Service API Request Examples

This document provides examples of API requests to the service, along with their responses.

## Projects

### Create a Project

```shell
curl -X POST http://localhost:8080/api/v1/projects \
-H "Content-Type: application/json" \
-d '{"project-name": "demo1"}'
```

Response (project already exists):
```json
{"error":"Failed to create project: project 'demo1' already exists"}
```

### List Projects

```shell
curl -s -X GET http://localhost:8080/api/v1/projects | jq
```

Response:
```json
{
  "projects": [
    {
      "name": "demo1",
      "namespace": "argocd",
      "default_cluster": "https://kubernetes.default.svc"
    },
    {
      "name": "testing",
      "namespace": "argocd",
      "default_cluster": "https://kubernetes.default.svc"
    }
  ]
}
```

### Delete a Project

```shell
curl -X DELETE "http://localhost:8080/api/v1/projects?project=demo1" | jq
```

Response:
```json
{
  "message": "Project 'demo1' deleted successfully"
}
```

## Applications

### Create an Application

```shell
curl -s -X POST http://localhost:8080/api/v1/applications \
-H "Content-Type: application/json" \
-d '{
  "project-name": "testing",
  "app-name": "demo1",
  "app": "github.com/h4-poc/demo-app"
}' | jq
```

Response:
```json
{
  "application": {
    "project-name": "testing",
    "app-name": "demo3",
    "app": "github.com/h4-poc/demo-app",
    "wait-timeout": ""
  },
  "message": "Application created successfully"
}
```

### List Applications in a Project

```shell
curl -s -X GET "http://localhost:8080/api/v1/applications?project=testing" | jq
```

Response:
```json
{
  "project_name": "testing",
  "apps": [
    {
      "name": "demo1",
      "dest_namespace": "default",
      "dest_server": "https://kubernetes.default.svc",
      "creator": "Unknown",
      "last_updater": "Unknown",
      "last_commit_id": "Unknown",
      "last_commit_message": "Unknown",
      "pod_count": 0,
      "secret_count": 0,
      "resource_usage": {
        "cpu_cores": "0",
        "memory_usage": "0"
      },
      "status": "Succeeded",
      "health": "Healthy",
      "sync_status": "Synced"
    }
  ]
}
```

### Delete an Application

```shell
curl -s -X DELETE "http://localhost:8080/api/v1/applications?project=testing&app=demo3" | jq
```

Response:
```json
{
  "message": "Application 'demo3' deleted from project 'testing'"
}
```

Attempting to delete a non-existent application:
```shell
curl -s -X DELETE "http://localhost:8080/api/v1/applications?project=testing&app=demo1" | jq
```

Response:
```json
{
  "error": "Failed to delete application: application 'demo1' not found"
}
```

## Helm Templates

``shell
$ curl -s -X GET "http://localhost:8080/api/v1/helm/templates" | jq
``

```json
[
  {
    "name": "h4-loki",
    "description": "Loki is a horizontally scalable, highly available, multi-tenant log aggregation system inspired by Prometheus. It is designed to be very cost effective and easy to operate. It is the best solution for large-scale microservices based systems.",
    "url": "https://github.com/h4-poc/manifest/blob/main/loki/values.yaml",
    "maintainers": [
      {
        "name": "h4-loki",
        "email": "h4-loki@h4.com"
      }
    ]
  },
  {
    "name": "h4-logging-operator",
    "description": "Logging operator is a tool for managing logging resources in Kubernetes. It is designed to be very cost effective and easy to operate. It is the best solution for large-scale microservices based systems.",
    "url": "https://github.com/h4-poc/manifest/blob/main/logging-operator/values.yaml",
    "maintainers": [
      {
        "name": "h4-logging-operator",
        "email": "h4-logging-operator@h4.com"
      }
    ]
  }
]
```