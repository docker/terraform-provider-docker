# Contributing to Docker Terraform Provider

Thank you for considering contributing to the Docker Terraform Provider! We welcome contributions from everyone. To ensure a smooth process, please follow the guidelines below.

## Reporting Issues and Suggesting Enhancements

1. Before creating a new issue, please check if the issue has already been reported in the list of [existing issues](https://github.com/docker/terraform-provider-docker/issues)
2. If you don't find an existing issue, open a new one and fill out the template based on the issue type - it's important to provide as much detail as possible

## Making Code Contributions

1. Fork the repository
2. Check with the authors in an issue ticket before doing anything big
3. Contribute improvements or fixes using a [Pull Request](https://github.com/docker/terraform-provider-docker/pulls)
4. Provide a clear description of what your PR does, why is it important, and list the related issues
5. Make small commits that are easy to merge
6. Make sure to test your changes
7. When updating documentation, please see our guidance for documentation contributions (TODO)

## Local Development Setup

### Prerequisites

- `make`, `git`, `bash`
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.1
- [Go](https://golang.org/doc/install) >= 1.21
  - Ensure your [GOPATH](http://golang.org/doc/code.html#GOPATH) is correctly setup, as well as adding `$GOPATH/bin` to your `$PATH`

1. Clone `terraform-provider-docker`

```shell
git clone https://github.com/docker/terraform-provider-docker
```

2. Install provider & setup `~/.terrraformrc` making `registry.terraform.io/docker/docker` reference your local installation

```shell
make local-build
```

3. Setup your Docker Hub username & password

```shell
export DOCKER_USERNAME=$(yourusername)
export DOCKER_PASSWORD=$(yourpassword)
```

4. Run an example build!

```shell
cd examples && terraform plan
```

Note, when you run the `terraform plan` you should see a warning, this ensures that you using the locally installed provider and not the publically available provider.

> │ Warning: Provider development overrides are in effect

Happy developing!

## Testing

Run full test suite:

```shell
make testacc
```

Run an specific test(s):

```shell
make testacc TESTS=TestAccXXX
```

which is equivalent to:

```shell
TF_ACC=1 go test ./... -v -count 1 -parallel 20 -timeout 120m -run TestAccXXX
```

where `TestFuncName` is the testing function within the `_test.go` file.

## Environment Variables

```shell
# enable debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH="/PATH/TO/YOUR/LOG_FILE.log"\
# acceptance testing overrides
export ACCTEST_DOCKER_ORG=myorgname
...
```

## Getting Help

If you have any questions or need assistance, please reach out to us through in the DTP Slack Channel (TODO)

## Thanks for Contributing

Your contributions help make the Docker Terraform Provider better for everyone. We appreciate your support and effort!
