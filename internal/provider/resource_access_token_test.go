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
					resource.TestCheckResourceAttrSet("dockerhub_access_token.test", "uuid"),
					resource.TestCheckResourceAttr("dockerhub_access_token.test", "is_active", "true"),
					resource.TestCheckResourceAttr("dockerhub_access_token.test", "token_label", "test-label"),
					resource.TestCheckResourceAttr("dockerhub_access_token.test", "scopes.#", "2"), // Assuming there are 2 scopes
					resource.TestCheckResourceAttrSet("dockerhub_access_token.test", "token"),      // Check if the token is set
				),
			},
			{
				ResourceName:                         "dockerhub_access_token.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateVerifyIgnore: []string{
					"token", // Ignore the token attribute during import state verification
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["dockerhub_access_token.test"].Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}

const testAccessTokenResourceConfig = `
provider "dockerhub" {
  host     = "https://hub-stage.docker.com/v2"
}

resource "dockerhub_access_token" "test" {
  token_label = "test-label"
  scopes      = ["repo:read", "repo:write"]
}
`
