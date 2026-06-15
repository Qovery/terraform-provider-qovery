package variable

import "github.com/google/uuid"

type ExternalSecretFile struct {
	ID                    uuid.UUID
	Key                   string
	Description           string
	MountPath             string
	Reference             string
	SecretManagerAccessId string
}

type ExternalSecretFiles []ExternalSecretFile

type ExternalSecretFileUpsertRequest struct {
	Key                   string `validate:"required"`
	Description           string
	MountPath             string `validate:"required"`
	Reference             string `validate:"required"`
	SecretManagerAccessId string `validate:"required"`
}

type ExternalSecretFileDiffRequest struct {
	Create []ExternalSecretFileDiffCreateRequest
	Update []ExternalSecretFileDiffUpdateRequest
	Delete []ExternalSecretFileDiffDeleteRequest
}

type ExternalSecretFileDiffCreateRequest struct {
	ExternalSecretFileUpsertRequest
}

type ExternalSecretFileDiffUpdateRequest struct {
	ExternalSecretFileUpsertRequest
	VariableID string
}

type ExternalSecretFileDiffDeleteRequest struct {
	VariableID string
}

func (r ExternalSecretFileDiffRequest) IsEmpty() bool {
	return len(r.Create) == 0 && len(r.Update) == 0 && len(r.Delete) == 0
}
