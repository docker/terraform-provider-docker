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
	_ resource.Resource                = &OrgSettingImageAccessAllowlistResource{}
	_ resource.ResourceWithConfigure   = &OrgSettingImageAccessAllowlistResource{}
	_ resource.ResourceWithImportState = &OrgSettingImageAccessAllowlistResource{}
)

func NewOrgSettingImageAccessAllowlistResource() resource.Resource {
	return &OrgSettingImageAccessAllowlistResource{}
}

type OrgSettingImageAccessAllowlistResource struct {
	client *hubclient.Client
}

type OrgSettingImageAccessAllowlistResourceModel struct {
	OrgName     types.String `tfsdk:"org_name"`
	Repositories types.Set   `tfsdk:"repositories"`
}

func (r *OrgSettingImageAccessAllowlistResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgSettingImageAccessAllowlistResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_setting_image_access_allowlist"
}

func (r *OrgSettingImageAccessAllowlistResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the repository allowlist for an organization namespace.

When the allowlist is enabled (via ` + "`repository_allowlist_enabled`" + ` in ` + "`docker_org_setting_namespace`" + `), only repositories listed here can be used by organization members.

## Example Usage

` + "```hcl" + `
resource "docker_org_setting_namespace" "example" {
  org_name                       = "my-organization"
  disable_public_repositories    = false
  disable_push_member_namespaces = false
  repository_allowlist_enabled   = true
}

resource "docker_org_setting_image_access_allowlist" "example" {
  org_name = docker_org_setting_namespace.example.org_name
  repositories = [
    "my-organization/allowed-repo",
    "my-organization/another-repo",
  ]
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
			"repositories": schema.SetAttribute{
				MarkdownDescription: "Set of repository names (in `namespace/name` format) to include in the allowlist.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *OrgSettingImageAccessAllowlistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgSettingImageAccessAllowlistResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repos []string
	resp.Diagnostics.Append(data.Repositories.ElementsAs(ctx, &repos, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(repos) > 0 {
		if err := r.client.AddNamespaceAllowlistItems(ctx, data.OrgName.ValueString(), repos); err != nil {
			resp.Diagnostics.AddError("Unable to create org_setting_image_access_allowlist resource", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingImageAccessAllowlistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgSettingImageAccessAllowlistResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := r.client.ListNamespaceAllowlistItems(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_setting_image_access_allowlist resource", err.Error())
		return
	}

	repos := make([]string, len(items))
	for i, item := range items {
		repos[i] = item.RepositoryName
	}

	repoSet, diags := types.SetValueFrom(ctx, types.StringType, repos)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Repositories = repoSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingImageAccessAllowlistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state OrgSettingImageAccessAllowlistResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planned, current []string
	resp.Diagnostics.Append(plan.Repositories.ElementsAs(ctx, &planned, false)...)
	resp.Diagnostics.Append(state.Repositories.ElementsAs(ctx, &current, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentSet := make(map[string]struct{}, len(current))
	for _, r := range current {
		currentSet[r] = struct{}{}
	}
	plannedSet := make(map[string]struct{}, len(planned))
	for _, r := range planned {
		plannedSet[r] = struct{}{}
	}

	var toAdd, toRemove []string
	for _, r := range planned {
		if _, exists := currentSet[r]; !exists {
			toAdd = append(toAdd, r)
		}
	}
	for _, r := range current {
		if _, exists := plannedSet[r]; !exists {
			toRemove = append(toRemove, r)
		}
	}

	orgName := plan.OrgName.ValueString()
	if len(toAdd) > 0 {
		if err := r.client.AddNamespaceAllowlistItems(ctx, orgName, toAdd); err != nil {
			resp.Diagnostics.AddError("Unable to update org_setting_image_access_allowlist resource", err.Error())
			return
		}
	}
	if len(toRemove) > 0 {
		if err := r.client.RemoveNamespaceAllowlistItems(ctx, orgName, toRemove); err != nil {
			resp.Diagnostics.AddError("Unable to update org_setting_image_access_allowlist resource", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrgSettingImageAccessAllowlistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgSettingImageAccessAllowlistResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repos []string
	resp.Diagnostics.Append(data.Repositories.ElementsAs(ctx, &repos, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(repos) > 0 {
		if err := r.client.RemoveNamespaceAllowlistItems(ctx, data.OrgName.ValueString(), repos); err != nil {
			resp.Diagnostics.AddError("Unable to delete org_setting_image_access_allowlist resource", err.Error())
			return
		}
	}
}

func (r *OrgSettingImageAccessAllowlistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), req.ID)...)
}
