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
					resource.TestCheckResourceAttr("data.dockerhub_org.test", "id", "dockerhackathon"),
				),
			},
		},
	})
}

const testOrgExampleDataSourceConfig = `
provider "dockerhub" {
  username = "username-placeholder"
  password = "PW"
  host = "https://hub-stage.docker.com/v2"
}
data "dockerhub_org" "test" {
  org_name = "dockerhackathon"
}
`
