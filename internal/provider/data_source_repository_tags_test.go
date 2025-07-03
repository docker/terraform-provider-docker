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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepositoryTagsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRepositoryTagsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_hub_repository_tags.test", "id", "library/hello-world"),
					resource.TestCheckResourceAttrSet("data.docker_hub_repository_tags.test", "tags.%"),
					resource.TestCheckResourceAttrSet("data.docker_hub_repository_tags.test", "tags.latest.name"),
					resource.TestCheckResourceAttrSet("data.docker_hub_repository_tags.test", "tags.latest.digest"),
					resource.TestCheckResourceAttrSet("data.docker_hub_repository_tags.test", "tags.latest.full_size"),
				),
			},
		},
	})
}

const testAccRepositoryTagsDataSourceConfig = `
data "docker_hub_repository_tags" "test" {
  namespace = "library"
  name      = "hello-world"
  page_size = 10
}
`
