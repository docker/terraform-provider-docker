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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepositoryResource(t *testing.T) {
	namespace := os.Getenv("DOCKER_USERNAME")
	name := "example-repo" + randString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRepositoryResourceConfig(namespace, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_hub_repository.test", "id"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "name", name),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "namespace", namespace),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "description", "Example repository"),
					resource.TestCheckNoResourceAttr("docker_hub_repository.test", "full_description"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "private", "false"),
				),
			},
			{
				Config: testRepositoryResourceConfigUpdated(namespace, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository.test", "description", "Updated example repository"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "full_description", "Full description update"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "private", "true"),
				),
			},
			{
				ResourceName:                         "docker_hub_repository.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["docker_hub_repository.test"].Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func testRepositoryResourceConfig(namespace, name string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Example repository"
  private         = false
}`, namespace, name)
}

func testRepositoryResourceConfigUpdated(namespace, name string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Updated example repository"
  full_description = "Full description update"
  private         = true
}`, namespace, name)
}

func TestAccRepositoryResourceImmutableTags(t *testing.T) {
	namespace := os.Getenv("DOCKER_USERNAME")
	name := "immutable-repo" + randString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRepositoryResourceConfigImmutableTags(namespace, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_hub_repository.test", "id"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "name", name),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "namespace", namespace),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.#", "2"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.0", "v*"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.1", "latest"),
				),
			},
			{
				Config: testRepositoryResourceConfigImmutableTagsUpdated(namespace, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.#", "1"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.0", "prod-*"),
				),
			},
			{
				Config: testRepositoryResourceConfigImmutableTagsDisabled(namespace, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.enabled", "false"),
				),
			},
		},
	})
}

func testRepositoryResourceConfigImmutableTags(namespace, name string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Repository with immutable tags"
  private         = false

  immutable_tags_settings = {
    enabled = true
    rules   = ["v*", "latest"]
  }
}`, namespace, name)
}

func testRepositoryResourceConfigImmutableTagsUpdated(namespace, name string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Repository with immutable tags"
  private         = false

  immutable_tags_settings = {
    enabled = true
    rules   = ["prod-*"]
  }
}`, namespace, name)
}

func testRepositoryResourceConfigImmutableTagsDisabled(namespace, name string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Repository with immutable tags disabled"
  private         = false

  immutable_tags_settings = {
    enabled = false
  }
}`, namespace, name)
}
