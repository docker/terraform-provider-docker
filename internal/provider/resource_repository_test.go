package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRepositoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRepositoryResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_repository.test", "id"),
					resource.TestCheckResourceAttr("docker_repository.test", "name", "example-repo"),
					resource.TestCheckResourceAttr("docker_repository.test", "namespace", os.Getenv("DOCKER_USERNAME")),
					resource.TestCheckResourceAttr("docker_repository.test", "description", "Example repository"),
					resource.TestCheckNoResourceAttr("docker_repository.test", "full_description"),
					resource.TestCheckResourceAttr("docker_repository.test", "private", "false"),
				),
			},
			{
				Config: testRepositoryResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_repository.test", "description", "Updated example repository"),
					resource.TestCheckResourceAttr("docker_repository.test", "full_description", "Full description update"),
					resource.TestCheckResourceAttr("docker_repository.test", "private", "true"),
				),
			},
			{
				ResourceName:                         "docker_repository.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["docker_repository.test"].Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func testRepositoryResourceConfig() string {
	return `
resource "docker_repository" "test" {
  name            = "example-repo"
  namespace       = "` + os.Getenv("DOCKER_USERNAME") + `"
  description     = "Example repository"
  private         = false
}
`
}

func testRepositoryResourceConfigUpdated() string {
	return `
resource "docker_repository" "test" {
  name            = "example-repo"
  namespace       = "` + os.Getenv("DOCKER_USERNAME") + `"
  description     = "Updated example repository"
  full_description = "Full description update"
  private         = true
}
`
}
