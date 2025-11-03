# delivery-route-system

A delivery route optimization system that finds the fastest routes between a source location and multiple destinations using the OSRM routing engine.

## Assignment Overview

### Problem Statement

We're building shipping solutions. One of the problems we face is presenting a list of pickup locations available in a customer's area so that they can have their package delivered to the nearest location in their neighborhood.

**The Question**: Which destination is closest to the source and how fast can one get there by car?

### Requirements

#### Functional Requirements

1. **API Endpoint**: Build a web service that accepts:
   - `src`: Source location (customer's home) - single latitude,longitude pair
   - `dst`: Destination locations (pickup points) - multiple latitude,longitude pairs

2. **Response Format**: Return a list of routes sorted by:
   - **Primary**: Driving time (duration in seconds)
   - **Secondary**: Distance (in meters, when duration is equal)

3. **External Service**: Must use [OSRM (Open Source Routing Machine)](http://project-osrm.org/) to calculate driving times and distances.


### Non-Functional Requirements

The assignment specified that this is a **highly business-critical production application** in a dynamic and rapidly changing environment. The following requirements are prioritized in order of importance:

#### 1. Fault Tolerance ⭐ (Highest Priority)

**Implementation**:
- **Retry Logic**: HTTP client automatically retries failed requests with exponential backoff
- **Connection Pooling**: Efficient connection reuse with configurable limits (100 max idle, 10 per host)
- **Error Classification**: Intelligent error handling that distinguishes retryable errors (503) from client errors (400)
- **Timeout Management**: Multiple timeout layers:
  - Request-level timeouts (30s default)
  - HTTP client timeouts (3s default)
  - Server read/write timeouts
- **Graceful Degradation**: Health check endpoint detects OSRM unavailability and reports service status
- **Panic Recovery**: Middleware recovers from panics and returns proper error responses
- **Response Validation**: Validates OSRM response structure before processing

#### 2. Maintainability

**Implementation**:
- **Layered Architecture**: Clear separation of concerns (Server → Service → OSRM Client → HTTP Client)
- **Comprehensive Error Handling**: Proper OSRM error code mapping with descriptive error messages
- **Extensive Testing**: Unit tests, integration tests, and validation tests
- **Documentation**: Inline code comments and comprehensive README
- **Logging**: Structured logging with context fields for debugging
- **Code Organization**: Modular design with single responsibility principle

#### 3. Extendability

**Implementation**:
- **Interface-Based Design**: Service layer uses interfaces, making it easy to swap implementations (e.g., different routing providers)
- **Dependency Injection**: Components are injected, not hardcoded
- **Configurable Timeouts**: All timeouts are configurable via server config
- **Mock Support**: OSRM client interface allows easy mocking for testing
- **Scalable Architecture**: Designed to handle increasing numbers of destinations efficiently

### Solution Approach

The solution addresses the assignment requirements through:

1. **API Compliance**: Implements the exact request/response format specified
2. **OSRM Integration**: Uses OSRM Table Service (chosen over Route Service for performance - see Design section)
3. **Sorting Logic**: Routes are sorted by duration first, then distance when durations are equal
4. **Production Readiness**: Implements all three non-functional requirements with production-grade patterns
5. **Developer Experience**: Easy to run, test, and deploy with Make commands and Docker support

## Design

### Architecture Overview

The system follows a layered architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────┐
│           HTTP Server Layer             │
│  (Request/Response, Validation, DTOs)    │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          Service Layer                   │
│  (Business Logic, Route Sorting)         │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│       OSRM Client Layer                  │
│  (External API Integration)             │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│       HTTP Client Layer                  │
│  (Retry Logic, Connection Pooling)       │
└──────────────────────────────────────────┘
```

### Component Layers

1. **Server Layer** (`server/`): Handles HTTP requests, validation, error responses, and middleware (logging, recovery, timeouts)
2. **Service Layer** (`service/`): Contains business logic for route sorting and service orchestration
3. **OSRM Client Layer** (`pkg/osrmclient/`): Encapsulates OSRM API communication and error handling
4. **HTTP Client Layer** (`pkg/httpclient/`): Provides reusable HTTP client with retry logic and connection pooling

### OSRM Table Service vs Route Service

After reading the OSRM API Documentation, the decision was made to use the **Table Service** instead of the Route Service. This design choice was driven by several key factors:

**Batch Processing**: The Table Service allows sending a single request with multiple destinations, enabling efficient batch processing. With the Route Service, we would need to make separate API calls for each destination, resulting in:
- Multiple network round trips
- Increased latency (serial execution or complex parallel coordination)
- Higher API rate limit consumption

**Performance**: The Table Service takes significantly less time to process multiple destinations compared to making individual Route Service requests. For example, requesting routes to 10 destinations:
- **Table Service**: 1 API call with all destinations → ~200-500ms total
- **Route Service**: 10 separate API calls → ~2-5 seconds total (serial) or requires complex async coordination

**Scalability**: The batch approach scales better as the number of destinations increases, making it ideal for delivery route optimization where multiple stops are common.

**API Efficiency**: The Table Service is specifically designed for distance/duration matrices, which is exactly what we need for finding the fastest routes. It returns duration and distance data without unnecessary routing details, making the response smaller and faster to process.


### Data Flow

1. Client sends request with source and destinations → `GET /routes?src=lat,lon&dst=lat,lon&dst=lat,lon`
2. Server validates input (coordinates, format, limits)
3. Service layer requests routes from OSRM client
4. OSRM client makes a single Table Service API call with all destinations
5. Table Service returns duration/distance matrix
6. Service layer sorts routes by duration/distance
7. Server returns sorted routes to client

## Future Improvements

While the current implementation meets all assignment requirements, there are several areas for potential enhancement in a production environment:

### API Design

**POST Instead of GET**: The current implementation uses GET requests with query parameters. For production use, consider switching to POST requests:

- **Limitations of GET**: URLs have practical length limits (typically 2-8 KB depending on server/browser), which restricts the number of destinations that can be sent in a single request
- **Benefits of POST**: Can handle much larger payloads (1-2 MB), allowing requests with many more destinations
- **Use Case**: When customers need to query routes to 100+ pickup locations, POST with JSON body would be more appropriate


### Protocol Selection

**gRPC for Internal Services**: If this service is primarily consumed by other internal services rather than web clients:

- **Performance**: gRPC uses HTTP/2 and Protocol Buffers, offering better performance and efficiency than REST
- **Type Safety**: Strong typing and code generation reduce integration errors

### Additional Enhancements

- **Caching**: Implement caching for frequently requested routes (e.g., Redis) to reduce OSRM API calls
- **Rate Limiting**: Add rate limiting to protect the service from abuse
- **Metrics & Observability**: Add Prometheus metrics and distributed tracing (e.g., OpenTelemetry)

## How to Run

This project uses a Makefile to simplify common operations. Here are the available commands:

### Run Locally

To run the application locally using Go:

```bash
make run
```

This will start the server on port 8000 by default. The server will be available at `http://localhost:8000`.

You can also customize the port by setting the `SERVER_PORT` environment variable:

```bash
SERVER_PORT=:9000 make run
```

### Run Tests

To run all tests:

```bash
make test
```

This executes all tests with verbose output and ensures tests are not cached (`-count=1`).

### Build Docker Image

To build the Docker image:

```bash
make build
```

This creates a Docker image tagged as `delivery-route-system`.

### Run with Docker

To run the application using Docker:

```bash
make run-docker
```

This starts a Docker container mapping port 8000 from the container to your host machine. The server will be available at `http://localhost:8000`.

You can customize the port mapping by running Docker directly:

```bash
docker run -p 9000:8000 delivery-route-system
```

Or override the server port inside the container:

```bash
docker run -p 8000:8000 -e SERVER_PORT=:8000 delivery-route-system
```

### API Endpoints

Once the server is running, you can access:

- **Health Check**: `GET http://localhost:8000/health` - Returns service health status
- **Routes**: `GET http://localhost:8000/routes?src=<lat>,<lon>&dst=<lat>,<lon>` - Get fastest routes to destinations

Example request:
```bash
curl "http://localhost:8000/routes?src=13.388860,52.517037&dst=13.397634,52.529407&dst=13.428555,52.523219"
```

Example response:
```json
{
  "source": "13.388860,52.517037",
  "routes": [
    {
      "destination": "13.397634,52.529407",
      "duration": 465.2,
      "distance": 1879.4
    },
    {
      "destination": "13.428555,52.523219",
      "duration": 712.6,
      "distance": 4123.0
    }
  ]
}
```

**Note**: The routes are automatically sorted by duration (fastest first), with distance used as a tiebreaker when durations are equal.