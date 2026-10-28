package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

const (
	wideCredentialsKeyCloudAccountID            = "wide_scope_cloud_account_id"
	wideCredentialsKeySlackWebhookURL           = "slack_webhook_url"
	wideCredentialsKeyRequireWideAccessToDeploy = "require_wide_access_to_deploy"
)

var wideCredentialsAttrTypes = map[string]attr.Type{
	wideCredentialsKeyCloudAccountID:            types.StringType,
	wideCredentialsKeySlackWebhookURL:           types.StringType,
	wideCredentialsKeyRequireWideAccessToDeploy: types.BoolType,
}

// wideCredentialsCloudAccountID returns the wide_scope_cloud_account_id of a using_wide_credentials
// object, null when the block is not set.
func wideCredentialsCloudAccountID(obj types.Object) types.String {
	if obj.IsNull() || obj.IsUnknown() {
		return types.StringNull()
	}
	value, ok := obj.Attributes()[wideCredentialsKeyCloudAccountID].(types.String)
	if !ok {
		return types.StringNull()
	}
	return value
}

// toQoveryWidePermissions converts using_wide_credentials into the wide_permissions of a cluster
// request. Setting the block turns restrictive permissions on. On update, a removed block sends
// them off explicitly so that q-core revokes the wide-scope account and the slack webhook.
//
// q-core ignores wide_scope_cloud_account_id on edit: the cloud provider info request carries it
// (see toUpsertClusterRequest).
func toQoveryWidePermissions(obj types.Object, isUpdate bool) *qovery.WidePermissionsRequestDto {
	if obj.IsUnknown() {
		return nil
	}
	if obj.IsNull() {
		if !isUpdate {
			return nil
		}
		return &qovery.WidePermissionsRequestDto{UsingRestrictivePermissions: qovery.PtrBool(false)}
	}

	request := &qovery.WidePermissionsRequestDto{UsingRestrictivePermissions: qovery.PtrBool(true)}
	if accountID, ok := nonEmptyStringAttr(obj.Attributes(), wideCredentialsKeyCloudAccountID); ok {
		request.SetWideScopeCloudAccountId(accountID)
	}
	// An omitted slack_webhook_url is read by q-core as null, which removes the webhook.
	if slackWebhookURL, ok := nonEmptyStringAttr(obj.Attributes(), wideCredentialsKeySlackWebhookURL); ok {
		request.SetSlackWebhookUrl(slackWebhookURL)
	}
	return request
}

// fromQoveryWidePermissions converts the wide permissions of a cluster into using_wide_credentials,
// null when restrictive permissions are off.
//
// The wide-scope account is read from the cloud provider info when it has one: on update it is
// fetched after the cloud provider info request that changes the account, whereas the cluster
// returned by the edit still holds the previous one.
func fromQoveryWidePermissions(widePermissions *qovery.WidePermissionsResponseDto, cloudProviderInfo *qovery.ClusterCloudProviderInfo, prior types.Object, mode clusterReadMode) types.Object {
	if widePermissions == nil || !widePermissions.GetUsingRestrictivePermissions() {
		return types.ObjectNull(wideCredentialsAttrTypes)
	}

	accountID := FromNullableString(widePermissions.WideScopeCloudAccountId)
	if cloudProviderInfo != nil && cloudProviderInfo.WidePermissionCredentials != nil && cloudProviderInfo.WidePermissionCredentials.Id != nil {
		accountID = types.StringValue(*cloudProviderInfo.WidePermissionCredentials.Id)
	}

	return types.ObjectValueMust(wideCredentialsAttrTypes, map[string]attr.Value{
		wideCredentialsKeyCloudAccountID:            accountID,
		wideCredentialsKeySlackWebhookURL:           slackWebhookURLFromAPI(widePermissions.SlackWebhookUrl, prior, mode),
		wideCredentialsKeyRequireWideAccessToDeploy: types.BoolValue(widePermissions.GetRequireWideAccessToDeploy()),
	})
}

// slackWebhookURLFromAPI reads the slack webhook url. The API never returns the real url, a secret,
// but a mask when a webhook is configured: the resource keeps the url of the prior state or plan
// then, and stores the mask only when it has none (on import). A webhook added or removed outside
// Terraform shows as a difference; one replaced by another url does not. The data source reports
// the mask as is.
func slackWebhookURLFromAPI(apiValue qovery.NullableString, prior types.Object, mode clusterReadMode) types.String {
	value := FromNullableString(apiValue)
	if value.IsNull() || mode == clusterReadModeDataSource || prior.IsNull() || prior.IsUnknown() {
		return value
	}
	priorValue, ok := prior.Attributes()[wideCredentialsKeySlackWebhookURL].(types.String)
	if !ok || priorValue.IsNull() || priorValue.IsUnknown() {
		return value
	}
	return priorValue
}
