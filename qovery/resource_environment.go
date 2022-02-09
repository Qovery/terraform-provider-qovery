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

const environmentAPIResource = "environment"

type environmentData struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
}

type environmentResourceType struct{}

func (r environmentResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"project_id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
		},
	}, nil
}

func (r environmentResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return environmentResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type environmentResource struct {
	client *qovery.APIClient
}

// Create qovery environment resource
func (r environmentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
}

// Read qovery environment resource
func (r environmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update qovery environment resource
func (r environmentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete qovery environment resource
func (r environmentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

// ImportState imports a qovery environment resource using its id
func (r environmentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStateNotImplemented(ctx, "", resp)
}

func environmentCreateAPIError(environmentName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentName, apierror.Create, res, err)
}

func environmentReadAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Read, res, err)
}

func environmentUpdateAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Update, res, err)
}

func environmentDeleteAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Delete, res, err)
}
