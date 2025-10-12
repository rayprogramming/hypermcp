# Weather Server Example

An MCP server that demonstrates HTTP client usage, caching, and metrics tracking.

## What This Example Demonstrates

- Enabling cache for responses
- Cache key management
- Cache hit/miss tracking
- Multiple tools
- Metrics collection and reporting
- Graceful shutdown with metrics logging

## Features

- **get_weather**: Get current weather for a city (cached for 5 minutes)
- **get_forecast**: Get 3-day forecast for a city

Note: This example returns mock data. In a real implementation, you would use `srv.HTTPClient()` to call an actual weather API.

## Running the Example

```bash
go run main.go
```

## Testing

Call the same city twice to see caching in action:

```json
// First call - cache miss
{"city": "San Francisco"}

// Second call within 5 minutes - cache hit
{"city": "San Francisco"}
```

## Metrics

The server tracks and reports:
- Tool invocations
- Cache hits and misses
- Cache hit rate
- Uptime

Metrics are logged on shutdown.

## Real Implementation

To integrate with a real weather API:

```go
// Use the HTTP client
type WeatherResponse struct {
    Temperature float64 `json:"temp"`
    Condition   string  `json:"condition"`
}

var result WeatherResponse
err := srv.HTTPClient().Get(ctx, 
    "https://api.weather.com/v1/current?city="+input.City,
    &result)
if err != nil {
    return nil, nil, fmt.Errorf("weather API error: %w", err)
}
```

## Next Steps

- Add more weather-related tools (alerts, historical data)
- Implement actual API integration
- Add error handling for API failures
- Use different cache TTLs for different data types

See the [fileserver example](../fileserver/) for resource usage.
