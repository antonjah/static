# static

A lightweight HTTP mock service with Kubernetes operator for dynamic endpoint configuration. Perfect for testing, development, and API mocking scenarios.

## Quick Start

### Kubernetes

The operator and static service images are automatically published to Docker Hub and will be pulled automatically during installation. The Helm chart is published to GitHub Container Registry (OCI).

Install the operator using Helm (from OCI registry):

```bash
helm install static-operator oci://ghcr.io/antonjah/static-operator
```

Or install a specific version:

```bash
helm install static-operator oci://ghcr.io/antonjah/static-operator --version 2.1.0
```

Or install from local chart:

```bash
helm install static-operator ./deployments/helm/static-operator
```

Deploy the static service:

```bash
kubectl apply -f deployments/examples/static-example.yaml
```

Create mock endpoints:

```bash
kubectl apply -f deployments/examples/staticapi-example.yaml
```

Access the service:

```bash
kubectl port-forward svc/static 8080:80
curl http://localhost:8080/api/v1/info
```

### Docker Compose

For local development without Kubernetes:

```bash
docker-compose up
curl http://localhost:8080/hello
```

## Architecture

The project consists of two components:

1. **Static Service** (`cmd/static`): HTTP server that serves mock responses
   - In Kubernetes: Watches StaticAPI CRDs directly via Kubernetes API (no ConfigMap needed)
   - In Docker/file mode: Watches local `staticapis.yaml` file for changes
   - Supports both HTTP and HTTPS with optional mTLS
   - Configuration updates are applied instantly (within 5 seconds in Kubernetes)

2. **Static Operator** (`cmd/operator`): Kubernetes controller managing Static resources
   - Creates and manages Deployments and Services for Static CRDs
   - Handles TLS secret mounting when using Kubernetes secrets
   - Does NOT manage StaticAPI resources (static service watches them directly)

### Custom Resource Definitions (CRDs)

#### Static CR

Deploys and manages the static HTTP service.

```yaml
apiVersion: static.io/v1alpha1
kind: Static
metadata:
  name: static
  namespace: default
spec:
  replicas: 1
  image: antonjah/static:latest
  logLevel: info
  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 50m
      memory: 64Mi
```

**Spec Fields:**

- `replicas`: Number of pod replicas (default: 1)
- `image`: Container image for static service
- `logLevel`: Log level - debug, info, warn, error (default: info)
- `logPretty`: Enable pretty-printed logs instead of JSON (default: false)
- `resources`: Kubernetes resource requests/limits
- `tls`: TLS configuration (optional)
  - `enabled`: Enable TLS (default: false)
  - `certificate`: Path to certificate file
  - `key`: Path to key file
  - `ca`: Path to CA certificate
  - `verifyClient`: Enable client certificate verification (default: false)
  - `secretName`: Kubernetes Secret containing TLS files (alternative to file paths)

#### StaticAPI CR

Defines mock HTTP endpoints with hot-reload support.

```yaml
apiVersion: static.io/v1alpha1
kind: StaticAPI
metadata:
  name: hello-endpoint
  namespace: default
spec:
  path: /hello
  methods:
  - method: GET
    statusCode: 200
    body: "Hello, World!"
    headers:
      content-type: "text/plain"
  - method: POST
    statusCode: 201
    body: '{"message": "Created"}'
    headers:
      content-type: "application/json"
```

**Spec Fields:**

- `path`: HTTP path (e.g., `/api/users`)
- `methods`: Array of HTTP method configurations
  - `method`: HTTP method
  - `statusCode`: HTTP status code (100-599)
  - `body`: Response body
  - `headers`: HTTP response headers

## TLS Configuration

### File-based TLS

```yaml
apiVersion: static.io/v1alpha1
kind: Static
metadata:
  name: static-tls
  namespace: default
spec:
  replicas: 1
  image: antonjah/static:latest
  tls:
    enabled: true
    certificate: /certs/tls.crt
    key: /certs/tls.key
    ca: /certs/ca.crt
    verifyClient: true
```

### Secret-based TLS

The operator automatically mounts TLS certificates from Kubernetes secrets. When `verifyClient: true` is set, the CA certificate is automatically loaded from `ca.crt` in the secret.

```bash
# Create TLS Secret (for mTLS, include ca.crt)
kubectl create secret generic static-tls \
  --from-file=tls.crt=/path/to/cert.pem \
  --from-file=tls.key=/path/to/key.pem \
  --from-file=ca.crt=/path/to/ca.pem
```

```yaml
# Simple TLS
apiVersion: static.io/v1alpha1
kind: Static
metadata:
  name: static-tls
  namespace: default
spec:
  replicas: 1
  image: antonjah/static:latest
  tls:
    enabled: true
    secretName: static-tls

# mTLS (client certificate verification)
---
apiVersion: static.io/v1alpha1
kind: Static
metadata:
  name: static-mtls
  namespace: default
spec:
  replicas: 1
  image: antonjah/static:latest
  tls:
    enabled: true
    secretName: static-tls
    verifyClient: true  # Automatically uses ca.crt from secret
```

## Environment Variables

The static service supports configuration via environment variables:

| Variable          | Default     | Description                                              |
|:------------------|:------------|:---------------------------------------------------------|
| HOSTNAME          | 0.0.0.0     | Bind hostname                                            |
| PORT              | 8080        | Bind port                                                |
| LOG_LEVEL         | info        | Log level (debug, info, warn, error)                     |
| LOG_PRETTY        | false       | Pretty-print logs (JSON by default)                      |
| STATICAPIS_PATH   | /config     | Path to staticapis.yaml configuration file               |
| TLS_ENABLED       | false       | Enable TLS                                               |
| TLS_CERTIFICATE   |             | Path to TLS certificate                                  |
| TLS_KEY           |             | Path to TLS key                                          |
| TLS_CA            |             | Path to CA certificate                                   |
| TLS_VERIFY_CLIENT | false       | Enable client certificate verification                   |

## Examples

### Multiple Methods on Same Path

```yaml
apiVersion: static.io/v1alpha1
kind: StaticAPI
metadata:
  name: user-api
  namespace: default
spec:
  path: /api/users
  methods:
  - method: GET
    statusCode: 200
    body: '[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]'
    headers:
      content-type: "application/json"
  - method: POST
    statusCode: 201
    body: '{"id":3,"name":"New User"}'
    headers:
      content-type: "application/json"
  - method: DELETE
    statusCode: 204
```

### Error Responses

```yaml
apiVersion: static.io/v1alpha1
kind: StaticAPI
metadata:
  name: error-endpoint
  namespace: default
spec:
  path: /error
  methods:
  - method: GET
    statusCode: 500
    body: '{"error": "Internal Server Error"}'
    headers:
      content-type: "application/json"
```

### Teapot Response

```yaml
apiVersion: static.io/v1alpha1
kind: StaticAPI
metadata:
  name: teapot
  namespace: default
spec:
  path: /tea/pot
  methods:
  - method: GET
    statusCode: 418
    body: "I'm a teapot"
    headers:
      content-type: "text/plain"
```

## Helm Chart

Install using Helm:

```bash
helm install static-operator ./deployments/helm/static-operator
```

Customize installation (different namespace, custom settings):

```bash
helm install static-operator ./deployments/helm/static-operator -n my-namespace --create-namespace \
  --set operator.image.tag=v1.0.0 \
  --set operator.replicaCount=2 \
  --set operator.leaderElect=true
```

### Uninstalling

```bash
# Uninstall the operator
helm uninstall static-operator
```

**Note**: CRDs are cluster-scoped and persist after uninstall. To fully remove:

```bash
kubectl delete crd statics.static.io staticapis.static.io
```
