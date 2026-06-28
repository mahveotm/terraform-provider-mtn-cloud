BINARY=terraform-provider-mtncloud

.PHONY: build test testacc fmt vet lint tidy check install-local docs docs-check

build:
	go build -o $(BINARY) .

# Unit tests (fast, no live API).
test:
	go test -race ./...

# Acceptance tests (provisions real infrastructure; needs MTN_CLOUD_TOKEN).
testacc:
	TF_ACC=1 go test -race -count=1 -timeout 120m ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# Static analysis. Install once with:
#   brew install golangci-lint   (or see https://golangci-lint.run/welcome/install/)
lint:
	golangci-lint run ./...

tidy:
	go mod tidy

# Run everything CI runs.
check: build vet lint test
	@gofmt -l . | (! grep .) || (echo "Run 'make fmt'"; exit 1)

install-local: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/mahveotm/mtncloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/mahveotm/mtncloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

# Regenerate docs/ from schema descriptions + examples/ via tfplugindocs.
docs:
	go generate ./...

# Verify docs/ is current with the schema (use in CI so stale docs fail the build).
docs-check: docs
	@git diff --exit-code -- docs/ || (echo "docs/ is stale — run 'make docs' and commit"; exit 1)
