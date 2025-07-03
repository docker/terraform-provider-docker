terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~> 0.2"
    }
  }
}

# Initialize provider
provider "docker" {}

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

# === Repository Tags Data Source Example ===
# This demonstrates the key security use case: converting human-friendly tags to digest-pinned references

# Get Alpine tags for our deployment
data "docker_hub_repository_tags" "alpine_tags" {
  namespace = "library"
  name      = "alpine"
  page_size = 20
}

# === Security-Hardened Image Reference ===
locals {
  # Define the human-friendly tag we want to use
  desired_tag = "latest"

  # Create fully qualified image reference with digest
  # This converts: alpine:latest → alpine@sha256:abc123...
  secure_image_ref = try(
    "${data.docker_hub_repository_tags.alpine_tags.tags[local.desired_tag].digest != "" ?
      "alpine@${data.docker_hub_repository_tags.alpine_tags.tags[local.desired_tag].digest}" :
      "alpine:${local.desired_tag}"
    }",
    "alpine:${local.desired_tag}"
  )
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

# === Security-Hardened Image Reference Output ===
output "secure_image_example" {
  description = "Demonstration of converting human-friendly tag to digest-pinned reference"
  value = {
    human_friendly = "alpine:${local.desired_tag}"
    digest_pinned  = local.secure_image_ref
    tag_details    = data.docker_hub_repository_tags.alpine_tags.tags[local.desired_tag]
    available_arches = [
      for image in data.docker_hub_repository_tags.alpine_tags.tags[local.desired_tag].images :
      image.architecture
    ]
  }
}

