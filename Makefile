ACCTEST_COUNT                ?= 1
ACCTEST_PARALLELISM          ?= 20
ACCTEST_TIMEOUT              ?= 120m

ifneq ($(origin TESTS), undefined)
	RUNARGS = -run='$(TESTS)'
endif

# Default target
default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -timeout $(ACCTEST_TIMEOUT) $(RUNARGS) $(TESTARGS)

# Install the provider binary to GOBIN
.PHONY: install
install:
	go install .

# Generate documentation
.PHONY: docs
docs:
	go generate ./...

# Setup local development environment
.PHONY: local-build
local-build: install setup-terraformrc

# Setup or update ~/.terraformrc with GOBIN path for docker/docker provider
.PHONY: setup-terraformrc
setup-terraformrc:
	@echo "Setting up ~/.terraformrc for local development..."
	@GOBIN_PATH=$$(go env GOBIN); \
	if [ -z "$$GOBIN_PATH" ]; then \
		echo "GOBIN is not set. Defaulting to GOPATH/bin"; \
		GOBIN_PATH=$$(go env GOPATH)/bin; \
	fi; \
	echo "Using GOBIN_PATH=$$GOBIN_PATH"; \
	if [ ! -f ~/.terraformrc ]; then \
		echo 'provider_installation {' > ~/.terraformrc; \
		echo '  dev_overrides {' >> ~/.terraformrc; \
		echo "    \"registry.terraform.io/docker/docker\" = \"$$GOBIN_PATH\"" >> ~/.terraformrc; \
		echo '  }' >> ~/.terraformrc; \
		echo '  direct {}' >> ~/.terraformrc; \
		echo '}' >> ~/.terraformrc; \
		echo "~/.terraformrc has been created and set up for local development."; \
	else \
		if grep -q 'registry.terraform.io/docker/docker' ~/.terraformrc; then \
			echo "The override for registry.terraform.io/docker/docker already exists in ~/.terraformrc."; \
		else \
			awk '/dev_overrides/ {print; print "    \"registry.terraform.io/docker/docker\" = \"'"$$GOBIN_PATH"'\""; next}1' ~/.terraformrc > ~/.terraformrc.tmp && mv ~/.terraformrc.tmp ~/.terraformrc; \
			echo "Added the override for registry.terraform.io/docker/docker to ~/.terraformrc."; \
		fi; \
	fi
