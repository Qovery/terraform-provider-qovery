# Implementation Plan: qovery_terraform_service Resource

**Status**: In Progress - 75% Complete (6/8 phases done)
**Last Updated**: 2025-11-19
**Last Commit**: 83da6ed - Phase 5 Data Source
**Reference Spec**: terraform-native.md

## ✅ Completed Phases

### ✅ Phase 1: Domain Layer (COMMITTED: 8482652, b53bc4a)

- ✅ Created domain entity (`terraformservice.go`) with validation
- ✅ Defined repository interface (`terraformservice_repository.go`)
- ✅ Defined service interface (`terraformservice_service.go`)
- ✅ Added comprehensive unit tests with validation scenarios
- ✅ Added domain constants (TERRAFORM service type, API resource)
- ✅ All tests passing

### ✅ Phase 2: Infrastructure Layer (COMMITTED: 50535f6)

- ✅ Implemented repository with Qovery API client integration
- ✅ Created API ↔ Domain model conversions
- ✅ Supports `TerraformsAPI` for Create/List operations
- ✅ Supports `TerraformMainCallsAPI` for Get/Edit/Delete
- ✅ Handles advanced settings via separate endpoint
- ✅ Properly converts oneOf backend types
- ✅ Compiles without errors

### ✅ Phase 3: Application Service Layer (COMMITTED: 870047b)

- ✅ Created `terraformservice_service.go` with simple CRUD + List pattern
- ✅ Wired up service in `services.go`
- ✅ Added repository field to `repositories.go`
- ✅ Initialized repository in `qoveryapi.go`
- ✅ UUID validation for IDs
- ✅ All files compile without errors
- ✅ Linter passes

### ✅ Phase 4: Terraform Resource (COMPLETE)

- ✅ Created `resource_terraform_service.go` with full schema (~400 lines)
  - Schema with all 20+ attributes (git_repository, backend, job_resources, etc.)
  - CRUD operations (Create, Read, Update, Delete)
  - Import support via ImportState
  - ModifyPlan for storage immutability check
- ✅ Created `resource_terraform_service_model.go` (~360 lines)
  - Terraform model structs with tfsdk tags
  - Terraform → Domain conversions (toUpsertServiceRequest)
  - Domain → Terraform conversions (convertDomainTerraformServiceToTerraformService)
  - Nested type conversions (git_repository, backend, job_resources, variables)
  - Variable Set ↔ Array conversions
  - ActionExtraArguments Map conversions
- ✅ Fixed all compilation errors:
  - Replaced external validators with local validators.Int64MinValidator
  - Fixed int32 → int64 type conversions
  - Added FromTime/FromTimePointer helpers to types_conversions.go
- ✅ Registered resource in provider.go:
  - Added terraformServiceService field
  - Initialized service in Configure()
  - Added newTerraformServiceResource to Resources()
- ✅ All code compiles successfully
- ✅ Linter passes with zero issues

### ✅ Phase 5: Terraform Data Source (COMMITTED: 83da6ed)

- ✅ Created `data_source_terraform_service.go` (~240 lines)
  - Schema with all fields as Computed (id is Required for lookup)
  - Read method implementation
  - Uses same TerraformService model as resource
  - Registered in provider DataSources()
- ✅ All code compiles successfully
- ✅ Linter passes with zero issues

### ✅ Phase 6: Wire up Provider (COMPLETE - done with Phase 4)

- ✅ Service registered in provider.go
- ✅ All layers properly connected

## 🔄 Remaining Phases

- ⏳ Phase 7: Create examples and generate documentation
- ⏳ Phase 8: Run all tests and validate

---

## Executive Summary

Implement `qovery_terraform_service` Terraform resource following DDD architecture patterns established in the codebase. This includes full domain layer, infrastructure layer, application layer, and presentation layer implementation with comprehensive testing.

### Scope Decisions

- ✅ Main `qovery_terraform_service` resource with **inline advanced settings**
- ✅ `data.qovery_terraform_service` data source
- ✅ Full test coverage: unit, integration, and mock tests
- ❌ Skip Deploy/Uninstall action operations (future work)

### Primary Reference Implementation

**Job Resource** (`resource_job.go`) - Most complete example with similar complexity

---

## Architecture Overview

### Layer Flow

```
Terraform Config
    ↓
qovery/resource_terraform_service.go (Presentation: Terraform CRUD)
    ↓
qovery/resource_terraform_service_model.go (Model: Terraform ↔ Domain conversion)
    ↓
internal/application/services/terraformservice_service.go (Application: Orchestration)
    ↓
internal/infrastructure/repositories/qoveryapi/terraformservice_qoveryapi.go (Infrastructure: API calls)
    ↓
internal/infrastructure/repositories/qoveryapi/terraformservice_qoveryapi_models.go (API ↔ Domain conversion)
    ↓
internal/domain/terraformservice/terraformservice.go (Domain: Business logic & validation)
    ↓
Qovery API Client (qovery-client-go)
```

---

## Phase 1: Domain Layer

**Location**: `internal/domain/terraformservice/`

### Files to Create

#### 1.1 `terraformservice.go`

**Purpose**: Core domain entity with business logic and validation

**Key Structures**:

```go
type TerraformService struct {
    ID                      uuid.UUID `validate:"required"`
    EnvironmentID           uuid.UUID `validate:"required"`
    Name                    string    `validate:"required"`
    Description             string    `validate:"required"`
    AutoDeploy              bool
    GitRepository           GitRepository
    TfVarFiles              []string
    Variables               []Variable
    Backend                 Backend
    Engine                  Engine
    EngineVersion           EngineVersion
    JobResources            JobResources
    TimeoutSec              *int32
    IconURI                 string
    UseClusterCredentials   bool
    ActionExtraArguments    map[string][]string
    AdvancedSettings        AdvancedSettings
    CreatedAt               time.Time
    UpdatedAt               *time.Time
}

type GitRepository struct {
    URL        string `validate:"required"`
    Branch     string
    RootPath   string // default: "/"
    GitTokenID *uuid.UUID
}

type Backend struct {
    Kubernetes    *KubernetesBackend
    UserProvided  *UserProvidedBackend
}

type EngineVersion struct {
    ExplicitVersion         string `validate:"required"`
    ReadFromTerraformBlock  bool
}

type JobResources struct {
    CPUMilli   int32 `validate:"required,min=10"`
    RAMMiB     int32 `validate:"required,min=1"`
    GPU        int32 `validate:"min=0"`
    StorageGiB int32 `validate:"required,min=1"`
}

type Variable struct {
    Key    string `validate:"required"`
    Value  string `validate:"required"`
    Secret bool
}

type AdvancedSettings struct {
    BuildTimeoutMaxSec                         *int32
    BuildCPUMaxInMilli                        *int32
    BuildRAMMaxInGiB                          *int32
    BuildEphemeralStorageInGiB                *int32
    DeploymentTerminationGracePeriodSeconds   *int32
    DeploymentAffinityNodeRequired            map[string]string
    SecurityServiceAccountName                string
    SecurityReadOnlyRootFilesystem            bool
}

type Engine string
const (
    EngineTerraform Engine = "TERRAFORM"
    EngineOpenTofu  Engine = "OPEN_TOFU"
)
```

**Constants to Define**:

```go
const (
    DefaultCPU         = 1000
    MinCPU             = 10
    DefaultRAM         = 1024
    MinRAM             = 1
    DefaultGPU         = 0
    MinGPU             = 0
    DefaultStorage     = 20
    MinStorage         = 1
    DefaultRootPath    = "/"
    DefaultIconURI     = "app://qovery-console/terraform"
)
```

**Methods to Implement**:

```go
func (t TerraformService) Validate() error
func (g GitRepository) Validate() error
func (b Backend) Validate() error
func (p EngineVersion) Validate() error
func (j JobResources) Validate() error
```

**Validation Logic**:

- Name must contain at least one ASCII letter
- Exactly ONE backend type (Kubernetes OR UserProvided)
- TfVar file paths must start with RootPath
- No directory traversal in paths (.., ~)
- Engine must be TERRAFORM or OPEN_TOFU
- ExplicitVersion is always required

**Error Constants**:

```go
var (
    ErrInvalidTerraformService             = errors.New("invalid terraform service")
    ErrInvalidTerraformServiceNameParam    = errors.New("invalid terraform service name")
    ErrInvalidGitRepositoryParam           = errors.New("invalid git repository")
    ErrInvalidBackendParam                 = errors.New("invalid backend configuration")
    ErrMissingBackendType                  = errors.New("exactly one backend type must be specified")
    ErrMultipleBackendTypes                = errors.New("cannot specify multiple backend types")
    ErrInvalidTfVarPath                    = errors.New("tfvar path must start with root_path")
    ErrInvalidEngineVersionParam           = errors.New("invalid engine version")
    ErrInvalidJobResourcesParam            = errors.New("invalid job resources")
)
```

#### 1.2 `terraformservice_repository.go`

**Purpose**: Repository interface defining data access contract

**Interface**:

```go
//go:generate mockery --testonly --with-expecter --name=Repository --structname=TerraformServiceRepository --filename=terraformservice_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

type Repository interface {
    Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*TerraformService, error)
    Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*TerraformService, error)
    Update(ctx context.Context, terraformServiceID string, request UpsertRepositoryRequest) (*TerraformService, error)
    Delete(ctx context.Context, terraformServiceID string) error
    List(ctx context.Context, environmentID string) ([]TerraformService, error)
}

type UpsertRepositoryRequest struct {
    Name                    string `validate:"required"`
    Description             string `validate:"required"`
    AutoDeploy              bool
    GitRepository           GitRepository
    TfVarFiles              []string
    Variables               []Variable
    Backend                 Backend
    Engine                  Engine
    EngineVersion           EngineVersion
    JobResources            JobResources
    TimeoutSec              *int32
    IconURI                 string
    UseClusterCredentials   bool
    ActionExtraArguments    map[string][]string
    AdvancedSettingsJson    string
}

func (r UpsertRepositoryRequest) Validate() error
```

#### 1.3 `terraformservice_service.go`

**Purpose**: Service interface for orchestration

**Interface**:

```go
type Service interface {
    Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*TerraformService, error)
    Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string) (*TerraformService, error)
    Update(ctx context.Context, terraformServiceID string, request UpsertServiceRequest) (*TerraformService, error)
    Delete(ctx context.Context, terraformServiceID string) error
    List(ctx context.Context, environmentID string) ([]TerraformService, error)
}

type UpsertServiceRequest struct {
    TerraformServiceUpsertRequest UpsertRepositoryRequest
    // Future: Add variable/secret diff requests if needed
}
```

#### 1.4 `terraformservice_test.go`

**Purpose**: Unit tests for domain logic

**Build Tag**: `//go:build unit && !integration`

**Test Cases**:

- Entity validation (all fields)
- Backend mutual exclusivity validation
- TfVar path validation
- Name validation (must have ASCII letter)
- Resource limits validation
- Nested struct validation

---

## Phase 2: Infrastructure Layer

**Location**: `internal/infrastructure/repositories/qoveryapi/`

### Files to Create

#### 2.1 `terraformservice_qoveryapi.go`

**Purpose**: Repository implementation with API client

**Structure**:

```go
type terraformServiceQoveryAPI struct {
    client *qovery.APIClient
}

func NewTerraformServiceQoveryAPI(client *qovery.APIClient) terraformservice.Repository {
    return &terraformServiceQoveryAPI{client: client}
}
```

**Methods to Implement**:

**Create**:

```go
func (c terraformServiceQoveryAPI) Create(ctx context.Context, environmentID string, request terraformservice.UpsertRepositoryRequest) (*terraformservice.TerraformService, error) {
    // 1. Validate request
    // 2. Convert to API request (toCreateTerraformRequest)
    // 3. Call: POST /environment/{environmentID}/terraform
    // 4. Handle advanced settings: PUT /terraform/{id}/advancedSettings
    // 5. Convert API response to domain entity
    // 6. Return
}
```

**Get**:

```go
func (c terraformServiceQoveryAPI) Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*terraformservice.TerraformService, error) {
    // 1. Call: GET /terraform/{terraformServiceID}
    // 2. Get advanced settings: GET /terraform/{id}/advancedSettings
    // 3. Convert API response to domain entity
    // 4. Return
}
```

**Update**:

```go
func (c terraformServiceQoveryAPI) Update(ctx context.Context, terraformServiceID string, request terraformservice.UpsertRepositoryRequest) (*terraformservice.TerraformService, error) {
    // 1. Validate request
    // 2. Convert to API request (toUpdateTerraformRequest)
    // 3. Call: PUT /terraform/{terraformServiceID}
    // 4. Update advanced settings: PUT /terraform/{id}/advancedSettings
    // 5. Refresh: GET /terraform/{terraformServiceID}
    // 6. Convert and return
}
```

**Delete**:

```go
func (c terraformServiceQoveryAPI) Delete(ctx context.Context, terraformServiceID string) error {
    // 1. Call: DELETE /terraform/{terraformServiceID}
    // 2. Handle errors
}
```

**List**:

```go
func (c terraformServiceQoveryAPI) List(ctx context.Context, environmentID string) ([]terraformservice.TerraformService, error) {
    // 1. Call: GET /environment/{environmentID}/terraform
    // 2. Convert all responses to domain entities
    // 3. Return list
}
```

**API Endpoints**:

- `EnvironmentActionsAPI.CreateTerraform(ctx, environmentID)`
- `TerraformActionsAPI.GetTerraform(ctx, terraformID)`
- `TerraformActionsAPI.EditTerraform(ctx, terraformID)`
- `TerraformActionsAPI.DeleteTerraform(ctx, terraformID)`
- `EnvironmentActionsAPI.ListEnvironmentTerraform(ctx, environmentID)`
- `TerraformAdvancedSettingsAPI.EditAdvancedSettings(ctx, terraformID)`
- `TerraformAdvancedSettingsAPI.GetAdvancedSettings(ctx, terraformID)`

#### 2.2 `terraformservice_qoveryapi_models.go`

**Purpose**: Conversion between API and domain models

**Key Functions**:

**Domain to API (Create/Update)**:

```go
func toCreateTerraformRequest(request terraformservice.UpsertRepositoryRequest) qovery.TerraformRequest {
    // Convert domain request to API request
    // Handle nested structure flattening:
    //   - Provider git_repository → API terraform_files_source.git_repository
    //   - Provider tfvar_files/variables → API terraform_variables_source
    // Return qovery.TerraformRequest
}

func toAdvancedSettingsRequest(jsonStr string) qovery.TerraformAdvancedSettings {
    // Parse JSON and convert to API advanced settings
}
```

**API to Domain (Read)**:

```go
func fromAPITerraformResponse(resp qovery.TerraformResponse, advancedSettings *qovery.TerraformAdvancedSettings) terraformservice.TerraformService {
    // Convert API response to domain entity
    // Flatten API nested structure to provider structure:
    //   - API terraform_files_source.git_repository → Provider git_repository
    //   - API terraform_variables_source → Provider tfvar_files/variables
    // Return terraformservice.TerraformService
}

func advancedSettingsToJson(settings *qovery.TerraformAdvancedSettings) string {
    // Convert advanced settings to JSON string
}
```

**Critical Transformations**:

1. **Git Repository Nesting**:

   ```
   Provider: git_repository { url, branch, root_path, git_token_id }
   API:      terraform_files_source.git_repository { url, branch, root_path, git_token_id }
   ```

2. **Variables Nesting**:

   ```
   Provider: tfvar_files[], variable[]
   API:      terraform_variables_source { tf_var_file_paths[], tf_vars[] }
   ```

3. **Backend Structure**:
   ```
   Provider: backend { kubernetes {} OR user_provided {} }
   API:      backend { kubernetes: {} OR user_provided: {} }
   ```

#### 2.3 `terraformservice_qoveryapi_test.go`

**Purpose**: Repository tests with mocked API client

**Build Tag**: `//go:build unit && !integration`

**Test Cases**:

- Create with successful API response
- Create with API error
- Get with successful response
- Update with successful response
- Delete with successful response
- API to domain conversions
- Domain to API conversions
- Advanced settings handling

---

## Phase 3: Application Layer

**Location**: `internal/application/services/`

### Files to Create

#### 3.1 `terraformservice_service.go`

**Purpose**: Orchestrate business logic across repositories

**Structure**:

```go
type terraformServiceService struct {
    terraformServiceRepository terraformservice.Repository
    // Future: add variable/secret services if needed
}

func NewTerraformServiceService(
    terraformServiceRepository terraformservice.Repository,
) (terraformservice.Service, error) {
    if terraformServiceRepository == nil {
        return nil, errors.New("terraformServiceRepository is required")
    }

    return &terraformServiceService{
        terraformServiceRepository: terraformServiceRepository,
    }, nil
}
```

**Methods**:

```go
func (s terraformServiceService) Create(ctx context.Context, environmentID string, request terraformservice.UpsertServiceRequest) (*terraformservice.TerraformService, error) {
    // 1. Validate request
    // 2. Call repository.Create()
    // 3. Handle any cross-service operations (future: variables, secrets)
    // 4. Return
}

func (s terraformServiceService) Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string) (*terraformservice.TerraformService, error) {
    // 1. Call repository.Get()
    // 2. Return
}

func (s terraformServiceService) Update(ctx context.Context, terraformServiceID string, request terraformservice.UpsertServiceRequest) (*terraformservice.TerraformService, error) {
    // 1. Validate request
    // 2. Call repository.Update()
    // 3. Handle any cross-service operations
    // 4. Return
}

func (s terraformServiceService) Delete(ctx context.Context, terraformServiceID string) error {
    // 1. Call repository.Delete()
    // 2. Return
}

func (s terraformServiceService) List(ctx context.Context, environmentID string) ([]terraformservice.TerraformService, error) {
    // 1. Call repository.List()
    // 2. Return
}
```

---

## Phase 4: Presentation Layer - Resource

**Location**: `qovery/`

### Files to Create

#### 4.1 `resource_terraform_service.go`

**Purpose**: Terraform resource implementation (~600-700 lines)

**Structure**:

```go
type terraformServiceResource struct {
    terraformServiceService terraformservice.Service
}

func newTerraformServiceResource() resource.Resource {
    return &terraformServiceResource{}
}

func (r terraformServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_terraform_service"
}

func (r *terraformServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Inject service from provider
    if req.ProviderData == nil {
        return
    }

    provider, ok := req.ProviderData.(*qProvider)
    if !ok {
        resp.Diagnostics.AddError(...)
        return
    }

    r.terraformServiceService = provider.terraformServiceService
}
```

**Schema Method**:

```go
func (r terraformServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Provides a Qovery Terraform service resource.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Id of the terraform service.",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "environment_id": schema.StringAttribute{
                Description: "Id of the environment.",
                Required:    true,
            },
            "name": schema.StringAttribute{
                Description: "Name of the terraform service.",
                Required:    true,
            },
            "description": schema.StringAttribute{
                Description: "Description of the terraform service.",
                Required:    true,
            },
            "auto_deploy": schema.BoolAttribute{
                Description: "Specify if the terraform service will be automatically updated on every new commit.",
                Required:    true,
            },
            // ... more attributes

            "git_repository": schema.SingleNestedAttribute{
                Description: "Terraform service git repository.",
                Required:    true,
                Attributes: map[string]schema.Attribute{
                    "url": schema.StringAttribute{
                        Description: "Git repository URL.",
                        Required:    true,
                    },
                    "branch": schema.StringAttribute{
                        Description: "Git branch.",
                        Optional:    true,
                    },
                    "root_path": schema.StringAttribute{
                        Description: "Git root path.",
                        Optional:    true,
                        Computed:    true,
                        Default:     stringdefault.StaticString("/"),
                    },
                    "git_token_id": schema.StringAttribute{
                        Description: "Git token ID for private repositories.",
                        Optional:    true,
                    },
                },
            },

            "tfvar_files": schema.ListAttribute{
                Description: "List of .tfvars file paths.",
                Required:    true,
                ElementType: types.StringType,
            },

            "variable": schema.SetNestedAttribute{
                Description: "Terraform variables.",
                Optional:    true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "key": schema.StringAttribute{
                            Description: "Variable key.",
                            Required:    true,
                        },
                        "value": schema.StringAttribute{
                            Description: "Variable value.",
                            Required:    true,
                            Sensitive:   true,
                        },
                        "secret": schema.BoolAttribute{
                            Description: "Is secret.",
                            Optional:    true,
                            Computed:    true,
                            Default:     booldefault.StaticBool(false),
                        },
                    },
                },
            },

            "backend": schema.SingleNestedAttribute{
                Description: "Terraform backend configuration.",
                Required:    true,
                Attributes: map[string]schema.Attribute{
                    "kubernetes": schema.SingleNestedAttribute{
                        Description: "Kubernetes backend configuration.",
                        Optional:    true,
                        Attributes:  map[string]schema.Attribute{}, // Empty
                        Validators: []validator.Object{
                            objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("user_provided")),
                        },
                    },
                    "user_provided": schema.SingleNestedAttribute{
                        Description: "User-provided backend configuration.",
                        Optional:    true,
                        Attributes:  map[string]schema.Attribute{}, // Empty
                        Validators: []validator.Object{
                            objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("kubernetes")),
                        },
                    },
                },
            },

            "engine": schema.StringAttribute{
                Description: "Terraform engine.",
                Required:    true,
                Validators: []validator.String{
                    stringvalidator.OneOf("TERRAFORM", "OPEN_TOFU"),
                },
            },

            "engine_version": schema.SingleNestedAttribute{
                Description: "Terraform provider version.",
                Required:    true,
                Attributes: map[string]schema.Attribute{
                    "explicit_version": schema.StringAttribute{
                        Description: "Explicit version.",
                        Required:    true,
                    },
                    "read_from_terraform_block": schema.BoolAttribute{
                        Description: "Read from terraform block.",
                        Optional:    true,
                        Computed:    true,
                        Default:     booldefault.StaticBool(false),
                    },
                },
            },

            "job_resources": schema.SingleNestedAttribute{
                Description: "Terraform job resources.",
                Required:    true,
                Attributes: map[string]schema.Attribute{
                    "cpu_milli": schema.Int64Attribute{
                        Description: "CPU in milli-cores.",
                        Required:    true,
                        Validators: []validator.Int64{
                            int64validator.AtLeast(10),
                        },
                    },
                    "ram_mib": schema.Int64Attribute{
                        Description: "RAM in MiB.",
                        Required:    true,
                        Validators: []validator.Int64{
                            int64validator.AtLeast(1),
                        },
                    },
                    "gpu": schema.Int64Attribute{
                        Description: "Number of GPUs.",
                        Optional:    true,
                        Computed:    true,
                        Default:     int64default.StaticInt64(0),
                        Validators: []validator.Int64{
                            int64validator.AtLeast(0),
                        },
                    },
                    "storage_gib": schema.Int64Attribute{
                        Description: "Storage in GiB.",
                        Required:    true,
                        Validators: []validator.Int64{
                            int64validator.AtLeast(1),
                        },
                    },
                },
            },

            "timeout_sec": schema.Int64Attribute{
                Description: "Timeout in seconds.",
                Optional:    true,
                Validators: []validator.Int64{
                    int64validator.AtLeast(0),
                },
            },

            "icon_uri": schema.StringAttribute{
                Description: "Icon URI.",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("app://qovery-console/terraform"),
            },

            "use_cluster_credentials": schema.BoolAttribute{
                Description: "Use cluster credentials.",
                Optional:    true,
                Computed:    true,
                Default:     booldefault.StaticBool(false),
            },

            "action_extra_arguments": schema.MapAttribute{
                Description: "Extra CLI arguments per action.",
                Optional:    true,
                ElementType: types.ListType{ElemType: types.StringType},
            },

            "advanced_settings_json": schema.StringAttribute{
                Description: "Advanced settings (JSON).",
                Optional:    true,
                Computed:    true,
            },

            "created_at": schema.StringAttribute{
                Description: "Created at.",
                Computed:    true,
            },

            "updated_at": schema.StringAttribute{
                Description: "Updated at.",
                Computed:    true,
            },
        },
    }
}
```

**CRUD Methods**:

```go
func (r terraformServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Get plan
    var plan TerraformService
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Convert to domain request
    request, err := plan.toUpsertServiceRequest(nil)
    if err != nil {
        resp.Diagnostics.AddError("Error creating terraform service", err.Error())
        return
    }

    // 3. Call service
    terraformService, err := r.terraformServiceService.Create(ctx, ToString(plan.EnvironmentID), *request)
    if err != nil {
        resp.Diagnostics.AddError("Error creating terraform service", err.Error())
        return
    }

    // 4. Convert to Terraform model
    state := convertDomainTerraformServiceToTerraformService(ctx, plan, terraformService)

    // 5. Set state
    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r terraformServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // 1. Get state
    var state TerraformService
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Detect import (EnvironmentID will be null)
    isImport := state.EnvironmentID.IsNull()

    // 3. Call service
    terraformService, err := r.terraformServiceService.Get(
        ctx,
        ToString(state.ID),
        ToString(state.AdvancedSettingsJson),
    )
    if err != nil {
        resp.Diagnostics.AddError("Error reading terraform service", err.Error())
        return
    }

    // 4. Convert to Terraform model
    state = convertDomainTerraformServiceToTerraformService(ctx, state, terraformService)

    // 5. Set state
    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r terraformServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Similar to Create
}

func (r terraformServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // 1. Get state
    var state TerraformService
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Call service
    err := r.terraformServiceService.Delete(ctx, ToString(state.ID))
    if err != nil {
        resp.Diagnostics.AddError("Error deleting terraform service", err.Error())
        return
    }
}

func (r terraformServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**ModifyPlan for Storage Immutability**:

```go
func (r terraformServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    // Prevent storage_gib reduction
    if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
        return
    }

    var plan, state TerraformService
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    if plan.JobResources != nil && state.JobResources != nil {
        planStorage := plan.JobResources.StorageGiB
        stateStorage := state.JobResources.StorageGiB

        if !planStorage.IsNull() && !stateStorage.IsNull() {
            if ToInt32(planStorage) < ToInt32(stateStorage) {
                resp.Diagnostics.AddError(
                    "Storage cannot be reduced",
                    fmt.Sprintf("Storage cannot be reduced from %d to %d GiB", ToInt32(stateStorage), ToInt32(planStorage)),
                )
            }
        }
    }
}
```

#### 4.2 `resource_terraform_service_model.go`

**Purpose**: Terraform model and type conversions (~400-500 lines)

**Main Model**:

```go
type TerraformService struct {
    ID                      types.String              `tfsdk:"id"`
    EnvironmentID           types.String              `tfsdk:"environment_id"`
    Name                    types.String              `tfsdk:"name"`
    Description             types.String              `tfsdk:"description"`
    AutoDeploy              types.Bool                `tfsdk:"auto_deploy"`
    GitRepository           *GitRepository            `tfsdk:"git_repository"`
    TfVarFiles              types.List                `tfsdk:"tfvar_files"`
    Variables               types.Set                 `tfsdk:"variable"`
    Backend                 *Backend                  `tfsdk:"backend"`
    Engine                  types.String              `tfsdk:"engine"`
    EngineVersion           *EngineVersion            `tfsdk:"engine_version"`
    JobResources            *JobResources             `tfsdk:"job_resources"`
    TimeoutSec              types.Int64               `tfsdk:"timeout_sec"`
    IconURI                 types.String              `tfsdk:"icon_uri"`
    UseClusterCredentials   types.Bool                `tfsdk:"use_cluster_credentials"`
    ActionExtraArguments    types.Map                 `tfsdk:"action_extra_arguments"`
    AdvancedSettingsJson    types.String              `tfsdk:"advanced_settings_json"`
    CreatedAt               types.String              `tfsdk:"created_at"`
    UpdatedAt               types.String              `tfsdk:"updated_at"`
}

type GitRepository struct {
    URL        types.String `tfsdk:"url"`
    Branch     types.String `tfsdk:"branch"`
    RootPath   types.String `tfsdk:"root_path"`
    GitTokenID types.String `tfsdk:"git_token_id"`
}

type Backend struct {
    Kubernetes   *KubernetesBackend   `tfsdk:"kubernetes"`
    UserProvided *UserProvidedBackend `tfsdk:"user_provided"`
}

type KubernetesBackend struct{}    // Empty
type UserProvidedBackend struct{}  // Empty

type EngineVersion struct {
    ExplicitVersion        types.String `tfsdk:"explicit_version"`
    ReadFromTerraformBlock types.Bool   `tfsdk:"read_from_terraform_block"`
}

type JobResources struct {
    CPUMilli   types.Int64 `tfsdk:"cpu_milli"`
    RAMMiB     types.Int64 `tfsdk:"ram_mib"`
    GPU        types.Int64 `tfsdk:"gpu"`
    StorageGiB types.Int64 `tfsdk:"storage_gib"`
}

type Variable struct {
    Key    types.String `tfsdk:"key"`
    Value  types.String `tfsdk:"value"`
    Secret types.Bool   `tfsdk:"secret"`
}
```

**Conversion Functions**:

```go
func (t TerraformService) toUpsertServiceRequest(state *TerraformService) (*terraformservice.UpsertServiceRequest, error) {
    // Convert Terraform model to domain request
    // Handle nested conversions
    // Return UpsertServiceRequest
}

func convertDomainTerraformServiceToTerraformService(ctx context.Context, state TerraformService, ts *terraformservice.TerraformService) TerraformService {
    // Convert domain entity to Terraform model
    // Preserve computed fields from state
    // Return TerraformService model
}

// Helper conversions for nested types
func toGitRepository(gr *GitRepository) terraformservice.GitRepository
func fromGitRepository(gr terraformservice.GitRepository) *GitRepository
func toBackend(b *Backend) terraformservice.Backend
func fromBackend(b terraformservice.Backend) *Backend
func toEngineVersion(pv *EngineVersion) terraformservice.EngineVersion
func fromEngineVersion(pv terraformservice.EngineVersion) *EngineVersion
func toJobResources(jr *JobResources) terraformservice.JobResources
func fromJobResources(jr terraformservice.JobResources) *JobResources
func toVariables(ctx context.Context, vars types.Set) []terraformservice.Variable
func fromVariables(ctx context.Context, vars []terraformservice.Variable) types.Set
```

#### 4.3 `resource_terraform_service_test.go`

**Purpose**: Integration tests

**Build Tag**: `//go:build integration && !unit`

**Test Structure**:

```go
func TestAcc_TerraformService(t *testing.T) {
    t.Parallel()

    testCases := []struct{
        name string
        config string
        checks []resource.TestCheckFunc
    }{
        {
            name: "create_with_kubernetes_backend",
            config: testAccQoveryTerraformServiceConfigKubernetes(),
            checks: []resource.TestCheckFunc{
                testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
                resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", "test-terraform"),
                resource.TestCheckResourceAttr("qovery_terraform_service.test", "backend.kubernetes", "true"),
            },
        },
        {
            name: "create_with_user_provided_backend",
            config: testAccQoveryTerraformServiceConfigUserProvided(),
            checks: []resource.TestCheckFunc{
                testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
                resource.TestCheckResourceAttr("qovery_terraform_service.test", "backend.user_provided", "true"),
            },
        },
        // More test cases
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            resource.Test(t, resource.TestCase{
                PreCheck:                 func() { testAccPreCheck(t) },
                ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
                CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
                Steps: []resource.TestStep{
                    {
                        Config: tc.config,
                        Check:  resource.ComposeAggregateTestCheckFunc(tc.checks...),
                    },
                },
            })
        })
    }
}

func TestAcc_TerraformService_Update(t *testing.T) {
    // Test update scenarios
}

func TestAcc_TerraformService_Import(t *testing.T) {
    // Test import functionality
}
```

---

## Phase 5: Presentation Layer - Data Source

**Location**: `qovery/`

### Files to Create

#### 5.1 `data_source_terraform_service.go`

**Purpose**: Data source for reading existing Terraform services

**Structure**:

```go
type terraformServiceDataSource struct {
    terraformServiceService terraformservice.Service
}

func newTerraformServiceDataSource() datasource.DataSource {
    return &terraformServiceDataSource{}
}

func (d terraformServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_terraform_service"
}

func (d *terraformServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    // Inject service from provider
}

func (d terraformServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    // Same schema as resource, but all fields Computed: true
}

func (d terraformServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    // Read by ID
}
```

#### 5.2 `data_source_terraform_service_test.go`

**Purpose**: Data source integration tests

**Build Tag**: `//go:build integration && !unit`

---

## Phase 6: Integration and Wiring

### 6.1 Update `internal/infrastructure/repositories/repositories.go`

```go
type Repositories struct {
    // ... existing repositories
    TerraformService terraformservice.Repository
}

// In NewRepositories():
terraformServiceRepository := qoveryapi.NewTerraformServiceQoveryAPI(client)
repositories.TerraformService = terraformServiceRepository
```

### 6.2 Update `internal/application/services/services.go`

```go
type Services struct {
    // ... existing services
    TerraformService terraformservice.Service
}

// In New():
terraformServiceService, err := NewTerraformServiceService(repos.TerraformService)
if err != nil {
    return nil, errors.Wrap(err, "cannot instantiate terraform service service")
}
services.TerraformService = terraformServiceService
```

### 6.3 Update `qovery/provider.go`

```go
type qProvider struct {
    // ... existing services
    terraformServiceService terraformservice.Service
}

// In Configure():
p.terraformServiceService = domainServices.TerraformService

// In Resources():
return []func() resource.Resource{
    // ... existing resources
    newTerraformServiceResource,
}

// In DataSources():
return []func() datasource.DataSource{
    // ... existing data sources
    newTerraformServiceDataSource,
}
```

---

## Phase 7: Testing Strategy

### 7.1 Unit Tests

**Location**: Domain layer (`internal/domain/terraformservice/`)

**Focus**:

- Entity validation logic
- Backend mutual exclusivity
- TfVar path validation
- Name validation rules
- Resource limits
- Error handling

**Run**: `task test`

### 7.2 Mock Tests

**Location**: Infrastructure layer (`internal/infrastructure/repositories/qoveryapi/`)

**Focus**:

- Repository method behavior with mocked API
- API request/response conversions
- Error scenarios
- Advanced settings handling

**Prerequisites**:

1. Generate mocks: `task mocks`
2. Run tests: `go test -tags=unit -v ./internal/infrastructure/repositories/qoveryapi/`

### 7.3 Integration Tests

**Location**: Presentation layer (`qovery/`)

**Focus**:

- Full resource lifecycle (create → read → update → delete)
- Import existing resources
- Different backend configurations
- Variables and secrets
- Advanced settings
- Validation errors

**Prerequisites**:

1. Set `QOVERY_API_TOKEN` in `.env`
2. Run: `task testacc -- -run 'TestAcc_TerraformService'`

**Test Matrix**:
| Test Case | Backend | Variables | Advanced Settings |
|-----------|---------|-----------|-------------------|
| Basic create | Kubernetes | None | None |
| With variables | Kubernetes | Yes | None |
| User backend | UserProvided | None | None |
| Full featured | Kubernetes | Yes | Yes |
| Update test | Kubernetes | Yes → Changed | None → Yes |
| Import test | Any | Any | Any |

---

## Phase 8: Documentation and Examples

### 8.1 Resource Examples

**Location**: `examples/resources/qovery_terraform_service/`

**Files**:

- `resource.tf` - Basic example with Kubernetes backend
- `resource_user_backend.tf` - Example with user-provided backend
- `resource_full.tf` - Full example with all options

### 8.2 Data Source Examples

**Location**: `examples/data-sources/qovery_terraform_service/`

**Files**:

- `data-source.tf` - Basic data source usage

### 8.3 Generate Documentation

```bash
task docs
```

---

## Implementation Checklist

### Domain Layer

- [ ] Create `terraformservice.go` with domain entity
- [ ] Create `terraformservice_repository.go` with interface
- [ ] Create `terraformservice_service.go` with service interface
- [ ] Add domain constants and defaults
- [ ] Add domain error definitions
- [ ] Implement `Validate()` methods
- [ ] Write unit tests
- [ ] Generate mocks: `task mocks`

### Infrastructure Layer

- [ ] Create `terraformservice_qoveryapi.go` with repository implementation
- [ ] Implement `Create()` method
- [ ] Implement `Get()` method
- [ ] Implement `Update()` method
- [ ] Implement `Delete()` method
- [ ] Implement `List()` method
- [ ] Create `terraformservice_qoveryapi_models.go` with conversions
- [ ] Implement API → Domain conversions
- [ ] Implement Domain → API conversions
- [ ] Handle advanced settings conversion
- [ ] Write mock tests

### Application Layer

- [ ] Create `terraformservice_service.go` with service implementation
- [ ] Implement all service methods
- [ ] Add service to `services.go`
- [ ] Wire up in service factory

### Presentation Layer - Resource

- [ ] Create `resource_terraform_service.go`
- [ ] Implement `Metadata()`, `Configure()` methods
- [ ] Implement comprehensive `Schema()` with all fields
- [ ] Add validators (ConflictsWith for backend)
- [ ] Add defaults (gpu, root_path, icon_uri)
- [ ] Add plan modifiers
- [ ] Implement `Create()` method
- [ ] Implement `Read()` method (handle import)
- [ ] Implement `Update()` method
- [ ] Implement `Delete()` method
- [ ] Implement `ImportState()` method
- [ ] Implement `ModifyPlan()` for storage immutability
- [ ] Create `resource_terraform_service_model.go`
- [ ] Define all Terraform model structs
- [ ] Implement `toUpsertServiceRequest()`
- [ ] Implement `convertDomainTerraformServiceToTerraformService()`
- [ ] Implement nested type conversions
- [ ] Write integration tests
- [ ] Test create scenarios
- [ ] Test update scenarios
- [ ] Test delete scenarios
- [ ] Test import scenarios
- [ ] Test validation errors

### Presentation Layer - Data Source

- [ ] Create `data_source_terraform_service.go`
- [ ] Implement data source methods
- [ ] Implement schema (all computed)
- [ ] Implement `Read()` method
- [ ] Write integration tests

### Integration

- [ ] Update `repositories.go` with new repository
- [ ] Update `services.go` with new service
- [ ] Update `provider.go` with resource and data source
- [ ] Test provider initialization

### Documentation

- [ ] Create resource examples
- [ ] Create data source examples
- [ ] Generate documentation: `task docs`
- [ ] Verify generated docs

### Final Validation

- [ ] Run all unit tests: `task test`
- [ ] Run all integration tests: `task testacc`
- [ ] Run linter: `task lint`
- [ ] Build provider: `task build`
- [ ] Test locally with `task install`
- [ ] Manual testing with example configurations

---

## Key Decisions and Trade-offs

### 1. Advanced Settings: Inline vs Separate Resource

**Decision**: Inline in main resource
**Rationale**: Simpler user experience, single resource to manage
**Trade-off**: Slightly more complex resource implementation

### 2. Backend Configuration: Booleans vs Blocks

**Decision**: Use nested blocks with ConflictsWith
**Rationale**: Better follows Terraform conventions, clearer intent
**Implementation**: Empty nested objects with mutual exclusivity validation

### 3. Variables: Set vs List

**Decision**: Use Set
**Rationale**: Prevents duplicate keys, better for key-value pairs
**Trade-off**: Slightly more complex ordering in state

### 4. Storage Immutability: Schema vs Runtime

**Decision**: Runtime check with ModifyPlan
**Rationale**: More flexible, better error messaging
**Implementation**: Compare state vs plan in ModifyPlan

### 5. Git Repository: Flattened vs Nested API Structure

**Decision**: Flatten for better UX
**Rationale**: Simpler Terraform configuration
**Trade-off**: Need conversion logic in repository layer

---

## Error Handling Strategy

### Domain Layer Errors

Define clear, actionable errors:

```go
var (
    ErrInvalidTerraformService          = errors.New("invalid terraform service")
    ErrInvalidTerraformServiceNameParam = errors.New("invalid terraform service name: must contain at least one ASCII letter")
    ErrMissingBackendType               = errors.New("exactly one backend type must be specified: kubernetes or user_provided")
    ErrInvalidTfVarPath                 = errors.New("tfvar path must start with root_path")
)
```

### Repository Layer Errors

Wrap API errors with context:

```go
if err != nil {
    return nil, errors.Wrap(err, "failed to create terraform service")
}
```

### Presentation Layer Errors

Convert to Terraform diagnostics:

```go
resp.Diagnostics.AddError(
    "Error creating terraform service",
    fmt.Sprintf("Could not create terraform service: %s", err.Error()),
)
```

---

## Performance Considerations

### 1. Advanced Settings

- Single API call after create/update
- Cache JSON in state to detect changes

### 2. List Operations

- Use for data source queries
- Consider pagination if API supports it

### 3. Import Operations

- Single GET call with advanced settings
- Minimize API calls during refresh

---

## Security Considerations

### 1. Sensitive Data

- Mark `variable.value` as `Sensitive: true`
- Store advanced settings JSON (may contain secrets)

### 2. Git Token

- Optional, only for private repos
- Store as UUID reference, not actual token

### 3. Backend Configuration

- User-provided backend may contain credentials
- Ensure proper state storage configuration

---

## API Endpoint Reference

### Main Resource Operations

| Operation | Method | Endpoint                                 | Request                                       | Response                    |
| --------- | ------ | ---------------------------------------- | --------------------------------------------- | --------------------------- |
| Create    | POST   | `/environment/{environmentId}/terraform` | TerraformRequest                              | TerraformResponse (201)     |
| Read      | GET    | `/terraform/{terraformId}`               | -                                             | TerraformResponse (200)     |
| Update    | PUT    | `/terraform/{terraformId}`               | TerraformRequest                              | TerraformResponse (200)     |
| Delete    | DELETE | `/terraform/{terraformId}`               | Query: resources_only, force_terraform_action | 204 No Content              |
| List      | GET    | `/environment/{environmentId}/terraform` | -                                             | TerraformResponseList (200) |

### Advanced Settings

| Operation | Method | Endpoint                                    | Request                   | Response                        |
| --------- | ------ | ------------------------------------------- | ------------------------- | ------------------------------- |
| Get       | GET    | `/terraform/{terraformId}/advancedSettings` | -                         | TerraformAdvancedSettings (200) |
| Update    | PUT    | `/terraform/{terraformId}/advancedSettings` | TerraformAdvancedSettings | TerraformAdvancedSettings (200) |

---

## Reference Files

### Templates to Follow

1. **Domain Entity**: `internal/domain/job/job.go`
2. **Repository Interface**: `internal/domain/job/job_repository.go`
3. **Service Interface**: `internal/domain/job/job_service.go`
4. **Repository Impl**: `internal/infrastructure/repositories/qoveryapi/job_qoveryapi.go`
5. **API Models**: `internal/infrastructure/repositories/qoveryapi/job_qoveryapi_models.go`
6. **Application Service**: `internal/application/services/job_service.go`
7. **Terraform Resource**: `qovery/resource_job.go`
8. **Terraform Model**: `qovery/resource_job_model.go`
9. **Integration Tests**: `qovery/resource_job_test.go`

### Helper Files

- `qovery/types_conversions.go` - Type conversion utilities
- `qovery/descriptions/descriptions.go` - Schema description helpers
- `qovery/validators/validators.go` - Custom validators

---

## Troubleshooting Guide

### Common Issues

**Issue**: Backend validation not working
**Solution**: Ensure ConflictsWith paths are correct: `path.MatchRelative().AtParent().AtName("user_provided")`

**Issue**: Storage reduction not blocked
**Solution**: Implement ModifyPlan method on resource

**Issue**: Import fails with null environment_id
**Solution**: Check `EnvironmentID.IsNull()` in Read method, fetch from API if import

**Issue**: Advanced settings not persisted
**Solution**: Ensure separate API call after create/update, store JSON in state

**Issue**: Variables not updating
**Solution**: Use Set type with proper hash function for variable blocks

**Issue**: API 422 Unprocessable Entity
**Solution**: Check domain validation, ensure all required fields present

---

## Next Steps After Implementation

1. **Manual Testing**: Test with real Terraform configurations
2. **Edge Case Testing**: Test error scenarios, validation failures
3. **Documentation Review**: Ensure all fields documented clearly
4. **Performance Testing**: Test with large variable sets
5. **Integration Testing**: Test with other Qovery resources (environment, cluster)
6. **User Acceptance**: Get feedback on schema design

---

## Future Enhancements

### Phase 10 (Future Work)

- [ ] Implement Deploy action resource/data source
- [ ] Implement Uninstall action resource/data source
- [ ] Add deployment status tracking
- [ ] Add state locking information
- [ ] Implement force unlock action
- [ ] Add migration state action
- [ ] Add deployment history tracking

---

## Notes and Observations

### API Quirks

1. **Nested Structure**: API uses `terraform_files_source.git_repository` but we flatten to `git_repository`
2. **Backend Blocks**: API accepts empty objects `{}` for backend types
3. **Advanced Settings**: Separate endpoint, requires additional API call
4. **Service Type**: API returns `service_type: "TERRAFORM"` in response

### Terraform SDK Patterns

1. **Computed + Optional**: Use when field has default but can be overridden
2. **UseStateForUnknown**: Use for ID and computed fields during creation
3. **ConflictsWith**: Use for mutually exclusive blocks
4. **Sensitive**: Use for secret values to hide in plan output

### DDD Patterns

1. **Dependencies flow inward**: Presentation → Application → Domain ← Infrastructure
2. **Domain is pure**: No external dependencies in domain layer
3. **Interfaces in domain**: Repository and Service interfaces defined in domain
4. **Validation in domain**: Business rules enforced at entity level

---

## Success Criteria

Implementation is complete when:

- ✅ All files created and properly structured
- ✅ All unit tests pass (`task test`)
- ✅ All integration tests pass (`task testacc`)
- ✅ Linter passes (`task lint`)
- ✅ Provider builds successfully (`task build`)
- ✅ Resource can be created, read, updated, deleted
- ✅ Data source can read existing resources
- ✅ Import functionality works
- ✅ Advanced settings are persisted
- ✅ Backend mutual exclusivity enforced
- ✅ Storage reduction blocked
- ✅ Documentation generated and accurate
- ✅ Examples work correctly

---

**END OF PLAN**
