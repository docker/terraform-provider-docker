---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "docker_hub_repositories Data Source - docker"
subcategory: ""
description: |-
  Retrieves a list of repositories within a specified Docker Hub namespace. All available repositories are automatically fetched using internal pagination (similar to AWS provider pattern).
  Example Usage
  
  data "docker_hub_repositories" "example" {
  	namespace = "my-organization"
  }
  output "repositories" {
  	value = data.docker_hub_repositories.example.repository
  }
---

# docker_hub_repositories (Data Source)

Retrieves a list of repositories within a specified Docker Hub namespace. All available repositories are automatically fetched using internal pagination (similar to AWS provider pattern).

## Example Usage

```hcl
data "docker_hub_repositories" "example" {
	namespace = "my-organization"
}
output "repositories" {
	value = data.docker_hub_repositories.example.repository
}

```



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `namespace` (String) Repository namespace

### Read-Only

- `id` (String) The namespace/name of the repository
- `repository` (Attributes List) List of repositories (see [below for nested schema](#nestedatt--repository))

<a id="nestedatt--repository"></a>
### Nested Schema for `repository`

Read-Only:

- `affiliation` (String)
- `description` (String)
- `is_private` (Boolean)
- `last_updated` (String)
- `name` (String)
- `namespace` (String)
- `pull_count` (Number)
