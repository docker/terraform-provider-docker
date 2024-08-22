package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepositoriesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testReposExampleDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_hub_repositories.test", "id", os.Getenv("DOCKER_USERNAME")+"/repositories"),
				),
			},
		},
	})
}

func testReposExampleDataSourceConfig() string {
	return `
provider "docker" {
  host = "https://hub-stage.docker.com/v2"
}

data "docker_hub_repositories" "test" {
  namespace         = "` + os.Getenv("DOCKER_USERNAME") + `"
  max_number_results = 10
}
`
}
