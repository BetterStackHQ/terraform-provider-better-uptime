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
	@echo "  # Install terraform-provider-betteruptime locally."
	@echo "  #"
	@echo "  # terraform {"
	@echo "  #   required_providers {"
	@echo "  #     custom = {"
	@echo "  #       source = \"registry.terraform.io/altinity/betteruptime\""
	@echo "  #       version = \"0.0.0-0\""
	@echo "  #     }"
	@echo "  #   }"
	@echo "  # }"
	@echo "  make install"
	@echo
	@echo "  # Upload terraform-provider-betteruptime to GitHub."
	@echo "  make VERSION=0.0.0-0 release"
	@echo

clean:
	rm -f cover.out coverage.html terraform-provider-betteruptime
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
	go build -gcflags "all=-N -l" -ldflags "-X main.version=0.0.0-0"

install: build
	PLUGIN_DIR="$$HOME/.terraform.d/plugins/registry.terraform.io/altinity/betteruptime/0.0.0-0/$$(go env GOOS)_$$(go env GOARCH)" && \
		mkdir -p "$$PLUGIN_DIR" && \
		cp terraform-provider-betteruptime "$$PLUGIN_DIR/"

debug: build
# https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./terraform-provider-betteruptime -- --debug

go-get-gox:
	@test -n "$$(which gox)" || (GO111MODULE=off go get github.com/mitchellh/gox)

release-build: go-get-gox
	test -n "$(VERSION)" # $$VERSION must be set
	env CGO_ENABLED=0 gox -verbose \
		-ldflags "-X main.version=${VERSION}" \
		-osarch="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64" \
		-output="release/${VERSION}/{{.OS}}_{{.Arch}}/terraform-provider-betteruptime_v${VERSION}" \
		.
	( \
		cd release/${VERSION} && \
		for q in $$(ls -d */ | cut -d/ -f1); do (cd $$q; zip "../terraform-provider-betteruptime_${VERSION}_$$q.zip" *); done && \
		shasum -a 256 *.zip > terraform-provider-betteruptime_${VERSION}_SHA256SUMS \
	)

release-sign:
	test -n "$(VERSION)" # $$VERSION must be set
	( \
		cd release/${VERSION} && \
		gpg --detach-sign terraform-provider-betteruptime_${VERSION}_SHA256SUMS \
	)

go-get-github-release:
	@test -n "$$(which github-release)" || (GO111MODULE=off go get github.com/aktau/github-release)

release: go-get-github-release clean gen release-build release-sign
	test -n "$(GITHUB_TOKEN)" # $$GITHUB_TOKEN must be set
	git tag -a v${VERSION} -m v${VERSION} && \
	git push origin v${VERSION} && \
	github-release release --user altinity --repo terraform-provider-betteruptime --tag "v${VERSION}" \
		--name "v${VERSION}" --description "[CHANGELOG](https://github.com/altinity/terraform-provider-betteruptime/blob/master/CHANGELOG.md)" && \
	\
	github-release upload --user altinity --repo terraform-provider-betteruptime --tag "v${VERSION}" \
		--name "terraform-provider-betteruptime_${VERSION}_SHA256SUMS" --file "release/${VERSION}/terraform-provider-betteruptime_${VERSION}_SHA256SUMS" && \
	github-release upload --user altinity --repo terraform-provider-betteruptime --tag "v${VERSION}" \
		--name "terraform-provider-betteruptime_${VERSION}_SHA256SUMS.sig" --file "release/${VERSION}/terraform-provider-betteruptime_${VERSION}_SHA256SUMS.sig" && \
	\
	for qualifier in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64; do \
		github-release upload --user altinity --repo terraform-provider-betteruptime --tag "v${VERSION}" \
			--name "terraform-provider-betteruptime_${VERSION}_$$qualifier.zip" --file "release/${VERSION}/terraform-provider-betteruptime_${VERSION}_$$qualifier.zip"; \
	done
