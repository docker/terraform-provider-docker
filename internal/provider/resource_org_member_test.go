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

func TestAccOrgMemberResource(t *testing.T) {
	org := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "docker_org_member" "test" {
  org_name  = "%[1]s"
  email = "newtest@example.com"
  role = "member"
}`, org),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_member.test", "invite_id"),
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", org),
					resource.TestCheckResourceAttr("docker_org_member.test", "email", "newtest@example.com"),
					resource.TestCheckResourceAttr("docker_org_member.test", "role", "member"),
				),
			},
			// TODO(nicks): Enable this once we support importing invites.
			// {
			// 	ResourceName:            "docker_org_member.test",
			// 	ImportState:             false,
			// 	ImportStateVerify:       false,
			// 	ImportStateVerifyIgnore: []string{"invite_id"},
			// },
		},
	})
}

// TestAccOrgMemberResource_ExistingMember tests managing the role of an existing
// organization member.
//
// By design, this test requires the user "nick20241127" to already be a member
// so that we don't have to deal with the invite flow.
func TestAccOrgMemberResource_ExistingMember(t *testing.T) {
	org := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "docker_org_member" "test" {
  org_name  = "%[1]s"
  user_name = "nick20241127"
  role      = "member"

  lifecycle {
    prevent_destroy = true
  }
}`, org),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", org),
					resource.TestCheckResourceAttr("docker_org_member.test", "user_name", "nick20241127"),
					resource.TestCheckResourceAttr("docker_org_member.test", "email", "nick.santos+nick20241127@docker.com"),
					resource.TestCheckResourceAttr("docker_org_member.test", "role", "member"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "docker_org_member" "test" {
  org_name  = "%[1]s"
  user_name = "nick20241127"
  role      = "editor"

  lifecycle{
    prevent_destroy = true
  }
}`, org),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", org),
					resource.TestCheckResourceAttr("docker_org_member.test", "user_name", "nick20241127"),
					resource.TestCheckResourceAttr("docker_org_member.test", "email", "nick.santos+nick20241127@docker.com"),
					resource.TestCheckResourceAttr("docker_org_member.test", "role", "editor"),
				),
			},
			{
				Config: `
removed {
  from = docker_org_member.test
  lifecycle {
    destroy = false
  }
}`,
			},
		},
	})
}

func TestAccOrgMemberResourceImport(t *testing.T) {
	username := os.Getenv("DOCKER_USERNAME")
	org := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: fmt.Sprintf(`
data "docker_login" "_" {}

import {
  id = "%[1]s/%[2]s"
  to = docker_org_member.test
}
resource "docker_org_member" "test" {
  org_name  = "%[1]s"
  user_name = "%[2]s"
  role      = "owner"
}`, org, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_member.test", "invite_id"),
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", org),
					resource.TestCheckResourceAttr("docker_org_member.test", "user_name", username),
				),
			},
		},
	})
}
