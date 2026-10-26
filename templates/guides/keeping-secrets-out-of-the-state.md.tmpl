---
page_title: "Keeping secrets out of the state - Qovery Provider"
subcategory: ""
description: |-
  How to keep secret values out of the Terraform state file: write-only arguments, secrets managed outside Terraform, identity-based authentication and state protection.
---

# Keeping secrets out of the state

Terraform stores the attributes of every resource in the state file, in plain text. `sensitive` only hides a value in the plan and apply output. The Qovery API never returns a secret, so the provider keeps the configured value in the state to detect its changes.

The options below keep secret values out of the state.

## Write-only arguments

Terraform 1.11 and later support write-only arguments. Terraform sends a write-only argument to Qovery on create and update, and stores it neither in the plan nor in the state. Feed it from an ephemeral resource or an ephemeral input variable, so that no other resource stores the value either.

| Resource | Secret | Write-only form |
|---|---|---|
| `qovery_container_registry` | `config.password` | `config.password_wo` and `config.password_wo_version` |
| `qovery_helm_repository` | `config.password` | `config.password_wo` and `config.password_wo_version` |

```terraform
ephemeral "vault_kv_secret_v2" "docker_hub" {
  mount = "kv"
  name  = "qovery/docker-hub"
}

resource "qovery_container_registry" "docker_hub" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-docker-hub"
  kind            = "DOCKER_HUB"
  url             = "https://docker.io"

  config = {
    username            = "my-docker-hub-user"
    password_wo         = ephemeral.vault_kv_secret_v2.docker_hub.data["token"]
    password_wo_version = "1"
  }
}
```

Terraform cannot compare a value it never stores, so changing the value alone does not trigger an update. To send a new value, change the version, any string. Any other update of the resource also sends the current value.

A write-only argument conflicts with its plain form and requires its version. Terraform 1.10 and earlier reject a configuration that sets one.

### Migrating a resource

Replace the plain argument with the write-only argument and its version in the same change. Terraform plans an in-place update, and the state no longer holds the secret once the apply completes.

Earlier versions of the state file, which some backends keep, still hold the secret. Rotate it if those versions are not protected.

An imported resource records no version, so its first plan updates it and sends the write-only value.

## Secrets managed outside Terraform

The API never returns a secret value, so a secret set from the Qovery Console, the CLI or the API never reaches the state. To manage the secrets of a service there, leave them out of the configuration and ignore the attribute, so that Terraform does not plan their removal:

```terraform
resource "qovery_container" "api" {
  # ...

  lifecycle {
    ignore_changes = [secrets]
  }
}
```

This hands every secret of the service over to the Console. A value that the state already holds stays there.

## Identity-based authentication

Some resources can authenticate without any secret:

- `qovery_aws_credentials`, and `qovery_container_registry` and `qovery_helm_repository` of kind `ECR` and `OCI_ECR`: `role_arn`, an IAM role that Qovery assumes.
- `qovery_gcp_credentials`: Workload Identity Federation, with `service_account_email` and `workload_identity_provider_resource`.
- `qovery_container_registry` of kind `GCP_ARTIFACT_REGISTRY`: Workload Identity Federation, with `gcp_credentials_type = "workload_identity_federation"`, `project_id`, `service_account_email` and `workload_identity_provider_resource`.

## Protecting the state

Whatever the configuration, the state also holds identifiers and other values worth protecting. Store it in a backend that encrypts it at rest and restricts who can read it, such as HCP Terraform or an S3 bucket with server-side encryption and a restrictive bucket policy.
