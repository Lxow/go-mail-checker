# Go Mail Checker 

A lightweight Go API for checking domain MX records. Perfect for validating email domains and understanding mail server configurations.

## Features

- **Fast & Lightweight** - Pure Go with stdlib only, no external dependencies
- **MX Record Lookup** - Check if a domain can receive emails
- **Production Ready** - Graceful shutdown, proper error handling, timeouts
- **Docker Support** - Multi-stage build for minimal container size
- **Well Tested** - Comprehensive unit tests with benchmarks
- **Health Monitoring** - Built-in health check endpoint

## Quick Start

### Local Development

```bash
# Clone the repository
git clone https://github.com/lxow/go-mail-checker.git
cd go-mail-checker

# Run the API
go run main.go

# Test the endpoints
curl http://localhost:8080/health
curl "http://localhost:8080/check-domain?domain=gmail.com"
```

### Docker

```bash
# Build the image
docker build -t go-mail-checker .

# Run the container
docker run -p 8080:8080 go-mail-checker

# Test the API
curl http://localhost:8080/health
```

## API Endpoints

### Health Check
```
GET /health
```

Returns API health status and timestamp.

**Example Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-03-12T14:30:45Z",
  "message": "Go Mail Checker API is running smoothly!"
}
```

### Domain MX Check
```
GET /check-domain?domain=<domain>
```

Checks if a domain has MX records (can receive emails).

**Parameters:**
- `domain` (required) - The domain to check (e.g., `gmail.com`)

**Example Response:**
```json
{
  "domain": "gmail.com",
  "has_mx": true,
  "mx_records": [
    "5 gmail-smtp-in.l.google.com",
    "10 alt1.gmail-smtp-in.l.google.com",
    "20 alt2.gmail-smtp-in.l.google.com"
  ],
  "message": "Found 3 MX record(s) for gmail.com"
}
```

**Error Response:**
```json
{
  "error": "Invalid domain format",
  "message": "'invalid_domain' doesn't look like a valid domain"
}
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## Configuration

The API runs on port `8080` by default. You can customize this by modifying the port in `main.go`.

**Timeouts:**
- DNS lookup: 5 seconds
- HTTP read/write: 15 seconds  
- Graceful shutdown: 30 seconds

## Use Cases

- **Email Validation** - Check if a domain can receive emails before sending
- **Domain Analysis** - Understand mail server configurations
- **Monitoring** - Verify mail services are properly configured
- **API Integration** - Embed domain checking into larger applications

## Technical Details

- **Language:** Go 1.23
- **Dependencies:** Standard library only
- **Architecture:** Simple HTTP server with JSON API
- **Container:** Alpine Linux base (~10MB final image)
- **Security:** Non-root user, static binary, minimal attack surface
