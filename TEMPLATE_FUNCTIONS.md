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