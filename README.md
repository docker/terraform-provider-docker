# Docker Hub Terraform Provider

This project is used to manage Docker resources (such as repositories, teams, organization settings, and more) using Terraform. It allows users to define Docker infrastructure as code, integrating Docker services into their Terraform workflows. The Terraform Registry page for this provider can be found [here](https://registry.terraform.io/providers/docker/docker/).

> [!WARNING]
> This project is **not** for managing objects in a local docker engine. If you would like to use Terraform to interact with a docker engine, [kreuzwerker/docker](https://registry.terraform.io/providers/kreuzwerker/docker/latest) is a fine provider.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.1
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Usage

Below is a basic example of how to use the Docker services Terraform provider to create a Docker repository. 

```hcl
terraform {
  required_providers {
    docker = {
      source  = "docker/docker"
      version = "~> 1.0"
    }
  }
}

provider "docker" { }

resource "docker_repository" "example" {
  name        = "example-repo"
  description = "This is an example Docker repository"
  private     = true
}
```

## Authentication

We have multiple ways to set your Docker credentials.

### Setting credentials

Use `docker login` to [log in to a
registry](https://docs.docker.com/reference/cli/docker/login/). The `docker` CLI
will store your credentials securely in your credential store, such as the
operating system native keychain. The Docker Terraform provider will
use these credentials automatically.

```
cat ~/my_password.txt | docker login --username my-username --password-stdin
```

If you'd like to use a different account for running the provider,
you can set credentials in the environment:

```
export DOCKER_USERNAME=my-username
export DOCKER_PASSWORD=my-secret-token
terraform plan ...
```

### Credential types

You can create a personal access token (PAT) to use as an alternative to your
password for Docker CLI authentication.

A "Read, Write, & Delete" PAT can be used to create, edit, and 
manage permissions for Docker Hub repositories.

The advantage of PATs is that they have [many security
benefits](https://docs.docker.com/security/for-developers/access-tokens/) over
passwords.

Unfortunately, PATs are limited to managing repositories. If you'd like to use
this provider to manage organizations and teams, you will need to authenticate
with a password.


## Contributing

We welcome contributions to the Docker services Terraform provider, detailed documentation for contributing & building the provider can be found [here](https://github.com/docker/terraform-provider-docker/blob/main/CONTRIBUTING.md)

## Roadmap

Our roadmap is managed through GitHub issues. You can view upcoming features and enhancements, as well as report bugs or request new features, by visiting our [issues page](https://github.com/docker/terraform-provider-docker/issues?q=sort%3Aupdated-desc+is%3Aissue+is%3Aopen).

## Support

TODO: how much will we be supporting this & at what cadence?

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](https://github.com/docker/terraform-provider-docker/blob/main/LICENSE) file for more information.
