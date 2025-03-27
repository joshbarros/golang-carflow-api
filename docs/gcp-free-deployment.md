# Deploying CarFlow API to GCP Free Tier

This guide explains how to deploy the CarFlow API to Google Cloud Platform using only free tier resources.

## GCP Free Tier Overview

GCP Free Tier includes:
- Cloud Run: 2 million requests per month, 360,000 GB-seconds of memory, 180,000 vCPU-seconds per month
- Cloud Storage: 5GB per month in US regions (used for Terraform state)
- Artifact Registry: 0.5GB of storage
- Cloud Build: 120 build-minutes per day

These free resources should be sufficient for running the CarFlow API with moderate traffic.

## Prerequisites

1. Google Account
2. GCP Project with billing account enabled (required for resource allocation, but you won't be charged if staying within free tier limits)
3. gcloud CLI installed locally

## Setting Up with Spending Controls

### 1. Create a Budget Alert

First, set up a budget alert to get notified if you approach spending limits:

1. Go to GCP Console > Billing > Budgets & Alerts
2. Create a new Budget
3. Set the budget amount to $0.00
4. Set alerts at 50%, 90%, and 100% of your budget
5. Enable email notifications

### 2. Create a GCP Project

```bash
# Create a new GCP project
gcloud projects create carflow-api-project --name="CarFlow API"

# Set the project as active
gcloud config set project carflow-api-project
```

### 3. Link a Billing Account with Spending Limits

Link your billing account but set a spending limit:

1. Go to GCP Console > Billing
2. Select your project and link your billing account
3. Under "Payment Settings", set up payment method (required)
4. Under "Budgets & Alerts", create a budget with a $0 threshold
5. Enable automatic email alerts

In CLI:
```bash
# List available billing accounts
gcloud billing accounts list

# Link a billing account to your project
gcloud billing projects link carflow-api-project --billing-account=YOUR_BILLING_ACCOUNT_ID
```

### 4. Enable Required APIs

```bash
# Enable required GCP APIs
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable storage.googleapis.com
```

### 5. Create a Service Account

```bash
# Create a service account
gcloud iam service-accounts create carflow-sa \
  --display-name="CarFlow Service Account"

# Grant necessary permissions to the service account
gcloud projects add-iam-policy-binding carflow-api-project \
  --member="serviceAccount:carflow-sa@carflow-api-project.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding carflow-api-project \
  --member="serviceAccount:carflow-sa@carflow-api-project.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding carflow-api-project \
  --member="serviceAccount:carflow-sa@carflow-api-project.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.admin"
```

### 6. Create a Cloud Storage Bucket for Terraform State

```bash
# Create a GCS bucket for storing Terraform state
gcloud storage buckets create gs://carflow-terraform-state \
  --location=us-central1 \
  --uniform-bucket-level-access
```

### 7. Create an Artifact Registry Repository

```bash
# Create a Docker repository
gcloud artifacts repositories create carflow-repo \
  --repository-format=docker \
  --location=us-central1 \
  --description="CarFlow Docker Repository"
```

## Deploying with Google Cloud Build

### 1. Set up Cloud Build Configuration

Create a file named `cloudbuild.yaml` in the root of your project:

```yaml
steps:
  # Build the container image
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'us-central1-docker.pkg.dev/$PROJECT_ID/carflow-repo/carflow:$COMMIT_SHA', '.']
  
  # Push the container image to Artifact Registry
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'us-central1-docker.pkg.dev/$PROJECT_ID/carflow-repo/carflow:$COMMIT_SHA']
  
  # Deploy container image to Cloud Run
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'carflow-api'
      - '--image=us-central1-docker.pkg.dev/$PROJECT_ID/carflow-repo/carflow:$COMMIT_SHA'
      - '--region=us-central1'
      - '--platform=managed'
      - '--allow-unauthenticated'
      - '--memory=128Mi'
      - '--cpu=1'
      - '--max-instances=1'
      - '--min-instances=0'
      - '--timeout=30s'

images:
  - 'us-central1-docker.pkg.dev/$PROJECT_ID/carflow-repo/carflow:$COMMIT_SHA'
```

### 2. Manual Deployment with Cloud Build

```bash
# Trigger a manual build and deployment
gcloud builds submit --config=cloudbuild.yaml
```

### 3. Setting up Automated Deployment with GitHub

1. Connect your GitHub repository to Cloud Build
2. Configure triggers to build on commits to main branch
3. Use the cloudbuild.yaml file for build configuration

```bash
# Create a trigger
gcloud builds triggers create github \
  --name="carflow-deploy" \
  --repo="YOUR_GITHUB_USERNAME/golang-carflow-api" \
  --branch-pattern="main" \
  --build-config="cloudbuild.yaml"
```

## Monitoring Usage

To ensure you stay within the free tier limits, regularly monitor your usage:

1. Go to GCP Console > Billing > Reports
2. Filter by service (Cloud Run, Artifact Registry, etc.)
3. Set the time period to the current billing cycle

You can also view Cloud Run usage specifically:
1. Go to GCP Console > Cloud Run
2. Select your service
3. Click on "Metrics" tab to view usage patterns

## Estimated Free Tier Duration

With the free tier limits and efficient usage, you should be able to run the CarFlow API indefinitely without charges, assuming:

- Less than 2 million requests per month
- Container set to 128MB RAM
- Minimal storage usage in Artifact Registry
- Auto-scaling to zero instances when not in use (for maximum efficiency)

## Cleanup to Avoid Charges

If you decide to stop using the service, clean up your resources:

```bash
# Delete Cloud Run service
gcloud run services delete carflow-api --region=us-central1

# Delete Docker images
gcloud artifacts docker images delete \
  us-central1-docker.pkg.dev/carflow-api-project/carflow-repo/carflow --delete-tags

# Delete repository
gcloud artifacts repositories delete carflow-repo \
  --location=us-central1

# Delete storage bucket
gcloud storage rm gs://carflow-terraform-state --recursive
``` 