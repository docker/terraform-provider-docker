/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgMembersDataSource(t *testing.T) {
	username := os.Getenv("DOCKER_USERNAME")
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgMembersDataSourceConfig(orgName, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("self", username),
				),
			},
		},
	})
}

func testAccOrgMembersDataSourceConfig(orgName string, username string) string {
	return fmt.Sprintf(`
data "docker_org_members" "_" {
  org_name  = "%s"
}

output "self" {
  value = [for member in data.docker_org_members._.members: member if member.username == "%s"][0].username
}
`, orgName, username)
}
