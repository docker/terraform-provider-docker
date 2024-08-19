package provider

import (
	"context"
	"fmt"

	"github.com/docker/terraform-provider-dockerhub/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OrgSettingImageAccessManagementResource{}
	_ resource.ResourceWithConfigure   = &OrgSettingImageAccessManagementResource{}
	_ resource.ResourceWithImportState = &OrgSettingImageAccessManagementResource{}
)

func NewOrgSettingImageAccessManagementResource() resource.Resource {
	return &OrgSettingImageAccessManagementResource{}
}

type OrgSettingImageAccessManagementResource struct {
	client *hubclient.Client
}

type OrgSettingImageAccessManagementResourceModel struct {
	OrgName                 types.String `tfsdk:"org_name"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	AllowOfficialImages     types.Bool   `tfsdk:"allow_official_images"`
	AllowVerifiedPublishers types.Bool   `tfsdk:"allow_verified_publishers"`
}

func (r *OrgSettingImageAccessManagementResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgSettingImageAccessManagementResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_setting_image_access_management"
}

func (r *OrgSettingImageAccessManagementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Image Access Management settings for an organization.",

		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether or not Image Access Management is enabled. When this feature is enabled, only images created by your organization or by Docker Official Images and Docker Verified Publishers are allowed. All community images are restricted.",
				Required:            true,
			},
			"allow_official_images": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to allow curated set of Docker Official Images repositories hosted on Docker Hub. Only takes effect when Image Access Management feature is enabled.⁠",
				Required:            true,
			},
			"allow_verified_publishers": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to allow High-quality images by Docker Verified Publishers. Only takes effect when Image Access Management feature is enabled.⁠",
				Required:            true,
			},
		},
	}
}

func (r *OrgSettingImageAccessManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgSettingImageAccessManagementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	iamReq := hubclient.OrgSettingImageAccessManagement{
		RestrictedImages: hubclient.ImageAccessManagementRestrictedImages{
			Enabled:                 data.Enabled.ValueBool(),
			AllowOfficialImages:     data.AllowOfficialImages.ValueBool(),
			AllowVerifiedPublishers: data.AllowVerifiedPublishers.ValueBool(),
		},
	}

	iamResp, err := r.client.SetOrgSettingImageAccessManagement(ctx, data.OrgName.ValueString(), iamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create org_setting_image_access_management resource", err.Error())
		return
	}

	data.Enabled = types.BoolValue(iamResp.RestrictedImages.Enabled)
	data.AllowOfficialImages = types.BoolValue(iamResp.RestrictedImages.AllowOfficialImages)
	data.AllowVerifiedPublishers = types.BoolValue(iamResp.RestrictedImages.AllowVerifiedPublishers)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingImageAccessManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgSettingImageAccessManagementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	iamResp, err := r.client.GetOrgSettingImageAccessManagement(ctx, data.OrgName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read org_setting_image_access_management resource", err.Error())
		return
	}

	data.Enabled = types.BoolValue(iamResp.RestrictedImages.Enabled)
	data.AllowOfficialImages = types.BoolValue(iamResp.RestrictedImages.AllowOfficialImages)
	data.AllowVerifiedPublishers = types.BoolValue(iamResp.RestrictedImages.AllowVerifiedPublishers)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingImageAccessManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrgSettingImageAccessManagementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	iamReq := hubclient.OrgSettingImageAccessManagement{
		RestrictedImages: hubclient.ImageAccessManagementRestrictedImages{
			Enabled:                 data.Enabled.ValueBool(),
			AllowOfficialImages:     data.AllowOfficialImages.ValueBool(),
			AllowVerifiedPublishers: data.AllowVerifiedPublishers.ValueBool(),
		},
	}

	iamResp, err := r.client.SetOrgSettingImageAccessManagement(ctx, data.OrgName.ValueString(), iamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update org_setting_image_access_management resource", err.Error())
		return
	}

	data.Enabled = types.BoolValue(iamResp.RestrictedImages.Enabled)
	data.AllowOfficialImages = types.BoolValue(iamResp.RestrictedImages.AllowOfficialImages)
	data.AllowVerifiedPublishers = types.BoolValue(iamResp.RestrictedImages.AllowVerifiedPublishers)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrgSettingImageAccessManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrgSettingImageAccessManagementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// delete resource just amounts to disabling IAM
	iamReq := hubclient.OrgSettingImageAccessManagement{
		RestrictedImages: hubclient.ImageAccessManagementRestrictedImages{
			Enabled:                 false,
			AllowOfficialImages:     true,
			AllowVerifiedPublishers: true,
		},
	}

	_, err := r.client.SetOrgSettingImageAccessManagement(ctx, data.OrgName.ValueString(), iamReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete org_setting_image_access_management resource", err.Error())
		return
	}
}

func (r *OrgSettingImageAccessManagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_name"), req.ID)...)
}
