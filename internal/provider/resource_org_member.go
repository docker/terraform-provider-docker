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
	"log"
	"strings"
	"sync"

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
	_ resource.Resource                = &OrgMemberResource{}
	_ resource.ResourceWithConfigure   = &OrgMemberResource{}
	_ resource.ResourceWithImportState = &OrgMemberResource{}
)

func NewOrgMemberResource() resource.Resource {
	return &OrgMemberResource{}
}

type OrgMemberResource struct {
	client *hubclient.Client

	mu         sync.Mutex
	orgMembers map[string][]hubclient.OrgMember
	orgInvites map[string][]hubclient.OrgInvite
}

type OrgMemberResourceModel struct {
	OrgName  types.String `tfsdk:"org_name"`
	UserName types.String `tfsdk:"user_name"`
	Email    types.String `tfsdk:"email"`
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

	r.orgMembers = make(map[string][]hubclient.OrgMember)
	r.orgInvites = make(map[string][]hubclient.OrgInvite)
	r.client = client
}

func (r *OrgMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_member"
}

func (r *OrgMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages members associated with an organization.

-> **Note** Only available when authenticated with a username and password as an owner of the org.

When a member is added to an organization, they don't have access to the
organization's repositories until they accept the invitation. The invitation is
sent to the email address associated with the user's Docker ID.

## Example Usage

` + "```hcl" + `
resource "docker_org_member" "example" {
	org_name = "org_name"
	role     = "member"
	email    = "orgmember@docker.com"
}
` + "```" + `

## Import State

` + "```hcl" + `

import {
  id = "org-name/user-name"
  to = docker_org_member.example
}

resource "docker_org_member" "example" {
	org_name  = "org-name"
	role      = "member"
	user_name = "user-name"
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
			"user_name": schema.StringAttribute{
				MarkdownDescription: "User name of the member. Either user_name or email must be specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the member. Either user_name or email must be specified.",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "The ID of the invite. Used for managing membership invites that haven't been accepted yet.",
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

	invitee := data.UserName.ValueString()
	if invitee == "" {
		invitee = data.Email.ValueString()
	}
	if invitee == "" {
		errMsg := "Either user_name or email must be specified."
		resp.Diagnostics.AddError("Missing Required Field", errMsg)
		return
	}

	current, found, err := r.orgMember(ctx, data.OrgName.ValueString(), invitee)
	if err != nil {
		resp.Diagnostics.AddError("Error Checking Existing Member", err.Error())
		return
	}

	if found {
		// Set computed fields.
		data.InviteID = current.InviteID
		data.Email = current.Email
	} else {
		// If the member is not found, invite them now.
		inviteResp, err := r.client.InviteOrgMember(ctx,
			data.OrgName.ValueString(), data.Role.ValueString(), []string{invitee}, false)
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
		if invite.Invite.ID != "" {
			data.InviteID = types.StringValue(invite.Invite.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrgMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invitee := state.UserName.ValueString()
	if invitee == "" {
		invitee = state.Email.ValueString()
	}
	if invitee == "" {
		errMsg := "Either user_name or email must be specified."
		resp.Diagnostics.AddError("Missing Required Field", errMsg)
		return
	}

	data, found, err := r.orgMember(ctx, state.OrgName.ValueString(), invitee)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Resource", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Resource Not Found",
			fmt.Sprintf("Member not found in %s: %s", state.OrgName.ValueString(), invitee))
	}

	if data.Role.ValueString() == "" {
		data.Role = state.Role
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// NOTE(nicks): currently, we treat all fields as immutable,
// so in theory update should never happen.
//
// Future work: update the provider to allow role changes.
func (r *OrgMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

func (r *OrgMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invitee := data.UserName.ValueString()
	if invitee == "" {
		invitee = data.Email.ValueString()
	}
	if invitee == "" {
		errMsg := "Either user_name or email must be specified."
		resp.Diagnostics.AddError("Missing Required Field", errMsg)
		return
	}

	// Deleting an established member (accepted inv) vs deleting an invited member has different API calls
	// Invited members that have not accepted do not have a recorded username in the org afaik
	// Attempt to delete by inviteID first
	deleted := false
	if data.InviteID.ValueString() != "" {
		err := r.client.DeleteOrgInvite(ctx, data.InviteID.ValueString())
		if err == nil {
			deleted = true
			return
		}
	}

	if !deleted {
		// If deleting by inviteID fails, try deleting by orgName and userName
		err := r.client.DeleteOrgMember(ctx, data.OrgName.ValueString(), invitee)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to delete org_member resource: %v", err)
			resp.Diagnostics.AddError("Error Deleting Resource", errMsg)
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Println("Successfully deleted OrgMemberResource.")
}

func (r *OrgMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_name/user_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_name"), idParts[1])...)
}

// orgMember returns the org member resource model for the given org and user name.
//
// Our org_member represents both invited and accepted members, so
// we need to merge the results from both endpoints.
//
// Returns true if the member was found, false if not found.
func (r *OrgMemberResource) orgMember(ctx context.Context, orgName string, userName string) (OrgMemberResourceModel, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	members, ok := r.orgMembers[orgName]
	if !ok {
		var err error
		members, err = r.client.ListOrgMembers(ctx, orgName)
		if err != nil {
			return OrgMemberResourceModel{}, false, err
		}
		r.orgMembers[orgName] = members
	}

	for _, member := range members {
		if member.Username == userName || member.Email == userName {
			return OrgMemberResourceModel{
				OrgName:  types.StringValue(orgName),
				UserName: types.StringValue(member.Username),
				Role:     types.StringValue(strings.ToLower(member.Role)),
				Email:    types.StringValue(member.Email),
			}, true, nil
		}
	}

	invites, ok := r.orgInvites[orgName]
	if !ok {
		var err error
		invites, err = r.client.ListOrgInvites(ctx, orgName)
		if err != nil {
			return OrgMemberResourceModel{}, false, err
		}
		r.orgInvites[orgName] = invites
	}

	for _, invite := range invites {
		if userName == invite.Invitee {
			result := OrgMemberResourceModel{
				OrgName:  types.StringValue(orgName),
				InviteID: types.StringValue(invite.ID),
			}
			if strings.Contains(userName, "@") {
				result.Email = types.StringValue(userName)
			} else {
				result.UserName = types.StringValue(userName)
			}
			return result, true, nil
		}
	}

	return OrgMemberResourceModel{}, false, nil
}
