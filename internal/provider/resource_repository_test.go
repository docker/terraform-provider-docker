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
provider "docker" {
  host = "hub-stage.docker.com"
}

resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Example repository"
  private         = false
}`, namespace, name)
}

func testRepositoryResourceConfigUpdated(namespace, name string) string {
	return fmt.Sprintf(`
provider "docker" {
  host = "hub-stage.docker.com"
}

resource "docker_hub_repository" "test" {
  name            = "%[2]s"
  namespace       = "%[1]s"
  description     = "Updated example repository"
  full_description = "Full description update"
  private         = true
}`, namespace, name)
}
