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

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &AccessTokenResource{}
	_ resource.ResourceWithConfigure   = &AccessTokenResource{}
	_ resource.ResourceWithImportState = &AccessTokenResource{}
)

func NewAccessTokenResource() resource.Resource {
	return &AccessTokenResource{}
}

type AccessTokenResource struct {
	client *hubclient.Client
}

type AccessTokenResourceModel struct {
	UUID       types.String `tfsdk:"uuid"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	TokenLabel types.String `tfsdk:"token_label"`
	Scopes     types.List   `tfsdk:"scopes"`
	Token      types.String `tfsdk:"token"`
}

func (r *AccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_token"
}

func (r *AccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages access tokens.",

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the token",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the token is active",
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"token_label": schema.StringAttribute{
				MarkdownDescription: "Token label",
				Required:            true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "List of scopes",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token itself",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *AccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes := []string{}
	data.Scopes.ElementsAs(ctx, &scopes, false)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := hubclient.AccessTokenCreateParams{
		Scopes:     scopes,
		TokenLabel: data.TokenLabel.ValueString(),
	}

	at, err := r.client.CreateAccessToken(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create access token", err.Error())
		return
	}

	data = r.toModel(ctx, at)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var fromState AccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &fromState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	at, err := r.client.GetAccessToken(ctx, fromState.UUID.ValueString())
	// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to read access token resource", err.Error())
		return
	}

	fromAPI := r.toModel(ctx, at)
	if resp.Diagnostics.HasError() {
		return
	}

	// The token is not returned by the API after initial creation,
	// so we need to copy it from the state
	fromAPI.Token = fromState.Token

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &fromAPI)...)
}

func (r *AccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var fromState AccessTokenResourceModel
	var fromPlan AccessTokenResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &fromState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &fromPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := hubclient.AccessTokenUpdateParams{
		TokenLabel: fromPlan.TokenLabel.ValueString(),
		IsActive:   fromPlan.IsActive.ValueBool(),
	}

	at, err := r.client.UpdateAccessToken(ctx, fromState.UUID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update access token", err.Error())
		return
	}

	fromAPI := r.toModel(ctx, at)
	if resp.Diagnostics.HasError() {
		return
	}

	// The token is not returned by the API after initial creation,
	// so we need to copy it from the state
	fromAPI.Token = fromState.Token

	resp.Diagnostics.Append(resp.State.Set(ctx, &fromAPI)...)
}

func (r *AccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteAccessToken(ctx, data.UUID.ValueString())
	if isNotFound(err) {
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to delete access token", err.Error())
		return
	}
}

func (r *AccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), req.ID)...)
}

func (r *AccessTokenResource) toModel(ctx context.Context, at hubclient.AccessToken) AccessTokenResourceModel {
	scopes, _ := types.ListValueFrom(ctx, types.StringType, at.Scopes)
	return AccessTokenResourceModel{
		UUID:       types.StringValue(at.UUID),
		IsActive:   types.BoolValue(at.IsActive),
		TokenLabel: types.StringValue(at.TokenLabel),
		Scopes:     scopes,
		Token:      types.StringValue(at.Token),
	}
}
