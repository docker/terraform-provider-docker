package provider

import (
	"context"
	"fmt"

	"github.com/docker/terraform-provider-docker/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &OrgSettingRegistryAccessManagementResource{}
	_ resource.ResourceWithConfigure   = &OrgSettingRegistryAccessManagementResource{}
	_ resource.ResourceWithImportState = &OrgSettingRegistryAccessManagementResource{}
)

func NewOrgSettingRegistryAccessManagementResource() resource.Resource {
	return &OrgSettingRegistryAccessManagementResource{}
}

type OrgSettingRegistryAccessManagementResource struct {
	client *hubclient.Client
}

type OrgSettingRegistryAccessManagementResourceModel struct {
	OrgName                types.String `tfsdk:"org_name"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	StandardRegistryDocker types.Object `tfsdk:"standard_registry_docker_hub"`
	CustomRegistries       types.Set    `tfsdk:"custom_registries"`
}

type StandardRegistryModel struct {
	Allowed types.Bool `tfsdk:"allowed"`
}

var StandardRegistryObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"allowed": types.BoolType,
	},
}

type CustomRegistryModel struct {
	Address      types.String `tfsdk:"address"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	Allowed      types.Bool   `tfsdk:"allowed"`
}

var CustomRegistryObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"address":       types.StringType,
		"friendly_name": types.StringType,
		"allowed":       types.BoolType,
	},
}

func (r *OrgSettingRegistryAccessManagementResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hubclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OrgSettingRegistryAccessManagementResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_setting_registry_access_management"
}

func (r *OrgSettingRegistryAccessManagementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Registry Access Management settings for an organization.",

		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether or not Registry Access Management is enabled. When this feature is enabled, only registrys created by your organization or by Docker Official Registrys and Docker Verified Publishers are allowed. All community registrys are restricted.",
				Required:            true,
			},
			"standard_registry_docker_hub": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of Docker hub standard registry.⁠",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"allowed": schema.BoolAttribute{
						Required:            true,
						MarkdownDescription: "Whether or not to allow the standard registry.",
					},
				},
			},
			"custom_registries": schema.SetNestedAttribute{
				MarkdownDescription: "Configuration of custom registries⁠",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The address of the registry.",
						},
						"friendly_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The friendly name of the registry.",
						},
						"allowed": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether or not to allow the registry.",
						},
					},
				},
			},
		},
	}
}

func (r *OrgSettingRegistryAccessManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgSettingRegistryAccessManagementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reamReq := getOrgSettingRegistryAccessManagementRequest(ctx, data, resp.Diagnostics)
	if reamReq == nil || resp.Diagnostics.HasError() {
		return
	}

	reamResp, err := r.client.SetOrgSettingRegistryAccessManagement(ctx, data.OrgName.ValueString(), *reamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org_setting_registry_access_management resource", err.Error())
		return
	}

	model := getOrgSettingRegistryAccessManagementModel(ctx, reamResp, resp.Diagnostics)
	if model == nil || resp.Diagnostics.HasError() {
		return
	}

	data.Enabled = model.Enabled
	data.StandardRegistryDocker = model.StandardRegistryDocker
	data.CustomRegistries = model.CustomRegistries

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryAccessManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgSettingRegistryAccessManagementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	reamResp, err := r.client.GetOrgSettingRegistryAccessManagement(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_setting_registry_access_management resource", err.Error())
		return
	}

	model := getOrgSettingRegistryAccessManagementModel(ctx, reamResp, resp.Diagnostics)
	if model == nil || resp.Diagnostics.HasError() {
		return
	}

	data.Enabled = model.Enabled
	data.StandardRegistryDocker = model.StandardRegistryDocker
	data.CustomRegistries = model.CustomRegistries

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryAccessManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrgSettingRegistryAccessManagementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reamReq := getOrgSettingRegistryAccessManagementRequest(ctx, data, resp.Diagnostics)
	if reamReq == nil || resp.Diagnostics.HasError() {
		return
	}

	reamResp, err := r.client.SetOrgSettingRegistryAccessManagement(ctx, data.OrgName.ValueString(), *reamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org_setting_registry_access_management resource", err.Error())
		return
	}

	model := getOrgSettingRegistryAccessManagementModel(ctx, reamResp, resp.Diagnostics)
	if model == nil || resp.Diagnostics.HasError() {
		return
	}

	data.Enabled = model.Enabled
	data.StandardRegistryDocker = model.StandardRegistryDocker
	data.CustomRegistries = model.CustomRegistries

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingRegistryAccessManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgSettingRegistryAccessManagementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// delete resource just amounts to disabling ReAM
	reamReq := hubclient.OrgSettingRegistryAccessManagement{
		Enabled: false,
		StandardRegistries: []hubclient.RegistryAccessManagementStandardRegistry{
			{
				ID:      hubclient.StandardRegistryDocker,
				Allowed: true,
			},
		},
		CustomRegistries: []hubclient.RegistryAccessManagementCustomRegistry{},
	}

	_, err := r.client.SetOrgSettingRegistryAccessManagement(ctx, data.OrgName.ValueString(), reamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete org_setting_registry_access_management resource", err.Error())
		return
	}
}

func (r *OrgSettingRegistryAccessManagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), req.ID)...)
}

func getOrgSettingRegistryAccessManagementRequest(ctx context.Context, data OrgSettingRegistryAccessManagementResourceModel, d diag.Diagnostics) *hubclient.OrgSettingRegistryAccessManagement {
	var standardRegistryHub StandardRegistryModel

	d.Append(data.StandardRegistryDocker.As(ctx, &standardRegistryHub, basetypes.ObjectAsOptions{})...)
	if d.HasError() {
		return nil
	}

	reamReq := hubclient.OrgSettingRegistryAccessManagement{
		Enabled: data.Enabled.ValueBool(),
		StandardRegistries: []hubclient.RegistryAccessManagementStandardRegistry{
			{
				ID:      hubclient.StandardRegistryDocker,
				Allowed: standardRegistryHub.Allowed.ValueBool(),
			},
		},
		CustomRegistries: []hubclient.RegistryAccessManagementCustomRegistry{},
	}

	var customRegistries []CustomRegistryModel
	d.Append(data.CustomRegistries.ElementsAs(ctx, &customRegistries, false)...)
	if d.HasError() {
		return nil
	}

	for _, c := range customRegistries {
		reamReq.CustomRegistries = append(reamReq.CustomRegistries, hubclient.RegistryAccessManagementCustomRegistry{
			Address:      c.Address.ValueString(),
			FriendlyName: c.FriendlyName.ValueString(),
			Allowed:      c.Allowed.ValueBool(),
		})
	}

	return &reamReq
}

func getOrgSettingRegistryAccessManagementModel(ctx context.Context, reamResp hubclient.OrgSettingRegistryAccessManagement, d diag.Diagnostics) *OrgSettingRegistryAccessManagementResourceModel {
	var data OrgSettingRegistryAccessManagementResourceModel
	data.Enabled = types.BoolValue(reamResp.Enabled)

	var diags diag.Diagnostics
	for _, v := range reamResp.StandardRegistries {
		if v.ID == hubclient.StandardRegistryDocker {
			data.StandardRegistryDocker, diags = types.ObjectValue(StandardRegistryObjectType.AttrTypes, map[string]attr.Value{
				"allowed": types.BoolValue(v.Allowed),
			})
			d.Append(diags...)
			if d.HasError() {
				return nil
			}
		}
	}

	customRegistryModels := []CustomRegistryModel{}
	for _, v := range reamResp.CustomRegistries {
		customRegistryModels = append(customRegistryModels, CustomRegistryModel{
			Address:      types.StringValue(v.Address),
			FriendlyName: types.StringValue(v.FriendlyName),
			Allowed:      types.BoolValue(v.Allowed),
		})
	}

	data.CustomRegistries, diags = types.SetValueFrom(
		context.Background(),
		CustomRegistryObjectType,
		customRegistryModels,
	)
	d.Append(diags...)
	if d.HasError() {
		return nil
	}

	return &data
}
