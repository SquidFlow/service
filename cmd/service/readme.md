# H4 Platform Service CLI

The Service CLI is a command-line tool for running and managing the H4 Platform API server.

## Quick Start

run the service

```shell
./output/service run -c deploy/service/templates/config.toml
```

## API Endpoints

The service provides several RESTful API endpoints:

- `/api/v1/applications`: Manage applications
- `/api/v1/project`: Manage project
- `/api/v1/status`: check H4 status

For detailed API documentation, please refer to the API specification in the main documentation.
