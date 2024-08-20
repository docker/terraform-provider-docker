package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/docker/terraform-provider-docker/internal/pkg/hubclient"
	"github.com/docker/terraform-provider-docker/internal/pkg/repositoryutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type RepositoryResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Namespace       types.String `tfsdk:"namespace"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	FullDescription types.String `tfsdk:"full_description"`
	Private         types.Bool   `tfsdk:"private"`
	PullCount       types.Int64  `tfsdk:"pull_count"`
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

	createReq := hubclient.CreateRepostoryRequest{
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
	plan.ID = types.StringValue(id)
	plan.Name = types.StringValue(createResp.Name)
	plan.Namespace = types.StringValue(createResp.Namespace)
	plan.Description = stringNullIfEmpty(createResp.Description)
	plan.FullDescription = stringNullIfEmpty(createResp.FullDescription)
	plan.Private = types.BoolValue(createResp.IsPrivate)
	plan.PullCount = types.Int64Value(createResp.PullCount)

	diags = resp.State.Set(ctx, plan)
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
	resp.TypeName = req.ProviderTypeName + "_repository"
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements resource.Resource.
func (r *RepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
		Description:     plan.Description.ValueString(),
		FullDescription: plan.FullDescription.ValueString(),
	}

	updateResp, err := r.client.UpdateRepository(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error updating repository", "Could not update repository, unexpected error: "+err.Error())
		return
	}

	if plan.Private.ValueBool() != state.Private.ValueBool() {
		err := r.client.SetRepositoryPrivacy(ctx, plan.ID.ValueString(), plan.Private.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Docker Hub API error setting repository privacy", "Could not set repository privacy, unexpected error: "+err.Error())
			return
		}
	}

	plan.Description = stringNullIfEmpty(updateResp.Description)
	plan.FullDescription = stringNullIfEmpty(updateResp.FullDescription)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func stringNullIfEmpty(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
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
	})...)
}
