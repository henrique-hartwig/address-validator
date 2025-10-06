# User guide: Execution - Address Validator API

This guide contains practical instructions to execute, test and deploy the application.

---

## Index

- [Requirements](#requirements)
- [Initial Configuration](#initial-configuration)
- [Local Execution](#local-execution)
- [Docker Execution](#docker-execution)
- [Tests](#tests)
- [Make Commands](#make-commands)
- [Troubleshooting](#troubleshooting)
- [Additional Resources](#additional-resources)

---

## Requirements

### Local Execution

- **Go 1.24+** ([download](https://go.dev/dl/))
- **Redis** (via Docker or installed locally)
- **Make** (optional, but recommended)

### Docker Execution

- **Docker 20.10+** ([download](https://docs.docker.com/get-docker/))
- **Docker Compose 2.0+** (included in Docker Desktop)

---

## Initial Configuration


### 1. Configure the Environment Variables

Create the .env file in the root of the project. Use .env.example as a template.


### 2. Install Go Dependencies

```bash
go mod download
go mod tidy
```

---

## Local Execution

### 1. Start Redis

#### Option A: Docker (Recommended)

```bash
docker run -d \
  --name redis-dev \
  -p 6379:6379 \
  redis:7-alpine
```

#### Option B: Redis Local

```bash
# Ubuntu/Debian
sudo apt-get install redis-server
sudo systemctl start redis

# macOS
brew install redis
brew services start redis
```

### 2. Execute the Application

#### With Make

```bash
make run
```

#### Without Make

```bash
go run cmd/api/main.go
```

### 3. Test the API

```bash
# Health check
curl http://localhost:3000/health

# Validate address
curl -X POST http://localhost:3000/api/v1/validate-address \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_token_here" \
  -d '{
    "address": "123 Main Street, New York, NY"
  }'
```

---

## Docker Execution

### Development (Docker Compose)

#### 1. Build and Start

```bash
docker compose up -d --build
```

#### 2. Check Logs

```bash
docker compose logs -f
```

#### 3. Check Status

```bash
docker compose ps
```

#### 4. Test the API

```bash
curl http://localhost:3000/health
```

#### 5. Stop containers

```bash
docker compose down

# Remove volumes
docker compose down -v
```

### Production (Docker Build Manual)

#### Build the Image

```bash
docker build -t address-validator:latest .
```

---

## Tests

### All Tests

```bash
# With Make
make test

# Without Make
go test -v ./...
```

### Only Unit Tests

Testes unitários são rápidos e não dependem de Docker/Redis:

```bash
# With Make
make test-unit

# Without Make
go test -v -short ./...
```

**Tests included:**
- Normalization of input
- Generation of cache key
- Correction of typos
- Normalization of states
- Matching fuzzy

### Only Integration Tests

Integration tests use testcontainers (require Docker running):

```bash
# With Make
make test-integration

# Without Make
go test -v -run Integration ./internal/services/
```

**Tests included:**
- Cache Set/Get/Delete
- Expiration of TTL
- Flush of cache
- Complex data
- Error handling

### Only Cache Tests

```bash
# With Make
make test-cache

# Without Make
go test -v ./internal/services/ -run TestCache
```

### Tests with Coverage

```bash
# With Make
make test-coverage

# Without Make
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Make Commands

The project includes a Makefile with useful commands:

### Build and Execution

```bash
make build
make run
```

### Tests

```bash
make test
make test-unit
make test-integration
make test-cache
make test-coverage
```

### Docker

```bash
make docker-build
make docker-up      
make docker-down    
make docker-logs    
make docker-restart 
```

### Maintenance

```bash
make clean
make deps 
make fmt  
make lint 
```

---

## Deploy in Production

### Checklist Pre-Deploy

- [ ] Environment variables configured
- [ ] Valid API keys
- [ ] Redis configured with password
- [ ] HTTPS configured
- [ ] Rate limiting implemented
- [ ] Logs centralized
- [ ] Active monitoring
- [ ] Redis backup


## Additional Resources

- **[README.md](README.md)**: Technical documentation
- **[Makefile](Makefile)**: Complete list of commands
- **[docker-compose.yml](docker-compose.yml)**: Container configuration
