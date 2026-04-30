---
paths:
  - "**/*_test.go"
  - "**/mocks_test/**"
---

# Testing Patterns

Detailed conventions for writing tests in this repo. CLAUDE.md covers the high-level layout (categories, directory structure, commands).

## Build Tags (required in every test file)

```go
// Unit test
//go:build unit && !integration
// +build unit,!integration

// Acceptance test
//go:build integration && !unit
// +build integration,!unit
```

## Unit Test Pattern (Table-Driven)

```go
func TestService_Method(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        TestName      string
        Input         Request
        SetupMock     func(*mocks_test.Repository)
        ExpectError   bool
        ErrorContains string
    }{
        {
            TestName:      "error_invalid_id",
            Input:         Request{ID: "invalid"},
            SetupMock:     func(m *mocks_test.Repository) {},
            ExpectError:   true,
            ErrorContains: "invalid",
        },
        {
            TestName: "error_repository_failure",
            SetupMock: func(m *mocks_test.Repository) {
                m.EXPECT().Create(mock.Anything, mock.Anything).
                    Return(nil, errors.New("db error"))
            },
            ExpectError:   true,
            ErrorContains: "failed to create",
        },
        {
            TestName: "success",
            SetupMock: func(m *mocks_test.Repository) {
                m.EXPECT().Create(mock.Anything, mock.Anything).
                    Return(expectedResult, nil)
            },
            ExpectError: false,
        },
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.TestName, func(t *testing.T) {
            t.Parallel()
            mockRepo := &mocks_test.Repository{}
            tc.SetupMock(mockRepo)
            // ... test logic
            mockRepo.AssertExpectations(t)
        })
    }
}
```

## Mock Generation

```go
// Add to repository interface file
//go:generate mockery --name=Repository --output=../infrastructure/repositories/mocks_test --outpkg=mocks_test --filename={entity}_repository_mock.go
```

```bash
task mocks  # Regenerate all mocks
```

## Test Coverage Priority

| Priority | Layer                | What to Test                  | Skip                                |
| -------- | -------------------- | ----------------------------- | ----------------------------------- |
| **P1**   | Application Services | ID validation, error wrapping | Success paths needing concrete deps |
| **P1**   | Domain Entities      | `Validate()`, business rules  | Simple getters                      |
| **P2**   | Infrastructure       | API model conversions         | Actual API calls                    |
| **P3**   | Terraform            | Use acceptance tests          | Unit tests redundant                |

## When to Write Tests

| Change        | Unit             | Acceptance           |
| ------------- | ---------------- | -------------------- |
| New resource  | Domain + Service | Required             |
| New entity    | Required         | -                    |
| Bug fix       | Regression test  | If Terraform-visible |
| New attribute | If validation    | Required             |

## Naming Convention

```
TestNew{Service}              # Constructor
Test{Service}_{Method}        # Method with subtests:
  └─ error_invalid_{field}    # Validation error
  └─ error_repository_failure # Dependency error
  └─ success                  # Happy path
```

## Known Limitations

1. **Concrete dependencies**: `DeploymentRestrictionService` is not an interface. Use acceptance tests for those success paths.
2. **Delete with wait**: `wait()` needs real 404 API responses. Skip unit tests for delete success.
3. **Complex services**: Container/Job/Helm error paths unit-testable, success paths need acceptance tests.

## Efficiency Rules

1. **Always `t.Parallel()`** in test functions and subtests
2. **Test validation first** before repository calls
3. **One acceptance test per resource** covering full lifecycle
4. **Mock at boundaries** (repositories), not internal functions
5. **Share fixtures** in `testdata/` directories
