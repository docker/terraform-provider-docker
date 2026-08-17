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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgSettingRegistryResource{}
	_ resource.ResourceWithConfigure   = &OrgSettingRegistryResource{}
	_ resource.ResourceWithImportState = &OrgSettingRegistryResource{}
)

func NewOrgSettingRegistryResource() resource.Resource {
	return &OrgSettingRegistryResource{}
}

type OrgSettingRegistryResource struct {
	client *hubclient.Client
}

type OrgSettingRegistryResourceModel struct {
	OrgName               types.String `tfsdk:"org_name"`
	DefaultRepoVisibility types.String `tfsdk:"default_repo_visibility"`
}

func (r *OrgSettingRegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgSettingRegistryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_setting_registry"
}

func (r *OrgSettingRegistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the registry settings for an organization namespace, including the default repository visibility.

> **Note:** Setting ` + "`default_repo_visibility`" + ` to ` + "`\"public\"`" + ` is blocked by the API when ` + "`disable_public_repositories`" + ` is enabled in ` + "`docker_org_setting_namespace`" + `.

## Example Usage

` + "```hcl" + `
resource "docker_org_setting_registry" "example" {
  org_name                = "my-organization"
  default_repo_visibility = "private"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_repo_visibility": schema.StringAttribute{
				MarkdownDescription: "Default visibility for new repositories. Must be `\"public\"` or `\"private\"`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "private"),
				},
			},
		},
	}
}

func (r *OrgSettingRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgSettingRegistryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetDefaultRepoVisibility(ctx, data.OrgName.ValueString(), data.DefaultRepoVisibility.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org_setting_registry resource", err.Error())
		return
	}

	data.DefaultRepoVisibility = types.StringValue(settings.DefaultRepoVisibility)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgSettingRegistryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetRegistrySettings(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_setting_registry resource", err.Error())
		return
	}

	data.DefaultRepoVisibility = types.StringValue(settings.DefaultRepoVisibility)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrgSettingRegistryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetDefaultRepoVisibility(ctx, data.OrgName.ValueString(), data.DefaultRepoVisibility.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org_setting_registry resource", err.Error())
		return
	}

	data.DefaultRepoVisibility = types.StringValue(settings.DefaultRepoVisibility)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgSettingRegistryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset to the default visibility (public).
	if _, err := r.client.SetDefaultRepoVisibility(ctx, data.OrgName.ValueString(), "public"); err != nil {
		resp.Diagnostics.AddError("Unable to delete org_setting_registry resource", err.Error())
	}
}

func (r *OrgSettingRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), req.ID)...)
}
