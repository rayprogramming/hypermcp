# Copilot Instructions for `hypermcp`

These instructions help AI coding agents work productively in this repository. They describe the architecture, key workflows, and project-specific patterns you should follow and extend.

## Big Picture

`hypermcp` is a small, reusable Go library that provides common infrastructure for building Model Context Protocol (MCP) servers. It wraps the MCP Go SDK with:
- A configured HTTP client (`httpx.Client`) with retries, timeouts, and limits
- A simple caching layer (`cache.Cache`) built on Ristretto
- Structured logging via `zap`
- Transport helpers for stdio (implemented) and Streamable HTTP (placeholder)

Core types:
- `Server` (`server.go`): owns `mcp.Server`, `httpx.Client`, `cache.Cache`, and `zap.Logger`; exposes helpers and registration stats.
- `TransportType` + `RunWithTransport` (`transport.go`): selects and starts transport (only `stdio` implemented).
- `ServerInfo` (`version.go`): simple version struct and formatter.

Typical usage (from `README.md`): create a `Server` with `Config`, build providers using `srv.HTTPClient()/Cache()/Logger()`, register tools/resources, log counts, then run with stdio.

## Project Conventions

- Package name: `hypermcp`; import path: `github.com/rayprogramming/hypermcp`.
- Prefer dependency injection via the `Server` getters over global singletons.
- Register MCP features via `srv.AddTool`, `srv.AddResource`, `srv.AddResourceTemplate` on `srv.MCP()` or helpers on `Server` (see README examples). After each registration, call `srv.IncrementToolCount()` or `srv.IncrementResourceCount()` so `LogRegistrationStats()` is accurate.
- Use `RunWithTransport(ctx, srv, TransportStdio, logger)` for starting servers; HTTP transport is not implemented and should return a clear error.
- Logging: use the `srv.Logger()`; keep logs structured and low-noise in hot paths. Cache/HTTP helpers already log at `Debug`/`Warn` where appropriate.

## httpx Client Patterns (`httpx/httpx.go`)

- Use `Client.Get(ctx, url, &result)` for JSON GET requests; it sets headers and delegates to `DoJSON`.
- `DoJSON` handles retries (3 max) for 429/5xx, request timeout (~6s), and caps response size at 10MB. It returns permanent errors for non-retryable statuses and JSON decode errors.
- Always pass `context.Context`; cancellation/timeouts are respected and covered by tests.
- Example:
  ```go
  type Payload struct { Message string `json:"message"` }
  var out Payload
  if err := srv.HTTPClient().Get(ctx, "https://api.example.com/status", &out); err != nil { /* handle */ }
  ```

## Cache Patterns (`cache/cache.go`)

- `cache.Cache` stores arbitrary values with TTL tracking (internal map) on top of Ristretto. Methods: `Get`, `Set(key, value, ttl)`, `Delete`, `Clear`, `Metrics`, `Close`.
- TTL is enforced via a background cleanup goroutine and checked on `Get`. If expired, the key is deleted lazily.
- Cost is roughly estimated (constant 64). If you store large payloads, prefer storing pointers or compact structs to avoid exceeding `MaxCost`.
- Example:
  ```go
  v, ok := srv.Cache().Get(cacheKey)
  if ok { /* use v */ }
  // fill and set for 1 minute
  srv.Cache().Set(cacheKey, data, time.Minute)
  ```

## Transport (`transport.go`)

- Use `TransportStdio` for most MCP servers. Starting with Streamable HTTP should return `fmt.Errorf("Streamable HTTP transport not yet implemented")` — don’t attempt to wire HTTP until implemented here.
- `RunWithTransport` logs selected transport, sets up `mcp.StdioTransport`, then calls `srv.Run(ctx, transport)`.

## Tests & Benchmarks

- Unit tests exist for `cache` and `httpx` packages:
  - `cache/cache_test.go`: Get/Set, expiration, delete, and basic benchmarks.
  - `httpx/httpx_test.go`: success path, retry behavior, context cancellation, and benchmark.
- Run with:
  - All: `go test ./...`
  - Specific pkg: `go test ./cache -v`
  - Benchmarks: `go test -bench=. ./cache`

## Developer Workflows

- Build: standard `go build ./...` (no custom build tags). This library is imported by downstream MCP servers.
- Logging setup: use `zap.NewProduction()` in binaries; tests use `zaptest`.
- Versioning: `ServerInfo` is available for embedding version/build info in your binaries.
- Error handling: return wrapped errors (`fmt.Errorf("context", %w)`) from helpers; transport selection returns explicit errors for unsupported modes.

## Integration Points

- MCP SDK: `github.com/modelcontextprotocol/go-sdk/mcp` — use `srv.MCP()` if you need direct access to register tools/resources beyond the helpers shown in README.
- HTTP deps: `cenkalti/backoff/v4` for retries; standard `net/http`; headers pre-set in `Get`.
- Cache dep: `github.com/dgraph-io/ristretto` with metrics available via `Cache.Metrics()`.

## Gotchas

- Cache is created even when `CacheEnabled=false` (as a minimal instance). Don’t assume `srv.Cache()` is nil.
- Response bodies over 10MB will be truncated by `httpx` and cause errors. Choose APIs accordingly or stream/process incrementally.
- Streamable HTTP transport is intentionally unimplemented — calling it should be treated as a configuration error.