package argoCdDestinationClusterMapping

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidArgoCdDestinationClusterMapping = errors.New("invalid argocd destination cluster mapping")
	ErrInvalidOrganizationIDParam             = errors.New("invalid organization id param")
	ErrInvalidAgentClusterIDParam             = errors.New("invalid agent cluster id param")
	ErrInvalidClusterIDParam                  = errors.New("invalid cluster id param")
	ErrInvalidUpsertRequest                   = errors.New("invalid argocd destination cluster mapping upsert request")
)

type ArgoCdDestinationClusterMapping struct {
	OrganizationID   uuid.UUID `validate:"required"`
	AgentClusterID   uuid.UUID `validate:"required"`
	ArgocdClusterUrl string    `validate:"required"`
	ClusterID        uuid.UUID `validate:"required"`
}

func (m ArgoCdDestinationClusterMapping) Validate() error {
	return validator.New().Struct(m)
}

func (m ArgoCdDestinationClusterMapping) IsValid() bool {
	return m.Validate() == nil
}

type UpsertRequest struct {
	AgentClusterId   string `validate:"required"`
	ArgocdClusterUrl string `validate:"required"`
	ClusterId        string `validate:"required"`
}

func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}
	return nil
}

func (r UpsertRequest) IsValid() bool {
	return r.Validate() == nil
}
