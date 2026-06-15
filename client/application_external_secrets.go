package client

import (
	"context"

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
