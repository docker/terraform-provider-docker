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
	"fmt"
	"regexp"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgAccessTokenDataSource_ByID(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)
	resourcePath := orgName + "/*"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgAccessTokenDataSourceConfigByID(orgName, label, resourcePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "org_name", orgName),
					resource.TestCheckResourceAttrSet("data.docker_org_access_token.test", "id"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "label", label),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "description", "test description"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "resources.#", "1"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "resources.0.type", hubclient.OrgAccessTokenTypeRepo),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "resources.0.path", resourcePath),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "resources.0.scopes.#", "1"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "resources.0.scopes.0", "scope-image-pull"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "expires_at", "2029-12-31T23:59:59Z"),
					resource.TestCheckNoResourceAttr("data.docker_org_access_token.test", "token"),
				),
			},
		},
	})
}

func TestAccOrgAccessTokenDataSource_FilterByLabel(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)
	resourcePath := orgName + "/*"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgAccessTokenDataSourceConfigByFilter(orgName, label, resourcePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "org_name", orgName),
					resource.TestCheckResourceAttrSet("data.docker_org_access_token.test", "id"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "label", label),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "description", "test description"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "filter.#", "1"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "filter.0.name", orgAccessTokenFilterNameLabel),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "filter.0.values.#", "1"),
					resource.TestCheckResourceAttr("data.docker_org_access_token.test", "filter.0.values.0", label),
					resource.TestCheckNoResourceAttr("data.docker_org_access_token.test", "token"),
				),
			},
		},
	})
}

func TestAccOrgAccessTokenDataSource_FilterMultipleMatches(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	labelOne := "test-" + randString(10)
	labelTwo := "test-" + randString(10)
	resourcePath := orgName + "/*"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrgAccessTokenDataSourceConfigWithMultipleMatches(orgName, labelOne, labelTwo, resourcePath),
				ExpectError: regexp.MustCompile(`configured filters matched 2 org access tokens`),
			},
		},
	})
}

func TestAccOrgAccessTokenDataSource_FilterNoMatch(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	label := "test-" + randString(10)
	resourcePath := orgName + "/*"
	missingLabel := "missing-" + randString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrgAccessTokenDataSourceConfigNoMatch(orgName, label, resourcePath, missingLabel),
				ExpectError: regexp.MustCompile(`no org access tokens matched the configured filters`),
			},
		},
	})
}

func testAccOrgAccessTokenDataSourceConfigByID(orgName, label, resourcePath string) string {
	return fmt.Sprintf(`
resource "docker_org_access_token" "test" {
  org_name    = "%s"
  label       = "%s"
  description = "test description"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2029-12-31T23:59:59Z"
}

data "docker_org_access_token" "test" {
  org_name = docker_org_access_token.test.org_name
  id       = docker_org_access_token.test.id
}
`, orgName, label, hubclient.OrgAccessTokenTypeRepo, resourcePath)
}

func testAccOrgAccessTokenDataSourceConfigByFilter(orgName, label, resourcePath string) string {
	return fmt.Sprintf(`
resource "docker_org_access_token" "test" {
  org_name    = "%s"
  label       = "%s"
  description = "test description"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2029-12-31T23:59:59Z"
}

data "docker_org_access_token" "test" {
  org_name = docker_org_access_token.test.org_name

  filter {
    name   = "label"
    values = [docker_org_access_token.test.label]
  }
}
`, orgName, label, hubclient.OrgAccessTokenTypeRepo, resourcePath)
}

func testAccOrgAccessTokenDataSourceConfigWithMultipleMatches(orgName, labelOne, labelTwo, resourcePath string) string {
	return fmt.Sprintf(`
resource "docker_org_access_token" "one" {
  org_name    = "%s"
  label       = "%s"
  description = "test description one"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2029-12-31T23:59:59Z"
}

resource "docker_org_access_token" "two" {
  org_name    = "%s"
  label       = "%s"
  description = "test description two"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2029-12-31T23:59:59Z"
}

data "docker_org_access_token" "test" {
  org_name = "%s"

  filter {
    name   = "label"
    values = ["%s", "%s"]
  }
}
`, orgName, labelOne, hubclient.OrgAccessTokenTypeRepo, resourcePath, orgName, labelTwo, hubclient.OrgAccessTokenTypeRepo, resourcePath, orgName, labelOne, labelTwo)
}

func testAccOrgAccessTokenDataSourceConfigNoMatch(orgName, label, resourcePath, missingLabel string) string {
	return fmt.Sprintf(`
resource "docker_org_access_token" "test" {
  org_name    = "%s"
  label       = "%s"
  description = "test description"
  resources = [
    {
      type   = "%s"
      path   = "%s"
      scopes = ["scope-image-pull"]
    }
  ]
  expires_at = "2029-12-31T23:59:59Z"
}

data "docker_org_access_token" "test" {
  org_name = "%s"

  filter {
    name   = "label"
    values = ["%s"]
  }
}
`, orgName, label, hubclient.OrgAccessTokenTypeRepo, resourcePath, orgName, missingLabel)
}
