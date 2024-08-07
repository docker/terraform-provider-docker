package provider

import (
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
				Config: testReposExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockerhub_repositories.test", "id", "username-placeholder/repositories"),
				),
			},
		},
	})
}

const testReposExampleDataSourceConfig = `
provider "dockerhub" {
  username = "username-placeholder"
  password = "PW"
  host = "https://hub-stage.docker.com/v2"
}
data "dockerhub_repositories" "test" {
  namespace = "username-placeholder"
  max_number_results = 10
}
`
