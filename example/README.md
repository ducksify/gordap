# RDAP Client Example

A simple example demonstrating how to use the RDAP client to query domain registration information.

## Quick Start

```bash
# Query example.com (default)
go run main.go

# Query a specific domain
go run main.go google.com
go run main.go github.com
go run main.go ducksify.ch
```

## What it does

This example:

1. **Creates an RDAP client** with a 10-second timeout
2. **Queries RDAP information** for the specified domain
3. **Displays the full JSON response** in a pretty-printed format
4. **Extracts key information** like domain name, status, events, and entities

## Sample Output

```
🔍 RDAP Client Example
Querying: example.com
──────────────────────────────────────────────────
✅ Success! RDAP data for example.com:

{
  "ldhName": "EXAMPLE.COM",
  "status": ["client delete prohibited", "client transfer prohibited"],
  "events": [...],
  "entities": [...],
  ...
}

📋 Key Information:
  Domain: EXAMPLE.COM
  Status: [client delete prohibited client transfer prohibited client update prohibited]
  Events: 4 events found
  Entities: 1 entities found
```

## Features Demonstrated

- ✅ **Default domain**: Uses `example.com` if no domain is specified
- ✅ **Error handling**: Shows clear error messages if queries fail
- ✅ **JSON parsing**: Pretty-prints the RDAP response
- ✅ **Key data extraction**: Highlights important domain information
- ✅ **Flexible input**: Accepts any domain name as command line argument

## Supported Domains

The RDAP client works with any domain that has RDAP servers, including:
- Generic TLDs: `.com`, `.org`, `.net`, `.cloud`, etc.
- Country TLDs: `.ch`, `.uk`, `.de`, etc.
- New gTLDs: `.app`, `.dev`, `.io`, etc.
