package argoCdCredentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidArgoCdCredentials = errors.New("invalid argocd credentials")
	ErrInvalidClusterIDParam    = errors.New("invalid cluster id param")
	ErrInvalidUpsertRequest     = errors.New("invalid argocd credentials upsert request")
)

type ArgoCdCredentials struct {
	ID          uuid.UUID `validate:"required"`
	ClusterID   uuid.UUID `validate:"required"`
	ArgocdUrl   string    `validate:"required"`
	ArgocdToken string    `validate:"required"`
}

func (c ArgoCdCredentials) Validate() error {
	return validator.New().Struct(c)
}

type UpsertRequest struct {
	ArgocdUrl   string `validate:"required"`
	ArgocdToken string `validate:"required"`
}

func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}
	return nil
}
