package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Organization struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Plan types.String `tfsdk:"plan"`
}

type AwsCredentials struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	OrganizationId  types.String `tfsdk:"organization_id"`
}

type Cluster struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Region         types.String `tfsdk:"region"`
	CredentialsId  types.String `tfsdk:"credentials_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
}
