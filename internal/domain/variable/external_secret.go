package variable

import (
	"github.com/google/uuid"
)

type ExternalSecret struct {
	ID                    uuid.UUID
	Key                   string
	Description           string
	Reference             string
	SecretManagerAccessId string
	Scope                 Scope
	VariableType          string
}

type ExternalSecrets []ExternalSecret

type ExternalSecretUpsertRequest struct {
	Key                   string `validate:"required"`
	Description           string
	Reference             string `validate:"required"`
	SecretManagerAccessId string `validate:"required"`
}

type ExternalSecretDiffRequest struct {
	Create []ExternalSecretDiffCreateRequest
	Update []ExternalSecretDiffUpdateRequest
	Delete []ExternalSecretDiffDeleteRequest
}

type ExternalSecretDiffCreateRequest struct {
	ExternalSecretUpsertRequest
}

type ExternalSecretDiffUpdateRequest struct {
	ExternalSecretUpsertRequest
	VariableID string
}

type ExternalSecretDiffDeleteRequest struct {
	VariableID string
}

func (r ExternalSecretDiffRequest) IsEmpty() bool {
	return len(r.Create) == 0 && len(r.Update) == 0 && len(r.Delete) == 0
}
