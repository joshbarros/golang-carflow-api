# 🏗️ CarFlow API Architecture

## 📋 System Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────────────┐
│                 │     │                 │     │                         │
│    Client       │────▶│    Cloud Run    │────▶│    In-Memory Storage    │
│  Applications   │◀────│    API Server   │◀────│     (Future: PostgreSQL)│
│                 │     │                 │     │                         │
└─────────────────┘     └─────────────────┘     └─────────────────────────┘
                              │     ▲
                              │     │
                              ▼     │
                        ┌─────────────────┐
                        │                 │
                        │   Middleware    │
                        │      Chain      │
                        │                 │
                        └─────────────────┘
```

## 🧩 Architecture Design Principles

### 📊 Clean Architecture

The CarFlow API follows the principles of Clean Architecture, separating concerns into distinct layers:

1. **Domain Layer** - Core business entities and logic
2. **Service Layer** - Application-specific business rules
3. **Handler Layer** - HTTP request/response handling
4. **Infrastructure Layer** - External systems integration

This separation ensures:
- ✅ Business logic is independent of frameworks and external systems
- ✅ Easy testability at all levels
- ✅ Flexibility to change infrastructure without affecting business logic

## 🔄 Request Flow

```
  ┌──────────┐      ┌────────────┐      ┌────────────┐      ┌─────────────┐      ┌────────────┐
  │          │      │            │      │            │      │             │      │            │
  │  Client  │─────▶│ Middleware │─────▶│  Handler   │─────▶│  Service    │─────▶│ Repository │
  │ Request  │      │   Chain    │      │  Layer     │      │  Layer      │      │ Layer      │
  │          │      │            │      │            │      │             │      │            │
  └──────────┘      └────────────┘      └────────────┘      └─────────────┘      └────────────┘
                                                                                       │
  ┌──────────┐      ┌────────────┐      ┌────────────┐      ┌─────────────┐           │
  │          │      │            │      │            │      │             │◀───────────┘
  │  Client  │◀─────│ Response   │◀─────│  Handler   │◀─────│  Service    │
  │ Response │      │ Processing │      │  Layer     │      │  Layer      │
  │          │      │            │      │            │      │             │
  └──────────┘      └────────────┘      └────────────┘      └─────────────┘
```

## 📦 Component Architecture

### 🚗 Car Domain Model

```go
type Car struct {
    ID    string
    Make  string
    Model string
    Year  int
    Color string
}
```

Design decisions:
- ✅ Simple, focused entity with clear purpose
- ✅ String ID for flexibility in ID generation strategies
- ✅ Core data model with essential car properties

### 🗄️ Repository Layer

```
┌─────────────────────┐     ┌─────────────────────────┐
│                     │     │                         │
│   Repository        │     │   In-Memory Storage     │
│   Interface         │◀───▶│   Implementation        │
│                     │     │                         │
└─────────────────────┘     └─────────────────────────┘
         ▲
         │
         │
┌─────────────────────┐
│                     │
│   Future Database   │
│   Implementation    │
│                     │
└─────────────────────┘
```

Design decisions:
- ✅ Repository pattern for data access abstraction
- ✅ Interface-based design for easy swapping of storage implementations
- ✅ In-memory implementation for initial development and testing
- ✅ Future-proof for database implementation

### 🛠️ Service Layer

```
┌─────────────────────┐     ┌─────────────────────────┐
│                     │     │                         │
│   Service Layer     │────▶│   Repository Layer      │
│                     │◀────│                         │
└─────────────────────┘     └─────────────────────────┘
         │
         │
         ▼
┌─────────────────────┐
│                     │
│   Business Logic    │
│   & Validation      │
│                     │
└─────────────────────┘
```

Design decisions:
- ✅ Service layer encapsulates business logic
- ✅ Input validation at the service level ensures data integrity
- ✅ Service layer orchestrates repository operations
- ✅ Clear separation of concerns from HTTP handling

### 🌐 HTTP Handler Layer

```
┌─────────────────────┐     ┌─────────────────────────┐
│                     │     │                         │
│   HTTP Handlers     │────▶│   Service Layer         │
│                     │◀────│                         │
└─────────────────────┘     └─────────────────────────┘
         │
         │
         ▼
┌─────────────────────┐
│                     │
│   Request/Response  │
│   Processing        │
│                     │
└─────────────────────┘
```

Design decisions:
- ✅ RESTful API design principles
- ✅ Clear handler responsibilities: parse request, validate, call service, format response
- ✅ Error handling standardized across handlers
- ✅ JSON response formatting with consistent structure

### 🔄 Middleware Architecture

```
┌────────────┐     ┌────────────┐     ┌────────────┐     ┌────────────┐     ┌────────────┐
│            │     │            │     │            │     │            │     │            │
│  Logging   │────▶│  Recovery  │────▶│   CORS     │────▶│    Rate    │────▶│   ETag     │
│ Middleware │     │ Middleware │     │ Middleware │     │  Limiting  │     │ Middleware │
│            │     │            │     │            │     │            │     │            │
└────────────┘     └────────────┘     └────────────┘     └────────────┘     └────────────┘
                                                                                   │
                   ┌────────────┐                                                  │
                   │            │                                                  │
                   │  Metrics   │◀─────────────────────────────────────────────────┘
                   │ Middleware │
                   │            │
                   └────────────┘
                        │
                        ▼
                   ┌────────────┐
                   │            │
                   │   Route    │
                   │  Handlers  │
                   │            │
                   └────────────┘
```

Design decisions:
- ✅ Middleware chain for cross-cutting concerns
- ✅ Composable design for flexibility
- ✅ Each middleware has a single responsibility
- ✅ Standard HTTP middleware pattern (take next handler, return handler)

## 📊 API Design

### RESTful Resource Modeling

```
┌─────────────────────────────────────────────┐
│                                             │
│  /cars                                      │
│  GET    - List all cars (filtering/sorting) │
│  POST   - Create a new car                  │
│                                             │
└─────────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
┌───────▼─────────┐     ┌───────▼─────────┐
│                 │     │                 │
│  /cars/{id}     │     │  /metrics       │
│  GET    - Read  │     │  GET - API      │
│  PUT    - Update│     │       metrics   │
│  DELETE - Delete│     │                 │
│                 │     │                 │
└─────────────────┘     └─────────────────┘
```

Design decisions:
- ✅ RESTful API design with standard HTTP methods
- ✅ Resource-based URL structure
- ✅ Query parameters for filtering and sorting
- ✅ Clear separation of resources

### 📋 Query Parameter Design

| Endpoint | Parameter | Description | Example |
|----------|-----------|-------------|---------|
| GET /cars | make | Filter by manufacturer | `?make=Toyota` |
| GET /cars | model | Filter by model | `?model=Corolla` |
| GET /cars | year | Filter by year | `?year=2020` |
| GET /cars | color | Filter by color | `?color=blue` |
| GET /cars | sort | Sort field (prefix with - for desc) | `?sort=-year` |
| GET /cars | page | Page number | `?page=2` |
| GET /cars | page_size | Items per page | `?page_size=10` |

## ☁️ Deployment Architecture

### Google Cloud Run Deployment

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │     │                    │
│  GitHub Actions    │────▶│   Cloud Build      │────▶│   Container        │
│     Workflow       │     │                    │     │   Registry         │
│                    │     │                    │     │                    │
└────────────────────┘     └────────────────────┘     └────────────────────┘
                                                              │
                                                              │
                                                              ▼
┌────────────────────┐                            ┌────────────────────┐
│                    │                            │                    │
│    Cloud IAM       │◀──────────────────────────▶│    Cloud Run       │
│                    │                            │    Service         │
│                    │                            │                    │
└────────────────────┘                            └────────────────────┘
        │                                                   ▲
        │                                                   │
        ▼                                                   │
┌────────────────────┐                            ┌────────────────────┐
│                    │                            │                    │
│   Terraform IaC    │───────────────────────────▶│   Auto-scaling     │
│   Configuration    │                            │   Configuration    │
│                    │                            │                    │
└────────────────────┘                            └────────────────────┘
```

Design decisions:
- ✅ Containerized deployment with Docker
- ✅ Cloud Run for serverless container execution
- ✅ Infrastructure as Code with Terraform
- ✅ CI/CD automation with GitHub Actions
- ✅ Auto-scaling configuration (0-1 instances)
- ✅ Cost optimization for free tier

## 📈 Observability Architecture

```
┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │
│   Application      │────▶│   Metrics          │
│      Logs          │     │   Endpoint         │
│                    │     │                    │
└────────────────────┘     └────────────────────┘
        │                           │
        │                           │
        ▼                           ▼
┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │
│   Cloud Logging    │     │   Health Check     │
│                    │     │   Endpoint         │
│                    │     │                    │
└────────────────────┘     └────────────────────┘
```

Design decisions:
- ✅ Built-in metrics collection
- ✅ Health check endpoint for monitoring
- ✅ Structured logging
- ✅ Request timing measurements
- ✅ Error tracking and reporting

## 🔐 Security Architecture

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │     │                    │
│   Rate Limiting    │────▶│   CORS Policy      │────▶│   Recovery         │
│   Middleware       │     │                    │     │   Middleware       │
│                    │     │                    │     │                    │
└────────────────────┘     └────────────────────┘     └────────────────────┘
                                                              │
                                                              │
                                                              ▼
┌────────────────────┐                            ┌────────────────────┐
│                    │                            │                    │
│   Input            │◀──────────────────────────▶│   Service Account  │
│   Validation       │                            │   Authentication   │
│                    │                            │                    │
└────────────────────┘                            └────────────────────┘
```

Design decisions:
- ✅ Rate limiting to prevent abuse
- ✅ CORS configuration for browser security
- ✅ Panic recovery to prevent service disruption
- ✅ Input validation for all operations
- ✅ Service account-based authentication for GCP resources

## 🔄 Future Architecture Enhancements

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │     │                    │
│   PostgreSQL       │────▶│   Redis Cache      │────▶│   Authentication   │
│   Database         │     │                    │     │   Service          │
│                    │     │                    │     │                    │
└────────────────────┘     └────────────────────┘     └────────────────────┘
        │                           │                          │
        │                           │                          │
        ▼                           ▼                          ▼
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│                    │     │                    │     │                    │
│   Event-Driven     │     │   Advanced         │     │   Stripe Payment   │
│   Architecture     │     │   Monitoring       │     │   Integration      │
│                    │     │                    │     │                    │
└────────────────────┘     └────────────────────┘     └────────────────────┘
```

Design decisions for future enhancements:
- ✅ Database persistence with PostgreSQL
- ✅ Caching layer with Redis
- ✅ JWT-based authentication and authorization
- ✅ Event-driven architecture for scalability
- ✅ Advanced monitoring and observability
- ✅ Integration with third-party services (Stripe, Brevo, Twilio)

## 🧠 Key Architectural Decisions

1. **Lightweight Initial Design**
   - Started with a minimal, in-memory implementation for rapid development
   - Designed with future database integration in mind

2. **Clean Separation of Concerns**
   - Core domain model independent of infrastructure
   - Repository pattern for data access abstraction
   - Service layer for business logic
   - Handler layer for HTTP concerns

3. **RESTful API Design**
   - Resource-based routing
   - Standard HTTP methods
   - Query parameters for filtering, sorting, and pagination
   - Consistent response formats

4. **Cloud-Native Deployment**
   - Containerized with Docker
   - Deployed to serverless Cloud Run
   - Infrastructure as Code with Terraform
   - CI/CD with GitHub Actions

5. **Observability First**
   - Built-in metrics collection
   - Health check endpoint
   - Structured logging

6. **Security Focus**
   - Rate limiting
   - Input validation
   - CORS configuration
   - Recovery middleware
   - Service account authentication

7. **Future Scaling Considerations**
   - Database persistence
   - Caching
   - Authentication
   - Monitoring
   - Third-party integrations

These architectural decisions ensure the CarFlow API is:
- ✅ Maintainable
- ✅ Testable
- ✅ Scalable
- ✅ Secure
- ✅ Observable
- ✅ Cloud-native
- ✅ Cost-effective 