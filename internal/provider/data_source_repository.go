package provider

import (
	"context"
	"fmt"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/docker/terraform-provider-docker/internal/repositoryutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &RepositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoryDataSource{}
)

func NewRepositoryDataSource() datasource.DataSource {
	return &RepositoryDataSource{}
}

type RepositoryDataSource struct {
	client *hubclient.Client
}

type RepositoryDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Namespace       types.String `tfsdk:"namespace"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	FullDescription types.String `tfsdk:"full_description"`
	Private         types.Bool   `tfsdk:"private"`
	PullCount       types.Int64  `tfsdk:"pull_count"`
}

func (d *RepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_repository"
}

func (d *RepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The namespace/name of the repository",
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Repository namespace",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Repository name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Repository description",
				Required:            false,
				Optional:            true,
			},
			"full_description": schema.StringAttribute{
				MarkdownDescription: "Repository name",
				Required:            false,
				Optional:            true,
			},
			"private": schema.BoolAttribute{
				MarkdownDescription: "Is the repository private",
				Required:            false,
				Optional:            true,
			},
			"pull_count": schema.Int64Attribute{
				Required: false,
				Optional: true,
			},
		},
	}
}

func (d *RepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *RepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	id := repositoryutils.NewID(data.Namespace.ValueString(), data.Name.ValueString())

	repository, err := d.client.GetRepository(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading repository", "Could not read repository, unexpected error: "+err.Error())
		return
	}

	data.ID = types.StringValue(id)
	data.Namespace = types.StringValue(repository.Namespace)
	data.Name = types.StringValue(repository.Name)
	data.Description = types.StringValue(repository.Description)
	data.FullDescription = types.StringValue(repository.FullDescription)
	data.Private = types.BoolValue(repository.IsPrivate)
	data.PullCount = types.Int64Value(repository.PullCount)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
