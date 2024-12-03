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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &OrgTeamDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgTeamDataSource{}
)

func NewOrgTeamDataSource() datasource.DataSource {
	return &OrgTeamDataSource{}
}

type OrgTeamDataSource struct {
	client *hubclient.Client
}

type OrgTeamDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	OrgName     types.String `tfsdk:"org_name"`
	TeamName    types.String `tfsdk:"team_name"`
	UUID        types.String `tfsdk:"uuid"`
	Description types.String `tfsdk:"description"`
	MemberCount types.Int64  `tfsdk:"member_count"`
}

func (d *OrgTeamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_team"
}

func (d *OrgTeamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Reads team information within a Docker Hub organization.

~> **Note**: This data source is only available when authenticated with a username and password.

## Example Usage

` + "```hcl" + `
data "docker_hub_org_team" "example" {
	org_name  = "my-organization"
	team_name = "dev-team"
}

output "team_info" {
value = {
	id           = data.docker_hub_org_team.example.id
	uuid         = data.docker_hub_org_team.example.uuid
	description  = data.docker_hub_org_team.example.description
	member_count = data.docker_hub_org_team.example.member_count
  }
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the team",
				Computed:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization namespace",
				Required:            true,
			},
			"team_name": schema.StringAttribute{
				MarkdownDescription: "Team name within the organization",
				Required:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the team",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the team",
				Computed:            true,
			},
			"member_count": schema.Int64Attribute{
				MarkdownDescription: "Number of members in the team",
				Computed:            true,
			},
		},
	}
}

func (d *OrgTeamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgTeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgTeamDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgTeam, err := d.client.GetOrgTeam(ctx, data.OrgName.ValueString(), data.TeamName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading organization team", fmt.Sprintf("%v", err))
		return
	}

	data.ID = types.Int64Value(int64(orgTeam.ID))
	data.UUID = types.StringValue(orgTeam.UUID)
	data.Description = types.StringValue(orgTeam.Description)
	data.MemberCount = types.Int64Value(int64(orgTeam.MemberCount))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
