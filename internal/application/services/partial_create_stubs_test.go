//go:build unit

package services_test

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// stubExternalSecretRepository / stubExternalSecretFileRepository replace the external
// secret repositories in the partial-create tests. There is no generated mock for those
// two interfaces, and these tests only need to choose between "succeeds" and "fails".
type stubExternalSecretRepository struct {
	listErr error
}

func (stubExternalSecretRepository) Create(_ context.Context, _ string, _ variable.ExternalSecretUpsertRequest) (*variable.ExternalSecret, error) {
	return nil, nil
}

func (stubExternalSecretRepository) Update(_ context.Context, _ string, _ variable.ExternalSecretUpsertRequest) (*variable.ExternalSecret, error) {
	return nil, nil
}

func (stubExternalSecretRepository) Delete(_ context.Context, _ string) error { return nil }

func (s stubExternalSecretRepository) List(_ context.Context, _ string) (variable.ExternalSecrets, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return variable.ExternalSecrets{}, nil
}

type stubExternalSecretFileRepository struct {
	listErr error
}

func (stubExternalSecretFileRepository) Create(_ context.Context, _ string, _ variable.ExternalSecretFileUpsertRequest) (*variable.ExternalSecretFile, error) {
	return nil, nil
}

func (stubExternalSecretFileRepository) Update(_ context.Context, _ string, _ variable.ExternalSecretFileUpsertRequest) (*variable.ExternalSecretFile, error) {
	return nil, nil
}

func (stubExternalSecretFileRepository) Delete(_ context.Context, _ string) error { return nil }

func (s stubExternalSecretFileRepository) List(_ context.Context, _ string) (variable.ExternalSecretFiles, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return variable.ExternalSecretFiles{}, nil
}
