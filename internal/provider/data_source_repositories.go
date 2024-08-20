package provider

import (
	"context"
	"fmt"

	"github.com/docker/terraform-provider-docker/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &RepositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoriesDataSource{}
)

func NewRepositoriesDataSource() datasource.DataSource {
	return &RepositoriesDataSource{}
}

type RepositoriesDataSource struct {
	client *hubclient.Client
}

type RepositoriesDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Namespace        types.String `tfsdk:"namespace"`
	MaxNumberResults types.Int64  `tfsdk:"max_number_results"`
	Repository       []Repository `tfsdk:"repository"`
}

type Repository struct {
	Name        types.String `tfsdk:"name"`
	Namespace   types.String `tfsdk:"namespace"`
	Description types.String `tfsdk:"description"`
	IsPrivate   types.Bool   `tfsdk:"is_private"`
	PullCount   types.Int64  `tfsdk:"pull_count"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Affiliation types.String `tfsdk:"affiliation"`
}

func (d *RepositoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *RepositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Docker Hub Repositories",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The namespace/name of the repository",
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Repository namespace",
				Required:            true,
			},
			"max_number_results": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of results",
				Optional:            true,
				// Default:             100,
			},
			"repository": schema.ListNestedAttribute{
				MarkdownDescription: "List of repositories",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"is_private": schema.BoolAttribute{
							Computed: true,
						},
						"pull_count": schema.Int64Attribute{
							Computed: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"affiliation": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *RepositoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoriesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	repositories, err := d.client.GetRepositories(ctx, data.Namespace.ValueString(), int(data.MaxNumberResults.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading repositories", fmt.Sprintf("%v", err))
		return
	}

	var repoList []Repository
	for _, repo := range repositories.Results {
		repoList = append(repoList, Repository{
			Name:        types.StringValue(repo.Name),
			Namespace:   types.StringValue(repo.Namespace),
			Description: types.StringValue(repo.Description),
			IsPrivate:   types.BoolValue(repo.IsPrivate),
			PullCount:   types.Int64Value(repo.PullCount),
			LastUpdated: types.StringValue(repo.LastUpdated),
			Affiliation: types.StringValue(repo.Affiliation),
		})
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/repositories", data.Namespace.ValueString()))
	data.Repository = repoList

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
