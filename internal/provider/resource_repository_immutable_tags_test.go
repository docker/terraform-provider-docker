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
				// update: change rules
				Config: testAccRepositoryImmutableTags(orgName, repoName, true, []string{".*"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("docker_hub_repository_immutable_tags.test", "rules.0", ".*"),
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
  private   = true
}

resource "docker_hub_repository_immutable_tags" "test" {
  namespace = docker_hub_repository.test.namespace
  name      = docker_hub_repository.test.name
  enabled   = %t%s
}
`, orgName, repoName, enabled, rulesHCL)
}

func testAccRepositoryImmutableTagsRepoOnly(orgName, repoName string) string {
	return fmt.Sprintf(`
resource "docker_hub_repository" "test" {
  namespace = %q
  name      = %q
  private   = true
}
`, orgName, repoName)
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
