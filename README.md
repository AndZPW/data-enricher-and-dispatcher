# API Client with Retry Logic

A Go service that processes users from an API with:

- Exponential backoff retry logic
- Docker containerization
- Context cancellation support

## Quick Start

### Run Locally

```
# Clone repository
git clone https://github.com/AndZPW/data-enricher-and-dispatcher
cd data-enricher-and-dispatcher

# Build and run
go run cmd/app/main.go
```

### Docker Setup

```
# Build image
docker build -t api-client .

# Run container
docker run api-client
```

## Configuration

Environment variables:

- ENV (default: DEV) – env verbosity (`DEV`, `PROD`)
- API_A_URL (default: https://jsonplaceholder.typicode.com/users) – Source API endpoint
- API_B_URL (default: https://webhook.site ) – Target API endpoint
- MAX_RETRIES (default: 3) – Max retry attempts
- RETRY_DELAY_MS (default: 1000) – Base retry delay in milliseconds
- TIMEOUT (default: 10) - Timeout for source API endpoint in seconds
- LOG_LEVEL (default: info) – Log verbosity (`debug`, `info`, `warn`, `error`)  