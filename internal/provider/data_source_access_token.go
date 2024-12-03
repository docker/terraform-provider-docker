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
	_ datasource.DataSource              = &AccessTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &AccessTokenDataSource{}
)

func NewAccessTokenDataSource() datasource.DataSource {
	return &AccessTokenDataSource{}
}

type AccessTokenDataSource struct {
	client *hubclient.Client
}

type AccessTokenDataSourceModel struct {
	UUID        types.String `tfsdk:"uuid"`
	ClientID    types.String `tfsdk:"client_id"`
	CreatorIP   types.String `tfsdk:"creator_ip"`
	CreatorUA   types.String `tfsdk:"creator_ua"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUsed    types.String `tfsdk:"last_used"`
	GeneratedBy types.String `tfsdk:"generated_by"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Token       types.String `tfsdk:"token"`
	TokenLabel  types.String `tfsdk:"token_label"`
	Scopes      types.List   `tfsdk:"scopes"`
}

func (d *AccessTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_token"
}

func (d *AccessTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Retrieves details of a specific Docker Hub access token by its UUID.

		## Example Usage

		` + "```hcl" + `
		data "docker_hub_access_token" "example" {
			uuid = "123e4567-e89b-12d3-a456-426614174000"
		}

		output "access_token_details" {
		  value = {
			label       = data.docker_hub_access_token.example.token_label
			scopes      = data.docker_hub_access_token.example.scopes
			created_at  = data.docker_hub_access_token.example.created_at
			last_used   = data.docker_hub_access_token.example.last_used
			is_active   = data.docker_hub_access_token.example.is_active
		  }
		}
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the access token",
				Required:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The client ID of the access token",
				Optional:            true,
			},
			"creator_ip": schema.StringAttribute{
				MarkdownDescription: "The IP address of the creator of the access token",
				Optional:            true,
			},
			"creator_ua": schema.StringAttribute{
				MarkdownDescription: "The user agent of the creator of the access token",
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation time of the access token",
				Optional:            true,
			},
			"last_used": schema.StringAttribute{
				MarkdownDescription: "The last time the access token was used",
				Optional:            true,
			},
			"generated_by": schema.StringAttribute{
				MarkdownDescription: "The user who generated the access token",
				Optional:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the access token is active",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The access token",
				Optional:            true,
			},
			"token_label": schema.StringAttribute{
				MarkdownDescription: "The label of the access token",
				Optional:            true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "The scopes of the access token",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *AccessTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccessTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccessTokenDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	at, err := d.client.GetAccessToken(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading repository", fmt.Sprintf("%v", err))
		return
	}

	data.ClientID = types.StringValue(at.ClientID)
	data.CreatorIP = types.StringValue(at.CreatorIP)
	data.CreatorUA = types.StringValue(at.CreatorUA)
	data.CreatedAt = types.StringValue(at.CreatedAt)
	data.LastUsed = types.StringValue(at.LastUsed)
	data.GeneratedBy = types.StringValue(at.GeneratedBy)
	data.IsActive = types.BoolValue(at.IsActive)
	data.Token = types.StringValue(at.Token)
	data.TokenLabel = types.StringValue(at.TokenLabel)
	data.Scopes, _ = types.ListValueFrom(ctx, types.StringType, at.Scopes)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
