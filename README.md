# 🚗 CarFlow API

A simple, lightweight car management microservice built with Go standard library only.

## 📋 Overview

CarFlow is a RESTful API microservice that allows management of car entities. It's built using only Go's standard library (`net/http`) with no external dependencies, and demonstrates modern software practices including testing, CI/CD, and observability.

## 🌟 Features

- **CRUD Operations** for car entities
- **In-Memory Storage** using Go maps
- **RESTful API** with JSON responses
- **OpenAPI Documentation**
- **Command-Line Interface** for API interaction
- **Web UI** built with Go standard library templates
- **Observability** with logging and custom metrics
- **Health Checks** for monitoring system status
- **Rate Limiting** to prevent abuse
- **Caching** for improved performance
- **ETag Support** for resource versioning
- **Automated Testing** using Go's testing packages
- **CI/CD Pipeline** with GitHub Actions
- **Cloud Deployment** using GCP free tier

## 🔧 Tech Stack

- **Backend**: Go (standard library only)
- **API**: RESTful JSON API using `net/http`
- **Storage**: In-memory map
- **Documentation**: OpenAPI 3.0
- **Testing**: Go testing package
- **CI/CD**: GitHub Actions
- **Cloud**: Google Cloud Run (free tier)

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile commands)
- Docker (optional, for containerization)
- GCP account (optional, for cloud deployment)

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/joshbarros/golang-carflow-api.git
   cd golang-carflow-api
   ```

2. Build and run the service:
   ```bash
   make build
   make run
   ```

   Or without Make:
   ```bash
   go build -o carflow ./cmd
   ./carflow
   ```

3. The service will be available at `http://localhost:8080`

### Using the CLI

CarFlow comes with a command-line interface for easy interaction with the API:

1. Build the CLI:
   ```bash
   make build-cli
   ```

2. Run CLI commands:
   ```bash
   # Show all commands
   ./carflow-cli help
   
   # List all cars
   ./carflow-cli list
   
   # Get a specific car
   ./carflow-cli get -id "1"
   
   # Create a new car
   ./carflow-cli create -make "Toyota" -model "Supra" -year 2022 -color "red"
   
   # Update a car
   ./carflow-cli update -id "1" -color "blue"
   
   # Delete a car
   ./carflow-cli delete -id "1"
   
   # Check API health
   ./carflow-cli health
   ```

### Using the Web UI

CarFlow also includes a web-based user interface:

1. Build the UI:
   ```bash
   make build-ui
   ```

2. Run the UI:
   ```bash
   make run-ui
   ```

3. Access the UI in your browser at `http://localhost:3000`

## 📡 API Endpoints

| Method | Path         | Description        | Status Codes      |
|--------|--------------|--------------------|-------------------|
| GET    | `/cars`      | List all cars      | 200               |
| GET    | `/cars/{id}` | Get car by ID      | 200, 404          |
| POST   | `/cars`      | Create new car     | 201, 400          |
| PUT    | `/cars/{id}` | Update existing    | 200, 400, 404     |
| DELETE | `/cars/{id}` | Delete existing    | 204, 404          |
| GET    | `/metrics`   | Service metrics    | 200               |
| GET    | `/healthz`   | Health check       | 200               |
| GET    | `/api-docs`  | API documentation  | 200               |

## 📦 API Examples

### Create a car
```bash
curl -X POST http://localhost:8080/cars \
  -H "Content-Type: application/json" \
  -d '{"make":"Tesla","model":"Model 3","year":2022,"color":"red"}'
```

### Get all cars
```bash
curl http://localhost:8080/cars
```

### Get a specific car
```bash
curl http://localhost:8080/cars/{id}
```

### Update a car
```bash
curl -X PUT http://localhost:8080/cars/{id} \
  -H "Content-Type: application/json" \
  -d '{"make":"Tesla","model":"Model 3","year":2023,"color":"blue"}'
```

### Filter and Sort
```bash
# Filter by make and sort by year descending
curl "http://localhost:8080/cars?make=Tesla&sort=year&order=desc"
```

### Pagination
```bash
# Get page 2 with 5 items per page
curl "http://localhost:8080/cars?page=2&page_size=5"
```

## 🧪 Testing

Run tests with:
```bash
make test
```

Or without Make:
```bash
go test ./... -v
```

Run benchmarks with:
```bash
go test -bench=. -benchmem ./test
```

## ☁️ Deployment Options

There are several ways to deploy the CarFlow API:

### Local Deployment

Run the application on your local machine:
```bash
go build -o carflow ./cmd
./carflow
```

### Docker Deployment

1. Build the Docker image:
   ```bash
   docker build -t carflow:latest .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 carflow:latest
   ```

### GCP Free Tier Deployment

You can deploy the API to Google Cloud Run using the free tier:

1. Set up GCP project with spending controls:
   ```bash
   # Create a new GCP project
   gcloud projects create carflow-api-project --name="CarFlow API"
   
   # Link billing account (required, but set spending limit to $0)
   gcloud billing projects link carflow-api-project --billing-account=YOUR_BILLING_ACCOUNT_ID
   ```

2. Enable required APIs and create resources:
   ```bash
   # Enable APIs
   gcloud services enable run.googleapis.com artifactregistry.googleapis.com cloudbuild.googleapis.com
   
   # Create Docker repository
   gcloud artifacts repositories create carflow-repo \
     --repository-format=docker \
     --location=us-central1
   ```

3. Deploy using Cloud Build:
   ```bash
   # Trigger build and deployment
   gcloud builds submit --config=cloudbuild.yaml
   ```

The Cloud Run service will auto-scale to zero when not in use, helping you stay within free tier limits. For detailed instructions, see [GCP Free Tier Deployment Guide](docs/gcp-free-deployment.md).

### GitHub Pages for UI

You can deploy the web UI component to GitHub Pages:
```bash
# Create a gh-pages branch
git checkout -b gh-pages

# Build the UI
go build -o carflow-ui ./cmd/ui

# Copy UI assets to root
cp -r cmd/ui/templates/* .

# Push to GitHub
git add .
git commit -m "Add GitHub Pages deployment"
git push origin gh-pages
```

## 🔄 CI/CD

The project uses GitHub Actions for continuous integration and deployment:

### CI Workflow

The CI workflow runs on every push and pull request to the main branch:
- Lints the code with golangci-lint
- Runs unit and integration tests
- Reports code coverage

### GCP Deployment with Cloud Build

Cloud Build can be configured to automatically deploy changes when you push to the main branch:

1. Connect GitHub repository to Cloud Build
2. Configure trigger to watch the main branch
3. Use the provided `cloudbuild.yaml` for deployment configuration

## 📂 Project Structure

```
/cmd
  main.go                   # API entry point
  /cli
    main.go                 # CLI entry point
    README.md               # CLI documentation
  /ui
    main.go                 # Web UI entry point
    README.md               # UI documentation
    /templates              # HTML templates
      layout.html           # Base template
      home.html             # Home page
      list.html             # Car listing page
      view.html             # Car details page
      new.html              # Create car form
      edit.html             # Edit car form
      delete.html           # Delete confirmation
      error.html            # Error page
      /static               # Static assets
        /css                # CSS styles
        /js                 # JavaScript files
/internal
  /car
    handler.go             # HTTP handlers
    service.go             # Business logic
    storage.go             # In-memory DB logic
    model.go               # Entity struct
  /middleware
    logger.go              # Logging middleware
    recovery.go            # Panic recovery
    ratelimit.go           # Rate limiting
    etag.go                # ETag support
  /metrics
    metrics.go             # Custom metrics tracking
    handler.go             # Metrics endpoint
  /health
    health.go              # Healthcheck handler
  /cache
    cache.go               # Caching mechanism
/docs
  openapi.json             # OpenAPI 3.0 Spec
  gcp-free-deployment.md   # GCP free tier deployment guide
/test
  car_test.go              # Integration tests
  service_test.go          # Unit tests
  benchmark_test.go        # Performance benchmarks
/terraform
  main.tf                  # Terraform configuration
  variables.tf             # Terraform variables
  terraform.tfvars         # Terraform variable values
  README.md                # Deployment options documentation
/.github
  /workflows
    ci.yml                 # CI workflow
    cd.yml                 # CD workflow
    pages.yml              # GitHub Pages deployment
cloudbuild.yaml            # Cloud Build configuration
```

## 📄 License

[MIT License](LICENSE) 