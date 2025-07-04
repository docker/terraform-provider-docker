terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~> 0.2"
    }
  }
}

# Initialize provider
provider "docker" {
  max_page_results = 5
}

# Define local variables for customization
locals {
  namespace     = "my-docker-namespace"
  repo_name     = "my-docker-repo"
  org_name      = "my-docker-org"
  team_name     = "my-team"
  my_team_users = ["user1", "user2"]
  token_label   = "my-pat-token"
  token_scopes  = ["repo:read", "repo:write"]
  permission    = "admin"
}

# Create repository
resource "docker_hub_repository" "org_hub_repo" {
  namespace        = local.namespace
  name             = local.repo_name
  description      = "This is a generic Docker repository."
  full_description = "Full description for the repository."
}

# Create team
resource "docker_org_team" "team" {
  org_name         = local.org_name
  team_name        = local.team_name
  team_description = "Team description goes here."
}

# Team association
resource "docker_org_team_member" "team_membership" {
  for_each = toset(local.my_team_users)

  org_name  = local.org_name
  team_name = docker_org_team.team.team_name
  user_name = each.value
}

# Create repository team permission
resource "docker_hub_repository_team_permission" "repo_permission" {
  repo_id    = docker_hub_repository.org_hub_repo.id
  team_id    = docker_org_team.team.id
  permission = local.permission
}

# Create access token
resource "docker_access_token" "access_token" {
  token_label = local.token_label
  scopes      = local.token_scopes
}

# Access tokens data source
data "docker_access_tokens" "access_tokens" {}

# === Repository Tags Data Source Example ===
# This demonstrates the key security use case: converting human-friendly tags to digest-pinned references
# Get Alpine tags for our deployment
data "docker_hub_repository_tags" "example" {
  namespace = "library"
  name      = "alpine"
}

# Get digest-pinned reference
locals {
  secure_image = "${data.docker_hub_repository_tags.example.name}@${data.docker_hub_repository_tags.example.tags["latest"].digest}"
}

# === Outputs ===

# Repository outputs
output "repo_output" {
  description = "Created repository information"
  value = {
    id          = docker_hub_repository.org_hub_repo.id
    name        = docker_hub_repository.org_hub_repo.name
    namespace   = docker_hub_repository.org_hub_repo.namespace
    description = docker_hub_repository.org_hub_repo.description
  }
}

# Team outputs
output "org_team_output" {
  description = "Created team information"
  value = {
    id   = docker_org_team.team.id
    name = docker_org_team.team.team_name
  }
}

# Access token output
output "access_token_uuid" {
  description = "Created access token UUID"
  value       = docker_access_token.access_token.uuid
  sensitive   = true
}
output "access_tokens" {
  value = data.docker_access_tokens.access_tokens
}

# Security-Hardened Image Reference Output
output "secure_image_example" {
  description = "Demonstration of converting human-friendly tag to digest-pinned reference"
  value = {
    human_friendly = "${data.docker_hub_repository_tags.example.name}:latest"
    digest_pinned  = local.secure_image
    tag_details    = data.docker_hub_repository_tags.example.tags["latest"]
    available_arches = [
      for image in data.docker_hub_repository_tags.example.tags["latest"].images :
      image.architecture
    ]
  }
}

