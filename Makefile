SHELL := /bin/bash
GOLANGCI_LINT := golangci-lint run --disable-all \
	-E deadcode \
	-E errcheck \
	-E goimports \
	-E gosimple \
	-E govet \
	-E ineffassign \
	-E maligned \
	-E staticcheck \
	-E structcheck \
	-E typecheck \
	-E unused \
	-E varcheck
.PHONY: test build

help:
	@echo Usage:
	@echo
	@echo "  make clean"
	@echo
	@echo "  # Regenerate docs/."
	@echo "  make gen"
	@echo
	@echo "  make lint"
	@echo "  make fmt"
	@echo
	@echo "  make test"
	@echo "  make test-coverage"
	@echo
	@echo "  make terraform CONFIGURATION=examples/basic ARGS=apply"
	@echo
	@echo "  # Run in \"Debug\" mode (connect debugger to port 2345)."
	@echo "  make debug"
	@echo
	@echo "  # Install terraform-provider-better-uptime locally."
	@echo "  #"
	@echo "  # terraform {"
	@echo "  #   required_providers {"
	@echo "  #     custom = {"
	@echo "  #       source = \"registry.terraform.io/BetterStackHQ/better-uptime\""
	@echo "  #       version = \"0.0.0-0\""
	@echo "  #     }"
	@echo "  #   }"
	@echo "  # }"
	@echo "  make install"
	@echo
	@echo "  # Upload terraform-provider-better-uptime to GitHub."
	@echo "  make VERSION=0.0.0-0 release"
	@echo

clean:
	rm -f cover.out coverage.html terraform-provider-better-uptime
	rm -rf release/

lint-init:
	@test -n "$$(which golangci-lint)" || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.27.0)

lint: lint-init
	$(GOLANGCI_LINT)
	terraform fmt -check -diff -recursive

fmt: lint-init
	$(GOLANGCI_LINT) --fix
	terraform fmt -recursive

gen:
	terraform fmt -check -diff -recursive
	go generate ./...
	@echo
	@echo "docs/ can be previewed at https://registry.terraform.io/tools/doc-preview"

test:
	go test ./...

test-race:
	go test -race ./...

test-coverage:
	go test -coverprofile cover.out ./...
	go tool cover -html=cover.out -o coverage.html
	rm -f cover.out
	@echo
	@echo "open coverage.html to review the report"

# https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

terraform: install
	cd $(CONFIGURATION) && rm -f .terraform.lock.hcl && terraform init && \
 		TF_LOG=DEBUG TF_PROVIDER_BETTERUPTIME_LOG_INSECURE=1 terraform $(ARGS)

build:
# -gcflags "all=-N -l" is here for delve (`go tool compile -help` for more)
	go build -gcflags "all=-N -l" -ldflags "-X main.version=0.3.8"

install: build
	PLUGIN_DIR="$$HOME/.terraform.d/plugins/registry.terraform.io/BetterStackHQ/better-uptime/0.3.8/$$(go env GOOS)_$$(go env GOARCH)" && \
		mkdir -p "$$PLUGIN_DIR" && \
		cp terraform-provider-better-uptime "$$PLUGIN_DIR/"

uninstall:
	rm -rf "$$HOME/.terraform.d/plugins/registry.terraform.io/BetterStackHQ/better-uptime/0.3.8"

debug: build
# https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./terraform-provider-better-uptime -- --debug
