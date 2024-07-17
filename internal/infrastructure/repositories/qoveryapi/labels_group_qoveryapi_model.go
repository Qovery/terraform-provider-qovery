package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func newDomainLabelsGroupFromQovery(labelsGroupResponse *qovery.OrganizationLabelsGroupResponse) (*labels_group.LabelsGroup, error) {
	if labelsGroupResponse == nil {
		return nil, variable.ErrNilVariable
	}

	labelsGroupID, err := uuid.Parse(labelsGroupResponse.Id)
	if err != nil {
		return nil, errors.Wrap(err, labels_group.ErrInvalidLabelsGroupIdParam.Error())
	}

	return &labels_group.LabelsGroup{
		Id:     labelsGroupID,
		Name:   labelsGroupResponse.Name,
		Labels: labelsGroupResponse.Labels,
	}, nil
}

func newQoveryLabelsGroupRequestFromDomain(request labels_group.UpsertRequest) *qovery.OrganizationLabelsGroupCreateRequest {
	labels := make([]qovery.Label, 0, len(request.Labels))
	for _, label := range request.Labels {
		labels = append(labels, qovery.Label{Key: label.Key, Value: label.Value, PropagateToCloudProvider: label.PropagateToCloudProvider})
	}

	return &qovery.OrganizationLabelsGroupCreateRequest{
		Name:   request.Name,
		Labels: labels,
	}
}

func NewQoveryServiceLabelsGroupRequestFromDomain(labelsGroupIds []string) ([]qovery.ServiceLabelRequest, error) {
	serviceLabelRequest := make([]qovery.ServiceLabelRequest, 0, len(labelsGroupIds))
	for _, id := range labelsGroupIds {
		newLabelsRequest := qovery.ServiceLabelRequest{
			Id: id,
		}

		serviceLabelRequest = append(serviceLabelRequest, newLabelsRequest)
	}

	return serviceLabelRequest, nil
}
