package provider

import (
	"context"
	"fmt"

	"github.com/docker/terraform-provider-dockerhub/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &AccessTokensDataSource{}
	_ datasource.DataSourceWithConfigure = &AccessTokensDataSource{}
)

func NewAccessTokensDataSource() datasource.DataSource {
	return &AccessTokensDataSource{}
}

type AccessTokensDataSource struct {
	client *hubclient.Client
}

type AccessTokensDataSourceModel struct {
	UUIDs types.List `tfsdk:"uuids"`
}

func (d *AccessTokensDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_tokens"
}

func (d *AccessTokensDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Docker Hub Access Token",

		Attributes: map[string]schema.Attribute{
			"uuids": schema.ListAttribute{
				MarkdownDescription: "The UUIDs of the access tokens",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *AccessTokensDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccessTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccessTokensDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	atPage, err := d.client.GetAccessTokens(ctx, hubclient.AccessTokenListParams{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading access token list", fmt.Sprintf("%v", err))
		return
	}

	uuids := []string{}
	for _, at := range atPage.Results {
		uuids = append(uuids, at.UUID)
	}
	data.UUIDs, _ = types.ListValueFrom(ctx, types.StringType, uuids)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
