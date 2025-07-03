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

	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/docker/terraform-provider-docker/internal/repositoryutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &RepositoryTagsDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoryTagsDataSource{}
)

func NewRepositoryTagsDataSource() datasource.DataSource {
	return &RepositoryTagsDataSource{}
}

type RepositoryTagsDataSource struct {
	client *hubclient.Client
}

type RepositoryTagsDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Repository types.String `tfsdk:"repository"`
	Namespace  types.String `tfsdk:"namespace"`
	Name       types.String `tfsdk:"name"`
	PageSize   types.Int64  `tfsdk:"page_size"`
	Tags       types.Map    `tfsdk:"tags"`
}

type RepositoryTagModel struct {
	Name            types.String    `tfsdk:"name"`
	FullSize        types.Int64     `tfsdk:"full_size"`
	ID              types.Int64     `tfsdk:"id"`
	Repository      types.Int64     `tfsdk:"repository"`
	Creator         types.Int64     `tfsdk:"creator"`
	LastUpdated     types.String    `tfsdk:"last_updated"`
	LastUpdater     types.Int64     `tfsdk:"last_updater"`
	LastUpdaterName types.String    `tfsdk:"last_updater_username"`
	ImageID         types.String    `tfsdk:"image_id"`
	V2              types.Bool      `tfsdk:"v2"`
	TagStatus       types.String    `tfsdk:"tag_status"`
	TagLastPulled   types.String    `tfsdk:"tag_last_pulled"`
	TagLastPushed   types.String    `tfsdk:"tag_last_pushed"`
	MediaType       types.String    `tfsdk:"media_type"`
	ContentType     types.String    `tfsdk:"content_type"`
	Digest          types.String    `tfsdk:"digest"`
	Images          []TagImageModel `tfsdk:"images"`
}

type TagImageModel struct {
	Architecture types.String `tfsdk:"architecture"`
	Features     types.String `tfsdk:"features"`
	Variant      types.String `tfsdk:"variant"`
	Digest       types.String `tfsdk:"digest"`
	OS           types.String `tfsdk:"os"`
	OSFeatures   types.String `tfsdk:"os_features"`
	OSVersion    types.String `tfsdk:"os_version"`
	Size         types.Int64  `tfsdk:"size"`
	Status       types.String `tfsdk:"status"`
	LastPulled   types.String `tfsdk:"last_pulled"`
	LastPushed   types.String `tfsdk:"last_pushed"`
}

func (d *RepositoryTagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_repository_tags"
}

func (d *RepositoryTagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Retrieves tags for a Docker Hub repository.

## Key Security Use Case: Digest-Pinned Image References

This data source enables converting human-friendly tags to digest-pinned references for security-hardened deployments:

- **Risky**: ` + "`image: alpine:latest`" + ` (mutable tag)
- **Secure**: ` + "`image: alpine@sha256:5c16ec53d312df1867044cc90abd951bf37fdad32cc9b4a1e1e25d2f8eaf343c`" + ` (immutable digest)

## Example Usage

` + "```hcl" + `
data "docker_hub_repository_tags" "example" {
  namespace = "my-organization"
  name      = "my-repo"
}

# Convert human-friendly tag to digest-pinned reference
locals {
  desired_tag = "latest"
  secure_image_ref = try(
    "${data.docker_hub_repository_tags.example.tags[local.desired_tag].digest != "" ?
      "alpine@${data.docker_hub_repository_tags.example.tags[local.desired_tag].digest}" :
      "alpine:${local.desired_tag}"
    }",
    "alpine:${local.desired_tag}"
  )
}

output "all_tags" {
  description = "All Tags"
  value       = data.docker_hub_repository_tags.example.tags
}

output "single_tag" {
  description = "Single Tag Information"
  value       = data.docker_hub_repository_tags.example.tags["latest"]
}

# Alternative usage with repository ID
data "docker_hub_repository" "main" {
  namespace = "example"
  name      = "hello-world"
}

data "docker_hub_repository_tags" "main" {
  repository = data.docker_hub_repository.main.id
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The namespace/name of the repository",
				Computed:            true,
			},
			"repository": schema.StringAttribute{
				MarkdownDescription: "Repository ID in format namespace/name",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Repository namespace",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Repository name",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Number of tags to retrieve (default: 100)",
				Optional:            true,
			},
			"tags": schema.MapNestedAttribute{
				MarkdownDescription: "Map of tag names to tag information",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"full_size": schema.Int64Attribute{
							Computed: true,
						},
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"repository": schema.Int64Attribute{
							Computed: true,
						},
						"creator": schema.Int64Attribute{
							Computed: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"last_updater": schema.Int64Attribute{
							Computed: true,
						},
						"last_updater_username": schema.StringAttribute{
							Computed: true,
						},
						"image_id": schema.StringAttribute{
							Computed: true,
						},
						"v2": schema.BoolAttribute{
							Computed: true,
						},
						"tag_status": schema.StringAttribute{
							Computed: true,
						},
						"tag_last_pulled": schema.StringAttribute{
							Computed: true,
						},
						"tag_last_pushed": schema.StringAttribute{
							Computed: true,
						},
						"media_type": schema.StringAttribute{
							Computed: true,
						},
						"content_type": schema.StringAttribute{
							Computed: true,
						},
						"digest": schema.StringAttribute{
							Computed: true,
						},
						"images": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"architecture": schema.StringAttribute{
										Computed: true,
									},
									"features": schema.StringAttribute{
										Computed: true,
									},
									"variant": schema.StringAttribute{
										Computed: true,
									},
									"digest": schema.StringAttribute{
										Computed: true,
									},
									"os": schema.StringAttribute{
										Computed: true,
									},
									"os_features": schema.StringAttribute{
										Computed: true,
									},
									"os_version": schema.StringAttribute{
										Computed: true,
									},
									"size": schema.Int64Attribute{
										Computed: true,
									},
									"status": schema.StringAttribute{
										Computed: true,
									},
									"last_pulled": schema.StringAttribute{
										Computed: true,
									},
									"last_pushed": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *RepositoryTagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepositoryTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoryTagsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var namespace, name string
	var id string

	// Handle different input methods
	if !data.Repository.IsNull() && !data.Repository.IsUnknown() {
		// Use repository ID (namespace/name format)
		id = data.Repository.ValueString()
		namespace, name = repositoryutils.SplitID(id)
	} else if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() && !data.Name.IsNull() && !data.Name.IsUnknown() {
		// Use namespace and name
		namespace = data.Namespace.ValueString()
		name = data.Name.ValueString()
		id = repositoryutils.NewID(namespace, name)
	} else {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'repository' or both 'namespace' and 'name' must be provided",
		)
		return
	}

	// Set default page size if not provided
	pageSize := 100
	if !data.PageSize.IsNull() && !data.PageSize.IsUnknown() {
		pageSize = int(data.PageSize.ValueInt64())
	}

	tags, err := d.client.GetRepositoryTags(ctx, namespace, name, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Docker Hub API error reading repository tags", "Could not read repository tags, unexpected error: "+err.Error())
		return
	}

	// Convert tags to map structure
	tagsMap := make(map[string]RepositoryTagModel)
	for _, tag := range tags.Results {
		// Convert images
		var images []TagImageModel
		for _, img := range tag.Images {
			images = append(images, TagImageModel{
				Architecture: types.StringValue(img.Architecture),
				Features:     types.StringValue(img.Features),
				Variant:      types.StringValue(img.Variant),
				Digest:       types.StringValue(img.Digest),
				OS:           types.StringValue(img.OS),
				OSFeatures:   types.StringValue(img.OSFeatures),
				OSVersion:    types.StringValue(img.OSVersion),
				Size:         types.Int64Value(img.Size),
				Status:       types.StringValue(img.Status),
				LastPulled:   types.StringValue(img.LastPulled),
				LastPushed:   types.StringValue(img.LastPushed),
			})
		}

		tagsMap[tag.Name] = RepositoryTagModel{
			Name:            types.StringValue(tag.Name),
			FullSize:        types.Int64Value(tag.FullSize),
			ID:              types.Int64Value(tag.ID),
			Repository:      types.Int64Value(tag.Repository),
			Creator:         types.Int64Value(tag.Creator),
			LastUpdated:     types.StringValue(tag.LastUpdated),
			LastUpdater:     types.Int64Value(tag.LastUpdater),
			LastUpdaterName: types.StringValue(tag.LastUpdaterName),
			ImageID:         types.StringValue(tag.ImageID),
			V2:              types.BoolValue(tag.V2),
			TagStatus:       types.StringValue(tag.TagStatus),
			TagLastPulled:   types.StringValue(tag.TagLastPulled),
			TagLastPushed:   types.StringValue(tag.TagLastPushed),
			MediaType:       types.StringValue(tag.MediaType),
			ContentType:     types.StringValue(tag.ContentType),
			Digest:          types.StringValue(tag.Digest),
			Images:          images,
		}
	}

	data.ID = types.StringValue(id)
	tagsMapValue, diags := types.MapValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                  types.StringType,
			"full_size":             types.Int64Type,
			"id":                    types.Int64Type,
			"repository":            types.Int64Type,
			"creator":               types.Int64Type,
			"last_updated":          types.StringType,
			"last_updater":          types.Int64Type,
			"last_updater_username": types.StringType,
			"image_id":              types.StringType,
			"v2":                    types.BoolType,
			"tag_status":            types.StringType,
			"tag_last_pulled":       types.StringType,
			"tag_last_pushed":       types.StringType,
			"media_type":            types.StringType,
			"content_type":          types.StringType,
			"digest":                types.StringType,
			"images": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"architecture": types.StringType,
						"features":     types.StringType,
						"variant":      types.StringType,
						"digest":       types.StringType,
						"os":           types.StringType,
						"os_features":  types.StringType,
						"os_version":   types.StringType,
						"size":         types.Int64Type,
						"status":       types.StringType,
						"last_pulled":  types.StringType,
						"last_pushed":  types.StringType,
					},
				},
			},
		},
	}, tagsMap)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Tags = tagsMapValue

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
