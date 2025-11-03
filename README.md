# GlitchMesh 🧪

**GlitchMesh** is a lightweight, developer-focused proxy tool designed for **testing microservice resilience**. It allows you to simulate network faults like latency, errors, and dropped requests, making it easier to validate the robustness of distributed systems, all without touching production code.

---
> ⚠️ **Note:** This project is **under active development**. Many features are still being implemented, and things may change frequently. Feedback and contributions are highly appreciated!
---


## 🚀 Current Features

### Core Proxy Functionality
- **Intelligent Request Routing** - Single-port proxy that routes requests to appropriate backend services based on URL patterns (`/redirect/{service-name}/{endpoint}`)
- **YAML Configuration** - Simple, declarative configuration using YAML files for service definitions and fault rules
- **HTTP Proxy** - Full HTTP request/response proxying with header preservation and query parameter support

### Fault Injection Capabilities
 **Latency Injection** - Add configurable artificial delays to simulate slow downstream services
  - Configurable delay in milliseconds
  - Per-service latency configuration
  
 **Error Simulation** - Simulate service failures and HTTP errors
  - Configurable HTTP status codes (500, 503, etc.)
  - Custom error messages
  - Request termination on error injection

### Configuration & Control
- **Priority-Based Fault Application** - Define fault execution order with priority arrays (`priority: ["error", "latency"]`)
- **Service-Specific Configuration** - Apply different fault profiles to different services
- **Enable/Disable Faults** - Toggle fault injection on/off per service configuration

### Observability
- **Request Logging** - Comprehensive logging of all proxied requests with timing information
- **Fault Logging** - Transparent logging of all injected faults for debugging and monitoring

### Current Configuration Format

```yaml
service:
  - name: "service-one"
    url: "http://localhost:8080/"
    fault:
      enabled: true
      priority: ["error", "latency"]
      types:
        latency:
          delay: 5000  # 5 second delay
        error:
          statuscode: 500
          message: "something went really wrong!!"
```

### Usage

```bash
# Start the proxy server on port 9000
go run ./cmd/glitchmesh/main.go start server

# Make requests through the proxy
curl http://localhost:9000/redirect/service-one/api/users
```

## Upcoming Features

**Network Faults**
  - [x] Connection Timeouts - Simulate services that accept connections but never respond
  - [ ] Connection Drop - Randomly Dropping connections mid request
  - [ ] Bandwidth Throttling - Limiting throughput to simulate network congestion

**Response Manupulation**
  - [ ] Partial response corruption - Corrupt random bytes in response bodies

**Advanced Simulations**
  - [ ] Circuit breaker simulation - Fail fast after consecutive failures

**Other Features**
  - [ ] Hot reloading of configs - reloading configs without restarts
  - [ ] Sidecar Mode for transparent proxy (mainly for K8s)
  - [ ] API Based control - to enable/disable faults using API endpoints
  - [x] Percentage Based Faults - Apply faults to only X% of the requests
