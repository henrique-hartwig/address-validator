# Address Validator API

API REST in Go for validating and normalizing free-form addresses (text typed naturally by the user) with automatic typo correction, error handling, and structured JSON response.

---

## Index

- [Overview](#overview)
- [Architectural Decisions](#architectural-decisions)
- [Technical Components](#technical-components)
- [System Architecture](#system-architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)

---

## Overview

This API accepts fre-form addresses (text typed naturally by the user), process common typing errors and returns normalized address components (street, number, city, state, postal code) in structured JSON format.

### Problem Solved

Users frequently make errors when typing addresses:
- Typos in street names: "Main Stret" → "Main Street"
- Inconsistent abbreviations: "Ave" → "Avenue", "St." → "Street"
- Errors in city names: "San Fransisco" → "San Francisco"
- States with incorrect spelling: "Californa" → "California"

This API automatically detects and corrects these errors before validating the address.

---

## Architectural Decisions

### Approach 1: Own Database + Cache (Not Chosen)

**Concept**: Maintain an own database with US address data and implement caching for fast lookups.

**Pros**:
- Total control over data and infrastructure
- Potentially lower cost at scale
- No concerns with rate limiting

**Cons**:
- Overhead of maintaining the database
- Need to update the address data constantly
- Still need to use external data sources for updates (making the approach circular)
- Higher initial infrastructure costs
- Complex data synchronization logic

**Verdict**: Rejected due to the complexity of maintenance and circular dependency on external data sources.

---

### Approach 2: API Wrapper with Intelligent Routing (Chosen)

**Concept**: Use this API as an intelligent wrapper around external geocoding services, adding value through:
- Correction of typos and normalization of input
- Optimization of requests via caching
- Multi-provider routing and fallback
- Unified response format

**Pros**:
- Utilizes existing and maintained address databases
- Focus on added value (typos correction, optimization)
- Minimal infrastructure maintenance
- Address data always updated

**Cons**:
- Dependency on external services
- Rate limits by provider
- Potential latency of external calls

**Solutions for the Cons**:

#### 1. Multi-Provider Strategy

Usage of two external geocoding providers with automatic fallback. There is many ohter APIs in market, and depending on project needs, we can easily add/replace the API providers. I've choose:

**Provider A (Primary): Geoapify**
- URL: `https://api.geoapify.com/v1/geocode/search`
- Free Tier: 3.000 requests/day
- Used as primary provider

**Provider B (Fallback): Smarty Streets**
- URL: `https://us-autocomplete-pro.api.smarty.com/lookup`
- Free Tier: Varies according to plan
- Used automatically if Provider A fails

**Benefits**:
- Distribute requests between providers to maximize free tier
- Fallback automatically if a provider fails or reaches limit
- Unified response format independent of the provider used
- Resilience: system continues to work even if a provider is offline

#### 2. Redis Cache Layer

**Why Redis?**

Initially we considered in-memory cache (go-cache), but we migrated to Redis for the following reasons:

**In-memory Cache (go-cache) - Limitations**:
- Cache lost upon application restart
- Impossible to scale horizontally
- Not shared between multiple instances

**Redis - Benefits**:
- **Cache Persistent**: Survives application restarts
- **Horizontal Scalability**: Allows multiple APIs to use the same cache
- **Distributed Cache**: Shared between multiple instances of the API
- **Production-Ready**: Battle-tested solution, used by large scale companies
- **Monitoring**: Mature monitoring tools
- **Performance**: Sub-millisecond latency even with distributed data

**Implementation**:
- Cache validation before making external requests, reducing external API calls
- Significant cost reduction for repeated queries
- Dramatic improvement in response times for cached addresses
- TTL configurable (default: 24h)

**Cache Key Generation**:
```go
// Key: "addr:" + MD5(lowercase(normalized_address))
// Example: "addr:5d41402abc4b2a76b9719d911017c592"
```

**Trade-off Acceptable**:
- Access to in-memory cache: ~1-10 µs
- Access to Redis local: ~100-500 µs
- Access to Redis in network: ~1-5 ms
- **Conclusion**: Small overhead acceptable for the benefits of scalability and persistence

#### 3. Normalization of Input

- Pre-process user input to correct common typos
- Standardize abbreviations (St. → Street, Ave → Avenue)
- Reduce cache misses and improve cache hit rate
- Apply corrections before even querying the cache

---

## Technical Components

### Strategy for Typos Correction

We use Levenshtein edit distance (deterministic, rule-driven) instead of NLP/ML to reduce complexity and improve performance.

**Open-Source Libraries**:
- github.com/agnivade/levenshtein — Edit distance calculation
- Custom normalization rules for common address abbreviations

**Why edit-distance instead of NLP (Natural Language Processing)?**

**Edit-distance + rules**:
- Well-defined problem with structured patterns
- Lightweight, fast, and deterministic
- Lower latency and resource usage
- No training or model maintenance
- Predictable and explainable results

**NLP/Machine Learning**:
- Would be overkill for this specific use case
- Adds unnecessary complexity
- Requires additional infrastructure (GPU, models, retraining)
- Higher latency and computational cost
- Non-deterministic results

### Normalization Dictionaries

**Street Types** (`CommonStreetTypes`):
- street, avenue, boulevard, road, drive, lane, court, place, terrace, way, parkway, circle, square, trail

**Common City Names** (`CommonCityNames`):
- San Francisco, Los Angeles, New York, Chicago, Houston, Phoenix, Philadelphia, etc.

**US States** (`USStates`):
- Complete mapping of full names → abbreviations
- Correction of typos in state names
- Normalization of case (California → CA, californa → CA)

**Cardinal Directions** (`DirectionAbbreviations`):
- N, S, E, W, NE, NW, SE, SW → North, South, East, West, etc.

**Street Abbreviations** (`StreetAbbreviations`):
- St → Street, Ave → Avenue, Blvd → Boulevard, Rd → Road, Dr → Drive, etc.

### Correction Algorithm

1. **Basic Normalization**: Trim and clean spaces
2. **Expansion of Abbreviations**: Convert known abbreviations
3. **Typos Correction** (Levenshtein distance ≤ 2):
   - For words ≥ 4 characters (except numbers)
   - Search in street type dictionary
   - Search in city dictionary (words ≥ 6 characters)
4. **Normalization of States**: Detect and correct states
5. **Final Formatting**: Remove duplicate spaces and normalize punctuation

---

## System Architecture

```
┌─────────────────────┐
│   User Input        │
│   (free-form text)  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Typo Correction &   │
│ Normalization       │
│ (Levenshtein-based) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Redis Cache        │◄──────┐
│  Lookup             │       │
└──────────┬──────────┘       │
           │                  │
           ▼                  │
      [Cache Hit?]            │
           │                  │
      ┌────┴────┐             │
     No        Yes            │
      │          │            │
      ▼          └────────────┘
┌──────────┐
│ Provider │
│ Routing  │
└────┬─────┘
     │
     ▼
┌──────────────────────┐
│  Multi-Provider      │
│  Load Balancing      │
│                      │
│  ┌────────────────┐  │
│  │ Geocoding API A│  │
│  │ (Primary)      │  │
│  └────────────────┘  │
│          │           │
│          ▼           │
│    [Success?]        │
│          │           │
│     ┌────┴────┐      │
│    No        Yes     │
│     │          │     │
│     ▼          │     │
│  ┌──────────┐  │     │
│  │Geocoding │  │     │
│  │ API B    │  │     │
│  │(Fallback)│  │     │
│  └──────────┘  │     │
│       │        │     │
│       └────┬───┘     │
└────────────┼─────────┘
             │
             ▼
    ┌────────────────┐
    │ Parse Response │
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │  Cache Result  │───► Redis
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │ Format JSON    │
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │ JSON Response  │
    └────────────────┘
```

### Data Flow

1. **Input**: User sends free-form address
2. **Normalization**: Typos correction and standardization
3. **Cache Check**: Check Redis with MD5 key
4. **Cache Hit**: Return result immediately
5. **Cache Miss**: Proceed to geocoding
6. **Provider Selection**: Choose available provider (A or B)
7. **Fallback**: If A fails, try B automatically
8. **Parse**: Convert provider response to standard format
9. **Cache Store**: Save result in Redis (TTL: 24h)
10. **Response**: Return structured JSON to client

---

## Project Structure

```
address-validator/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point
│
├── config/
│   └── config.go                      # Configuration management
|
├── docs/
│   └── docs.go                        # Documentation
│   └── swagger.json
│   └── swagger.yaml
│
├── internal/
│   ├── handlers/
│   │   └── address.go                 # HTTP handlers
│   │
│   └── middleware/
│   |   └── logger.go                  # Middleware for logging
│   |   └── auth.go                    # Middleware for authentication
│   │
│   ├── models/
│   │   └── address.go                 # Data structures
│   │
│   └── services/
│       ├── validator.go               # Validation logic
│       ├── address_dictionary.go      # Normalization dictionaries
│       ├── address_dictionary_test.go # Test with dictionary
│       ├── cache_integration_test.go  # Test with testcontainers
│       ├── cache_interface.go         # Cache interface
│       ├── cache_mock_test.go         # Mock for unit tests
│       ├── cache.go                   # Redis implementation
│       ├── geocoding.go               # Integration with external APIs
│       ├── validator_test.go          # Test with validation
│       └── validator.go               # Validation logic
│
├── .env.example                       # Environment variables template
├── docker-compose.yml                 # Docker orchestration (API + Redis)
├── Dockerfile                         # API container (Go 1.24)
├── go.mod                             # Go dependencies
├── go.sum                             # Checksums of dependencies
├── Makefile                           # Build, test, run commands
├── README.md                          # This file
└── RUN.md                             # Execution guide
```

### Architectural Patterns

**Clean Architecture**:
- `cmd/`: Entry layer
- `internal/handlers/`: Presentation layer (HTTP)
- `internal/services/`: Domain layer (business logic)
- `internal/models/`: Domain entities
- `config/`: External configuration

**Dependency Injection**:
- Services are injected into handlers
- Cache is injected into services
- Allows easy mock in tests

**Interface Segregation**:
- `Cache` interface for implementation abstraction
- Allows swapping Redis for Memcached, etc. using the interface

---

## Tech Stack

### Language and Framework

- **Go 1.24**: Performance, simple deploy
- **Gin**: Web framework for REST APIs

### Cache and Persistence

- **Redis 7 Alpine**: Distributed, persistent and scalable cache
- **github.com/redis/go-redis/v9**: Official Redis client

**Why Go?**
- Static compilation (single binary)
- Excellent performance in I/O and networking
- Goroutines for efficient concurrency (but not applied here)
- Simplified deploy (no runtime dependencies)
- Strong typing reduces bugs in production

**Why Gin?**
- One of the fastest frameworks in Go
- Efficient routing
- Flexible middleware chain
- Integrated JSON validation
- Excellent documentation


### Tests

- **testing (stdlib)**: Native Go test framework
- **github.com/testcontainers/testcontainers-go**: Integration tests with Docker
- **github.com/testcontainers/testcontainers-go/modules/redis**: Redis testcontainer

**Why Testcontainers?**
- Integration tests with Redis real (not mock)
- Containers ephemeral for each test
- Environment identical to production
- Cleanup automatic
- CI/CD friendly

### Infrastructure

- **Docker**: Containerization of the application
- **Docker Compose**: Local orchestration (API + Redis)
- **Alpine Linux**: Small Docker images (~15MB final)

### Configuration

- **github.com/joho/godotenv**: Environment variable management
- Configuration via ENV vars

### API Documentation

- **github.com/swaggo/swag**: Automatic generation of Swagger/OpenAPI documentation
- **github.com/swaggo/gin-swagger**: Swagger UI integrated with Gin
- **github.com/swaggo/files**: Serve Swagger files

**Why Swagger?**
- Documentation always synchronized with the code (automatically generated)
- Interactive interface to test the API
- Automatic generation from code comments
- OpenAPI 3.0 standard
- Simple regeneration: `make swagger`
- Easly import in Postman or other tools

---

## API Documentation (Swagger)

The API has interactive documentation generated automatically with Swagger/OpenAPI.

**Access**: `http://localhost:3000/swagger/index.html`

**Regenerate documentation** (after code changes):
```bash
make swagger
```

The Swagger files are generated in `/docs`:
- `docs.go` - Go code with documentation
- `swagger.json` - OpenAPI specification in JSON
- `swagger.yaml` - OpenAPI specification in YAML

---

## API Endpoints

### POST /api/v1/validate-address

Validate and normalize a free-form address.

**Authentication**: Bearer Token (required)

**Headers**:
```
Authorization: Bearer your_token_here
Content-Type: application/json
Accept: application/json
```

**Request**:
```json
{
  "address": "123 Main Stret, San Fransisco, CA, 94102"
}
```

**Response (Success)**:
```json
{
  "status": "success",
  "data": {
    "street": "Main Street",
    "number": "123",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94102",
    "country": "United States",
    "formatted": "123 Main Street, San Francisco, CA 94102"
  },
  "corrections": [
    "Stret → street (typo correction)",
    "Fransisco → Francisco (city correction)"
  ]
}
```

**Response (Error)**:
```json
{
  "status": "error",
  "error": "Failed to validate address: address not found"
}
```

### GET /health

Health check endpoint.

**Response**:
```json
{
  "status": "healthy",
  "service": "address-validator"
}
```

---

### Logging

- Request/Response logging via middleware
- Structured logging with contextual fields
- Log of external API errors
- Log of cache hits/misses

### Potential Future Improvements

- [ ] Rate limiting by IP/user
- [ ] Prometheus Metrics (latency, error rate, cache hit rate)
- [ ] Distributed tracing with Jaeger/OpenTelemetry

---


## Security

### Implemented

- **Bearer Token Authentication**: Required authentication via token
- **Header Validation**: Content-Type and Accept headers validation
- **Environment Variables**: Secrets managed via environment variables
- **No Hardcoded Credentials**: No credentials in code
- **Input Validation**: JSON schema validation


**Authentication Errors**:
- `400 Bad Request`: Invalid request: address field is required
- `401 Unauthorized`: Token missing, invalid or malformed
- `415 Unsupported Media Type`: Content-Type incorrect
- `406 Not Acceptable`: Accept header does not include application/json

### Production Recommendations

- [ ] **Rate Limiting**: Implement by IP/token
- [ ] **HTTPS Required**: TLS/SSL in production
- [ ] **Token Rotation**: Periodic token rotation
- [ ] **Audit Logging**: Log of all authenticated requests
- [ ] **Token with Expiration**: Implement JWT with expiration

---


## Additional Documentation

- **[RUN.md](RUN.md)**: Complete guide to execute, test and deploy the application
- **[PROMPT.go](PROMPT.go)**: Used prompts using Cursor to get speed and help to decide some technical and architectural decisions
