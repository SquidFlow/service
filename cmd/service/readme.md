# H4 Platform Service CLI

The Service CLI is a command-line tool for running and managing the H4 Platform API server.

## Quick Start

run the service

```shell
./output/service run -c deploy/service/templates/config.toml
```

## API Endpoints

The service provides the following RESTful API endpoints:

### System Health
- `GET /healthz` - Health check endpoint

### AppCode
- `GET /api/v1/appcode` - List AppCode

### Destination Clusters
- `GET /api/v1/destinationCluster` - List destination clusters
- `POST /api/v1/destinationCluster` - Create a destination cluster
- `PATCH /api/v1/destinationCluster/:name` - Update a destination cluster

### Application Templates
- `GET /api/v1/applications/templates` - List application templates
- `POST /api/v1/applications/templates` - Create an application template
- `POST /api/v1/applications/templates/validate` - Validate an application template
- `PATCH /api/v1/applications/templates/:template_id` - Update a template
- `DELETE /api/v1/applications/templates/:template_id` - Delete a template

### ArgoCD Applications
- `POST /api/v1/deploy/argocdapplications` - Create an ArgoCD application
- `GET /api/v1/deploy/argocdapplications` - List ArgoCD applications
- `POST /api/v1/deploy/argocdapplications/sync` - Sync ArgoCD applications
- `POST /api/v1/deploy/argocdapplications/dryrun` - Dry run ArgoCD applications

Individual Application Operations:
- `GET /api/v1/deploy/argocdapplications/:appName` - Get application details
- `PATCH /api/v1/deploy/argocdapplications/:appName` - Update an application
- `DELETE /api/v1/deploy/argocdapplications/:appName` - Delete an application

### Tenants (ArgoCD Projects)
- `POST /api/v1/tenants` - Create a new tenant
- `GET /api/v1/tenants` - List all tenants
- `GET /api/v1/tenants/:tenantName` - Get tenant details
- `DELETE /api/v1/tenants/:tenantName` - Delete a tenant

### Security Management
External Secrets:
- `POST /api/v1/security/externalsecrets/secretstore` - Create a secret store
- `GET /api/v1/security/externalsecrets/secretstore` - List secret stores
- `PATCH /api/v1/security/externalsecrets/secretstore/:name` - Update a secret store
- `DELETE /api/v1/security/externalsecrets/secretstore/:name` - Delete a secret store

For detailed API documentation, please refer to the API specification in the main documentation.
