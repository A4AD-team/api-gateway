# AGENTS.md

This file contains guidelines for agentic coding agents working in the API Gateway service.

## Service Overview

**API Gateway** is a Go-based service using Fiber framework that serves as the unified entry point for all external and internal clients of our business process engine.

**Key Responsibilities:**
- Authentication and authorization (JWT + RBAC)
- Request routing to backend services
- Rate limiting and throttling
- CORS and request/response transformation
- Logging, tracing, and metrics (OpenTelemetry + Prometheus)
- Health checks for Kubernetes
- Basic attack protection

## Build/Test Commands

```bash
# Build
go build -o bin/gateway ./cmd/gateway

# Run locally
go run ./cmd/gateway
# or
make run

# Test
go test ./...
go test -v ./internal/middleware          # Run specific package tests
go test -run TestJWTAuth ./internal/auth   # Run specific test function
go test -race ./...                        # Run with race detection

# Lint/Format
go fmt ./...
go vet ./...
goimports -w .
golangci-lint run                           # if configured

# Dependencies
go mod tidy
go mod download
```

## Project Structure

```
api-gateway/
├── cmd/gateway/           # Application entry point
│   └── main.go
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── middleware/       # Fiber middlewares
│   │   ├── auth/        # JWT authentication
│   │   ├── logging/     # Structured logging
│   │   ├── ratelimit/   # Rate limiting
│   │   ├── tracing/     # OpenTelemetry tracing
│   │   └── cors/        # CORS handling
│   ├── proxy/           # Reverse proxy logic
│   ├── routes/          # Route registration
│   └── health/          # Health check endpoints
├── pkg/                  # Public library code
│   └── limiter/         # Rate limiting utilities
├── config/              # Configuration files
│   └── config.yaml
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Code Style Guidelines

### Go Code Style
- **Imports**: Group in order: standard library, third-party packages, internal packages
- **Formatting**: Use `gofmt` and `goimports` consistently
- **Naming Conventions**:
  - Package names: lowercase, single words when possible (`auth`, `proxy`, `config`)
  - Functions: `CamelCase` for exported, `camelCase` for unexported
  - Variables: `camelCase`, avoid abbreviations (`req` not `r`, `resp` not `w`)
  - Constants: `UPPER_SNAKE_CASE` for untyped, `camelCase` for typed constants
- **Error Handling**: Always handle errors explicitly, use `fmt.Errorf` for wrapping, never use panic for flow control
- **Interfaces**: Keep interfaces small, "accept interfaces, return structs" principle

### Fiber-Specific Patterns
- Use `fiber.Ctx` as the primary context parameter
- Follow Fiber middleware signature: `func(c *fiber.Ctx) error`
- Use `c.Next()` to pass control to next middleware
- Return proper HTTP status codes with `c.Status(code).JSON(data)`
- Use `c.Locals()` and `c.Query()` for request data extraction

### Configuration Management
- Use Viper for configuration loading
- Support both environment variables and YAML config files
- Default values should be sensible for local development
- Validate configuration at startup using struct tags

### Authentication & Authorization
- Validate JWT tokens using `golang-jwt` package
- Extract user info from token claims (user_id, roles, permissions)
- Implement RBAC checks in middleware
- Use JWKS endpoint for token validation when available

### Rate Limiting
- Implement both user-based and IP-based rate limiting
- Use Redis for distributed rate limiting in production
- Support different rate limits per role/endpoint
- Return proper rate limit headers (`X-RateLimit-*`)

## Testing Guidelines

### Unit Tests
- Use table-driven tests for multiple test cases
- Use `testify/assert` for assertions
- Mock external dependencies (HTTP clients, Redis)
- Test both success and error paths
- Target 80%+ code coverage

### Integration Tests
- Test middleware chains end-to-end
- Use testcontainers for external dependencies
- Test actual HTTP request/response cycles
- Verify rate limiting behavior

### Test Structure
```go
func TestJWTAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        setupAuth      string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "valid token",
            setupAuth:      "Bearer valid.jwt.token",
            expectedStatus: 200,
        },
        {
            name:           "missing token",
            setupAuth:      "",
            expectedStatus: 401,
            expectedError:  "missing authorization header",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Observability Standards

### Logging
- Use structured logging with `zerolog` or `slog`
- Include correlation ID in all log entries
- Log levels: DEBUG, INFO, WARN, ERROR
- Never log sensitive data (tokens, passwords, PII)

### Tracing
- Use OpenTelemetry for distributed tracing
- Create spans for significant operations
- Include relevant tags and attributes
- Propagate trace context to upstream services

### Metrics
- Use Prometheus client library
- Track request count, duration, error rate
- Include rate limiting metrics
- Expose metrics on `/metrics` endpoint

## Security Guidelines

- Validate all input parameters and headers
- Implement proper timeout handling (read/write timeouts)
- Use HTTPS in production
- Implement request size limits
- Never trust upstream service responses
- Sanitize error messages to avoid information leakage

## Environment Configuration

### Required Environment Variables
```bash
# Server
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=30s

# Auth Service
AUTH_SERVICE_URL=http://auth-service:8081
JWT_SECRET=your-secret-key
JWKS_URL=http://auth-service:8081/.well-known/jwks.json

# Rate Limiting
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
RATE_LIMIT_RPS=50
RATE_LIMIT_BURST=100

# Observability
LOG_LEVEL=info
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
PROMETHEUS_ENABLED=true
```

### Docker Development
```bash
# Start dependencies
docker compose up -d redis auth-service

# Run gateway
go run ./cmd/gateway
```

## Common Patterns

### Middleware Pattern
```go
func AuthMiddleware(config *config.AuthConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "missing authorization header",
            })
        }
        
        claims, err := validateToken(token, config)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "invalid token",
            })
        }
        
        c.Locals("user_id", claims.UserID)
        c.Locals("roles", claims.Roles)
        return c.Next()
    }
}
```

### Error Response Pattern
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details any    `json:"details,omitempty"`
}

func sendError(c *fiber.Ctx, status int, message string) error {
    return c.Status(status).JSON(ErrorResponse{
        Error: message,
    })
}
```

## Health Checks

Implement proper health checks:
- `/health` - Basic liveness check
- `/ready` - Readiness check (dependencies)
- `/metrics` - Prometheus metrics

## Performance Considerations

- Use connection pooling for HTTP clients
- Implement proper timeouts for all external calls
- Use context with timeout for long-running operations
- Consider response caching for static endpoints
- Monitor memory usage and goroutine leaks