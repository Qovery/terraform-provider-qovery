package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainProjectFromQovery(p *qovery.Project) (*project.Project, error) {
	if p == nil {
		return nil, variable.ErrNilVariable
	}

	return project.NewProject(project.NewProjectParams{
		ProjectID:      p.Id,
		OrganizationID: p.Organization.Id,
		Name:           p.Name,
		Description:    p.Description,
	})
}

// newQoveryEnvironmentVariableRequestFromDomain takes the domain request variable.UpsertRequest and turns it into a qovery.EnvironmentVariableRequest to make the api call.
func newQoveryProjectRequestFromDomain(request project.UpsertRepositoryRequest) qovery.ProjectRequest {
	return qovery.ProjectRequest{
		Name:        request.Name,
		Description: request.Description,
	}
}
