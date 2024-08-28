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
	_ datasource.DataSource              = &OrgDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgDataSource{}
)

func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

type OrgDataSource struct {
	client *hubclient.Client
}

type OrgDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	OrgName    types.String `tfsdk:"org_name"`
	FullName   types.String `tfsdk:"full_name"`
	Location   types.String `tfsdk:"location"`
	Company    types.String `tfsdk:"company"`
	DateJoined types.String `tfsdk:"date_joined"`
}

func (d *OrgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *OrgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Docker Hub Organization",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization",
				Computed:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization namespace",
				Required:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "Full name of the organization",
				Optional:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location of the organization",
				Optional:            true,
			},
			"company": schema.StringAttribute{
				MarkdownDescription: "Company name",
				Optional:            true,
			},
			"date_joined": schema.StringAttribute{
				MarkdownDescription: "Date the organization joined",
				Optional:            true,
			},
		},
	}
}

func (d *OrgDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	org, err := d.client.GetOrg(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading organization", fmt.Sprintf("%v", err))
		return
	}

	data.ID = types.StringValue(data.OrgName.ValueString())
	data.FullName = types.StringValue(org.FullName)
	data.Location = types.StringValue(org.Location)
	data.Company = types.StringValue(org.Company)
	data.DateJoined = types.StringValue(org.DateJoined)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
