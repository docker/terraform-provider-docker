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
					resource.TestCheckResourceAttr("data.docker_hub_repository.test", "id", "ryanhristovski/data-source-example"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "docker_hub_repository" "test" {
  namespace = "ryanhristovski"
  name = "data-source-example"
}
`
