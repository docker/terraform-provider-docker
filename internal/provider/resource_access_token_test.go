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
provider "docker" {
  host = "hub-stage.docker.com"
}

resource "docker_access_token" "test" {
  token_label = "test-label"
  scopes      = ["repo:read", "repo:write"]
}
`
