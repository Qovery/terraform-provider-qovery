package qovery

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
)

const projectAPIResource = "project"

type projectData struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

type projectResourceType struct{}

func (r projectResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery project resource. This can be used to create and manage Qovery projects.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the project.",
				Type:        types.StringType,
				Required:    true,
			},
		},
	}, nil
}

func (r projectResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return projectResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type projectResource struct {
	client *qovery.APIClient
}

// Create qovery project resource
func (r projectResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
}

// Read qovery project resource
func (r projectResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update qovery project resource
func (r projectResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete qovery project resource
func (r projectResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

// ImportState imports a qovery project resource using its id
func (r projectResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStateNotImplemented(ctx, "", resp)
}

func projectCreateAPIError(projectName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectName, apierror.Create, res, err)
}

func projectReadAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Read, res, err)
}

func projectUpdateAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Update, res, err)
}

func projectDeleteAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Delete, res, err)
}
