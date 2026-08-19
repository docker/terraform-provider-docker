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
	"strings"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &RepositoryImmutableTagsResource{}
	_ resource.ResourceWithConfigure   = &RepositoryImmutableTagsResource{}
	_ resource.ResourceWithImportState = &RepositoryImmutableTagsResource{}
)

func NewRepositoryImmutableTagsResource() resource.Resource {
	return &RepositoryImmutableTagsResource{}
}

type RepositoryImmutableTagsResource struct {
	client *hubclient.Client
}

type RepositoryImmutableTagsResourceModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Rules     types.List   `tfsdk:"rules"`
}

func (r *RepositoryImmutableTagsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*hubclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hubclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *RepositoryImmutableTagsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_repository_immutable_tags"
}

func (r *RepositoryImmutableTagsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages tag immutability settings for a Docker Hub repository.

When enabled, tags matching the specified rules cannot be overwritten or deleted. Up to 5 rules (Go-compatible regular expressions) are supported. Rules default to ` + "`[\".*\"]`" + ` (all tags immutable) when none are specified.

This resource can be used independently of ` + "`docker_hub_repository`" + ` to manage immutable tag settings on repositories that were not created by Terraform.

> **Note:** Requires a Personal Access Token (PAT) with the ` + "`repo:admin`" + ` scope, or an OAT with the ` + "`scope-repository-settings-admin`" + ` permission.

## Example Usage

` + "```hcl" + `
resource "docker_hub_repository_immutable_tags" "example" {
  namespace = "my-organization"
  name      = "my-repo"
  enabled   = true
  rules     = ["latest", "v[0-9]+\\.[0-9]+.*"]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Repository namespace (organization or username)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Repository name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether tag immutability is enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"rules": schema.ListAttribute{
				MarkdownDescription: "List of Go-compatible regular expressions defining which tags are immutable. Up to 5 rules. Defaults to `[\".*\"]` (all tags) when enabled and no rules are specified.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 5),
					listvalidator.ValueStringsAre(
						stringvalidator.NoneOf(","),
					),
				},
			},
		},
	}
}

func (r *RepositoryImmutableTagsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepositoryImmutableTagsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateRepositoryImmutableTags(ctx, data.Namespace.ValueString(), data.Name.ValueString(), buildImmutableTagsRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create hub_repository_immutable_tags resource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readImmutableTagsModel(data.Namespace.ValueString(), data.Name.ValueString(), result))...)
}

func (r *RepositoryImmutableTagsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepositoryImmutableTagsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetRepository(ctx, data.Namespace.ValueString()+"/"+data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read hub_repository_immutable_tags resource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readImmutableTagsModel(data.Namespace.ValueString(), data.Name.ValueString(), result))...)
}

func (r *RepositoryImmutableTagsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepositoryImmutableTagsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateRepositoryImmutableTags(ctx, data.Namespace.ValueString(), data.Name.ValueString(), buildImmutableTagsRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update hub_repository_immutable_tags resource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readImmutableTagsModel(data.Namespace.ValueString(), data.Name.ValueString(), result))...)
}

func (r *RepositoryImmutableTagsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepositoryImmutableTagsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disable immutable tags; preserve existing rules on the server side.
	_, err := r.client.UpdateRepositoryImmutableTags(ctx, data.Namespace.ValueString(), data.Name.ValueString(),
		hubclient.UpdateRepositoryImmutableTagsRequest{ImmutableTags: false})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete hub_repository_immutable_tags resource", err.Error())
	}
}

func (r *RepositoryImmutableTagsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format 'namespace/name', got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func buildImmutableTagsRequest(data RepositoryImmutableTagsResourceModel) hubclient.UpdateRepositoryImmutableTagsRequest {
	req := hubclient.UpdateRepositoryImmutableTagsRequest{
		ImmutableTags: data.Enabled.ValueBool(),
	}
	if data.Enabled.ValueBool() && !data.Rules.IsNull() && !data.Rules.IsUnknown() {
		var rules []string
		data.Rules.ElementsAs(context.Background(), &rules, false)
		req.ImmutableTagsRules = rules
	}
	return req
}

func readImmutableTagsModel(namespace, name string, repo hubclient.Repository) RepositoryImmutableTagsResourceModel {
	settings := repo.ImmutableTagsSettings
	model := RepositoryImmutableTagsResourceModel{
		Namespace: types.StringValue(namespace),
		Name:      types.StringValue(name),
		Enabled:   types.BoolValue(settings.Enabled),
	}
	if settings.Enabled && len(settings.Rules) > 0 {
		model.Rules = typesListFromStringSlice(settings.Rules)
	} else {
		model.Rules = types.ListValueMust(types.StringType, nil)
	}
	return model
}
