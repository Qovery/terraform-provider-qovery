//go:build unit || !integration

package qovery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestConvertDomainCredentialsToAWSCredentials_StaticCredentials(t *testing.T) {
	t.Parallel()

	credID := uuid.New()
	orgID := uuid.New()
	creds := &credentials.Credentials{
		ID:             credID,
		OrganizationID: orgID,
		Name:           "test-aws-creds",
	}

	plan := AWSCredentials{
		Id:              types.StringValue(""),
		OrganizationId:  types.StringValue(orgID.String()),
		Name:            types.StringValue("test-aws-creds"),
		AccessKeyId:     types.StringValue("AKIAIOSFODNN7EXAMPLE"),
		SecretAccessKey: types.StringValue("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		RoleArn:         types.StringNull(),
	}

	result := convertDomainCredentialsToAWSCredentials(creds, plan)

	assert.Equal(t, credID.String(), result.Id.ValueString())
	assert.Equal(t, orgID.String(), result.OrganizationId.ValueString())
	assert.Equal(t, "test-aws-creds", result.Name.ValueString())
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", result.AccessKeyId.ValueString())
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", result.SecretAccessKey.ValueString())
	assert.True(t, result.RoleArn.IsNull())
}

func TestConvertDomainCredentialsToAWSCredentials_RoleCredentials(t *testing.T) {
	t.Parallel()

	credID := uuid.New()
	orgID := uuid.New()
	creds := &credentials.Credentials{
		ID:             credID,
		OrganizationID: orgID,
		Name:           "test-role-creds",
	}

	plan := AWSCredentials{
		Id:              types.StringValue(""),
		OrganizationId:  types.StringValue(orgID.String()),
		Name:            types.StringValue("test-role-creds"),
		AccessKeyId:     types.StringNull(),
		SecretAccessKey: types.StringNull(),
		RoleArn:         types.StringValue("arn:aws:iam::123456789012:role/qovery-role"),
	}

	result := convertDomainCredentialsToAWSCredentials(creds, plan)

	assert.Equal(t, credID.String(), result.Id.ValueString())
	assert.Equal(t, orgID.String(), result.OrganizationId.ValueString())
	assert.Equal(t, "test-role-creds", result.Name.ValueString())
	assert.True(t, result.AccessKeyId.IsNull())
	assert.True(t, result.SecretAccessKey.IsNull())
	assert.Equal(t, "arn:aws:iam::123456789012:role/qovery-role", result.RoleArn.ValueString())
}

func TestToUpsertAwsRequest_StaticCredentials(t *testing.T) {
	t.Parallel()

	creds := AWSCredentials{
		Name:            types.StringValue("test-creds"),
		AccessKeyId:     types.StringValue("AKIAIOSFODNN7EXAMPLE"),
		SecretAccessKey: types.StringValue("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		RoleArn:         types.StringNull(),
	}

	req := creds.toUpsertAwsRequest()

	assert.Equal(t, "test-creds", req.Name)
	assert.NotNil(t, req.StaticCredentials)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", req.StaticCredentials.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", req.StaticCredentials.SecretAccessKey)
	assert.Nil(t, req.RoleCredentials)
}

func TestToUpsertAwsRequest_RoleCredentials(t *testing.T) {
	t.Parallel()

	creds := AWSCredentials{
		Name:    types.StringValue("test-role-creds"),
		RoleArn: types.StringValue("arn:aws:iam::123456789012:role/qovery-role"),
	}

	req := creds.toUpsertAwsRequest()

	assert.Equal(t, "test-role-creds", req.Name)
	assert.Nil(t, req.StaticCredentials)
	assert.NotNil(t, req.RoleCredentials)
	assert.Equal(t, "arn:aws:iam::123456789012:role/qovery-role", req.RoleCredentials.RoleArn)
}
