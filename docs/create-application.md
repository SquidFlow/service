# Application

## Create (POST)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "app-name": "hello-world",
    "repo": "github.com/argoproj-labs/argocd-autopilot/examples/demo-app/",
    "wait-timeout": "2m"
  }' \
  http://localhost:8080/api/v1/applications

## Read (GET)
# Get a specific application
curl -X GET \
  http://localhost:8080/api/v1/applications/hello-world

# Get all applications with a filter
curl -X GET \
  "http://localhost:8080/api/v1/applications?filter=name=hello-world"

# Get applications with multiple filters
curl -X GET \
  "http://localhost:8080/api/v1/applications?filter=name=hello-world&filter=project=default"

## Update (PUT)
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{
    "app-name": "hello-world",
    "repo": "github.com/argoproj-labs/argocd-autopilot/examples/updated-app/",
    "wait-timeout": "3m"
  }' \
  http://localhost:8080/api/v1/applications/hello-world

## Delete (DELETE)
curl -X DELETE \
  http://localhost:8080/api/v1/applications/hello-world

