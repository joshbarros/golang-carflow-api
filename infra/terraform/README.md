# CarFlow API - Deployment Options

This directory contains configurations for deploying the CarFlow API.

## Free Deployment Options

### 1. GitHub Pages (Static UI Only)

You can deploy the web UI component to GitHub Pages for free:

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

### 2. Local Development

Run the API locally for development and testing:

```bash
# Build and run the API
go build -o carflow ./cmd
./carflow
```

### 3. Alternative Free Hosting Options

Consider these free hosting options that don't require payment information:

#### Railway.app (Free Tier)
- Offers limited free resources
- No credit card required for basic usage
- Simple GitHub integration

#### Render.com (Free Tier)
- Free web services
- Automatic deploys from GitHub
- No credit card required for basic usage

#### Fly.io (Free Tier)
- 3 shared-cpu VMs with 256MB RAM
- 3GB persistent volume storage
- No credit card required for free usage

## GCP Deployment (Requires Billing Account)

If you decide to use GCP in the future, follow these steps:

### 1. Create a GCP Project

```bash
# Create a new GCP project
gcloud projects create carflow-api-project --name="CarFlow API"

# Set the project as active
gcloud config set project carflow-api-project
```

### 2. Link a Billing Account

You must link a billing account to use GCP services. Note that you can set budget alerts and limits to avoid unexpected charges.

```bash
# List available billing accounts
gcloud billing accounts list

# Link a billing account to your project
gcloud billing projects link carflow-api-project --billing-account=YOUR_BILLING_ACCOUNT_ID
```

### 3. Enable Required APIs

```bash
# Enable required GCP APIs
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable run.googleapis.com
```

### 4. Create Terraform Resources

Follow the steps in the original Terraform configuration if you choose to use GCP.

## Local Terraform Testing

You can test Terraform configurations locally without deploying:

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Plan changes without applying
terraform plan -var="project_id=carflow-api-project"
``` 