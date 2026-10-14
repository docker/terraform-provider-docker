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
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrgAccessTokenResource(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)
	updatedLabel := "test-" + randString(10)
	allReposPath := orgName + "/*"
	publicReposPath := "*/*/public"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgAccessTokenResourceConfig(orgName, label, "test description", allReposPath, "2029-12-31T23:59:59Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "id"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "label", label),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "description", "test description"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.#", "1"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.type", hubclient.OrgAccessTokenTypeRepo),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.path", allReposPath),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.scopes.#", "1"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.scopes.0", "scope-image-pull"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "expires_at", "2029-12-31T23:59:59Z"),
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "token"),
					resource.TestCheckNoResourceAttr("docker_org_access_token.test", "is_active"),
					resource.TestCheckNoResourceAttr("docker_org_access_token.test", "last_used_at"),
				),
			},
			{
				ResourceName:                         "docker_org_access_token.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateVerifyIgnore: []string{
					"token",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return orgName + "/" + state.RootModule().Resources["docker_org_access_token.test"].Primary.Attributes["id"], nil
				},
			},
			{
				Config: testAccOrgAccessTokenResourceConfig(orgName, updatedLabel, "updated description", publicReposPath, "2029-12-31T23:59:59Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "id"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "label", updatedLabel),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "description", "updated description"),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.path", publicReposPath),
					resource.TestCheckResourceAttr("docker_org_access_token.test", "resources.0.scopes.0", "scope-image-pull"),
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "token"),
				),
			},
		},
	})
}

func TestAccOrgAccessTokenResource_ExpiresAtReplaces(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)
	allReposPath := orgName + "/*"
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgAccessTokenResourceConfig(orgName, label, "test description", allReposPath, "2029-12-31T23:59:59Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "id"),
					storeResourceAttribute("docker_org_access_token.test", "id", &firstID),
				),
			},
			{
				Config: testAccOrgAccessTokenResourceConfig(orgName, label, "test description", allReposPath, "2030-12-31T23:59:59Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_access_token.test", "expires_at", "2030-12-31T23:59:59Z"),
					assertResourceAttributeChanged("docker_org_access_token.test", "id", firstID),
				),
			},
		},
	})
}

func TestAccOrgAccessTokenResource_NoExpiresAt(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "docker_org_access_token" "test" {
  org_name = "%s"
  label    = "%s"
  resources = [
    {
      type   = "%s"
      path   = "%s/*"
      scopes = ["scope-image-pull"]
    }
  ]
}
`, orgName, label, hubclient.OrgAccessTokenTypeRepo, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "id"),
					resource.TestCheckNoResourceAttr("docker_org_access_token.test", "expires_at"),
					resource.TestCheckNoResourceAttr("docker_org_access_token.test", "description"),
					resource.TestCheckResourceAttrSet("docker_org_access_token.test", "token"),
				),
			},
		},
	})
}

func TestOrgAccessTokenResource_toModelOptionalAttributes(t *testing.T) {
	r := &OrgAccessTokenResource{}
	token := hubclient.OrgAccessToken{ID: "id", Label: "label", Token: "secret"}
	withValues := hubclient.OrgAccessToken{
		ID:          "id",
		Label:       "label",
		Description: "desc",
		ExpiresAt:   "2029-12-31T23:59:59Z",
	}

	tests := []struct {
		name        string
		token       hubclient.OrgAccessToken
		state       *OrgAccessTokenResourceModel
		wantExpires types.String
		wantDesc    types.String
	}{
		{
			name:        "no prior state and no values from the API stay null",
			token:       token,
			state:       nil,
			wantExpires: types.StringNull(),
			wantDesc:    types.StringNull(),
		},
		{
			name:  "null prior values stay null",
			token: token,
			state: &OrgAccessTokenResourceModel{
				ExpiresAt:   types.StringNull(),
				Description: types.StringNull(),
			},
			wantExpires: types.StringNull(),
			wantDesc:    types.StringNull(),
		},
		{
			name:  "empty strings from earlier provider versions are healed to null",
			token: token,
			state: &OrgAccessTokenResourceModel{
				ExpiresAt:   types.StringValue(""),
				Description: types.StringValue(""),
			},
			wantExpires: types.StringNull(),
			wantDesc:    types.StringNull(),
		},
		{
			name:  "values returned by the API are kept",
			token: withValues,
			state: &OrgAccessTokenResourceModel{
				ExpiresAt:   types.StringValue("2029-12-31T23:59:59Z"),
				Description: types.StringValue("desc"),
			},
			wantExpires: types.StringValue("2029-12-31T23:59:59Z"),
			wantDesc:    types.StringValue("desc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, diags := r.toModel(context.Background(), tt.token, tt.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if !model.ExpiresAt.Equal(tt.wantExpires) {
				t.Errorf("expires_at: got %s, want %s", model.ExpiresAt, tt.wantExpires)
			}
			if !model.Description.Equal(tt.wantDesc) {
				t.Errorf("description: got %s, want %s", model.Description, tt.wantDesc)
			}
		})
	}
}

func testAccOrgAccessTokenResourceConfig(orgName, label, description, resourcePath, expiresAt string) string {
	return fmt.Sprintf(`
resource "docker_org_access_token" "test" {
  org_name    = "%s"
  label       = "%s"
  description = "%s"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "%s"
}
`, orgName, label, description, hubclient.OrgAccessTokenTypeRepo, resourcePath, expiresAt)
}

func storeResourceAttribute(resourceName, attributeName string, destination *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		value, ok := resourceState.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found for resource %s", attributeName, resourceName)
		}

		*destination = value
		return nil
	}
}

func assertResourceAttributeChanged(resourceName, attributeName, previousValue string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		value, ok := resourceState.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found for resource %s", attributeName, resourceName)
		}

		if value == previousValue {
			return fmt.Errorf("expected %s for resource %s to change", attributeName, resourceName)
		}

		return nil
	}
}
