package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testOrgExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_org.test", "id", "dockerhackathon"),
				),
			},
		},
	})
}

const testOrgExampleDataSourceConfig = `
provider "docker" {
  host = "https://hub-stage.docker.com/v2"
}
data "docker_org" "test" {
  org_name = "dockerhackathon"
}
`
