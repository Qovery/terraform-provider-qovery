# Qovery Terraform and OpenTofu Provider

[![Terraform Registry](https://img.shields.io/badge/Terraform-registry-7B42BC?logo=terraform)](https://registry.terraform.io/providers/qovery/qovery/latest)
[![OpenTofu Registry](https://img.shields.io/badge/OpenTofu-registry-FFDA18?logo=opentofu)](https://search.opentofu.org/provider/qovery/qovery/latest)
[![Tests](https://github.com/qovery/terraform-provider-qovery/actions/workflows/tests.yml/badge.svg)](https://github.com/qovery/terraform-provider-qovery/actions/workflows/tests.yml)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](LICENSE)

Run applications, databases, Helm charts and Terraform modules on Kubernetes in your own AWS, GCP, Azure or Scaleway account, described in Terraform or OpenTofu.

[Qovery](https://www.qovery.com) creates and operates the Kubernetes cluster (EKS, GKE, AKS, Kapsule) in your cloud account, then builds and deploys your services on it, with DNS, TLS, autoscaling, preview environments and rollbacks. This provider lets you declare all of it as code, so you do not have to write and maintain EKS modules, Helm charts and Kubernetes manifests yourself.

- **Documentation:** [registry.terraform.io/providers/qovery/qovery](https://registry.terraform.io/providers/qovery/qovery/latest/docs)
- **Start a new project:** [Getting started guide](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/getting-started), from an empty cloud account to an app and a database in staging and production
- **Changelog:** [CHANGELOG.md](CHANGELOG.md). Upgrading from 0.x? Read the [1.0 upgrade guide](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/upgrade-to-1.0).

## Quick start

```terraform
terraform {
  required_providers {
    qovery = {
      source  = "qovery/qovery"
      version = "~> 1.0"
    }
  }
}

# Reads the API token from the QOVERY_API_TOKEN environment variable.
provider "qovery" {}

# Both IDs are in the Console URL, or in `qovery organization list` and `qovery cluster list`.
variable "organization_id" {
  type = string
}

variable "cluster_id" {
  type = string
}

resource "qovery_project" "app" {
  organization_id = var.organization_id
  name            = "my-app"
}

resource "qovery_environment" "production" {
  project_id = qovery_project.app.id
  cluster_id = var.cluster_id
  name       = "production"
  mode       = "PRODUCTION"
}

resource "qovery_database" "postgres" {
  environment_id = qovery_environment.production.id
  name           = "postgres"
  type           = "POSTGRESQL"
  version        = "17"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 20
}

resource "qovery_application" "api" {
  environment_id  = qovery_environment.production.id
  name            = "api"
  dockerfile_path = "Dockerfile"

  git_repository = {
    url    = "https://github.com/my-org/my-app.git"
    branch = "main"
  }

  ports = [
    {
      internal_port       = 8080
      external_port       = 443
      publicly_accessible = true
    }
  ]

  healthchecks = {
    readiness_probe = {
      type = {
        tcp = {
          port = 8080
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  environment_variables = [
    {
      key   = "DATABASE_HOST"
      value = qovery_database.postgres.internal_host
    }
  ]
}

resource "qovery_deployment" "production" {
  environment_id = qovery_environment.production.id
  desired_state  = "RUNNING"
  depends_on     = [qovery_application.api, qovery_database.postgres]
}
```

```shell
export QOVERY_API_TOKEN="<your token>"   # Qovery Console > Settings > API Tokens
terraform init && terraform apply        # or: tofu init && tofu apply
```

This example uses an existing cluster. To create the cluster from Terraform too, with `qovery_aws_credentials` and `qovery_cluster`, follow the [Getting started guide](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/getting-started).

## Terraform and OpenTofu

The provider is published on both the [Terraform Registry](https://registry.terraform.io/providers/qovery/qovery/latest) and the [OpenTofu Registry](https://search.opentofu.org/provider/qovery/qovery/latest) under the same address, `qovery/qovery`. The configuration is the same for both: run `terraform` or `tofu`. The test suite runs against Terraform 1.15.

## What you can manage

| Area | Resources |
|------|-----------|
| Cloud accounts and clusters | `qovery_aws_credentials`, `qovery_gcp_credentials`, `qovery_scaleway_credentials`, `qovery_cluster` (EKS, GKE, AKS, Kapsule, EKS Anywhere) |
| Projects and environments | `qovery_project`, `qovery_environment`, `qovery_deployment_stage`, `qovery_deployment` |
| Services | `qovery_application` (built from git), `qovery_container` (from an image), `qovery_job` (cron and lifecycle), `qovery_helm`, `qovery_terraform_service`, `qovery_database` |
| Managed cloud resources | `qovery_blueprint` (RDS, ElastiCache, S3, MSK, CloudFront and more, from the [service catalog](https://github.com/Qovery/service-catalog)) |
| Organization and access | `qovery_organization`, `qovery_organization_member`, `qovery_custom_role`, `qovery_api_token` |

See the [full list of resources and data sources](https://registry.terraform.io/providers/qovery/qovery/latest/docs).

## Examples

- [Getting started](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/getting-started): AWS account, EKS cluster, Postgres and an application in staging and production.
- [Qovery AWS starter module](https://github.com/Qovery/terraform-qovery-aws-starter): the same setup as the getting-started guide in one `module` block, for Terraform and OpenTofu.
- [Airbyte on Qovery](https://github.com/evoxmusic/qovery-airbyte): a Helm-based stack with a database, deployment stages and a proxy.
- [Terraform provider documentation on qovery.com](https://www.qovery.com/docs/terraform-provider/overview): multi-environment setups, linking services, modules and remote state.
- Already configured in the Console? [Export your setup as Terraform](https://www.qovery.com/docs/terraform-provider/exporter), or ask your coding agent: the [qovery-terraform skill](https://github.com/Qovery/qovery-skills) writes the configuration, imports the resources and checks that the plan shows no change.

## Contributing

To build, test and document the provider, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Mozilla Public License 2.0](LICENSE)
