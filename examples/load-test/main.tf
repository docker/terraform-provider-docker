terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~>1.0"
    }
  }

  required_version = "~>1.9"
}

provider "docker" {
  host = "https://hub-stage.docker.com/v2"
}

# Variables to create variations
variable "user_names" {
  type    = list(string)
  default = ["forrestloomis371", "username-placeholder"]
}

variable "team_names" {
  type    = list(string)
  default = ["terraformhack"]
}

variable "repo_names" {
  type    = list(string)
  default = ["docker-terraform-repo-demo"]
}

variable "token_labels" {
  type    = list(string)
  default = ["terraform-created PAT-v2"]
}

# Create 200 teams with variations
resource "docker_org_team" "terraform_team" {
  count            = 200
  org_name         = "dockerterraform"
  team_name        = format("tfteam%03d", count.index + 1) # Ensures the team name is within limits
  team_description = format("Terraform Hackathon Demo - 2024 - Variation %03d", count.index + 1)
}

# Team associations with variations
resource "docker_org_team_member_association" "example_association" {
  count      = 200
  org_name   = "dockerterraform"
  team_name  = docker_org_team.terraform_team[count.index].team_name
  user_names = var.user_names
}

# Create 200 repositories with variations
resource "docker_hub_repository" "org_hub_repo" {
  count            = 200
  namespace        = "dockerterraform"
  name             = format("%s-%03d", element(var.repo_names, 0), count.index + 1)
  description      = format("This is a repo demo - Variation %03d", count.index + 1)
  full_description = "Lorem ipsum"
}

# Repository team permissions with variations
resource "docker_hub_repository_team_permission" "test" {
  count      = 200
  repo_id    = docker_hub_repository.org_hub_repo[count.index].id
  team_id    = docker_org_team.terraform_team[count.index].id
  permission = "admin"
}

# Create 200 access tokens with variations
resource "docker_access_token" "new_token_v2" {
  count       = 200
  token_label = format("%s-%03d", element(var.token_labels, 0), count.index + 1)
  scopes      = ["repo:read", "repo:write"]
}
