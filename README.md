# Terraform Provider for Vezor

This Terraform provider allows you to access secrets stored in [Vezor](https://vezor.io), a GitOps-native secrets management platform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    vezor = {
      source  = "vezor/vezor"
      version = "~> 1.0"
    }
  }
}
```

### From Source

```bash
# Clone the repository
git clone https://github.com/vezor/terraform-provider-vezor.git
cd terraform-provider-vezor

# Build the provider
go build -o terraform-provider-vezor

# Install locally (Linux/macOS)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/vezor/vezor/1.0.0/$(go env GOOS)_$(go env GOARCH)
mv terraform-provider-vezor ~/.terraform.d/plugins/registry.terraform.io/vezor/vezor/1.0.0/$(go env GOOS)_$(go env GOARCH)/
```

## Authentication

The provider requires an API key to authenticate with Vezor. You can create an API key in the Vezor web UI under Settings > API Keys.

### Option 1: Environment Variable (Recommended)

```bash
export VEZOR_API_KEY="vz_xxxxxxxxxxxxxxxxxxxx"
```

### Option 2: Provider Configuration

```hcl
provider "vezor" {
  api_key = var.vezor_api_key
}
```

**Note:** Never commit API keys to version control. Use environment variables or Terraform variables with sensitive values.

## Configuration

```hcl
provider "vezor" {
  api_key = var.vezor_api_key  # Optional if VEZOR_API_KEY is set
  api_url = "https://api.vezor.io"  # Optional, defaults to production
}
```

### Configuration Reference

| Attribute | Description | Default | Environment Variable |
|-----------|-------------|---------|---------------------|
| `api_key` | API key for authentication | - | `VEZOR_API_KEY` |
| `api_url` | Vezor API URL | `https://api.vezor.io` | `VEZOR_API_URL` |

## Data Sources

### vezor_secret

Fetches a single secret by name and tags.

```hcl
data "vezor_secret" "database_url" {
  name = "DATABASE_URL"
  tags = {
    env = "production"
    app = "my-api"
  }
}

# Access the secret value
resource "some_resource" "example" {
  connection_string = data.vezor_secret.database_url.value
}
```

#### Argument Reference

| Attribute | Description | Required |
|-----------|-------------|----------|
| `name` | The name (key) of the secret | Yes |
| `tags` | Map of tags to filter the secret | Yes |

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | The unique identifier of the secret |
| `value` | The decrypted secret value (sensitive) |
| `description` | The secret description |
| `version` | The secret version number |

### vezor_group

Fetches all secrets from a group. Groups are saved tag queries that match multiple secrets.

```hcl
data "vezor_group" "prod_api" {
  name = "prod-api-secrets"
}

# Access all secrets as a map
resource "kubernetes_secret" "app" {
  metadata {
    name = "app-secrets"
  }
  data = data.vezor_group.prod_api.secrets
}
```

#### Argument Reference

| Attribute | Description | Required |
|-----------|-------------|----------|
| `name` | The name of the group | Yes |

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | The unique identifier of the group |
| `description` | The group description |
| `tags` | Map of tags that define the group's query |
| `secrets` | Map of secret names to values (sensitive) |
| `count` | Number of secrets in the group |

## Examples

### Kubernetes Secrets

```hcl
data "vezor_group" "app_secrets" {
  name = "my-app-production"
}

resource "kubernetes_secret" "app" {
  metadata {
    name      = "app-secrets"
    namespace = "production"
  }
  data = data.vezor_group.app_secrets.secrets
}
```

### AWS Secrets Manager

```hcl
data "vezor_secret" "db_password" {
  name = "DATABASE_PASSWORD"
  tags = {
    env = "production"
    app = "my-api"
  }
}

resource "aws_secretsmanager_secret_version" "db" {
  secret_id     = aws_secretsmanager_secret.db.id
  secret_string = data.vezor_secret.db_password.value
}
```

### Environment Variables for ECS/Docker

```hcl
data "vezor_group" "app_config" {
  name = "my-app-config"
}

resource "aws_ecs_task_definition" "app" {
  family = "my-app"

  container_definitions = jsonencode([{
    name  = "app"
    image = "my-app:latest"
    environment = [
      for key, value in data.vezor_group.app_config.secrets : {
        name  = key
        value = value
      }
    ]
  }])
}
```

## Development

### Building

```bash
go build -o terraform-provider-vezor
```

### Testing

```bash
go test ./...
```

### Generating Documentation

```bash
go generate ./...
```

## License

MPL-2.0
