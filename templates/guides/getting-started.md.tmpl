---
page_title: "Getting started: from an empty cloud account to a running app - Qovery Provider"
subcategory: ""
description: |-
  Start a new infrastructure project with Terraform or OpenTofu: connect an AWS account, create a Kubernetes cluster, then run an app and a Postgres database in staging and production.
---

# Getting started: from an empty cloud account to a running app

This guide starts a new infrastructure project from nothing. At the end, one configuration holds:

- the AWS account Qovery manages, and an EKS cluster in it
- a project with a `staging` and a `production` environment
- a Postgres database and an application built from your git repository, in each environment
- the deployment of both environments

Terraform keeps describing your infrastructure. Qovery runs it on your own cloud account: it creates and upgrades the cluster, builds and deploys the services, and handles DNS, TLS, autoscaling and rollbacks. You do not write Kubernetes manifests, Helm charts or EKS modules.

The same configuration works with [OpenTofu](https://opentofu.org): the provider is published on both the [Terraform Registry](https://registry.terraform.io/providers/qovery/qovery/latest) and the [OpenTofu Registry](https://search.opentofu.org/provider/qovery/qovery/latest). Run `tofu` wherever this guide runs `terraform`.

## Before you start

- A Qovery organization. Create one for free at [console.qovery.com](https://console.qovery.com).
- An API token, exported as `QOVERY_API_TOKEN`. See [Authentication](https://registry.terraform.io/providers/qovery/qovery/latest/docs#authentication).
- An IAM role in your AWS account that Qovery can assume. The [AWS installation guide](https://www.qovery.com/docs/getting-started/installation/aws) creates it with a CloudFormation stack and gives you its ARN.
- A git repository with a `Dockerfile`, and its application listening on port 8080. To deploy a public image instead, use `qovery_container`.

## 1. Declare the provider

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

variable "organization_id" {
  type = string
}

variable "aws_role_arn" {
  type = string
}
```

## 2. Connect the cloud account and create the cluster

```terraform
resource "qovery_aws_credentials" "main" {
  organization_id = var.organization_id
  name            = "main-aws-account"
  role_arn        = var.aws_role_arn
}

resource "qovery_cluster" "main" {
  organization_id = var.organization_id
  credentials_id  = qovery_aws_credentials.main.id
  name            = "main"
  cloud_provider  = "AWS"
  region          = "us-east-2"

  features = {
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        # Allow many instance families and sizes: a short list makes node allocation fail more often.
        requirements = [
          {
            key      = "InstanceFamily"
            operator = "In"
            values   = ["c6i", "c7i", "m6i", "m7i", "r6i", "r7i", "t3", "t3a"]
          },
          {
            key      = "InstanceSize"
            operator = "In"
            values   = ["medium", "large", "xlarge", "2xlarge", "4xlarge"]
          },
          {
            key      = "Arch"
            operator = "In"
            values   = ["AMD64"]
          }
        ]
      }
    }
  }
}
```

Qovery creates the VPC, the EKS cluster and its add-ons in your account, then keeps them upgraded. For GCP, Azure or Scaleway, see the examples of [qovery_cluster](https://registry.terraform.io/providers/qovery/qovery/latest/docs/resources/cluster).

## 3. Create the project and its environments

```terraform
locals {
  environments = {
    staging    = "STAGING"
    production = "PRODUCTION"
  }
}

resource "qovery_project" "app" {
  organization_id = var.organization_id
  name            = "my-app"
}

resource "qovery_environment" "app" {
  for_each = local.environments

  project_id = qovery_project.app.id
  cluster_id = qovery_cluster.main.id
  name       = each.key
  mode       = each.value
}
```

Both environments share the cluster here. To isolate production, declare a second `qovery_cluster` and point the production environment at it.

## 4. Add the database and the application

```terraform
# A CONTAINER database runs on the cluster. For Amazon RDS, declare a qovery_blueprint
# with blueprint = "AWS/postgres/17" instead.
resource "qovery_database" "postgres" {
  for_each = local.environments

  environment_id = qovery_environment.app[each.key].id
  name           = "postgres"
  type           = "POSTGRESQL"
  version        = "17"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 20
}

resource "qovery_application" "api" {
  for_each = local.environments

  environment_id = qovery_environment.app[each.key].id
  name           = "api"

  git_repository = {
    # Use another branch for staging once you have one.
    url    = "https://github.com/my-org/my-app.git"
    branch = "main"
  }
  dockerfile_path = "Dockerfile"

  ports = [
    {
      internal_port       = 8080
      external_port       = 443
      publicly_accessible = true
    }
  ]

  # Qovery routes traffic to the application once the readiness probe passes.
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

  # DATABASE_URL is an alias of the database's built-in DATABASE_URL_INTERNAL secret.
  # Qovery resolves it at deploy time, so the password never enters the Terraform state.
  secret_aliases = [
    {
      key   = "DATABASE_URL"
      value = "QOVERY_POSTGRESQL_Z${upper(split("-", qovery_database.postgres[each.key].id)[0])}_DATABASE_URL_INTERNAL"
    }
  ]
}
```

## 5. Deploy

```terraform
resource "qovery_deployment" "app" {
  for_each = local.environments

  environment_id = qovery_environment.app[each.key].id
  desired_state  = "RUNNING"

  depends_on = [
    qovery_application.api,
    qovery_database.postgres,
  ]
}
```

At runtime, `DATABASE_URL` holds `postgresql://<login>:<password>@<host>:<port>/<database>`.

Then run:

```shell
terraform init    # or: tofu init
terraform apply   # or: tofu apply
```

The cluster takes the longest to create. Once it is ready, Qovery builds the application, deploys the database and the application in each environment, and gives the application a public URL with TLS. Follow the progress in the [Qovery Console](https://console.qovery.com).

~> **Note:** Destroying a `qovery_deployment` deletes its environment and all its services.

## Next steps

- Keep secrets out of the Terraform state: [Keeping secrets out of the state](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/keeping-secrets-out-of-the-state).
- See how the provider plans changes made in the Console: [Managing changes](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/managing-changes).
- Run Helm charts next to your apps: [qovery_helm](https://registry.terraform.io/providers/qovery/qovery/latest/docs/resources/helm). The [Airbyte on Qovery example](https://github.com/evoxmusic/qovery-airbyte) deploys a full Helm-based stack with a database and deployment stages.
- Provision managed cloud resources, such as RDS, ElastiCache or S3, from the service catalog: [qovery_blueprint](https://registry.terraform.io/providers/qovery/qovery/latest/docs/resources/blueprint).
- Already set up in the Console? [Export it as Terraform](https://www.qovery.com/docs/terraform-provider/exporter).
