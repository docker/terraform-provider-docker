---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "docker_org_members Data Source - docker"
subcategory: ""
description: |-
  Reads members of an organization.
  -> Note Only available when authenticated with a username and password.
  Example Usage
  
  data "docker_org_members" "_" {
    org_name  = "my-org"
  }
  
  output "usernames" {
    value = [for member in data.docker_org_members._.members: member.username]
  }
---

# docker_org_members (Data Source)

Reads members of an organization.

-> **Note** Only available when authenticated with a username and password.

## Example Usage

```hcl
data "docker_org_members" "_" {
  org_name  = "my-org"
}

output "usernames" {
  value = [for member in data.docker_org_members._.members: member.username]
}
```



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `org_name` (String) Organization name

### Read-Only

- `members` (Attributes List) List of members (see [below for nested schema](#nestedatt--members))

<a id="nestedatt--members"></a>
### Nested Schema for `members`

Read-Only:

- `company` (String)
- `date_joined` (String)
- `email` (String)
- `full_name` (String)
- `gravatar_email` (String)
- `gravatar_url` (String)
- `groups` (List of String)
- `id` (String)
- `is_guest` (Boolean)
- `location` (String)
- `profile_url` (String)
- `role` (String)
- `type` (String)
- `username` (String)
