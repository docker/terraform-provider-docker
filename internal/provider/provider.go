package provider

import (
	"context"
	"os"

	"github.com/docker/terraform-provider-docker/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure DockerProvider satisfies various provider interfaces.
var (
	_ provider.Provider              = &DockerProvider{}
	_ provider.ProviderWithFunctions = &DockerProvider{}
)

// DockerProvider defines the provider implementation.
type DockerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DockerProviderModel describes the provider data model.
type DockerProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Host     types.String `tfsdk:"host"`
}

func (p *DockerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "docker"
	resp.Version = p.version
}

func (p *DockerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Docker Hub host URL",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *DockerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Docker Hub client")

	var data DockerProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Docker Hub API Host",
			"The provider cannot create the Docker Hub API client as there is an unknown configuration value for the Docker Hub API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOCKER_HOST environment variable.",
		)
	}

	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Docker Hub API Username",
			"The provider cannot create the Docker Hub API client as there is an unknown configuration value for the Docker Hub API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOCKER_USERNAME environment variable.",
		)
	}

	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Docker Hub API Password",
			"The provider cannot create the Docker Hub API client as there is an unknown configuration value for the Docker Hub API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOCKER_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		// once this is ready for the lime-light, we should default this to prod
		// host = "https://hub.docker.com/v2"
		host = "https://hub-stage.docker.com/v2"
	}
	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	username := os.Getenv("DOCKER_USERNAME")
	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	password := os.Getenv("DOCKER_PASSWORD")
	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Docker Hub API Host",
			"The provider cannot create the Docker Hub API client as there is a missing or empty value for the Docker Hub API host. "+
				"Set the host value in the configuration or use the DOCKER_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Docker Hub API Username",
			"The provider cannot create the Docker Hub API client as there is a missing or empty value for the Docker Hub API username. "+
				"Set the username value in the configuration or use the DOCKER_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Docker Hub API Password",
			"The provider cannot create the Docker Hub API client as there is a missing or empty value for the Docker Hub API password. "+
				"Set the password value in the configuration or use the DOCKER_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "docker_host", host)
	ctx = tflog.SetField(ctx, "docker_username", username)
	ctx = tflog.SetField(ctx, "docker_password", password)

	tflog.Debug(ctx, "Creating Docker Hub client")

	client := hubclient.NewClient(hubclient.Config{
		Host:             host,
		Username:         username,
		Password:         password,
		UserAgentVersion: p.version,
	})
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DockerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessTokenResource,
		NewOrgSettingImageAccessManagementResource,
		NewOrgSettingRegistryAccessManagementResource,
		NewOrgTeamResource,
		NewOrgTeamMemberAssociationResource,
		NewRepositoryResource,
		NewRepositoryTeamPermissionResource,
	}
}

func (p *DockerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrgDataSource,
		NewOrgTeamMemberDataSource,
		NewRepositoryDataSource,
		NewRepositoriesDataSource,
		NewAccessTokenDataSource,
		NewAccessTokensDataSource,
	}
}

func (p *DockerProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DockerProvider{
			version: version,
		}
	}
}
