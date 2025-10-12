# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow these steps:

### 1. **Do Not** Open a Public Issue

Please do not report security vulnerabilities through public GitHub issues. This helps protect users before a fix is available.

### 2. Report Privately

**Preferred Method**: Use GitHub's private security vulnerability reporting:
- Go to the repository's Security tab
- Click "Report a vulnerability"
- Fill out the form with details

**Alternative Method**: Email the maintainer directly at:
- Email: [Create an issue and request contact information]
- Include "SECURITY" in the subject line

### 3. Provide Details

Include as much information as possible:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)
- Your contact information

### 4. What to Expect

- **Acknowledgment**: We'll acknowledge receipt within 48 hours
- **Assessment**: We'll assess the vulnerability within 5 business days
- **Updates**: We'll keep you informed of our progress
- **Disclosure**: We'll work with you on responsible disclosure timing
- **Credit**: We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices for Users

### When Building MCP Servers with hypermcp

1. **Input Validation**
   ```go
   // Always validate user input
   if input.Path != "" && filepath.IsAbs(input.Path) {
       return nil, nil, fmt.Errorf("absolute paths not allowed")
   }
   ```

2. **Path Traversal Prevention**
   ```go
   // Sanitize file paths
   cleanPath := filepath.Clean(input.Path)
   if strings.Contains(cleanPath, "..") {
       return nil, nil, fmt.Errorf("directory traversal detected")
   }
   ```

3. **Error Handling**
   ```go
   // Don't expose internal details in errors
   if err != nil {
       srv.Metrics().IncrementErrors()
       return nil, nil, fmt.Errorf("operation failed")
   }
   ```

4. **Rate Limiting**
   ```go
   // Implement rate limiting for expensive operations
   // Use cache and track metrics to detect abuse
   if srv.GetMetrics().ToolInvocations > threshold {
       return nil, nil, fmt.Errorf("rate limit exceeded")
   }
   ```

5. **Secrets Management**
   ```go
   // Never hardcode secrets
   apiKey := os.Getenv("API_KEY")
   if apiKey == "" {
       return fmt.Errorf("API_KEY not set")
   }
   ```

6. **Logging**
   ```go
   // Don't log sensitive information
   logger.Info("API request", 
       zap.String("endpoint", "/api/user"),
       // DON'T: zap.String("password", password),
   )
   ```

### HTTP Client Security

The `httpx.Client` includes:
- Request timeouts (prevents hanging)
- Response size limits (prevents memory exhaustion)
- Retry logic (handles transient failures)

Additional recommendations:
```go
// Validate URLs before making requests
if !strings.HasPrefix(url, "https://") {
    return fmt.Errorf("only HTTPS URLs allowed")
}

// Set reasonable timeouts
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := srv.HTTPClient().Get(ctx, url, &result)
```

### Cache Security

The cache implementation:
- Uses memory limits (prevents exhaustion)
- Supports TTL (prevents stale data)
- Is thread-safe (prevents race conditions)

Recommendations:
```go
// Don't cache sensitive data
if !isSensitive(data) {
    srv.Cache().Set(key, data, ttl)
}

// Use appropriate TTLs
shortTTL := 5 * time.Minute  // For frequently changing data
longTTL := 24 * time.Hour     // For stable data
```

## Known Security Considerations

### 1. Stdio Transport

The stdio transport is designed for local use with trusted clients. It:
- Does not include authentication
- Does not encrypt communication
- Assumes a trusted local environment

**Recommendation**: Only use stdio transport for:
- Local development
- Claude Desktop integration
- Trusted local tools

### 2. Streamable HTTP Transport (Not Yet Implemented)

When using HTTP transport (future):
- Use HTTPS only
- Implement authentication (API keys, OAuth, etc.)
- Use CORS appropriately
- Implement rate limiting
- Consider using a reverse proxy

### 3. External API Integration

When calling external APIs:
- Validate and sanitize all inputs
- Use proper error handling
- Don't expose API errors to users
- Implement circuit breakers for failing services
- Use the provided retry logic appropriately

### 4. Resource Access

When implementing resources:
- Validate all URI parameters
- Prevent path traversal attacks
- Implement access controls
- Audit resource access
- Set appropriate MIME types

## Dependency Security

We use:
- **Dependabot**: Automatic dependency updates (weekly)
- **Go mod**: Dependency verification
- **golangci-lint**: Security linting with gosec

To check dependencies:
```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
go mod verify
```

## Security Update Process

1. **Detection**: Vulnerability reported or detected
2. **Assessment**: Evaluate severity and impact
3. **Fix**: Develop and test patch
4. **Release**: Create security release
5. **Announcement**: Publish security advisory
6. **Disclosure**: Full details after users have time to update

## Security Contacts

- **GitHub Security**: Use private vulnerability reporting
- **Maintainer**: Request contact via issue (for sensitive matters)

## Responsible Disclosure

We believe in responsible disclosure and will:
- Work with security researchers
- Provide credit for discoveries
- Keep reporters informed
- Disclose after fixes are available

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security/)
- [MCP Specification](https://modelcontextprotocol.io/)

---

**Last Updated**: October 11, 2025

Thank you for helping keep hypermcp and its users safe!
