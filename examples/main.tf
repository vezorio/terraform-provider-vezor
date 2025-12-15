terraform {
  required_providers {
    vezor = {
      source = "vezor/vezor"
    }
  }
}

# Configure the Vezor provider
# API key can be set here or via VEZOR_API_KEY environment variable
provider "vezor" {
  # api_key = var.vezor_api_key  # Uncomment to use variable
  # api_url = "https://api.vezor.io"  # Optional, defaults to production
}

# ============================================================================
# Example 1: Fetch a single secret
# ============================================================================

data "vezor_secret" "database_url" {
  name = "DATABASE_URL"
  tags = {
    env = "production"
    app = "my-api"
  }
}

output "database_url_version" {
  value       = data.vezor_secret.database_url.version
  description = "The version of the DATABASE_URL secret"
}

# Use the secret value (marked sensitive)
# output "database_url_value" {
#   value     = data.vezor_secret.database_url.value
#   sensitive = true
# }

# ============================================================================
# Example 2: Fetch all secrets from a group
# ============================================================================

data "vezor_group" "prod_api" {
  name = "prod-api-secrets"
}

output "prod_api_secret_count" {
  value       = data.vezor_group.prod_api.count
  description = "Number of secrets in the prod-api-secrets group"
}

output "prod_api_tags" {
  value       = data.vezor_group.prod_api.tags
  description = "Tags that define the prod-api-secrets group"
}

# ============================================================================
# Example 3: Use with Kubernetes
# ============================================================================

# resource "kubernetes_secret" "app_secrets" {
#   metadata {
#     name      = "app-secrets"
#     namespace = "production"
#   }
#
#   # Use all secrets from the group directly
#   data = data.vezor_group.prod_api.secrets
# }

# ============================================================================
# Example 4: Use with AWS Secrets Manager
# ============================================================================

# resource "aws_secretsmanager_secret" "db_credentials" {
#   name = "production/database"
# }
#
# resource "aws_secretsmanager_secret_version" "db_credentials" {
#   secret_id = aws_secretsmanager_secret.db_credentials.id
#   secret_string = jsonencode({
#     url      = data.vezor_secret.database_url.value
#     password = data.vezor_secret.database_password.value
#   })
# }

# ============================================================================
# Example 5: Use with Docker/ECS
# ============================================================================

# resource "aws_ecs_task_definition" "app" {
#   family = "my-app"
#
#   container_definitions = jsonencode([{
#     name  = "app"
#     image = "my-app:latest"
#     environment = [
#       for key, value in data.vezor_group.prod_api.secrets : {
#         name  = key
#         value = value
#       }
#     ]
#   }])
# }
