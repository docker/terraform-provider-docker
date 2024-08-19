terraform {
  required_providers {
    dockerhub = {
      source  = "docker/dockerhub"
      version = "~>1.0"
    }
  }

  required_version = "~>1.9"
}

provider "dockerhub" {
  host = "https://hub-stage.docker.com/v2"
}

# Resources Demo
# Create team
resource "dockerhub_org_team" "terraform-team" {
  org_name         = "dockerterraform"
  team_name        = "terraformhackk"
  team_description = "Terraform Hackathon Demo - 2024"
}

# Team association
resource "dockerhub_org_team_member_association" "example_association" {
  org_name   = "dockerterraform"
  team_name  = resource.dockerhub_org_team.terraform-team.team_name
  user_names = ["forrestloomis371", "username-placeholder"]
}

# Create repository
resource "dockerhub_repository" "org_repo" {
  namespace        = "dockerterraform"
  name             = "docker-terraform-repo-demo-"
  description      = "This is a repo demo"
  full_description = "Lorem ipsum"
}

# Create repository team permission
resource "dockerhub_repository_team_permission" "test" {
  repo_id    = dockerhub_repository.org_repo.id
  team_id    = dockerhub_org_team.terraform-team.id
  permission = "admin"
}

# Create access token
resource "dockerhub_access_token" "new_token_v2" {
  token_label = "terraform-created PAT-v2 t"
  scopes      = ["repo:read", "repo:write"]
}


# Output Demos
output "repo_output" {
  value = resource.dockerhub_repository.org_repo
}

output "org_team_output" {
  value = resource.dockerhub_org_team.terraform-team
}

output "org_team_association_output" {
  value = resource.dockerhub_org_team_member_association.example_association
}

# output "access_tokens_uuids_output" {
#   value = resource.dockerhub_access_token.new_token.uuid
# }

