package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepositoryDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockerhub_repository.test", "id", "ryanhristovski/data-source-example"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
provider "dockerhub" {
  host = "https://hub-stage.docker.com/v2"
}
data "dockerhub_repository" "test" {
  namespace = "ryanhristovski"
  name = "data-source-example"
}
`
