# Docker Deployment Guide for hypermcp Servers

This guide covers containerizing and deploying MCP servers built with hypermcp.

## 📋 Table of Contents
- [Quick Start](#quick-start)
- [Dockerfile Explained](#dockerfile-explained)
- [Building Images](#building-images)
- [Running Containers](#running-containers)
- [Docker Compose](#docker-compose)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### 1. Copy the Example Dockerfile
```bash
cp Dockerfile.example ./Dockerfile
```

### 2. Build Your Image
```bash
docker build -t my-mcp-server:latest .
```

### 3. Run Your Server
```bash
# For stdio transport (recommended for most MCP servers)
docker run -i my-mcp-server:latest

# For Streamable HTTP transport (if implemented, for servers handling multiple clients)
docker run -p 8080:8080 my-mcp-server:latest -transport streamable-http
```

## 📖 Dockerfile Explained

The example Dockerfile uses **multi-stage builds** for optimal image size and security:

### Stage 1: Builder
```dockerfile
FROM golang:1.21-alpine AS builder
```
- Uses Alpine Linux for smaller image
- Installs only build-time dependencies (git, ca-certificates)
- Downloads Go modules first (layer caching optimization)
- Compiles a static binary with stripped debug info

### Stage 2: Runtime
```dockerfile
FROM scratch
```
- Minimal image (no OS, just your binary)
- Copies only essential files (binary, CA certs, timezone data)
- Runs as non-root user (UID 65534)
- ~10-20MB final image size

## 🔨 Building Images

### Basic Build
```bash
docker build -t my-mcp-server:latest .
```

### Build with Version Info
```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t my-mcp-server:v1.0.0 \
  .
```

### Build for Multiple Platforms
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t my-mcp-server:latest \
  --push \
  .
```

## 🏃 Running Containers

### Stdio Transport (Default & Recommended)
Most MCP servers use stdio transport where the client launches the server as a subprocess.
MCP servers using stdio need `-i` (interactive) flag:

```bash
# Run interactively
docker run -i my-mcp-server:latest

# Run with custom log level
docker run -i -e LOG_LEVEL=debug my-mcp-server:latest

# Run in background (less common for stdio)
docker run -d -i --name mcp-server my-mcp-server:latest
```

### Streamable HTTP Transport  
If your server needs to handle multiple concurrent client connections, you can use Streamable HTTP transport
(this replaces the old deprecated HTTP+SSE transport):

```bash
# Run with port mapping
docker run -p 8080:8080 my-mcp-server:latest -transport streamable-http

# Run in background
docker run -d -p 8080:8080 --name mcp-server my-mcp-server:latest -transport streamable-http
```

### With Environment Variables
```bash
docker run -i \
  -e LOG_LEVEL=debug \
  -e CACHE_MAX_COST=52428800 \
  my-mcp-server:latest
```

### With Volume Mounts
```bash
# Mount config file
docker run -i \
  -v $(pwd)/config.yaml:/config.yaml \
  my-mcp-server:latest -config /config.yaml

# Mount cache directory
docker run -i \
  -v $(pwd)/cache:/cache \
  my-mcp-server:latest
```

## 🐳 Docker Compose

For more complex setups with multiple services:

### docker-compose.yml
```yaml
version: '3.8'

services:
  mcp-server:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        VERSION: ${VERSION:-dev}
        COMMIT: ${COMMIT:-unknown}
        BUILD_TIME: ${BUILD_TIME:-unknown}
    image: my-mcp-server:${VERSION:-latest}
    container_name: mcp-server
    restart: unless-stopped
    
    # For stdio transport
    stdin_open: true
    tty: true
    
    # For SSE transport (uncomment if needed)
    # ports:
    #   - "8080:8080"
    
    environment:
      LOG_LEVEL: ${LOG_LEVEL:-info}
      TRANSPORT: ${TRANSPORT:-stdio}
    
    # Optional: Add health check if you implement one
    # healthcheck:
    #   test: ["CMD", "/mcp-server", "-health"]
    #   interval: 30s
    #   timeout: 3s
    #   retries: 3
    #   start_period: 5s
    
    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
    
    # Logging
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # Example: Add Redis for caching
  # redis:
  #   image: redis:7-alpine
  #   container_name: mcp-redis
  #   restart: unless-stopped
  #   ports:
  #     - "6379:6379"
  #   volumes:
  #     - redis-data:/data

# volumes:
#   redis-data:
```

### .env file
```bash
VERSION=v1.0.0
COMMIT=abc123
BUILD_TIME=2025-01-10T12:00:00Z
LOG_LEVEL=info
TRANSPORT=stdio
```

### Usage
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f mcp-server

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

## ✅ Best Practices

### Security
1. **Run as non-root**: Always use `USER` directive
2. **Use scratch or distroless**: Minimal attack surface
3. **Scan for vulnerabilities**: Use `docker scan my-mcp-server:latest`
4. **No secrets in image**: Use environment variables or secrets management

### Performance
1. **Multi-stage builds**: Smaller images, faster deployments
2. **Layer caching**: Copy `go.mod` before source code
3. **Static binaries**: `CGO_ENABLED=0` for portability
4. **Strip debug info**: `-ldflags="-w -s"` reduces size

### Maintainability
1. **Version tagging**: Use semantic versioning
2. **Build args**: Inject version info at build time
3. **Health checks**: Implement and use in production
4. **Logging**: Use structured logging, JSON format for log aggregation

### Example Production Dockerfile
```dockerfile
# Production-ready example with all best practices

FROM golang:1.21-alpine AS builder

# Security: Install only necessary packages
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Dependency caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source
COPY . .

# Build with optimizations
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o /build/mcp-server \
    ./cmd/server

# Runtime stage: Use distroless for better security than scratch
FROM gcr.io/distroless/static:nonroot

# Copy timezone and CA certs
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/mcp-server /mcp-server

# Already non-root in distroless:nonroot
# USER nonroot:nonroot

ENV LOG_LEVEL=info
ENV TRANSPORT=stdio

ENTRYPOINT ["/mcp-server"]
CMD ["-transport", "stdio"]
```

## 🐛 Troubleshooting

### Issue: "Cannot connect to stdio"
**Solution**: Make sure you're using `-i` flag:
```bash
docker run -i my-mcp-server:latest
```

### Issue: "Permission denied"
**Solution**: Check file permissions in container:
```bash
# Debug with shell (if not using scratch)
docker run --rm -it --entrypoint /bin/sh my-mcp-server:latest

# Or use distroless debug image
FROM gcr.io/distroless/static:debug-nonroot
```

### Issue: "Cannot find ca-certificates"
**Solution**: Make sure you copied them in Dockerfile:
```dockerfile
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
```

### Issue: Large image size
**Solution**: Check layers and optimize:
```bash
# Analyze image
docker history my-mcp-server:latest

# Use dive for detailed analysis
dive my-mcp-server:latest
```

### Issue: Container exits immediately
**Solution**: Check logs and entry point:
```bash
# View logs
docker logs mcp-server

# Override entrypoint for debugging
docker run --rm -it --entrypoint /bin/sh my-mcp-server:latest
```

## 📚 Additional Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Docker Compose](https://docs.docker.com/compose/)
- [Container Security](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)

## 🎯 Next Steps

1. **Test locally**: Build and run your container
2. **Push to registry**: Docker Hub, GitHub Container Registry, or private registry
3. **Deploy**: Kubernetes, Docker Swarm, or cloud services (ECS, Cloud Run, etc.)
4. **Monitor**: Set up logging and metrics collection
5. **Scale**: Use orchestration for multiple instances

---

**Need help?** Check the main hypermcp documentation or open an issue!
