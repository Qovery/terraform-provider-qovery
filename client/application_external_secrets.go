package client

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func (c *Client) updateApplicationExternalSecrets(ctx context.Context, applicationID string, diff variable.ExternalSecretDiffRequest) *apierrors.APIError {
	for _, item := range diff.Delete {
		res, err := c.api.VariableMainCallsAPI.
			DeleteVariable(ctx, item.VariableID).
			Execute()
		if err != nil || res.StatusCode >= 300 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationExternalSecret, item.VariableID, res, err)
		}
	}

	for _, item := range diff.Update {
		smAccessID := item.SecretManagerAccessId
		_, res, err := c.api.VariableMainCallsAPI.
			EditVariable(ctx, item.VariableID).
			VariableEditRequest(qovery.VariableEditRequest{
				Key:                   item.Key,
				Value:                 *qovery.NewNullableString(&item.Reference),
				SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
				Description:           *qovery.NewNullableString(&item.Description),
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationExternalSecret, item.VariableID, res, err)
		}
	}

	for _, item := range diff.Create {
		smAccessID := item.SecretManagerAccessId
		_, res, err := c.api.VariableMainCallsAPI.
			CreateVariable(ctx).
			VariableRequest(qovery.VariableRequest{
				Key:                   item.Key,
				Value:                 item.Reference,
				IsSecret:              false,
				VariableScope:         qovery.APIVARIABLESCOPEENUM_APPLICATION,
				VariableParentId:      applicationID,
				SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
				Description:           *qovery.NewNullableString(&item.Description),
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationExternalSecret, item.Key, res, err)
		}
	}

	return nil
}

func (c *Client) updateApplicationExternalSecretFiles(ctx context.Context, applicationID string, diff variable.ExternalSecretFileDiffRequest) *apierrors.APIError {
	for _, item := range diff.Delete {
		res, err := c.api.VariableMainCallsAPI.
			DeleteVariable(ctx, item.VariableID).
			Execute()
		if err != nil || res.StatusCode >= 300 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationExternalSecretFile, item.VariableID, res, err)
		}
	}

	for _, item := range diff.Update {
		smAccessID := item.SecretManagerAccessId
		_, res, err := c.api.VariableMainCallsAPI.
			EditVariable(ctx, item.VariableID).
			VariableEditRequest(qovery.VariableEditRequest{
				Key:                   item.Key,
				Value:                 *qovery.NewNullableString(&item.Reference),
				SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
				Description:           *qovery.NewNullableString(&item.Description),
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationExternalSecretFile, item.VariableID, res, err)
		}
	}

	for _, item := range diff.Create {
		smAccessID := item.SecretManagerAccessId
		_, res, err := c.api.VariableMainCallsAPI.
			CreateVariable(ctx).
			VariableRequest(qovery.VariableRequest{
				Key:                   item.Key,
				Value:                 item.Reference,
				IsSecret:              false,
				VariableScope:         qovery.APIVARIABLESCOPEENUM_APPLICATION,
				VariableParentId:      applicationID,
				SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
				MountPath:             *qovery.NewNullableString(&item.MountPath),
				Description:           *qovery.NewNullableString(&item.Description),
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationExternalSecretFile, item.Key, res, err)
		}
	}

	return nil
}

func (c *Client) getApplicationExternalSecretsAndFiles(ctx context.Context, applicationID string) (variable.ExternalSecrets, variable.ExternalSecretFiles, *apierrors.APIError) {
	vars, resp, err := c.api.VariableMainCallsAPI.
		ListVariables(ctx).
		ParentId(applicationID).
		Scope(qovery.APIVARIABLESCOPEENUM_APPLICATION).
		IsSecret(false).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, nil, apierrors.NewReadError(apierrors.APIResourceApplicationExternalSecret, applicationID, resp, err)
	}

	secrets := make(variable.ExternalSecrets, 0)
	files := make(variable.ExternalSecretFiles, 0)
	for _, v := range vars.GetResults() {
		switch {
		case strings.EqualFold(string(v.VariableType), "EXTERNAL_SECRET"):
			s, convErr := newApplicationExternalSecretFromQovery(&v)
			if convErr != nil {
				return nil, nil, apierrors.NewReadError(apierrors.APIResourceApplicationExternalSecret, applicationID, nil, convErr)
			}
			secrets = append(secrets, *s)
		case strings.EqualFold(string(v.VariableType), "FILE_EXTERNAL_SECRET"):
			f, convErr := newApplicationExternalSecretFileFromQovery(&v)
			if convErr != nil {
				return nil, nil, apierrors.NewReadError(apierrors.APIResourceApplicationExternalSecretFile, applicationID, nil, convErr)
			}
			files = append(files, *f)
		}
	}
	return secrets, files, nil
}

func newApplicationExternalSecretFromQovery(v *qovery.VariableResponse) (*variable.ExternalSecret, error) {
	reference := ""
	if v.Value.IsSet() && v.Value.Get() != nil {
		reference = *v.Value.Get()
	}

	smAccessID := ""
	if v.SecretManagerAccessId.IsSet() && v.SecretManagerAccessId.Get() != nil {
		smAccessID = *v.SecretManagerAccessId.Get()
	}

	description := ""
	if v.Description != nil {
		description = *v.Description
	}

	scope, err := variable.NewScopeFromString(string(v.Scope))
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrInvalidScopeParam.Error())
	}
	return &variable.ExternalSecret{
		ID:                    uuid.MustParse(v.GetId()),
		Key:                   v.Key,
		Description:           description,
		Reference:             reference,
		SecretManagerAccessId: smAccessID,
		Scope:                 *scope,
		VariableType:          string(v.VariableType),
	}, nil
}

func newApplicationExternalSecretFileFromQovery(v *qovery.VariableResponse) (*variable.ExternalSecretFile, error) {
	reference := ""
	if v.Value.IsSet() && v.Value.Get() != nil {
		reference = *v.Value.Get()
	}

	smAccessID := ""
	if v.SecretManagerAccessId.IsSet() && v.SecretManagerAccessId.Get() != nil {
		smAccessID = *v.SecretManagerAccessId.Get()
	}

	mountPath := ""
	if v.MountPath.IsSet() && v.MountPath.Get() != nil {
		mountPath = *v.MountPath.Get()
	}

	description := ""
	if v.Description != nil {
		description = *v.Description
	}

	scope, err := variable.NewScopeFromString(string(v.Scope))
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrInvalidScopeParam.Error())
	}
	return &variable.ExternalSecretFile{
		ID:                    uuid.MustParse(v.GetId()),
		Key:                   v.Key,
		Description:           description,
		MountPath:             mountPath,
		Reference:             reference,
		SecretManagerAccessId: smAccessID,
		Scope:                 *scope,
		VariableType:          string(v.VariableType),
	}, nil
}
