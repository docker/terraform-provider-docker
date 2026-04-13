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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const orgAccessTokenFilterNameLabel = "label"

var (
	_ datasource.DataSource              = &OrgAccessTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &OrgAccessTokenDataSource{}
)

func NewOrgAccessTokenDataSource() datasource.DataSource {
	return &OrgAccessTokenDataSource{}
}

type OrgAccessTokenDataSource struct {
	client *hubclient.Client
}

type OrgAccessTokenDataSourceModel struct {
	OrgName     types.String                       `tfsdk:"org_name"`
	ID          types.String                       `tfsdk:"id"`
	Filter      []OrgAccessTokenFilterModel        `tfsdk:"filter"`
	Label       types.String                       `tfsdk:"label"`
	Description types.String                       `tfsdk:"description"`
	CreatedBy   types.String                       `tfsdk:"created_by"`
	IsActive    types.Bool                         `tfsdk:"is_active"`
	CreatedAt   types.String                       `tfsdk:"created_at"`
	ExpiresAt   types.String                       `tfsdk:"expires_at"`
	LastUsedAt  types.String                       `tfsdk:"last_used_at"`
	Resources   []OrgAccessTokenResourceEntryModel `tfsdk:"resources"`
}

type OrgAccessTokenFilterModel struct {
	Name   types.String `tfsdk:"name"`
	Values types.List   `tfsdk:"values"`
}

type orgAccessTokenFilter struct {
	name   string
	values map[string]struct{}
}

func (d *OrgAccessTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_access_token"
}

func (d *OrgAccessTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Reads metadata for an existing organization access token.

Provide either ` + "`id`" + ` for a direct lookup or a ` + "`filter`" + ` block to resolve exactly one token by metadata. The ` + "`filter`" + ` block follows the common Terraform name/value pattern and currently supports ` + "`label`" + ` only.

-> **Note**: This data source is only available when authenticated with a username and password as an owner of the org.

-> **Note**: This data source returns metadata only. Docker Hub only returns the secret token value during creation, so ` + "`token`" + ` is intentionally not exposed here.

-> **Note**: Filter-based lookups scan all organization access token pages to guarantee unique matching, even if the provider ` + "`max_page_results`" + ` setting is lower.

## Example Usage

` + "```hcl" + `
data "docker_org_access_token" "by_id" {
  org_name = "my-organization"
  id       = "a7a5ef25-8889-43a0-8cc7-f2a94268e861"
}

data "docker_org_access_token" "by_label" {
  org_name = "my-organization"

  filter {
    name   = "label"
    values = ["ci-token"]
  }
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The organization namespace.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization access token. Set this for a direct lookup, or omit it when using `filter`.",
				Optional:            true,
				Computed:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "The label of the access token.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the access token.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user that created the access token.",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the access token is active.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the access token.",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration timestamp of the access token, if set.",
				Computed:            true,
			},
			"last_used_at": schema.StringAttribute{
				MarkdownDescription: "The last time the access token was used, if available.",
				Computed:            true,
			},
			"resources": schema.ListNestedAttribute{
				MarkdownDescription: "Resources this token has access to.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of resource.",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The path of the resource.",
							Computed:            true,
						},
						"scopes": schema.ListAttribute{
							MarkdownDescription: "The scopes this token has access to.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.ListNestedBlock{
				MarkdownDescription: "One or more name/value filter blocks. Exactly one of `id` or `filter` must be set. Filters are applied with OR semantics within `values` and AND semantics across blocks. Only `label` is supported in v1.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the field to filter by. Only `label` is supported.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(orgAccessTokenFilterNameLabel),
							},
						},
						"values": schema.ListAttribute{
							MarkdownDescription: "Accepted values for the given filter name.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *OrgAccessTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgAccessTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgAccessTokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != ""
	hasFilter := len(data.Filter) > 0
	if hasID == hasFilter {
		resp.Diagnostics.AddError(
			"Invalid org access token lookup",
			"Exactly one of `id` or `filter` must be set.",
		)
		return
	}

	var at hubclient.OrgAccessToken
	var err error
	if hasID {
		at, err = d.client.GetOrgAccessToken(ctx, data.OrgName.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Docker Hub API error reading org access token", fmt.Sprintf("%v", err))
			return
		}
	} else {
		at, err = d.findOrgAccessTokenByFilter(ctx, data.OrgName.ValueString(), data.Filter)
		if err != nil {
			resp.Diagnostics.AddError("Docker Hub API error reading org access token", err.Error())
			return
		}
	}

	resources, diags := flattenOrgAccessTokenResources(ctx, at.Resources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(at.ID)
	data.Label = types.StringValue(at.Label)
	data.Description = types.StringValue(at.Description)
	data.CreatedBy = types.StringValue(at.CreatedBy)
	data.IsActive = types.BoolValue(at.IsActive)
	data.CreatedAt = types.StringValue(at.CreatedAt)
	data.ExpiresAt = types.StringValue(at.ExpiresAt)
	data.LastUsedAt = types.StringValue(at.LastUsedAt)
	data.Resources = resources

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *OrgAccessTokenDataSource) findOrgAccessTokenByFilter(
	ctx context.Context,
	orgName string,
	filters []OrgAccessTokenFilterModel,
) (hubclient.OrgAccessToken, error) {
	expandedFilters, diags := expandOrgAccessTokenFilters(ctx, filters)
	if diags.HasError() {
		return hubclient.OrgAccessToken{}, fmt.Errorf("%s", diags[0].Detail())
	}

	accessTokens, err := d.client.ListOrgAccessTokens(ctx, orgName)
	if err != nil {
		return hubclient.OrgAccessToken{}, err
	}

	matches := filterOrgAccessTokens(accessTokens, expandedFilters)
	switch len(matches) {
	case 0:
		return hubclient.OrgAccessToken{}, fmt.Errorf("no org access tokens matched the configured filters")
	case 1:
		return d.client.GetOrgAccessToken(ctx, orgName, matches[0].ID)
	default:
		return hubclient.OrgAccessToken{}, fmt.Errorf("configured filters matched %d org access tokens; refine the filter or use `id`", len(matches))
	}
}

func expandOrgAccessTokenFilters(ctx context.Context, filters []OrgAccessTokenFilterModel) ([]orgAccessTokenFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	expanded := make([]orgAccessTokenFilter, 0, len(filters))
	for _, filterModel := range filters {
		var values []string
		diags.Append(filterModel.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return nil, diags
		}

		if len(values) == 0 {
			diags.AddError(
				"Invalid org access token filter",
				fmt.Sprintf("Filter %q must include at least one value.", filterModel.Name.ValueString()),
			)
			return nil, diags
		}

		valueSet := make(map[string]struct{}, len(values))
		for _, value := range values {
			valueSet[value] = struct{}{}
		}

		expanded = append(expanded, orgAccessTokenFilter{
			name:   filterModel.Name.ValueString(),
			values: valueSet,
		})
	}

	return expanded, diags
}

func filterOrgAccessTokens(tokens []hubclient.OrgAccessToken, filters []orgAccessTokenFilter) []hubclient.OrgAccessToken {
	matches := make([]hubclient.OrgAccessToken, 0, len(tokens))

	for _, token := range tokens {
		if orgAccessTokenMatchesAllFilters(token, filters) {
			matches = append(matches, token)
		}
	}

	return matches
}

func orgAccessTokenMatchesAllFilters(token hubclient.OrgAccessToken, filters []orgAccessTokenFilter) bool {
	for _, filter := range filters {
		switch filter.name {
		case orgAccessTokenFilterNameLabel:
			if _, ok := filter.values[token.Label]; !ok {
				return false
			}
		default:
			return false
		}
	}

	return true
}
