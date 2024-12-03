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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &OrgMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgMembersDataSource{}
)

func NewOrgMembersDataSource() datasource.DataSource {
	return &OrgMembersDataSource{}
}

type OrgMembersDataSource struct {
	client *hubclient.Client
}

type OrgMembersDataSourceModel struct {
	OrgName types.String `tfsdk:"org_name"`
	Members []OrgMember  `tfsdk:"members"`
}

type OrgMember struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	FullName      types.String `tfsdk:"full_name"`
	Location      types.String `tfsdk:"location"`
	Company       types.String `tfsdk:"company"`
	ProfileURL    types.String `tfsdk:"profile_url"`
	DateJoined    types.String `tfsdk:"date_joined"`
	GravatarURL   types.String `tfsdk:"gravatar_url"`
	GravatarEmail types.String `tfsdk:"gravatar_email"`
	Type          types.String `tfsdk:"type"`
	Email         types.String `tfsdk:"email"`
	Role          types.String `tfsdk:"role"`
	Groups        types.List   `tfsdk:"groups"`
	IsGuest       types.Bool   `tfsdk:"is_guest"`
}

func (d *OrgMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_members"
}

func (d *OrgMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Reads members of an organization.

~> **Note** Only available when authenticated with a username and password.

## Example Usage

` + "```hcl" + `
data "docker_org_members" "_" {
  org_name  = "my-org"
}

output "usernames" {
  value = [for member in data.docker_org_members._.members: member.username]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "List of members",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"username": schema.StringAttribute{
							Computed: true,
						},
						"full_name": schema.StringAttribute{
							Computed: true,
						},
						"location": schema.StringAttribute{
							Computed: true,
						},
						"company": schema.StringAttribute{
							Computed: true,
						},
						"profile_url": schema.StringAttribute{
							Computed: true,
						},
						"date_joined": schema.StringAttribute{
							Computed: true,
						},
						"gravatar_url": schema.StringAttribute{
							Computed: true,
						},
						"gravatar_email": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
						"groups": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"is_guest": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *OrgMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hubclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OrgMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgMembersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	members, err := d.client.ListOrgMembers(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading org members", fmt.Sprintf("%v", err))
		return
	}

	var memberList []OrgMember
	for _, member := range members {
		memberGroups := make([]attr.Value, len(member.Groups))
		for i, group := range member.Groups {
			memberGroups[i] = types.StringValue(group)
		}
		memberList = append(memberList, OrgMember{
			ID:            types.StringValue(member.ID),
			Username:      types.StringValue(member.Username),
			FullName:      types.StringValue(member.FullName),
			Location:      types.StringValue(member.Location),
			Company:       types.StringValue(member.Company),
			ProfileURL:    types.StringValue(member.ProfileURL),
			DateJoined:    types.StringValue(member.DateJoined),
			GravatarURL:   types.StringValue(member.GravatarURL),
			GravatarEmail: types.StringValue(member.GravatarEmail),
			Type:          types.StringValue(member.Type),
			Email:         types.StringValue(member.Email),
			Role:          types.StringValue(member.Role),
			Groups:        types.ListValueMust(types.StringType, memberGroups),
			IsGuest:       types.BoolValue(member.IsGuest),
		})
	}

	data.Members = memberList

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
