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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccessTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccessTokenResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_access_token.test", "uuid"),
					resource.TestCheckResourceAttr("docker_access_token.test", "is_active", "true"),
					resource.TestCheckResourceAttr("docker_access_token.test", "token_label", "test-label"),
					resource.TestCheckResourceAttr("docker_access_token.test", "scopes.#", "2"), // Assuming there are 2 scopes
					resource.TestCheckResourceAttrSet("docker_access_token.test", "token"),      // Check if the token is set
				),
			},
			{
				ResourceName:                         "docker_access_token.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateVerifyIgnore: []string{
					"token", // Ignore the token attribute during import state verification
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["docker_access_token.test"].Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}

const testAccessTokenResourceConfig = `
resource "docker_access_token" "test" {
  token_label = "test-label"
  scopes      = ["repo:read", "repo:write"]
}
`
