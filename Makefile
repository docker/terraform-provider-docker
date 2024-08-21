default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: install
install:
	go install .

.PHONY: docs
docs:
	go generate ./...

