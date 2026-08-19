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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgSettingNamespaceResource{}
	_ resource.ResourceWithConfigure   = &OrgSettingNamespaceResource{}
	_ resource.ResourceWithImportState = &OrgSettingNamespaceResource{}
)

func NewOrgSettingNamespaceResource() resource.Resource {
	return &OrgSettingNamespaceResource{}
}

type OrgSettingNamespaceResource struct {
	client *hubclient.Client
}

type OrgSettingNamespaceResourceModel struct {
	OrgName                     types.String `tfsdk:"org_name"`
	DisablePublicRepositories   types.Bool   `tfsdk:"disable_public_repositories"`
	DisablePushMemberNamespaces types.Bool   `tfsdk:"disable_push_member_namespaces"`
	RepositoryAllowlistEnabled  types.Bool   `tfsdk:"repository_allowlist_enabled"`
}

func (r *OrgSettingNamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *OrgSettingNamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_setting_namespace"
}

func (r *OrgSettingNamespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages namespace settings for an organization.

## Example Usage

` + "```hcl" + `
resource "docker_org_setting_namespace" "example" {
  org_name                       = "my-organization"
  disable_public_repositories    = true
  disable_push_member_namespaces = true
  repository_allowlist_enabled   = true
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
			"disable_public_repositories": schema.BoolAttribute{
				MarkdownDescription: "When true, prevents organization members from creating public repositories.",
				Required:            true,
			},
			"disable_push_member_namespaces": schema.BoolAttribute{
				MarkdownDescription: "When true, prevents organization members from pushing images to their personal namespaces.",
				Required:            true,
			},
			"repository_allowlist_enabled": schema.BoolAttribute{
				MarkdownDescription: "When true, enables the repository allowlist for the organization. Only repositories in the allowlist can be used. Manage allowlist entries with `docker_org_setting_image_access_allowlist`.",
				Required:            true,
			},
		},
	}
}

func (r *OrgSettingNamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgSettingNamespaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetOrgSettingNamespace(ctx, data.OrgName.ValueString(), hubclient.OrgSettingNamespace{
		DisablePublicRepositories:   data.DisablePublicRepositories.ValueBool(),
		DisablePushMemberNamespaces: data.DisablePushMemberNamespaces.ValueBool(),
		RepositoryAllowlistEnabled:  data.RepositoryAllowlistEnabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org_setting_namespace resource", err.Error())
		return
	}

	data.DisablePublicRepositories = types.BoolValue(settings.DisablePublicRepositories)
	data.DisablePushMemberNamespaces = types.BoolValue(settings.DisablePushMemberNamespaces)
	data.RepositoryAllowlistEnabled = types.BoolValue(settings.RepositoryAllowlistEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingNamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgSettingNamespaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	settings, err := r.client.GetOrgSettingNamespace(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_setting_namespace resource", err.Error())
		return
	}

	data.DisablePublicRepositories = types.BoolValue(settings.DisablePublicRepositories)
	data.DisablePushMemberNamespaces = types.BoolValue(settings.DisablePushMemberNamespaces)
	data.RepositoryAllowlistEnabled = types.BoolValue(settings.RepositoryAllowlistEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingNamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrgSettingNamespaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetOrgSettingNamespace(ctx, data.OrgName.ValueString(), hubclient.OrgSettingNamespace{
		DisablePublicRepositories:   data.DisablePublicRepositories.ValueBool(),
		DisablePushMemberNamespaces: data.DisablePushMemberNamespaces.ValueBool(),
		RepositoryAllowlistEnabled:  data.RepositoryAllowlistEnabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org_setting_namespace resource", err.Error())
		return
	}

	data.DisablePublicRepositories = types.BoolValue(settings.DisablePublicRepositories)
	data.DisablePushMemberNamespaces = types.BoolValue(settings.DisablePushMemberNamespaces)
	data.RepositoryAllowlistEnabled = types.BoolValue(settings.RepositoryAllowlistEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingNamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgSettingNamespaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Reset all settings to false (permissive defaults)
	_, err := r.client.SetOrgSettingNamespace(ctx, data.OrgName.ValueString(), hubclient.OrgSettingNamespace{
		DisablePublicRepositories:   false,
		DisablePushMemberNamespaces: false,
		RepositoryAllowlistEnabled:  false,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete org_setting_namespace resource", err.Error())
		return
	}
}

func (r *OrgSettingNamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), req.ID)...)
}
