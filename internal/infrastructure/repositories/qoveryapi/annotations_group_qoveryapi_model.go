package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func newDomainAnnotationsGroupFromQovery(annotationsGroupResponse *qovery.OrganizationAnnotationsGroupResponse) (*annotations_group.AnnotationsGroup, error) {
	if annotationsGroupResponse == nil {
		return nil, variable.ErrNilVariable
	}

	annotationsGroupID, err := uuid.Parse(annotationsGroupResponse.Id)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupIdParam.Error())
	}

	return &annotations_group.AnnotationsGroup{
		Id:          annotationsGroupID,
		Name:        annotationsGroupResponse.Name,
		Annotations: annotationsGroupResponse.Annotations,
		Scopes:      annotationsGroupResponse.Scopes,
	}, nil
}

func newQoveryAnnotationsGroupRequestFromDomain(request annotations_group.UpsertRequest) (*qovery.OrganizationAnnotationsGroupCreateRequest, error) {
	annotations := make([]qovery.Annotation, 0, len(request.Annotations))
	for _, annotation := range request.Annotations {
		annotations = append(annotations, qovery.Annotation{Key: annotation.Key, Value: annotation.Value})
	}

	scopes := make([]qovery.OrganizationAnnotationsGroupScopeEnum, 0, len(request.Scopes))
	for _, scope := range request.Scopes {
		s, err := qovery.NewOrganizationAnnotationsGroupScopeEnumFromValue(scope)
		if err != nil {
			return nil, annotations_group.ErrInvalidScope
		}
		scopes = append(scopes, *s)
	}

	return &qovery.OrganizationAnnotationsGroupCreateRequest{
		Name:        request.Name,
		Annotations: annotations,
		Scopes:      scopes,
	}, nil
}

func NewQoveryServiceAnnotationsGroupRequestFromDomain(annotationsGroupIds []string) ([]qovery.ServiceAnnotationRequest, error) {
	serviceAnnotationRequest := make([]qovery.ServiceAnnotationRequest, 0, len(annotationsGroupIds))
	for _, id := range annotationsGroupIds {
		newAnnotationsRequest := qovery.ServiceAnnotationRequest{
			Id: id,
		}

		serviceAnnotationRequest = append(serviceAnnotationRequest, newAnnotationsRequest)
	}

	return serviceAnnotationRequest, nil
}
