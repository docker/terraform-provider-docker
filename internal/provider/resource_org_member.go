package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgMemberResource{}
	_ resource.ResourceWithConfigure   = &OrgMemberResource{}
	_ resource.ResourceWithImportState = &OrgMemberResource{}
)

func NewOrgMemberResource() resource.Resource {
	return &OrgMemberResource{}
}

type OrgMemberResource struct {
	client *hubclient.Client
}

type OrgMemberResourceModel struct {
	OrgName  types.String `tfsdk:"org_name"`
	TeamName types.String `tfsdk:"team_name"`
	UserName types.String `tfsdk:"user_name"`
	Role     types.String `tfsdk:"role"`      // New field for role
	InviteID types.String `tfsdk:"invite_id"` // This is needed for deletion
}

func (r *OrgMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)
	if !ok {
		errMsg := fmt.Sprintf("Expected *hubclient.Client, got: %T", req.ProviderData)
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", errMsg)
		return
	}

	r.client = client
}

func (r *OrgMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_member"
}

func (r *OrgMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the  of a member with a team in an organization.",

		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				MarkdownDescription: "Team name within the organization",
				Required:            false,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "User name (email) of the member being associated with the team",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role assigned to the user within the organization (e.g., 'member', 'editor', 'owner').",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("member", "editor", "owner"),
				},
			},
			"invite_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invite. Used for managing the , especially for deletion.",
				Computed:            true,
			},
		},
	}
}

func (r *OrgMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var data OrgMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inviteResp, err := r.client.InviteOrgMember(ctx, data.OrgName.ValueString(), data.TeamName.ValueString(), data.Role.ValueString(), []string{data.UserName.ValueString()}, false)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to create org_member resource: %v", err)
		resp.Diagnostics.AddError("Error Creating Resource", errMsg)
		return
	}

	if len(inviteResp.OrgInvitees) == 0 {
		errMsg := "No invitees were returned from the Docker Hub API."
		resp.Diagnostics.AddError("Invite Failed", errMsg)
		return
	}

	invite := inviteResp.OrgInvitees[0]
	data.InviteID = types.StringValue(invite.Invite.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: finish read
func (r *OrgMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data OrgMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// TODO: setup update
func (r *OrgMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

func (r *OrgMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Deleting an established member (accepted inv) vs deleting an invited member has different API calls
	// Invited members that have not accepted do not have a recorded username in the org afaik
	// Attempt to delete by inviteID first
	err := r.client.DeleteOrgInvite(ctx, data.InviteID.ValueString())
	if err != nil {
		// If deleting by inviteID fails, try deleting by orgName and userName
		err = r.client.DeleteOrgMember(ctx, data.OrgName.ValueString(), data.UserName.ValueString())
		if err != nil {
			errMsg := fmt.Sprintf("Unable to delete org_member resource: %v", err)
			log.Println(errMsg)
			resp.Diagnostics.AddError("Error Deleting Resource", errMsg)
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Println("Successfully deleted OrgMemberResource.")
}

// TODO: setup import state
func (r *OrgMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	return
}
