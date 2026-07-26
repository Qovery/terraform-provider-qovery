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

## Development Workflow

### Adding a New Resource

The step-by-step guide, required service-resource attributes, the deployment-stage API pattern, and the completion checklist live in [`.claude/rules/new-resource.md`](.claude/rules/new-resource.md). **Read that file before adding or modifying any resource.** (Claude Code loads it automatically when editing resource/domain/repository files; other agents must open it explicitly.)

## Testing Architecture

### Test Categories

| Category       | Build Tag     | Location    | Purpose                           | Command        |
| -------------- | ------------- | ----------- | --------------------------------- | -------------- |
| **Unit**       | `unit`        | `internal/` | Fast, isolated, mock-based        | `task test`    |
| **Acceptance** | `integration` | `qovery/`   | Terraform lifecycle with real API | `task testacc` |

**Rule**: Write unit tests first. Whenever feasible, add an acceptance test alongside — especially for Terraform-visible behavior (plan/apply errors, replacement triggers, state consistency). Unit-only is fine for purely internal logic; acceptance-only is reserved for paths that can't be unit-tested.

Detailed test patterns (build tags, table-driven scaffold, mock generation, coverage priority, naming) live in [`.claude/rules/testing.md`](.claude/rules/testing.md) — read it when writing or editing tests. (Claude Code loads it automatically when editing test files; other agents must open it explicitly.)

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
