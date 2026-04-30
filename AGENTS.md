# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Codex, etc.) when working with code in this repository.

## Project Overview

This is a **Terraform Provider for Qovery** built with the `terraform-plugin-framework` (not the legacy SDK). The project follows **Domain-Driven Design (DDD)** principles with clean architecture patterns.

- **Terraform Registry**: https://registry.terraform.io/providers/qovery/qovery/latest

## Essential Commands

### Development

- **Build**: `task build` - Compiles provider binary to `bin/terraform-provider-qovery`
- **Install for local dev**: `task install` - Builds and creates dev override in `~/.terraformrc`
- **Remove dev override**: `task uninstall-dev-override`
- **Update API client**: `go get github.com/qovery/qovery-client-go && go mod tidy` - Update to latest Qovery API client (do this frequently!)

### Testing

- **Unit tests**: `task test` or `go test -tags=unit -v -cover ./...`
- **Integration tests**: `task testacc` or `TF_ACC=true go test -tags=integration -v -cover -timeout 2h ./...`
- **Run specific test**: `task testacc -- -run 'TestAcc_Organization*'`

### Code Quality

- **Lint**: `task lint` - Runs golangci-lint with `--fix` on `./internal/...`, `./qovery/...`, and `./client/...`
- **Generate mocks**: `task mocks` - Uses mockery to generate test mocks
- **Generate docs**: `task docs` - Generates Terraform provider documentation
- **Clean test resources**: `task clean-tests` - Cleans up leftover test resources from Qovery organization

### After Every Change

Always run these commands before committing:

```bash
go mod tidy
task docs
```

### Commit Message Format

- **One-liner only** - No multi-line commit messages
- **No author** - Don't add author information
- **Format**: `type(TICKET): description`
- **Ticket required** - If no `QOV-XXXX` ticket is provided, ask the user for it before proposing a commit message

Examples:

```
fix(QOV-1557): preserve descriptions from state for built-in environment variables
feat(QOV-1234): add support for GCP credentials resource
refactor(QOV-5678): simplify environment variable model conversion
```

### Environment Requirements

- Integration tests require `QOVERY_API_TOKEN` in `.env` file at repo root
- Use `.env.example` as template
- Acceptance tests create real Qovery resources and may incur costs

## Architecture Overview

### Layer Structure (DDD)

```
qovery/                  → Presentation: Terraform resources & data sources
internal/application/    → Application: Service orchestration
internal/domain/         → Domain: Business logic, entities, interfaces
internal/infrastructure/ → Infrastructure: Repository implementations, API clients
client/                  → Qovery API client and error handling
```

### Key Patterns

**Repository Pattern**:

- Interfaces defined in domain: `internal/domain/{entity}/{entity}_repository.go`
- Implementations in infrastructure: `internal/infrastructure/repositories/{entity}_qovery_repository.go`
- Mocks generated in: `internal/infrastructure/repositories/mocks_test/`

**Domain Entity Structure** (in `internal/domain/{entity}/`):

- `{entity}.go` - Domain entity with validation
- `{entity}_service.go` - Domain service interface and implementation
- `{entity}_repository.go` - Repository interface
- `{entity}_test.go` - Unit tests with `//go:build unit && !integration`

**Terraform Resource Structure** (in `qovery/`):

- `resource_{entity}.go` - Resource CRUD implementation
- `resource_{entity}_model.go` - Terraform model with type conversions
- `resource_{entity}_test.go` - Integration tests with `//go:build integration && !unit`
- `data_source_{entity}.go` - Data source implementation

### Dependencies & Key Libraries

- **Terraform Plugin**: `github.com/hashicorp/terraform-plugin-framework`
- **Qovery API Client**: `github.com/qovery/qovery-client-go`
- **UUID**: `github.com/google/uuid`
- **Validation**: `github.com/go-playground/validator/v10`
- **Error wrapping**: `github.com/pkg/errors`
- **Testing**: `github.com/stretchr/testify`

## Development Workflow

### Adding a New Resource

1. Define domain entity in `internal/domain/{entity}/`
2. Create repository interface in domain layer
3. Implement repository in `internal/infrastructure/repositories/`
4. Create application service in `internal/application/services/`
5. Implement Terraform resource in `qovery/resource_{entity}.go`
6. Create model in `qovery/resource_{entity}_model.go`
7. Write tests with proper build tags
8. Add examples in `examples/resources/qovery_{entity}/`

### Qovery Service Resource Patterns

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

**Checklist for New Service Resources:**

- [ ] Domain entity has `DeploymentStageID string` field
- [ ] `UpsertRepositoryRequest` has `DeploymentStageID string` field
- [ ] Repository Create/Update calls `AttachServiceToDeploymentStage()` if provided
- [ ] Repository Create/Update/Get calls `GetServiceDeploymentStage()` to retrieve
- [ ] Model conversion function accepts and uses `deploymentStageID` parameter
- [ ] Terraform schema has `deployment_stage_id` as Optional + Computed
- [ ] Terraform model struct has `DeploymentStageId types.String` field
- [ ] Run `task docs` to regenerate documentation
- [ ] Add acceptance tests for the new resource/attribute

## Testing Architecture

### Test Categories

| Category       | Build Tag     | Location    | Purpose                           | Command        |
| -------------- | ------------- | ----------- | --------------------------------- | -------------- |
| **Unit**       | `unit`        | `internal/` | Fast, isolated, mock-based        | `task test`    |
| **Acceptance** | `integration` | `qovery/`   | Terraform lifecycle with real API | `task testacc` |

**Rule**: Write unit tests first. Acceptance tests only for critical Terraform paths.

### Directory Structure

```
internal/
├── application/services/*_test.go   → Service orchestration tests
├── domain/{entity}/*_test.go        → Business logic tests
└── infrastructure/repositories/
    ├── mocks_test/                  → Generated mocks (mockery)
    └── qoveryapi/*_test.go          → API model conversion tests

qovery/
├── resource_{entity}_test.go        → Acceptance tests
└── data_source_{entity}_test.go     → Acceptance tests
```

### Commands

```bash
task test                                    # All unit tests (~10s)
task testacc                                 # All acceptance (~2h)
task testacc -- -run 'TestAcc_Container'     # Specific test
```

Detailed test patterns (build tags, table-driven scaffold, mock generation, coverage priority, naming) live in [`.claude/rules/testing.md`](.claude/rules/testing.md) and load automatically when editing test files.

## Code Standards

### Package Naming

Use singular nouns for packages (`project`, not `projects`). File-name patterns are listed under "Key Patterns" above.

### Validation

- Domain entities must implement `Validate() error`
- Use struct tags with validator: `validate:"required,min=1,max=255"`
- Validate at domain layer before repository operations

### Error Handling

- Domain errors defined in each domain package
- API errors handled through `client/apierrors/`
- Use `errors.Wrap()` to add context to errors
- Terraform diagnostics for user-facing errors

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party packages
    "github.com/hashicorp/terraform-plugin-framework/resource"

    // Internal - domain first, then infrastructure
    "github.com/qovery/terraform-provider-qovery/internal/domain/{entity}"
    "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories"
)
```
