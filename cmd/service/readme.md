# H4 Platform Service CLI

The Service CLI is a command-line tool for running and managing the H4 Platform API server.

## Quick Start

1. Set environment variable (optional):

```shell
export SERVICE_CONFIG=deploy/service/templates/config.toml
```
If not set, use `-c <config-path>` in commands.

2. Run the API server:

```shell
service run
```

## Available Commands

- `run`: Start the H4 Platform API server
- `version`: Display the current version of the service

## Detailed Usage

### Configuration

Before running the service, ensure your configuration file is properly set up. The default location is `deploy/service/templates/config.toml`.

### Running the API Server

To start the API server:

```shell
service -c /path/to/config.toml run
```

### Checking Version

To display the current version of the service:

```shell
service version
```

## Examples

1. Run the service with a custom config file:

```shell
service -c /path/to/custom/config.toml run
```

2. Check the service version:

```shell
service version
```

## API Endpoints

The service provides several RESTful API endpoints:

- `/api/v1/repos`: Manage repositories
- `/api/v1/applications`: Manage applications
- `/api/v1/project`: Manage project
- `/api/v1/status`: check H4 status

For detailed API documentation, please refer to the API specification in the main documentation.

## Notes

- The service relies on a properly configured Kubernetes environment.

## Troubleshooting

If you encounter issues:

1. Check the service logs for error messages.
2. Verify that your configuration file is correct.
3. Ensure that all required services (Kubernetes cluster, etc.) are running and accessible.

For more detailed information and advanced usage, please refer to the main documentation.
