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
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepositoryImmutableTags(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	repoName := "tf-acc-immutable-tags"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create repository + enable immutable tags with rules
				Config: testAccRepositoryImmutableTags(orgName, repoName, true, []string{"latest", "v[0-9]+.*"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "namespace", orgName),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "name", repoName),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.#", "2"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.0", "latest"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.1", "v[0-9]+.*"),
				),
			},
			{
				// import
				Config:        testAccRepositoryImmutableTags(orgName, repoName, true, []string{"latest", "v[0-9]+.*"}),
				ImportState:   true,
				ImportStateId: orgName + "/" + repoName,
				ResourceName:  "docker_hub_repository_immutable_tags.test",
			},
			{
				// update the repository itself: docker_hub_repository does not
				// configure immutable_tags_settings, so it must not reset them
				Config: testAccRepositoryImmutableTagsWithDescription(orgName, repoName, "updated", true, []string{"latest", "v[0-9]+.*"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository.test", "description", "updated"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.#", "2"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.0", "latest"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.rules.1", "v[0-9]+.*"),
				),
			},
			{
				// update: change rules
				Config: testAccRepositoryImmutableTags(orgName, repoName, true, []string{".*"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.0", ".*"),
				),
			},
			{
				// remove the immutable tags resource (which disables immutable
				// tags) and update the repository in the same apply: the
				// repository update must not re-enable them from cached state
				Config: testAccRepositoryImmutableTagsRepoOnlyWithDescription(orgName, repoName, "removed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository.test", "description", "removed"),
					resource.TestCheckResourceAttr("docker_hub_repository.test", "immutable_tags_settings.enabled", "false"),
				),
			},
			{
				// update: disable
				Config: testAccRepositoryImmutableTags(orgName, repoName, false, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "enabled", "false"),
				),
			},
			{
				// delete immutable tags resource + repo
				Config: testAccRepositoryImmutableTagsRepoOnly(orgName, repoName),
			},
		},
	})
}

func testAccRepositoryImmutableTags(orgName, repoName string, enabled bool, rules []string) string {
	return testAccRepositoryImmutableTagsWithDescription(orgName, repoName, "", enabled, rules)
}

func testAccRepositoryImmutableTagsWithDescription(orgName, repoName, description string, enabled bool, rules []string) string {
	descriptionHCL := ""
	if description != "" {
		descriptionHCL = fmt.Sprintf("\n  description = %q", description)
	}
	rulesHCL := ""
	if rules != nil {
		quoted := make([]string, len(rules))
		for i, r := range rules {
			quoted[i] = fmt.Sprintf("%q", r)
		}
		rulesHCL = fmt.Sprintf("\n  rules   = [%s]", join(quoted, ", "))
	}
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  namespace = %q
  name      = %q
  private   = true%s
}

resource "docker_hub_repository_immutable_tags" "test" {
  namespace = docker_hub_repository.test.namespace
  name      = docker_hub_repository.test.name
  enabled   = %t%s
}
`, orgName, repoName, descriptionHCL, enabled, rulesHCL)
}

func testAccRepositoryImmutableTagsRepoOnly(orgName, repoName string) string {
	return testAccRepositoryImmutableTagsRepoOnlyWithDescription(orgName, repoName, "")
}

func testAccRepositoryImmutableTagsRepoOnlyWithDescription(orgName, repoName, description string) string {
	descriptionHCL := ""
	if description != "" {
		descriptionHCL = fmt.Sprintf("\n  description = %q", description)
	}
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  namespace = %q
  name      = %q
  private   = true%s
}
`, orgName, repoName, descriptionHCL)
}

func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
