# FullStackArkham Terraform Variables

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "gke_machine_type" {
  description = "GKE node machine type"
  type        = string
  default     = "n2-standard-4"
}

variable "gke_min_nodes" {
  description = "Minimum GKE nodes"
  type        = number
  default     = 3
}

variable "gke_max_nodes" {
  description = "Maximum GKE nodes"
  type        = number
  default     = 20
}

variable "postgres_tier" {
  description = "Cloud SQL tier"
  type        = string
  default     = "db-custom-4-15360"
}

variable "redis_tier" {
  description = "Redis tier (BASIC or STANDARD_HA)"
  type        = string
  default     = "BASIC"
}

variable "redis_memory_gb" {
  description = "Redis memory in GB"
  type        = number
  default     = 4
}
