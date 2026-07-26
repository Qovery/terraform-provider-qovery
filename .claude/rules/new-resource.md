---
paths:
  - "qovery/resource_*.go"
  - "qovery/data_source_*.go"
  - "internal/domain/**"
  - "internal/infrastructure/repositories/**"
  - "internal/application/services/**"
---

# Adding a New Resource

1. Define domain entity in `internal/domain/{entity}/`
2. Create repository interface in domain layer
3. Implement repository in `internal/infrastructure/repositories/`
4. Create application service in `internal/application/services/`
5. Implement Terraform resource in `qovery/resource_{entity}.go`
6. Create model in `qovery/resource_{entity}_model.go`
7. Mirror the schema in `data_source_{entity}.go` — the resource and data source **share the same model struct**, so every model field needs a matching data source attribute (usually `Computed: true`). A missing attribute is a **runtime** error (`mismatch between struct and object: Struct defines fields not found in object`), not a compile error, so it slips past `go build`.
8. Write tests with proper build tags
9. Add examples in `examples/resources/qovery_{entity}/`

## Qovery Service Resource Patterns

When creating or modifying a **service resource** (application, container, job, helm, terraform_service), ensure these common attributes are present:

**Required Service Attributes:**
| Attribute | Type | Notes |
|-----------|------|-------|
| `environment_id` | Required | With `RequiresReplace()` plan modifier |
| `deployment_stage_id` | Optional + Computed | Separate API call pattern (see below) |
| `name` | Required | Service name |
| `description` | Optional | Service description |
| `icon_uri` | Optional + Computed | Default icon URI |
| `auto_deploy` | Required/Optional | Auto-deploy on commit |
| `advanced_settings_json` | Optional + Computed | JSON advanced settings |

**Deployment Stage Pattern (IMPORTANT):**

The `deployment_stage_id` is **NOT** included in the service create/update API request. It requires separate API calls:

```go
// To SET deployment stage (in Create/Update):
if len(request.DeploymentStageID) > 0 {
    c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, serviceID).Execute()
}

// To GET deployment stage (in Create/Update/Get):
deploymentStage, _, _ := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, serviceID).Execute()
```

**Reference Implementation:** `internal/infrastructure/repositories/qoveryapi/{container,job,helm}_qoveryapi.go` — all three follow the same deployment stage pattern.

## Checklist for New Service Resources

- [ ] Domain entity has `DeploymentStageID string` field
- [ ] `UpsertRepositoryRequest` has `DeploymentStageID string` field
- [ ] Repository Create/Update calls `AttachServiceToDeploymentStage()` if provided
- [ ] Repository Create/Update/Get calls `GetServiceDeploymentStage()` to retrieve
- [ ] Model conversion function accepts and uses `deploymentStageID` parameter
- [ ] Terraform schema has `deployment_stage_id` as Optional + Computed
- [ ] Terraform model struct has `DeploymentStageId types.String` field
- [ ] Mirror every new attribute in `data_source_{entity}.go` schema (see step 7 above — fails at runtime, not `go build`)
- [ ] Update the matching `Terraform*Resource.kt` model in the q-core exporter (no automatic sync; audit helper-function schema attrs too, not just inline `schema.X`)
- [ ] Run `task docs` to regenerate documentation
- [ ] Add acceptance tests for the new resource/attribute
