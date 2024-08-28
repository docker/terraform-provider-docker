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
	_ resource.Resource                = &OrgTeamMemberAssociationResource{}
	_ resource.ResourceWithConfigure   = &OrgTeamMemberAssociationResource{}
	_ resource.ResourceWithImportState = &OrgTeamMemberAssociationResource{}
)

func NewOrgTeamMemberAssociationResource() resource.Resource {
	return &OrgTeamMemberAssociationResource{}
}

type OrgTeamMemberAssociationResource struct {
	client *hubclient.Client
}

type OrgTeamMemberAssociationResourceModel struct {
	ID       types.String `tfsdk:"id"`
	OrgName  types.String `tfsdk:"org_name"`
	TeamName types.String `tfsdk:"team_name"`
	UserName types.String `tfsdk:"user_name"`
}

func (r *OrgTeamMemberAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgTeamMemberAssociationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_team_member_association"
}

func (r *OrgTeamMemberAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Docker team member associations for an organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the team member association",
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
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "User name to be added to the team",
				Required:            true,
			},
		},
	}
}

func (r *OrgTeamMemberAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgTeamMemberAssociationResourceModel

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

func (r *OrgTeamMemberAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO:
	// The endpoint to list team members is returning a 503

	var data OrgTeamMemberAssociationResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.ListOrgTeamMembers(ctx, data.OrgName.ValueString(), data.TeamName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_team_member_association resource", err.Error())
		return
	}

	// Check if the specified user is in the team
	found := false
	for _, member := range members.Results {
		if member.Username == data.UserName.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("User not found", fmt.Sprintf("User %s is not a member of team %s", data.UserName.ValueString(), data.TeamName.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgTeamMemberAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrgTeamMemberAssociationResourceModel
	var state OrgTeamMemberAssociationResourceModel

	// Read Terraform plan and state data into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the username has changed, update the team member
	if plan.UserName.ValueString() != state.UserName.ValueString() {
		// Remove the old user
		err := r.client.DeleteOrgTeamMember(ctx, state.OrgName.ValueString(), state.TeamName.ValueString(), state.UserName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to delete old team member", err.Error())
			return
		}

		// Add the new user
		err = r.client.AddOrgTeamMember(ctx, plan.OrgName.ValueString(), plan.TeamName.ValueString(), plan.UserName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to add new team member", err.Error())
			return
		}
	}

	// Save the updated state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.OrgName.ValueString(), plan.TeamName.ValueString(), plan.UserName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrgTeamMemberAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgTeamMemberAssociationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteOrgTeamMember(ctx, data.OrgName.ValueString(), data.TeamName.ValueString(), data.UserName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete team member", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *OrgTeamMemberAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
