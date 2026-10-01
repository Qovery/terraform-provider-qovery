//go:build unit || !integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
