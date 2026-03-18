/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgAccessTokenResource{}
	_ resource.ResourceWithConfigure   = &OrgAccessTokenResource{}
	_ resource.ResourceWithImportState = &OrgAccessTokenResource{}
)

func NewOrgAccessTokenResource() resource.Resource {
	return &OrgAccessTokenResource{}
}

type OrgAccessTokenResource struct {
	client *hubclient.Client
}

type OrgAccessTokenResourceModel struct {
	ID          types.String                       `tfsdk:"id"`
	OrgName     types.String                       `tfsdk:"org_name"`
	Label       types.String                       `tfsdk:"label"`
	Description types.String                       `tfsdk:"description"`
	Resources   []OrgAccessTokenResourceEntryModel `tfsdk:"resources"`
	ExpiresAt   types.String                       `tfsdk:"expires_at"`
	Token       types.String                       `tfsdk:"token"`
	CreatedBy   types.String                       `tfsdk:"created_by"`
	CreatedAt   types.String                       `tfsdk:"created_at"`
}

type OrgAccessTokenResourceEntryModel struct {
	Type   types.String `tfsdk:"type"`
	Path   types.String `tfsdk:"path"`
	Scopes types.List   `tfsdk:"scopes"`
}

func (r *OrgAccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hubclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *OrgAccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_access_token"
}

func (r *OrgAccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages organization access tokens.

-> **Note**: This resource is only available when authenticated with a username and password as an owner of the org.

## Example Usage

` + "```hcl" + `
resource "docker_org_access_token" "example" {
  org_name    = "my-organization"
  label       = "ci-token"
  description = "Token for CI pulls"
  resources = [
    {
      type   = "TYPE_REPO"
      path   = "my-organization/*"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2027-12-31T23:59:59Z"
}
` + "```" + `

For ` + "`TYPE_REPO`" + ` resources, ` + "`path`" + ` must point to an existing repository or a supported glob such as ` + "`my-organization/*`" + `.

## Public-Only Repositories

Use the special path ` + "`*/*/public`" + ` to scope the token to public repositories only.

` + "```hcl" + `
resource "docker_org_access_token" "public_pull" {
  org_name = "my-organization"
  label    = "public-pull-token"

  resources = [
    {
      type   = "TYPE_REPO"
      path   = "*/*/public"
      scopes = ["scope-image-pull"]
    }
  ]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization access token",
				Computed:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The organization namespace",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Label for the access token",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the access token",
				Optional:            true,
			},
			"resources": schema.ListNestedAttribute{
				MarkdownDescription: "Resources this token has access to",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of resource",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(hubclient.OrgAccessTokenTypeRepo, hubclient.OrgAccessTokenTypeOrg),
							},
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The path of the resource. For TYPE_REPO, this must point to an existing repository or a supported glob such as `my-organization/*`. Use `*/*/public` for public repositories only.",
							Required:            true,
						},
						"scopes": schema.ListAttribute{
							MarkdownDescription: "The scopes this token has access to",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Expiration date for the token. Changing this value recreates the token.",
				Optional:            true,
				Validators: []validator.String{
					accessTokenExpiresAtValidator,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The organization access token. This value is only returned during creation.",
				Computed:            true,
				Sensitive:           true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user that created the access token",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation time of the access token",
				Computed:            true,
			},
		},
	}
}

func (r *OrgAccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgAccessTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenResources, diags := expandOrgAccessTokenResources(ctx, data.Resources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := hubclient.OrgAccessTokenCreateParams{
		Label:       data.Label.ValueString(),
		Description: stringValueOrEmpty(data.Description),
		Resources:   tokenResources,
	}
	if !data.ExpiresAt.IsNull() {
		createReq.ExpiresAt = data.ExpiresAt.ValueString()
	}

	at, err := r.client.CreateOrgAccessToken(ctx, data.OrgName.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org access token", err.Error())
		return
	}

	model, diags := r.toModel(ctx, at, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *OrgAccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var fromState OrgAccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &fromState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	at, err := r.client.GetOrgAccessToken(ctx, fromState.OrgName.ValueString(), fromState.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to read org access token", err.Error())
		return
	}

	model, diags := r.toModel(ctx, at, &fromState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *OrgAccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var fromState OrgAccessTokenResourceModel
	var fromPlan OrgAccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &fromState)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &fromPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenResources, diags := expandOrgAccessTokenResources(ctx, fromPlan.Resources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := hubclient.OrgAccessTokenUpdateParams{
		Label:       fromPlan.Label.ValueString(),
		Description: stringValueOrEmpty(fromPlan.Description),
		Resources:   tokenResources,
	}

	at, err := r.client.UpdateOrgAccessToken(ctx, fromState.OrgName.ValueString(), fromState.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org access token", err.Error())
		return
	}

	model, diags := r.toModel(ctx, at, &fromState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *OrgAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgAccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrgAccessToken(ctx, data.OrgName.ValueString(), data.ID.ValueString())
	if isNotFound(err) {
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to delete org access token", err.Error())
		return
	}
}

func (r *OrgAccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_name/id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func expandOrgAccessTokenResources(ctx context.Context, resources []OrgAccessTokenResourceEntryModel) ([]hubclient.OrgAccessTokenResource, diag.Diagnostics) {
	var diags diag.Diagnostics

	tokenResources := make([]hubclient.OrgAccessTokenResource, 0, len(resources))
	for _, resourceModel := range resources {
		var scopes []string
		diags.Append(resourceModel.Scopes.ElementsAs(ctx, &scopes, false)...)
		if diags.HasError() {
			return nil, diags
		}

		tokenResources = append(tokenResources, hubclient.OrgAccessTokenResource{
			Type:   resourceModel.Type.ValueString(),
			Path:   resourceModel.Path.ValueString(),
			Scopes: scopes,
		})
	}

	return tokenResources, diags
}

func flattenOrgAccessTokenResources(ctx context.Context, resources []hubclient.OrgAccessTokenResource) ([]OrgAccessTokenResourceEntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	resourceModels := make([]OrgAccessTokenResourceEntryModel, 0, len(resources))
	for _, resource := range resources {
		scopes, scopesDiags := types.ListValueFrom(ctx, types.StringType, resource.Scopes)
		diags.Append(scopesDiags...)
		if diags.HasError() {
			return nil, diags
		}

		resourceModels = append(resourceModels, OrgAccessTokenResourceEntryModel{
			Type:   types.StringValue(resource.Type),
			Path:   types.StringValue(resource.Path),
			Scopes: scopes,
		})
	}

	return resourceModels, diags
}

func (r *OrgAccessTokenResource) toModel(
	ctx context.Context,
	at hubclient.OrgAccessToken,
	currentState *OrgAccessTokenResourceModel,
) (OrgAccessTokenResourceModel, diag.Diagnostics) {
	resources, diags := flattenOrgAccessTokenResources(ctx, at.Resources)
	if diags.HasError() {
		return OrgAccessTokenResourceModel{}, diags
	}

	model := OrgAccessTokenResourceModel{
		ID:          types.StringValue(at.ID),
		Label:       types.StringValue(at.Label),
		Description: types.StringValue(at.Description),
		Resources:   resources,
		ExpiresAt:   types.StringValue(at.ExpiresAt),
		Token:       types.StringValue(at.Token),
		CreatedBy:   types.StringValue(at.CreatedBy),
		CreatedAt:   types.StringValue(at.CreatedAt),
	}

	if currentState != nil {
		model.OrgName = currentState.OrgName

		// The token is not returned by the API after initial creation,
		// so we need to preserve it from state on subsequent reads.
		if !currentState.Token.IsUnknown() {
			model.Token = currentState.Token
		}
	}

	return model, diags
}

func stringValueOrEmpty(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}
