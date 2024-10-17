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
	"strconv"
	"strings"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &RepositoryTeamPermissionResource{}
	_ resource.ResourceWithConfigure   = &RepositoryTeamPermissionResource{}
	_ resource.ResourceWithImportState = &RepositoryTeamPermissionResource{}
)

func NewRepositoryTeamPermissionResource() resource.Resource {
	return &RepositoryTeamPermissionResource{}
}

type RepositoryTeamPermissionResource struct {
	client *hubclient.Client
}

type RepositoryTeamPermissionResourceModel struct {
	RepoID     types.String `tfsdk:"repo_id"`
	TeamID     types.Int64  `tfsdk:"team_id"`
	Permission types.String `tfsdk:"permission"`
}

func (r *RepositoryTeamPermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryTeamPermissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_repository_team_permission"
}

func (r *RepositoryTeamPermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Grants team permissions to an image repository.

~> **Note** When used with a Personal Access Token authentication (PAT), the PAT should
   have the "Read, Write, and Delete" scope to create and delete team permissions. The
   owner of the PAT must be an editor of the org.

## Example Usage

` + "```hcl" + `
resource "docker_hub_repository" "my_repo" {
  namespace        = "my-namespace"
  name             = "my-repo"
}

resource "docker_hub_repository_team_permission" "my_repo" {
  repo_id    = docker_hub_repository.my_repo.id
  team_id    = 123456
  permission = "admin"
}
` + "```" + `

`,

		Attributes: map[string]schema.Attribute{
			"repo_id": schema.StringAttribute{
				MarkdownDescription: "The namespace/name of the repository",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric id of the team",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "The permission to assign to the team and repository.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						hubclient.TeamRepoPermissionLevelRead,
						hubclient.TeamRepoPermissionLevelWrite,
						hubclient.TeamRepoPermissionLevelAdmin,
					),
				},
			},
		},
	}
}

func (r *RepositoryTeamPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepositoryTeamPermissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	perm, err := r.client.CreatePermissionForTeamAndRepo(
		ctx,
		data.RepoID.ValueString(),
		data.TeamID.ValueInt64(),
		data.Permission.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create repository_team_permission resource", err.Error())
		return
	}

	data.Permission = types.StringValue(perm.Permission)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryTeamPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepositoryTeamPermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	perm, err := r.client.GetPermissionForTeamAndRepo(
		ctx,
		data.RepoID.ValueString(),
		data.TeamID.ValueInt64(),
	)
	// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to read repository_team_permission resource", err.Error())
		return
	}

	data.Permission = types.StringValue(perm.Permission)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryTeamPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepositoryTeamPermissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	perm, err := r.client.UpdatePermissionForTeamAndRepo(
		ctx,
		data.RepoID.ValueString(),
		data.TeamID.ValueInt64(),
		data.Permission.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update repository_team_permission resource", err.Error())
		return
	}

	data.Permission = types.StringValue(perm.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryTeamPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepositoryTeamPermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeletePermissionForTeamAndRepo(
		ctx,
		data.RepoID.ValueString(),
		data.TeamID.ValueInt64(),
	)
	if isNotFound(err) {
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to delete repository_team_permission resource", err.Error())
		return
	}
}

func (r *RepositoryTeamPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: repo_namespace/repo_name/team_id. Got: %q", req.ID),
		)
		return
	}

	teamID, err := strconv.ParseInt(idParts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected numeric team id in identifier. Got: %q", idParts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repo_id"), idParts[0]+"/"+idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), teamID)...)
}
