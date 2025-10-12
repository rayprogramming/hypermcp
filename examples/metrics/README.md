# Metrics Demo Example

This example demonstrates how to use the metrics system in hypermcp to track server performance and usage statistics.

## Features Demonstrated

- Tool invocation tracking
- Cache hit/miss tracking
- Metrics retrieval via tool
- Periodic metrics logging
- Final metrics on shutdown
- Cache-specific metrics

## Running the Example

```bash
go run main.go
```

## Testing with MCP Inspector

Use the MCP Inspector to test the tools:

1. Call `process_data` with different keys to populate the cache:
   ```json
   {"key": "test1"}
   {"key": "test2"}
   {"key": "test1"}  // This will hit the cache
   ```

2. Call `get_metrics` to see current statistics:
   ```json
   {}
   ```

## What to Observe

- First call to `process_data` with a key will be a cache miss
- Subsequent calls with the same key (within 2 minutes) will be cache hits
- `get_metrics` shows real-time statistics
- Every 30 seconds, metrics are logged to console
- On shutdown (Ctrl+C), final metrics and cache statistics are displayed

## Metrics Tracked

- **Uptime**: How long the server has been running
- **Tool Invocations**: Number of times tools have been called
- **Resource Reads**: Number of resource read operations
- **Cache Hits**: Number of successful cache retrievals
- **Cache Misses**: Number of cache misses
- **Cache Hit Rate**: Percentage of cache hits vs total cache operations
- **Errors**: Number of errors encountered

## Cache Metrics

The example also demonstrates accessing Ristretto cache metrics:
- Hits and misses
- Hit ratio
- Other internal cache statistics
