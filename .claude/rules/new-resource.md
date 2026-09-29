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

A **service resource** (application, container, job, helm, terraform_service) carries the common attributes below. The table records each one's service contract; every other attribute follows "HCL Config Is the Source of Truth" in `AGENTS.md`.

| Attribute | Schema | Notes |
|-----------|--------|-------|
| `environment_id` | Required | `RequiresReplace()` plan modifier |
| `name` | Required | |
| `description` | Optional | |
| `icon_uri` | Optional + Computed, `Default` = the type's q-core icon (`app://qovery-console/<type>`) | Job: the `JobIconUriDefault()` plan modifier picks the cron or lifecycle icon |
| `auto_deploy` | Optional + Computed, `Default` `true` (helm `false`) | Required on terraform_service |
| `auto_preview` | Optional + Computed, `Default` `false` | Not on terraform_service |
| `ephemeral_storage` | Optional + Computed, `Default` `0` | The read maps an API null to `0` (`ephemeralStorageFromAPI`). Not on helm and terraform_service |
| `deployment_stage_id` | Optional + Computed, `UseStateForUnknown` | Exception, commented in code: q-core cannot detach a service from its stage. Append `deploymentStageIDRemovalNote` to the description. Separate API calls (see below) |
| `advanced_settings_json` | Optional + Computed, `UseStateForUnknown` | Exception, commented in code: the QOV-2028 contract (removing a key keeps its value) until QOV-2315 exposes ownership |

The default values are constants in `qovery/service_schema_defaults.go`.

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

- [ ] Every attribute follows "HCL Config Is the Source of Truth" in `AGENTS.md`, apart from the exceptions in the table above
- [ ] Each exception is commented at the spot and explained in the attribute description
- [ ] Descriptions set `MarkdownDescription` only, in one or two sentences: shared wording comes from `qovery/schema_descriptions.go`, defaults and limits from the `qovery/descriptions` helpers, and the data source reuses the resource's base sentence
- [ ] Domain entity has `DeploymentStageID string` field
- [ ] `UpsertRepositoryRequest` has `DeploymentStageID string` field
- [ ] Repository Create/Update calls `AttachServiceToDeploymentStage()` if provided
- [ ] Repository Create/Update/Get calls `GetServiceDeploymentStage()` to retrieve
- [ ] Model conversion function accepts and uses `deploymentStageID` parameter
- [ ] Terraform model struct has `DeploymentStageId types.String` field
- [ ] Mirror every new attribute in `data_source_{entity}.go` schema (see step 7 above — fails at runtime, not `go build`)
- [ ] Update the matching `Terraform*Resource.kt` model in the q-core exporter (no automatic sync; audit helper-function schema attrs too, not just inline `schema.X`)
- [ ] Add the resource to the table in `templates/index.md.tmpl` (`docs/index.md` is generated from it)
- [ ] Run `task docs` to regenerate documentation
- [ ] Add an entry under `## [Unreleased]` in `CHANGELOG.md`
- [ ] Add acceptance tests for the new resource/attribute, with steps showing that removing the attribute from the config plans a difference, a change made outside Terraform shows on refresh, and import records the remote value
