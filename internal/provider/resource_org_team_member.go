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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgTeamMemberResource{}
	_ resource.ResourceWithConfigure   = &OrgTeamMemberResource{}
	_ resource.ResourceWithImportState = &OrgTeamMemberResource{}
)

func NewOrgTeamMemberResource() resource.Resource {
	return &OrgTeamMemberResource{}
}

type OrgTeamMemberResource struct {
	client *hubclient.Client
}

type OrgTeamMemberResourceModel struct {
	ID       types.String `tfsdk:"id"`
	OrgName  types.String `tfsdk:"org_name"`
	TeamName types.String `tfsdk:"team_name"`
	UserName types.String `tfsdk:"user_name"`
}

func (r *OrgTeamMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgTeamMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_team_member"
}
func (r *OrgTeamMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages team members associated with an organization.

	~> **Note**: This resource is only available when authenticated with a username and password as an owner of the org.

	## Example Usage

	` + "```hcl" + `
	resource "docker_org_team_member" "example" {
	org_name  = "my-organization"
	team_name = "dev-team"
	user_name = "johndoe"
	}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the team member",
				Computed:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`), "Team name must be 3-30 characters long and can only contain letters, numbers, underscores, or hyphens."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "User name to be added to the team",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *OrgTeamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgTeamMemberResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddOrgTeamMember(ctx, data.OrgName.ValueString(), data.TeamName.ValueString(), data.UserName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to add team member", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", data.OrgName.ValueString(), data.TeamName.ValueString(), data.UserName.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgTeamMemberResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call the new API to list members of the team
	membersResponse, err := r.client.ListOrgTeamMembers(ctx, data.OrgName.ValueString(), data.TeamName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_team_member resource", fmt.Sprintf("Error retrieving team members: %v", err))
		return
	}

	// Check if the specified user is in the team
	found := false
	for _, member := range membersResponse.Results {
		if member.Username == data.UserName.ValueString() {
			found = true
			break
		}
	}

	if !found {
		// If the user is not found in the team, remove the resource from state
		resp.Diagnostics.AddWarning("User not found", fmt.Sprintf("User %s is not a member of team %s. Removing from state.", data.UserName.ValueString(), data.TeamName.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since we added RequiresReplace() to user_name, there's no need to handle
	// the update logic manually here
}

func (r *OrgTeamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgTeamMemberResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteOrgTeamMember(ctx, data.OrgName.ValueString(), data.TeamName.ValueString(), data.UserName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete team member", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *OrgTeamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import logic for the resource
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_name/team_name/user_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_name"), idParts[2])...)
}
