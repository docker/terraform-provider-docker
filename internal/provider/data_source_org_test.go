package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgDataSource(t *testing.T) {
	org := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testOrgExampleDataSourceConfig(org),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_org.test", "id", org),
				),
			},
		},
	})
}

func testOrgExampleDataSourceConfig(org string) string {
	return fmt.Sprintf(`
provider "docker" {
  host = "hub-stage.docker.com"
}
data "docker_org" "test" {
  org_name = "%[1]s"
}`, org)
}
