# Google Cloud Platform (GCP) Infrastructure

This document outlines the GCP infrastructure setup for the CarFlow API project. All services are configured to use only GCP Free Tier resources.

## Project Structure

- **Project ID**: `carflow-api-project`
- **Project Name**: CarFlow API
- **Region**: us-central1

## Services Used

### 1. Cloud Run

Cloud Run is used to host the CarFlow API as a serverless container. This allows the application to automatically scale to zero when not in use, which helps stay within the free tier limits.

**Configuration**:
- **Service Name**: carflow-api
- **Region**: us-central1
- **Memory**: 128Mi (minimal to stay within free tier)
- **CPU**: 1
- **Max Instances**: 1 (prevents scaling beyond free tier)
- **Min Instances**: 0 (scales to zero when not in use)
- **Timeout**: 30s
- **Access**: Public (--allow-unauthenticated)

**Free Tier Limits**:
- 2 million requests per month
- 360,000 GB-seconds of memory
- 180,000 vCPU-seconds per month

### 2. Artifact Registry / Container Registry

Used to store Docker container images for the application.

**Configuration**:
- **Repository**: carflow-repo
- **Format**: Docker
- **Location**: us-central1

**Free Tier Limits**:
- 0.5GB storage (Artifact Registry)

### 3. Cloud Build

Used for building and deploying the application from source code.

**Free Tier Limits**:
- 120 build-minutes per day

### 4. IAM & Service Accounts

A dedicated service account is used for GitHub Actions to deploy the application.

**Service Account**:
- **Name**: github-actions-sa
- **Purpose**: Used by GitHub Actions for deployment

**Roles**:
- Cloud Run Admin (`roles/run.admin`) - Allows deploying to Cloud Run
- Storage Admin (`roles/storage.admin`) - Allows managing Docker images
- Service Account User (`roles/iam.serviceAccountUser`) - Allows using service accounts

## CI/CD Pipeline

### GitHub Actions Workflow

The GitHub Actions workflow automatically deploys the application to Cloud Run whenever changes are pushed to the `main` branch.

**Workflow File**: `.github/workflows/cloud-run-deploy.yml`

**Trigger**:
- Push to `main` branch
- Changes to `simple-app/**` or the workflow file

**Steps**:
1. Checkout code
2. Authenticate to GCP using service account key
3. Set up Google Cloud SDK
4. Deploy to Cloud Run
5. Get the service URL
6. Test the deployment
7. Display the deployment URL

## Budget Controls

To ensure we don't incur any charges beyond the free tier, we've implemented the following budget controls:

1. Budget alert set to $0
2. Notifications configured for 50%, 90%, and 100% of budget
3. Cloud Run configured to prevent excessive scaling

## Deployment Architecture

```
                     ┌─────────────────┐
                     │  GitHub Actions │
                     │     Workflow    │
                     └────────┬────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │   Cloud Build   │
                     │    Process      │
                     └────────┬────────┘
                              │
                     ┌────────┴────────┐
                     │                 │
            ┌────────▼─────┐   ┌───────▼───────┐
            │  Container   │   │   Cloud Run   │
            │  Registry    │   │   Service     │
            └──────────────┘   └───────────────┘
```

## Cleaning Up Resources

To avoid any potential charges, the following commands can be used to clean up resources when they're no longer needed:

```bash
# Delete Cloud Run service
gcloud run services delete carflow-api --region=us-central1

# Delete Docker images
gcloud artifacts docker images delete us-central1-docker.pkg.dev/carflow-api-project/carflow-repo/carflow --delete-tags

# Delete repository
gcloud artifacts repositories delete carflow-repo --location=us-central1

# Delete service account key (if any)
gcloud iam service-accounts keys list --iam-account=github-actions-sa@carflow-api-project.iam.gserviceaccount.com
# Then delete each key:
# gcloud iam service-accounts keys delete KEY_ID --iam-account=github-actions-sa@carflow-api-project.iam.gserviceaccount.com

# Delete service account
gcloud iam service-accounts delete github-actions-sa@carflow-api-project.iam.gserviceaccount.com
```

## Monitoring & Management

The following commands are useful for monitoring and managing the deployment:

```bash
# View deployed service info
gcloud run services describe carflow-api --region=us-central1

# View logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=carflow-api" --limit=10

# View billing
gcloud billing accounts list
# Then:
# gcloud billing accounts get-iam-policy ACCOUNT_ID
``` 