package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

func newDomainStatusFromQovery(qoveryStatus *qovery.Status) (*status.Status, error) {
	if qoveryStatus == nil {
		return nil, status.ErrNilStatus
	}

	return status.NewStatus(status.NewStatusParams{
		StatusID:                qoveryStatus.Id,
		ServiceDeploymentStatus: string(qoveryStatus.ServiceDeploymentStatus),
		State:                   string(qoveryStatus.State),
		LastDeploymentDate:      qoveryStatus.LastDeploymentDate,
		Message:                 qoveryStatus.Message,
	})
}
