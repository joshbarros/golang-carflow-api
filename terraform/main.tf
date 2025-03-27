terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
  backend "gcs" {
    bucket = "carflow-terraform-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Enable required APIs
resource "google_project_service" "cloud_run_api" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "artifact_registry_api" {
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}

# Create Artifact Registry Repository for Docker images
resource "google_artifact_registry_repository" "carflow_repo" {
  provider      = google
  location      = var.region
  repository_id = "carflow-repo"
  format        = "DOCKER"
  depends_on    = [google_project_service.artifact_registry_api]
}

# Cloud Run service
resource "google_cloud_run_service" "carflow" {
  name     = "carflow-api"
  location = var.region

  template {
    spec {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.carflow_repo.repository_id}/carflow:latest"
        
        resources {
          limits = {
            cpu    = "1000m"
            memory = "256Mi"
          }
        }
        
        ports {
          container_port = 8080
        }
        
        env {
          name  = "PORT"
          value = "8080"
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [google_project_service.cloud_run_api]
}

# Make the Cloud Run service publicly accessible
resource "google_cloud_run_service_iam_member" "public_access" {
  service  = google_cloud_run_service.carflow.name
  location = google_cloud_run_service.carflow.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Output the service URL
output "service_url" {
  value = google_cloud_run_service.carflow.status[0].url
} 