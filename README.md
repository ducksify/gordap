# RDAP Client for Go

A Go module for querying domain RDAP (Registration Data Access Protocol) information using the IANA RDAP bootstrap file.

## Features

- **Automatic Server Discovery**: Uses the IANA RDAP bootstrap file to automatically find the correct RDAP server for any TLD
- **Special Case Handling**: Handles special cases like `.ch` domains that use a different URL structure
- **Caching**: Caches bootstrap data and server mappings for improved performance
- **Thread-Safe**: All operations are thread-safe with proper mutex protection
- **Configurable**: Customizable timeouts, HTTP clients, and bootstrap URLs
- **Simple API**: Easy-to-use API similar to the WHOIS client

## Installation

```bash
go get github.com/ducksify/rdap
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/ducksify/rdap"
)

func main() {
    // Simple usage
    result, err := rdap.RDAP("example.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}
```

## API Reference

### Functions

#### `RDAP(domain string) Default (string, error)`

Performs an RDAP query for the given domain using the default client.

```go
result, err := rdap.RDAP("example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)
```

#### `Version() string`

Returns the package version.

#### `Author() string`

Returns the package author information.

#### `License() string`

Returns the package license information.

### Client Methods

#### `NewClient() *Client`

Creates a new RDAP client with default settings.

```go
client := rdap.NewClient()
```

#### `SetTimeout(timeout time.Duration) *Client`

Sets the HTTP request timeout.

```go
client := rdap.NewClient().SetTimeout(10 * time.Second)
```

#### `SetHTTPClient(httpClient *http.Client) *Client`

Sets a custom HTTP client.

```go
customClient := &http.Client{
    Timeout: 5 * time.Second,
}
client := rdap.NewClient().SetHTTPClient(customClient)
```

#### `SetBootstrapURL(url string) *Client`

Sets a custom bootstrap URL (useful for testing).

```go
client := rdap.NewClient().SetBootstrapURL("https://example.com/rdap.json")
```

#### `SetBootstrapFile(filepath string) *Client`

Sets the path to a local bootstrap file (useful for Docker/Lambda environments).

```go
// Use local bootstrap file
client := rdap.NewClient().SetBootstrapFile("/app/bootstrap.json")
```

#### `SetDisableCache(disabled bool) *Client`

Disables caching for Lambda environments or when fresh data is always needed.

```go
// For AWS Lambda environments
client := rdap.NewClient().SetDisableCache(true)

// Re-enable caching for other environments
client := rdap.NewClient().SetDisableCache(false)
```

#### `SetCacheBootstrapOnly(enabled bool) *Client`

Enables caching only for bootstrap data, not for domain queries. This is useful when you want to cache the IANA bootstrap file (which changes rarely) but always fetch fresh domain information.

```go
// Cache bootstrap data but always fetch fresh domain info
client := rdap.NewClient().SetCacheBootstrapOnly(true)

// Disable bootstrap-only caching
client := rdap.NewClient().SetCacheBootstrapOnly(false)
```

#### `RDAP(domain string) (string, error)`

Performs an RDAP query for the given domain.

```go
client := rdap.NewClient()
result, err := client.RDAP("example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)
```

#### `ClearCache()`

Clears the bootstrap data and server mapping cache.

```go
client := rdap.NewClient()
client.ClearCache()
```

## How It Works

1. **Bootstrap Data**: The client fetches the IANA RDAP bootstrap file from [https://data.iana.org/rdap/dns.json](https://data.iana.org/rdap/dns.json)
2. **Server Mapping**: For each TLD, it maps to the appropriate RDAP server from the bootstrap data
3. **Special Cases**: Handles special cases like `.ch` domains that use a different URL structure
4. **Caching**: Caches bootstrap data for 24 hours and server mappings for improved performance
5. **Query**: Performs the actual RDAP query to the appropriate server

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/ducksify/rdap"
)

func main() {
    domains := []string{"example.com", "example.org", "example.ch"}
    
    for _, domain := range domains {
        result, err := rdap.RDAP(domain)
        if err != nil {
            log.Printf("Error querying %s: %v", domain, err)
            continue
        }
        fmt.Printf("RDAP result for %s:\n%s\n\n", domain, result)
    }
}
```

### Custom Client Configuration

```go
package main

import (
    "fmt"
    "log"
    "time"
    "net/http"
    "github.com/ducksify/rdap"
)

func main() {
    // Create a custom HTTP client with shorter timeout
    httpClient := &http.Client{
        Timeout: 5 * time.Second,
    }
    
    // Create RDAP client with custom settings
    client := rdap.NewClient().
        SetHTTPClient(httpClient).
        SetTimeout(5 * time.Second)
    
    result, err := client.RDAP("example.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}
```

### AWS Lambda Configuration

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/ducksify/rdap"
)

func main() {
    // Option 1: Always fetch fresh data (no caching)
    client1 := rdap.NewClient().
        SetDisableCache(true).  // Always fetch fresh bootstrap data
        SetTimeout(10 * time.Second)
    
    // Option 2: Cache bootstrap data only (recommended for Lambda)
    client2 := rdap.NewClient().
        SetCacheBootstrapOnly(true).  // Cache bootstrap, fresh domain data
        SetTimeout(10 * time.Second)
    
    result, err := client2.RDAP("example.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}
```

### Error Handling

```go
package main

import (
    "fmt"
    "log"
    "github.com/ducksify/rdap"
)

func main() {
    result, err := rdap.RDAP("example.com")
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "domain cannot be empty"):
            log.Fatal("Please provide a valid domain")
        case strings.Contains(err.Error(), "no RDAP server found"):
            log.Fatal("No RDAP server available for this TLD")
        case strings.Contains(err.Error(), "RDAP query failed"):
            log.Fatal("Failed to query RDAP server")
        default:
            log.Fatal("Unexpected error:", err)
        }
    }
    fmt.Println(result)
}
```

The client returns descriptive errors for various failure scenarios:

- Empty or invalid domains
- Network connectivity issues
- Bootstrap data fetch failures
- RDAP server query failures
- Unsupported TLDs

## Supported TLDs

The client supports all TLDs listed in the IANA RDAP bootstrap file, including:

- Generic TLDs: `.com`, `.org`, `.net`, `.info`, etc.
- Country code TLDs: `.us`, `.uk`, `.de`, `.fr`, etc.
- New gTLDs: `.cloud`, `.app`, `.dev`, etc.

### Special Cases

- **`.ch` domains**: Uses `https://rdap.nic.ch/domain/{domain}` format
- Other TLDs: Use the standard format from the bootstrap file

## Testing

Run the tests:

```bash
go test
```

Run tests with verbose output:

```bash
go test -v
```

## Performance

- **Thread-Safe**: All operations are thread-safe
- **Connection Reuse**: Uses Go's standard HTTP client for connection pooling

## License

Licensed under the Apache License 2.0. See the LICENSE file for details.

## Author

François "@Ducksify"

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
