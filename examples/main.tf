terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~> 0.2"
    }
  }

  required_version = "~>1.9"
}

provider "docker" {
  host = "hub-stage.docker.com"
}

# Resources Demo
# Create team
resource "docker_org_team" "terraform-team" {
  org_name         = "dockerterraform"
  team_name        = "terraformhackk"
  team_description = "Terraform Hackathon Demo - 2024"
}

# Team association
resource "docker_org_team_member" "example_association" {
  org_name  = "dockerterraform"
  team_name = resource.docker_org_team.terraform-team.team_name
  user_name = "username-placeholder"
}

# Create repository
resource "docker_hub_repository" "org_hub_repo" {
  namespace        = "dockerterraform"
  name             = "docker-terraform-repo-demo"
  description      = "This is a repo demo"
  full_description = "Lorem ipsum"
}

# Create repository team permission
resource "docker_hub_repository_team_permission" "test" {
  repo_id    = docker_hub_repository.org_hub_repo.id
  team_id    = docker_org_team.terraform-team.id
  permission = "admin"
}

# Create access token
resource "docker_access_token" "new_token_v2" {
  token_label = "terraform-created PAT-v2"
  scopes      = ["repo:read", "repo:write"]
}


# Output Demos
output "repo_output" {
  value = resource.docker_hub_repository.org_hub_repo
}

output "org_team_output" {
  value = resource.docker_org_team.terraform-team
}

output "org_team_association_output" {
  value = resource.docker_org_team_member.example_association
}

# output "access_tokens_uuids_output" {
#   value = resource.docker_access_token.new_token.uuid
# }

