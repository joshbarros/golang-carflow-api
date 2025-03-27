# 🧾 PRD - Car Management Microservice (Standard Library Only)

### 🏷️ Project Name: **CarFlow**

---

## 🎯 Purpose

The purpose of **CarFlow** is to provide a simple, self-contained backend microservice written in Go, to manage car entities via a RESTful API using **only the Go standard library** (`net/http`). The system will be used to demonstrate code quality, architecture principles, automated testing, and observability practices without relying on any external dependencies.

This PRD outlines the technical specs, endpoints, architecture, and observability plan that adheres strictly to the original coding challenge requirements.

---

## 🧩 Features

| Feature                     | Description |
|-----------------------------|-------------|
| 🔹 **CRUD for Cars**         | Create, list, retrieve and update cars |
| 🔹 **In-Memory Store**       | Store cars in a simple Go map |
| 🔹 **OpenAPI Docs**          | Document endpoints (static JSON/YAML file) |
| 🔹 **Logging**               | Standard library logging middleware |
| 🔹 **Metrics Endpoint**      | Basic metrics exposed via custom `/metrics` JSON |
| 🔹 **Tests**                 | Automated tests using `testing` and `httptest` |
| 🔹 **Healthcheck**           | Simple endpoint for system readiness |

---

## 🏛️ Architecture Overview

### 🧱 Base Stack

| Layer           | Technology                         |
|-----------------|-------------------------------------|
| Language        | Golang (net/http only)              |
| DB              | In-memory `map[string]Car`          |
| API Format      | JSON (application/json)             |
| Docs            | OpenAPI 3.0 (static file)           |
| Observability   | Standard `log` package + custom metrics |
| Tests           | Go `testing` + `httptest`           |

---

## 🧠 Entity Design

### 🛻 Car

```go
type Car struct {
  ID     string `json:"id"`
  Make   string `json:"make"`
  Model  string `json:"model"`
  Year   int    `json:"year"`
  Color  string `json:"color"`
}
```

---

## 🌐 REST API

| Method | Path         | Description        | Status Codes      |
|--------|--------------|--------------------|-------------------|
| GET    | `/cars`      | List all cars      | 200               |
| GET    | `/cars/{id}` | Get car by ID      | 200, 404          |
| POST   | `/cars`      | Create new car     | 201, 400          |
| PUT    | `/cars/{id}` | Update existing    | 200, 400, 404     |

---

## ⚙️ Additional Endpoints

| Endpoint      | Description                       |
|---------------|-----------------------------------|
| `/metrics`    | Custom JSON metrics format        |
| `/healthz`    | Health check                      |
| `/api-docs`   | Static OpenAPI JSON/YAML file     |

---

## 📁 Folder Structure

Simple structure using standard Go conventions:

```
/cmd
  main.go                  # Entry point
/internal
  /car
    handler.go             # HTTP handlers
    service.go             # Business logic
    storage.go             # In-memory DB logic
    model.go               # Entity struct
  /middleware
    logger.go              # Logging middleware
    recovery.go            # Panic recovery
  /metrics
    metrics.go             # Custom metrics tracking
    handler.go             # Metrics endpoint
  /health
    health.go              # Healthcheck handler
/docs
  openapi.json             # OpenAPI 3.0 Spec (static)
/test
  car_test.go              # Integration tests
  service_test.go          # Unit tests
Makefile                   # Build and run commands
README.md                  # Documentation
```

---

## 📈 Observability

| Feature        | Tool / Strategy                    |
|----------------|-----------------------------------|
| Logs           | Standard `log` package with custom format |
| Request Logs   | Middleware with timing, status, method, path |
| Metrics        | Custom metrics handler returning JSON |
| Request IDs    | Generate and log UUID v4 for tracing requests |
| Healthcheck    | `/healthz` endpoint                |

### Custom Metrics Implementation

```go
// Simple in-memory metrics without external dependencies
type Metrics struct {
  RequestCount   int64
  ErrorCount     int64
  ResponseTimes  []time.Duration // For calculating percentiles
  LastRequests   []RequestInfo   // Limited circular buffer
  mu             sync.Mutex
}

type RequestInfo struct {
  Path      string
  Method    string
  Status    int
  Duration  time.Duration
  Timestamp time.Time
}
```

---

## 🧪 Testing Strategy

| Layer        | Type              | Tools Used         |
|--------------|-------------------|--------------------|
| Handlers     | Integration tests | `httptest`         |
| Business     | Unit tests        | `testing`          |
| Middleware   | Coverage          | `net/http` mocks   |

✅ Add sample test data using a `seed()` function

### Test Example:

```go
func TestGetCar(t *testing.T) {
  // Setup test server with handler
  store := NewInMemoryStore()
  store.Add(Car{ID: "test-1", Make: "Tesla", Model: "Model 3", Year: 2022, Color: "red"})
  
  handler := NewCarHandler(store)
  server := httptest.NewServer(handler)
  defer server.Close()
  
  // Make request
  resp, err := http.Get(server.URL + "/cars/test-1")
  if err != nil {
    t.Fatalf("Error making GET request: %v", err)
  }
  defer resp.Body.Close()
  
  // Assert status code
  if resp.StatusCode != http.StatusOK {
    t.Errorf("Expected status 200, got %d", resp.StatusCode)
  }
  
  // Assert response body
  var car Car
  err = json.NewDecoder(resp.Body).Decode(&car)
  if err != nil {
    t.Fatalf("Error decoding response: %v", err)
  }
  
  if car.ID != "test-1" || car.Make != "Tesla" {
    t.Errorf("Unexpected car data: %+v", car)
  }
}
```

---

## 📖 README Highlights

- Project intro
- How to run locally
- Example cURL commands
- Testing approach
- Implementation details

---

## 🧠 Stretch Goals (Standard Library Only)

- Add `DELETE /cars/{id}` endpoint
- Implement sorting and filtering using URL query parameters
- Add validation logic for car data
- Implement basic caching mechanism
- Add request throttling/rate limiting
- Implement a simple CLI tool to interact with the API
- Add a basic HTML UI using standard library templates
- Add context handling for timeouts and cancellation
- Implement simple ETags for caching