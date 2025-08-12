# Enhanced Template Functions for Caddy-Docker-Proxy

This document describes the comprehensive template function system that enables dynamic configuration of ALL Caddy directives using Docker container metadata.

## Overview

Template functions extract data from Docker containers and services to dynamically generate Caddyfile configurations. These functions unlock the full power of Caddy by making container metadata available for any directive, not just reverse proxying.

**Compatibility**: All functions work with both Docker containers (`types.Container`) and Docker Swarm services (`swarm.Service`), with graceful fallback for missing data.

## Existing Functions

### upstreams

Returns all addresses for the current Docker resource separated by whitespace.

**Usage:** `{{upstreams [protocol] [port]}}`

```yaml
labels:
  caddy.reverse_proxy: "{{upstreams}}"           # → 192.168.0.1 192.168.0.2
  caddy.reverse_proxy: "{{upstreams https}}"     # → https://192.168.0.1 https://192.168.0.2  
  caddy.reverse_proxy: "{{upstreams 8080}}"      # → 192.168.0.1:8080 192.168.0.2:8080
  caddy.reverse_proxy: "{{upstreams http 8080}}" # → http://192.168.0.1:8080 http://192.168.0.2:8080
```

### Protocol Helpers

Simple protocol string helpers:

```yaml
labels:
  caddy.reverse_proxy: "{{upstreams {{http}}}}"    # → http://192.168.0.1
  caddy.reverse_proxy: "{{upstreams {{https}}}}"   # → https://192.168.0.1
  caddy.reverse_proxy: "{{upstreams {{h2c}}}}"     # → h2c://192.168.0.1
```

## Enhanced Functions (NEW)

### Environment Variable Functions

Access container environment variables for dynamic configuration.

#### env

Returns the value of an environment variable.

**Usage:** `{{env "KEY"}}`

```yaml
labels:
  caddy: "{{env "DOMAIN_NAME"}}"
  caddy.respond: "App: {{env "APP_NAME"}} Version: {{env "APP_VERSION"}}"
  caddy.header.X-Environment: "{{env "NODE_ENV"}}"
```

#### hasEnv

Checks if an environment variable exists.

**Usage:** `{{hasEnv "KEY"}}`

```yaml
labels:
  # Only add debug headers if DEBUG env var exists
  caddy.@debug.expression: "{{hasEnv "DEBUG"}}"
  caddy.header.X-Debug-Mode: "@debug true"
```

**Note**: For containers, environment variables require container inspection (not yet implemented). Currently works fully for Swarm services only.

### Container Metadata Functions

Access container identification and image information.

#### containerName

Returns the container or service name.

**Usage:** `{{containerName}}`

```yaml
labels:
  caddy.header.X-Container: "{{containerName}}"
  caddy.respond: "Served by: {{containerName}}"
```

#### imageName  

Returns the full image name including registry.

**Usage:** `{{imageName}}`

```yaml
labels:
  caddy.header.X-Image: "{{imageName}}"
  caddy.respond: "Image: {{imageName}}"
```

#### imageTag

Returns just the tag portion of the image.

**Usage:** `{{imageTag}}`

```yaml
labels:
  caddy.header.X-Version: "{{imageTag}}"
  # Conditional config based on tag
  caddy.@production.expression: '{{imageTag}} == "latest"'
```

### Label Functions

Access Docker labels for advanced conditional logic.

#### label

Returns the value of a Docker label.

**Usage:** `{{label "key"}}`

```yaml
labels:
  app.environment: "production"
  app.team: "backend"
  caddy: "{{label "app.environment"}}.example.com"
  caddy.header.X-Team: "{{label "app.team"}}"
```

#### hasLabel

Checks if a Docker label exists.

**Usage:** `{{hasLabel "key"}}`

```yaml
labels:
  # Only enable auth if auth.enabled label exists
  caddy.@auth.expression: "{{hasLabel "auth.enabled"}}"
  caddy.basic_auth: "@auth"
```

### Network Functions

Access container network information for dynamic routing.

#### primaryIP

Returns the first available IP address of the container.

**Usage:** `{{primaryIP}}`

```yaml
labels:
  caddy.respond: "Container IP: {{primaryIP}}"
  caddy.reverse_proxy: "{{primaryIP}}:8080"
```

#### networkIP

Returns the IP address on a specific network.

**Usage:** `{{networkIP "networkName"}}`

```yaml
labels:
  caddy.reverse_proxy: "{{networkIP "backend"}}:3000"
  caddy.header.X-Backend-IP: "{{networkIP "backend"}}"
```

#### networks

Returns a list of all networks the container is connected to.

**Usage:** `{{networks}}`

```yaml
labels:
  caddy.header.X-Networks: "{{networks}}"
  # Example output: ["web", "backend", "database"]
```

### Volume and Mount Functions

Access Docker volume and mount information for static file serving.

#### mountSource

Returns the host source path for a specific mount point.

**Usage:** `{{mountSource "/container/path"}}`

```yaml
labels:
  caddy: static.example.com
  caddy.root: "* {{mountSource "/app/static"}}"
  caddy.file_server: ""
```

#### bindMounts

Returns all bind mount source paths from the host.

**Usage:** `{{bindMounts}}`

```yaml
labels:
  caddy: static.example.com
  # Use first bind mount as static file root
  caddy.root: "* {{index (bindMounts) 0}}"
  caddy.file_server: ""
```

#### volumeMounts

Returns all Docker volume names mounted to the container.

**Usage:** `{{volumeMounts}}`

```yaml
labels:
  caddy.header.X-Volumes: "{{volumeMounts}}"
  # Example output: ["app-data", "shared-storage"]
```

#### hasMount

Checks if a specific mount point exists.

**Usage:** `{{hasMount "/path"}}`

```yaml
labels:
  # Only serve static files if mount exists
  caddy.@static.expression: "{{hasMount "/app/static"}}"
  caddy.file_server: "@static"
  caddy.root: "@static {{mountSource "/app/static"}}"
```

### Port Mapping Functions

Access container port information for dynamic configuration.

#### portMapping

Returns the host port mapped to a container port.

**Usage:** `{{portMapping containerPort}}`

```yaml
labels:
  caddy.respond: "External port: {{portMapping 8080}}"
  # Use mapped port for health checks
  caddy.reverse_proxy.health_uri: "http://localhost:{{portMapping 3000}}/health"
```

#### exposedPorts

Returns all host ports exposed by the container.

**Usage:** `{{exposedPorts}}`

```yaml
labels:
  caddy.header.X-Exposed-Ports: "{{exposedPorts}}"
  # Example output: ["8080", "9090"]
```

### Container State Functions

Access container runtime state for conditional configurations.

#### isRunning

Checks if the container is in running state.

**Usage:** `{{isRunning}}`

```yaml
labels:
  caddy.@running.expression: "{{isRunning}}"
  caddy.respond: "@running Container is healthy"
  caddy.respond: "Container is stopped"
```

#### isHealthy

Checks if the container is healthy (basic implementation).

**Usage:** `{{isHealthy}}`

```yaml
labels:
  caddy.header.X-Healthy: "{{isHealthy}}"
  caddy.@healthy.expression: "{{isHealthy}}"
  caddy.reverse_proxy: "@healthy {{upstreams}}"
```

**Note**: Currently uses basic state checking. Full health check inspection not yet implemented.

## Real-World Examples

### Pure Caddy Django Deployment

Replace Nginx entirely with Caddy for Django static/media serving:

```yaml
services:
  caddy:
    image: caddy-docker-proxy:local
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - django-static:/srv/static:ro
      - django-media:/srv/media:ro
    labels:
      # Static files served directly by Caddy
      caddy_0: "static.{{env "DOMAIN_NAME"}}"
      caddy_0.root: "* {{mountSource "/srv/static"}}"
      caddy_0.file_server: ""
      caddy_0.header.Cache-Control: "public, max-age=31536000"
      
      # Media files with proper headers
      caddy_1: "media.{{env "DOMAIN_NAME"}}"  
      caddy_1.root: "* {{mountSource "/srv/media"}}"
      caddy_1.file_server: ""
      caddy_1.header.X-Content-Type-Options: "nosniff"

  django:
    image: my-django-app:{{env "VERSION"}}
    labels:
      caddy: "{{env "DOMAIN_NAME"}}"
      caddy.reverse_proxy: "{{upstreams}}"
      caddy.header.X-Django-Container: "{{containerName}}"
```

### Conditional Configuration

Dynamic configuration based on container metadata:

```yaml
labels:
  # Different config for different environments
  caddy: "{{label "app.environment"}}.example.com"
  
  # Enable auth only for production
  caddy.@prod.expression: '{{label "app.environment"}} == "production"'
  caddy.basic_auth: "@prod"
  
  # Different upstreams based on health
  caddy.@healthy.expression: "{{isHealthy}}"
  caddy.reverse_proxy: "@healthy {{upstreams}}"
  caddy.respond: "Service temporarily unavailable"
```

### Multi-Service Applications

Complex applications with multiple services:

```yaml
# API Gateway
api-gateway:
  labels:
    caddy: "api.example.com"
    caddy.header.X-Gateway: "{{containerName}}"
    
    # Route to different services based on path
    caddy.@users.path: "/users/*"
    caddy.@orders.path: "/orders/*"
    
    caddy.reverse_proxy: "@users {{networkIP "backend"}}:3001"
    caddy.reverse_proxy: "@orders {{networkIP "backend"}}:3002"
    caddy.reverse_proxy: "{{upstreams}}"  # Default fallback

# Frontend with dynamic asset serving
frontend:
  labels:
    caddy: "app.example.com"
    
    # Serve static assets from mounted volume
    caddy.@assets.path: "/assets/*"
    caddy.file_server: "@assets"
    caddy.root: "@assets {{mountSource "/app/build"}}"
    
    # API proxy
    caddy.@api.path: "/api/*"
    caddy.reverse_proxy: "@api api-gateway:8080"
    
    # Default to serving React app
    caddy.reverse_proxy: "{{upstreams}}"
```

## Testing Infrastructure

Comprehensive test suite validates all template functions across multiple real-world scenarios.

### Test Structure

The testing infrastructure is located in `tests/template-functions/` and includes:

```
tests/template-functions/
├── compose.yaml              # Main test configuration (15+ services)
├── simple-test.yaml         # Basic functionality tests
├── enhanced-test.yaml       # Advanced template function tests
├── run.sh                   # Primary test runner script
├── simple-run.sh           # Basic functionality test runner
├── validate.sh             # Template function validation
├── api-responses.sh        # API response testing
├── test-data/              # Test data for static/media serving
│   ├── django/             # Django static/media test files
│   ├── react-build/        # React build artifacts
│   ├── static/             # Basic static files
│   └── webapp/             # Web application assets
├── README.md               # Test suite documentation
├── CADDY_NATIVE_DJANGO.md  # Pure Caddy Django deployment guide
└── DJANGO_SCENARIO.md      # Django scenario documentation
```

### Test Coverage

#### Core Template Functions
- **Environment Variables**: Tests `{{env}}` and `{{hasEnv}}` with service containers
- **Container Metadata**: Validates `{{containerName}}`, `{{imageName}}`, `{{imageTag}}`
- **Network Functions**: Tests `{{primaryIP}}`, `{{networks}}`, `{{networkIP}}`
- **Volume Functions**: Validates `{{mountSource}}`, `{{bindMounts}}`, `{{volumeMounts}}`
- **Port Functions**: Tests `{{portMapping}}`, `{{exposedPorts}}`
- **State Functions**: Validates `{{isRunning}}`, `{{isHealthy}}`

#### Real-World Scenarios

##### Django Deployment Test
Complete Django application with pure Caddy (no Nginx):

```yaml
# Django Combined App - ASGI/WSGI with Hypercorn
django-combined:
  image: python:3.11-slim
  command: >
    sh -c "pip install hypercorn django &&
           # [Dynamic Django app creation with template functions]"
  labels:
    caddy: "{{env "DOMAIN_NAME"}}"
    caddy.header.X-Container: "{{containerName}}"
    caddy.header.X-Django-App: "combined"
    caddy.reverse_proxy: "{{upstreams 8000}}"

# Static files served directly by Caddy main container
caddy:
  volumes:
    - django-static:/srv/django-static:ro
    - django-media:/srv/django-media:ro
  labels:
    caddy_0: "static.{{env "DOMAIN_NAME"}}"
    caddy_0.root: "* {{mountSource "/srv/django-static"}}"
    caddy_0.file_server: ""
```

##### Multi-Service Application Test
Tests complex service interactions:

```yaml
# API Backend with container metadata
api-backend:
  labels:
    caddy.route.0_respond: 'API Data from {{containerName}}'

# React Frontend with volume serving  
react-frontend:
  labels:
    caddy: webapp.example.com
    caddy.header.X-Container: "{{containerName}}"
    caddy.file_server: ""
    caddy.root: "* {{mountSource "/app/build"}}"

# Network-aware routing
network-app:
  labels:
    caddy.respond: 'Primary IP: {{primaryIP}} Networks: {{networks}}'
```

### SSL/TLS Testing

Tests include Let's Encrypt staging environment configuration:

```yaml
caddy:
  labels:
    # Use Let's Encrypt staging for testing
    caddy_global: |
      {
        acme_ca https://acme-staging-v02.api.letsencrypt.org/directory
      }
```

This eliminates SSL certificate errors during testing while validating HTTPS functionality.

### Test Execution

#### Run All Tests
```bash
cd tests/template-functions/
./run.sh
```

#### Basic Functionality Test
```bash
./simple-run.sh
```

#### Template Function Validation
```bash
./validate.sh
```

### Test Validation

The test suite validates:

1. **Template Function Processing**: No "function not defined" errors
2. **Generated Caddyfile**: Correct template function resolution  
3. **HTTP Responses**: All endpoints return expected content
4. **Container Metadata**: Template functions return correct values
5. **Volume Serving**: Static files served correctly from volumes
6. **Network Functions**: IP addresses and network lists populated
7. **Error Handling**: Graceful fallback for missing data

### Expected Test Results

#### Successful Template Processing
```
Container Name: template-functions-test-container-name-1
IP: 192.168.48.2 Networks: [template-functions_caddy]
X-Container: template-functions-django-combined-1
Primary IP: 172.29.0.17 Networks: caddy_test template-functions_backend
```

#### Static File Serving
```
✅ Django static files served at: https://static.myapp.local:9443/
✅ Django media files served at: https://media.myapp.local:9443/
✅ React build artifacts served correctly
```

#### Template Function Headers
```
X-Container: template-functions-api-backend-1
X-Django-App: combined  
X-Environment: production
X-Networks: [web, backend, database]
```

### Test Configuration Features

- **Port Conflict Resolution**: Uses ports 9080:80, 9443:443 to avoid conflicts
- **Network Isolation**: Separate networks for frontend, backend, management
- **Volume Management**: Docker volumes for Django static/media files
- **Service Dependencies**: Proper startup order and health checks
- **Template Data**: Comprehensive test data for all scenarios

### Debugging Failed Tests

Common issues and solutions:

```bash
# Check template function processing
docker compose logs caddy | grep -E "(Container Name|IP:|Networks)"

# Validate generated Caddyfile
docker compose logs caddy | grep "New Caddyfile"

# Test specific endpoints
curl -k --resolve example.com:9443:127.0.0.1 https://example.com:9443/
```

The comprehensive test suite ensures all template functions work correctly across containers and services, with real-world deployment scenarios validating the complete feature set.

## Error Handling

All template functions handle errors gracefully:

- **Missing data**: Returns empty string or false for boolean functions
- **Invalid parameters**: Returns empty values instead of failing
- **Network issues**: Graceful fallback to available networks
- **Mount issues**: Empty strings for missing mounts

This ensures containers are never excluded from the configuration due to template function errors.

## Function Compatibility

| Function | Containers | Swarm Services | Notes |
|----------|------------|----------------|-------|
| `env` | ⚠️ Partial | ✅ Full | Container inspection needed |
| `hasEnv` | ⚠️ Partial | ✅ Full | Container inspection needed |
| `containerName` | ✅ Full | ✅ Full | |
| `imageName` | ✅ Full | ✅ Full | |
| `imageTag` | ✅ Full | ✅ Full | |
| `label` | ✅ Full | ✅ Full | |
| `hasLabel` | ✅ Full | ✅ Full | |
| `primaryIP` | ✅ Full | ❌ N/A | Containers only |
| `networkIP` | ✅ Full | ❌ N/A | Containers only |
| `networks` | ✅ Full | ❌ N/A | Containers only |
| `mountSource` | ✅ Full | ✅ Full | |
| `bindMounts` | ✅ Full | ✅ Full | |
| `volumeMounts` | ✅ Full | ✅ Full | |
| `hasMount` | ✅ Full | ✅ Full | |
| `portMapping` | ✅ Full | ❌ N/A | Containers only |
| `exposedPorts` | ✅ Full | ❌ N/A | Containers only |
| `isRunning` | ✅ Full | ✅ Full | Basic implementation |
| `isHealthy` | ✅ Full | ✅ Full | Basic implementation |

## Migration from Basic Setup

### Before (Basic Reverse Proxy)
```yaml
labels:
  caddy: example.com
  caddy.reverse_proxy: "{{upstreams}}"
```

### After (Enhanced with Metadata)
```yaml  
labels:
  caddy: example.com
  caddy.reverse_proxy: "{{upstreams}}"
  
  # Add dynamic headers
  caddy.header.X-Container: "{{containerName}}"
  caddy.header.X-Version: "{{imageTag}}"
  caddy.header.X-Environment: "{{env "NODE_ENV"}}"
  
  # Conditional static file serving
  caddy.@static.expression: "{{hasMount "/app/public"}}"
  caddy.handle: "@static"
  caddy.handle.file_server: ""
  caddy.handle.root: "{{mountSource "/app/public"}}"
```

The enhanced template functions are fully backward compatible - existing configurations continue to work unchanged while new functions unlock advanced Caddy features.