package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// newDomainCredentialsFromQovery takes a qovery.Secret returned by the API client and turns it into the domain model secret.Secret.
func newDomainSecretsFromQovery(list *qovery.SecretResponseList) (secret.Secrets, error) {
	vars := make(secret.Secrets, 0, len(list.GetResults()))
	for _, it := range list.GetResults() {
		v, err := newDomainSecretFromQovery(&it)
		if err != nil {
			return nil, err
		}

		vars = append(vars, *v)
	}

	return vars, nil
}

// newDomainCredentialsFromQovery takes a qovery.Secret returned by the API client and turns it into the domain model secret.Secret.
func newDomainSecretFromQovery(v *qovery.Secret) (*secret.Secret, error) {
	if v == nil {
		return nil, secret.ErrNilSecret
	}

	description := ""
	// shouldnt we have a description on qovery.Secret?
	//if v.Description.IsSet() {
	//	description = v.Description.Get()
	//}
	return secret.NewSecret(secret.NewSecretParams{
		SecretID:    v.GetId(),
		Scope:       string(v.Scope),
		Key:         v.GetKey(),
		Type:        string(*v.VariableType),
		Description: description,
	})
}

// newQoverySecretRequestFromDomain takes the domain request secret.UpsertRequest and turns it into a qovery.SecretRequest to make the api call.
func newQoverySecretRequestFromDomain(request secret.UpsertRequest) qovery.SecretRequest {
	return qovery.SecretRequest{
		Key:   request.Key,
		Value: &request.Value,
		// shouldnt we have a description here?
		// Description: *qovery.NewNullableString(&request.Description),
	}
}

// newQoverySecretEditRequestFromDomain takes the domain request secret.UpsertRequest and turns it into a qovery.SecretEditRequest to make the api call.
func newQoverySecretEditRequestFromDomain(request secret.UpsertRequest) qovery.SecretEditRequest {
	return qovery.SecretEditRequest{
		Key:   request.Key,
		Value: &request.Value,
		// shouldnt we have a description here?
		// Description: *qovery.NewNullableString(&request.Description),
	}
}
