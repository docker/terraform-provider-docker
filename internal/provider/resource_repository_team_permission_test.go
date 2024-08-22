package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepositoryTeamPermission(t *testing.T) {
	orgName := "dockerterraform"
	teamName := "test" + randString(10)
	repoName := "test" + randString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccRepositoryTeamPermission(orgName, teamName, repoName, hubclient.TeamRepoPermissionLevelRead),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "repo_id", "docker_hub_repository.test", "id"),
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "team_id", "docker_org_team.test", "id"),
					resource.TestCheckResourceAttr("docker_hub_repository_team_permission.test", "permission", hubclient.TeamRepoPermissionLevelRead),
				),
			},
			{
				// import
				Config:      testAccRepositoryTeamPermission(orgName, teamName, repoName, hubclient.TeamRepoPermissionLevelRead),
				ImportState: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					teamID := state.RootModule().Resources["docker_org_team.test"].Primary.Attributes["id"]
					return orgName + "/" + repoName + "/" + teamID, nil
				},
				ResourceName: "docker_hub_repository_team_permission.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "repo_id", "docker_hub_repository.test", "id"),
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "team_id", "docker_org_team.test", "id"),
					resource.TestCheckResourceAttr("docker_hub_repository_team_permission.test", "permission", hubclient.TeamRepoPermissionLevelRead),
				),
			},
			{
				// update permission
				Config: testAccRepositoryTeamPermission(orgName, teamName, repoName, hubclient.TeamRepoPermissionLevelAdmin),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "repo_id", "docker_hub_repository.test", "id"),
					resource.TestCheckResourceAttrPair("docker_hub_repository_team_permission.test", "team_id", "docker_org_team.test", "id"),
					resource.TestCheckResourceAttr("docker_hub_repository_team_permission.test", "permission", hubclient.TeamRepoPermissionLevelAdmin),
				),
			},
			{
				// delete
				Config: testAccRepositoryTeamPermissionBase(orgName, teamName, repoName),
			},
		},
	})
}

func testAccRepositoryTeamPermissionBase(orgName, teamName, repoName string) string {
	return fmt.Sprintf(`
provider "docker" {
  host = "https://hub-stage.docker.com/v2"
}

resource "docker_org_team" "test" {
  org_name         = "%[1]s"
  team_name        = "%[2]s"
}

resource "docker_hub_repository" "test" {
  namespace = "%[1]s"
  name      = "%[3]s"
}`, orgName, teamName, repoName)
}

func testAccRepositoryTeamPermission(orgName, teamName, repoName string, permission hubclient.TeamRepoPermissionLevel) string {
	return fmt.Sprintf(`
%[1]s

resource "docker_hub_repository_team_permission" "test" {
  repo_id    = docker_hub_repository.test.id
  team_id    = docker_org_team.test.id
  permission = "%[2]s"
}
`, testAccRepositoryTeamPermissionBase(orgName, teamName, repoName), permission)
}
