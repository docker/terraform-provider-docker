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
	"regexp"
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
	_ resource.Resource                = &OrgTeamResource{}
	_ resource.ResourceWithConfigure   = &OrgTeamResource{}
	_ resource.ResourceWithImportState = &OrgTeamResource{}
)

func NewOrgTeamResource() resource.Resource {
	return &OrgTeamResource{}
}

type OrgTeamResource struct {
	client *hubclient.Client
}

type OrgTeamResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	OrgName  types.String `tfsdk:"org_name"`
	TeamName types.String `tfsdk:"team_name"`
	TeamDesc types.String `tfsdk:"team_description"`
}

func (r *OrgTeamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgTeamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_team"
}
func (r *OrgTeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages teams for an organization.

	~> **Note**: This resource is only available when authenticated with a username and password as an owner of the org.

	## Example Usage

	` + "```hcl" + `
	resource "docker_org_team" "example" {
	org_name         = "my-organization"
	team_name        = "dev-team"
	team_description = "Development team responsible for backend services"
	}
	` + "```" + `

`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID associated with the team",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The name of the organization",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				MarkdownDescription: "The name of the team",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`), "Team name must be 3-30 characters long and can only contain letters, numbers, underscores, or hyphens."),
				},
			},
			"team_description": schema.StringAttribute{
				MarkdownDescription: "A description of the team's purpose or responsibilities",
				Optional:            true,
			},
		},
	}
}

func (r *OrgTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgTeamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := hubclient.OrgTeam{
		Name:        data.TeamName.ValueString(),
		Description: data.TeamDesc.ValueString(),
	}

	orgTeam, err := r.client.CreateOrgTeam(ctx, data.OrgName.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org_team resource", err.Error())
		return
	}

	data.ID = types.Int64Value(orgTeam.ID)
	data.TeamName = types.StringValue(orgTeam.Name)
	if len(orgTeam.Description) > 0 {
		data.TeamDesc = types.StringValue(orgTeam.Description)
	} else {
		data.TeamDesc = types.StringNull()
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgTeamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	orgTeam, err := r.client.GetOrgTeam(ctx, data.OrgName.ValueString(), data.TeamName.ValueString())
	// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to read org_team resource", err.Error())
		return
	}

	data.ID = types.Int64Value(orgTeam.ID)
	data.TeamName = types.StringValue(orgTeam.Name)
	if len(orgTeam.Description) > 0 {
		data.TeamDesc = types.StringValue(orgTeam.Description)
	} else {
		data.TeamDesc = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrgTeamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	updateReq := hubclient.OrgTeam{
		Name:        data.TeamName.ValueString(),
		Description: data.TeamDesc.ValueString(),
	}

	// Updates to Team Names are a bit awkward.
	// It takes in the old team name in path, but needs new team name in body
	// It does not use/accept the numeric id as "key" although this stays consistent
	var stateData OrgTeamResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	orgTeam, err := r.client.UpdateOrgTeam(ctx, data.OrgName.ValueString(), stateData.TeamName.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org_team resource", err.Error())
		return
	}

	data.ID = types.Int64Value(orgTeam.ID)
	data.TeamName = types.StringValue(orgTeam.Name)
	if len(orgTeam.Description) > 0 {
		data.TeamDesc = types.StringValue(orgTeam.Description)
	} else {
		data.TeamDesc = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgTeamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteOrgTeam(ctx, data.OrgName.ValueString(), data.TeamName.ValueString())
	if isNotFound(err) {
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to delete org_team resource", err.Error())
		return
	}
}

func (r *OrgTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// we can't import on the numeric id because we can't READ by numeric id in API (only by org_name/team_name)
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_name/team_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_name"), idParts[1])...)
}

func isNotFound(err error) bool {
	// todo: better
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}
