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
	"log"
	"regexp"
	"strings"

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/docker/terraform-provider-docker/internal/repositoryutils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &RepositoryResource{}
	_ resource.ResourceWithConfigure = &RepositoryResource{}
)

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

type RepositoryResource struct {
	client *hubclient.Client
}

type ImmutableTagsSettings struct {
	Enabled types.Bool `tfsdk:"enabled"`
	Rules   types.List `tfsdk:"rules"`
}

type RepositoryResourceModel struct {
	ID                    types.String           `tfsdk:"id"`
	Namespace             types.String           `tfsdk:"namespace"`
	Name                  types.String           `tfsdk:"name"`
	Description           types.String           `tfsdk:"description"`
	FullDescription       types.String           `tfsdk:"full_description"`
	Private               types.Bool             `tfsdk:"private"`
	PullCount             types.Int64            `tfsdk:"pull_count"`
	ImmutableTagsSettings *ImmutableTagsSettings `tfsdk:"immutable_tags_settings"`
}

func immutableTagsSettingsSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Immutable tags settings for the repository",
		Required:            false,
		Optional:            true,
		Computed:            true,
		Default: objectdefault.StaticValue(
			types.ObjectValueMust(
				map[string]attr.Type{
					"enabled": types.BoolType,
					"rules":   types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"enabled": types.BoolValue(false),
					"rules":   types.ListValueMust(types.StringType, nil),
				})),
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether immutable tags are enabled for the repository",
				Required:            false,
				Optional:            true,
			},
			"rules": schema.ListAttribute{
				MarkdownDescription: "List of immutable tag rules for the repository",
				Required:            false,
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, nil)),
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *RepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Create implements resource.Resource.
func (r *RepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep plan and data separate so we can record partial state.
	var data RepositoryResourceModel
	diags = req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := hubclient.CreateRepositoryRequest{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		FullDescription: plan.FullDescription.ValueString(),
		IsPrivate:       plan.Private.ValueBool(),
	}
	createResp, err := r.client.CreateRepository(ctx, plan.Namespace.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error creating repository", "Could not create repository, unexpected error: "+err.Error())
		return
	}

	id := repositoryutils.NewID(createResp.Namespace, createResp.Name)
	data.ID = types.StringValue(id)
	data.Name = types.StringValue(createResp.Name)
	data.Namespace = types.StringValue(createResp.Namespace)
	data.Description = stringNullIfEmpty(createResp.Description)
	data.FullDescription = stringNullIfEmpty(createResp.FullDescription)
	data.Private = types.BoolValue(createResp.IsPrivate)
	data.PullCount = types.Int64Value(createResp.PullCount)
	data.ImmutableTagsSettings = deserializeImmutableTagsSettings(false, nil)

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE(nicks): Immutable settings cannot be set on creation, so we handle
	// them separately as an update request.
	//
	// If the update requests fail, we save empty settings.
	immutableTagsEnabled := serializeImmutableTagsEnabled(plan)
	if immutableTagsEnabled {
		immutableTagsReq := hubclient.UpdateRepositoryRequest{
			Description:        plan.Description.ValueString(),
			FullDescription:    plan.FullDescription.ValueString(),
			ImmutableTags:      immutableTagsEnabled,
			ImmutableTagsRules: serializeImmutableTagsRules(plan),
		}
		id := repositoryutils.NewID(createResp.Namespace, createResp.Name)
		updateResult, err := r.client.UpdateRepository(ctx, id, immutableTagsReq)
		if err != nil {
			resp.Diagnostics.AddError("Docker Hub API error setting immutable tags",
				"Could not set immutable tags, unexpected error: "+err.Error())
			return
		}

		data.ImmutableTagsSettings = deserializeImmutableTagsSettings(
			true, updateResult.ImmutableTagsSettings.Rules)
	}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Println("Error occurred while fetching the state")
		return
	}

	err := r.client.DeleteRepository(ctx, state.ID.ValueString())
	if err != nil {
		log.Printf("Failed to delete repository with ID: %s, error: %v", state.ID.ValueString(), err)
		resp.Diagnostics.AddError("Docker Hub API error deleting repository", "Could not delete repository, unexpected error: "+err.Error())
		return
	}
}

// Metadata implements resource.Resource.
func (r *RepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_repository"
}

// Read implements resource.Resource.
func (r *RepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.client.GetRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Docker Hub repository",
			"Could not read Docker Hub repository "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(repositoryutils.NewID(getResp.Namespace, getResp.Name))
	state.Name = types.StringValue(getResp.Name)
	state.Namespace = types.StringValue(getResp.Namespace)
	state.Description = stringNullIfEmpty(getResp.Description)
	state.FullDescription = stringNullIfEmpty(getResp.FullDescription)
	state.Private = types.BoolValue(getResp.IsPrivate)
	state.PullCount = types.Int64Value(getResp.PullCount)

	state.ImmutableTagsSettings = deserializeImmutableTagsSettings(
		getResp.ImmutableTagsSettings.Enabled,
		getResp.ImmutableTagsSettings.Rules)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements resource.Resource.
func (r *RepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages an image repository in your account or organization.

-> **Note** When used with a Personal Access Token authentication (PAT), the PAT should
have the "Read, Write, and Delete" scope to create and delete repositories. The
owner of the PAT must be an editor of the org.

## Example Usage

` + "```hcl" + `
resource "docker_hub_repository" "example" {
	namespace        = "my-organization"
	name             = "my-repo"
	description      = "A repository for storing container images"
	full_description = "This repository stores container images for the development team."
	private          = true
}
` + "```" + `

## Import

Use an import block to import a repository into your Terraform state, using
an id in the format of ` + "`<namespace>/<repository>`" + `.

` + "```hcl" + `
import {
  to = docker_hub_repository.docker-repo
  id = 'docker-namespace/docker-repo'
}
` + "```" + `

Or using the ` + "`terraform import`" + ` command:

` + "```bash" + `
terraform import docker_hub_repository.docker-repo docker-namespace/docker-repo
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The namespace/name of the repository",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Repository namespace",
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
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`),
						"Name must only contain lowercase alphanumeric characters, '.', or '-', and must start and end with a lowercase alphanumeric character",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Repository description",
				Required:            false,
				Optional:            true,
			},
			"full_description": schema.StringAttribute{
				MarkdownDescription: "Repository full description",
				Required:            false,
				Optional:            true,
			},
			"private": schema.BoolAttribute{
				MarkdownDescription: "Is the repository private",
				Required:            false,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"pull_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"immutable_tags_settings": immutableTagsSettingsSchema(),
		},
	}
}

// Update implements resource.Resource.
func (r *RepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RepositoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RepositoryResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := hubclient.UpdateRepositoryRequest{
		Description:        plan.Description.ValueString(),
		FullDescription:    plan.FullDescription.ValueString(),
		ImmutableTags:      serializeImmutableTagsEnabled(plan),
		ImmutableTagsRules: serializeImmutableTagsRules(plan),
	}

	updateResp, err := r.client.UpdateRepository(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error updating repository", "Could not update repository, unexpected error: "+err.Error())
		return
	}

	state.Description = stringNullIfEmpty(updateResp.Description)
	state.FullDescription = stringNullIfEmpty(updateResp.FullDescription)
	state.ImmutableTagsSettings = deserializeImmutableTagsSettings(
		updateResp.ImmutableTagsSettings.Enabled,
		updateResp.ImmutableTagsSettings.Rules)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Private.ValueBool() != state.Private.ValueBool() {
		err := r.client.SetRepositoryPrivacy(ctx, plan.ID.ValueString(), plan.Private.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Docker Hub API error setting repository privacy", "Could not set repository privacy, unexpected error: "+err.Error())
			return
		}

		state.Private = plan.Private
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func stringNullIfEmpty(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func serializeImmutableTagsEnabled(plan RepositoryResourceModel) bool {
	return plan.ImmutableTagsSettings.Enabled.ValueBool()
}

func serializeImmutableTagsRules(plan RepositoryResourceModel) string {
	enabled := serializeImmutableTagsEnabled(plan)
	if !enabled {
		return "" // If immutable tags are disabled, return an empty string.
	}
	rules := plan.ImmutableTagsSettings.Rules
	if len(rules.Elements()) == 0 {
		return ".*" // Default to all tags immutable.
	}
	return strings.Join(
		stringSliceFromTypesList(rules),
		hubclient.ImmutableTagRulesSeparator)
}

func deserializeImmutableTagsSettings(enabled bool, rules []string) *ImmutableTagsSettings {
	if !enabled {
		return &ImmutableTagsSettings{
			Enabled: types.BoolValue(false),
			Rules:   types.ListValueMust(types.StringType, nil),
		}
	}
	return &ImmutableTagsSettings{
		Enabled: types.BoolValue(true),
		Rules:   typesListFromStringSlice(rules),
	}
}

func stringSliceFromTypesList(typesStrings types.List) []string {
	result := make([]string, len(typesStrings.Elements()))
	for i, ts := range typesStrings.Elements() {
		result[i] = ts.(types.String).ValueString()
	}
	return result
}

func typesListFromStringSlice(strings []string) types.List {
	result := make([]attr.Value, len(strings))
	for i, s := range strings {
		result[i] = types.StringValue(s)
	}
	return types.ListValueMust(types.StringType, result)
}

func (r *RepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID passed during the import operation is expected to be in the form of "namespace/name"
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format 'namespace/name', got: %s", req.ID),
		)
		return
	}

	namespace := idParts[0]
	name := idParts[1]

	// Set the ID, namespace, and name in the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &RepositoryResourceModel{
		ID:        types.StringValue(req.ID),
		Namespace: types.StringValue(namespace),
		Name:      types.StringValue(name),

		// Set a default value to avoid type conversion problems.
		ImmutableTagsSettings: deserializeImmutableTagsSettings(false, nil),
	})...)
}
