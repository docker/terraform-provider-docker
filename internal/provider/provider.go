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
	"os"
	"regexp"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/docker/terraform-provider-docker/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var hostRegexp = regexp.MustCompile(`^[a-zA-Z0-9:.-]+$`)

const dockerHubConfigfileKey = "https://index.docker.io/v1/"
const dockerHubStageConfigfileKey = "index-stage.docker.io"
const dockerHubHost = "hub.docker.com"
const dockerHubStageHost = "hub-stage.docker.com"

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
	Host     types.String `tfsdk:"host"`
}

func (p *DockerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "docker"
	resp.Version = p.version
}

func (p *DockerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manage Docker-hosted resources (such as repositories,
teams, organization settings, and more) using Terraform.

> [!WARNING]
> This project is **not** for managing objects in a local docker engine. If you would like to use Terraform to interact with a docker engine, [kreuzwerker/docker](https://registry.terraform.io/providers/kreuzwerker/docker/latest) is a fine provider.

## Usage

Below is a basic example of how to use the Docker services Terraform provider to create a Docker repository.

` + "```" + `hcl
terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~> 0.2"
    }
  }
}

provider "docker" { }

resource "docker_repository" "example" {
  name        = "example-repo"
  description = "This is an example Docker repository"
  private     = true
}
` + "```" + `


## Authentication

We have multiple ways to set your Docker credentials.

### Setting credentials

Use ` + "`docker login`" + ` to [log in to a
registry](https://docs.docker.com/reference/cli/docker/login/). The ` + "`docker`" + ` CLI
will store your credentials securely in your credential store, such as the
operating system native keychain. The Docker Terraform provider will
use these credentials automatically.

` + "```" + `
cat ~/my_password.txt | docker login --username my-username --password-stdin
` + "```" + `

If you'd like to use a different account for running the provider,
you can set credentials in the environment:

` + "```" + `
export DOCKER_USERNAME=my-username
export DOCKER_PASSWORD=my-secret-token
terraform plan ...
` + "```" + `

### Credential types

You can create a personal access token (PAT) to use as an alternative to your
password for Docker CLI authentication.

A "Read, Write, & Delete" PAT can be used to create, edit, and
manage permissions for Docker Hub repositories.

The advantage of PATs is that they have [many security
benefits](https://docs.docker.com/security/for-developers/access-tokens/) over
passwords.

Unfortunately, PATs are limited to managing repositories. If you'd like to use
this provider to manage organizations and teams, you will need to authenticate
`,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Docker Hub API Host. Default is `hub.docker.com`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(hostRegexp, "Must be a valid host"),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication",
				Optional:            true,
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
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOCKER_HUB_HOST environment variable.",
		)
	}

	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Docker Hub API Username",
			"The provider cannot create the Docker Hub API client as there is an unknown configuration value for the Docker Hub API username.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	host := os.Getenv("DOCKER_HUB_HOST")
	if host == "" {
		host = "hub.docker.com"
	}
	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	username := os.Getenv("DOCKER_USERNAME")
	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	password := os.Getenv("DOCKER_PASSWORD")

	// If DOCKER_USERNAME and DOCKER_PASSWORD are not set, or if they are empty,
	// retrieve them from the credential store
	if username == "" || password == "" {
		// Loosely adapted from
		// https://github.com/moby/buildkit/blob/b9a3e7b31958b83f9ab1850a8c2ab1c66bf21f1f/session/auth/authprovider/authprovider.go#L243
		//
		// The Docker Hub host is a special case
		// that stores its credentials differently in the store.
		configfileKey := host
		if host == dockerHubHost {
			configfileKey = dockerHubConfigfileKey
		} else if host == dockerHubStageHost {
			configfileKey = dockerHubStageConfigfileKey
		}

		// Use the getUserCreds function to retrieve credentials from Docker config
		var err error
		username, password, err = tools.GetUserCreds(configfileKey)
		if err != nil {
			resp.Diagnostics.AddError("Credential Store Error",
				fmt.Sprintf("Failed to retrieve credentials from the Docker config file: %v", err))
		}
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Docker Hub API Host",
			"The provider cannot create the Docker Hub API client as there is a missing or empty value for the Docker Hub API host. "+
				"Set the host value in the configuration or use the DOCKER_HUB_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	} else if !hostRegexp.MatchString(host) {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Invalid Docker Hub API Host",
			"DOCKER_HUB_HOST must be a valid host (of the form 'hub.docker.com').")
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Docker Hub API Username",
			"Missing valid login credentials. More details: https://github.com/docker/terraform-provider-docker#authentication.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Docker Hub API Password",
			"Missing valid login credentials. More details: https://github.com/docker/terraform-provider-docker#authentication.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "docker_hub_host", host)
	ctx = tflog.SetField(ctx, "docker_username", username)
	ctx = tflog.SetField(ctx, "docker_password", password)

	tflog.Debug(ctx, "Creating Docker Hub client")

	client := hubclient.NewClient(hubclient.Config{
		BaseURL:          fmt.Sprintf("https://%s/v2", host),
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
		NewOrgTeamMemberResource,
		NewRepositoryResource,
		NewRepositoryTeamPermissionResource,
		NewOrgMemberResource,
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
		NewOrgTeamDataSource,
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
