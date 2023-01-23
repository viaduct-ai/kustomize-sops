PLUGIN_NAME="ksops"

# Default to installing KSOPS
default: install

.PHONY: install
install: go-install clean build install-plugin

.PHONY: install-plugin
install-plugin:
	./scripts/install-ksops.sh

.PHONY: build
build:
	go build -o $(PLUGIN_NAME)

.PHONY: clean
clean:
	rm -f $(PLUGIN_NAME)
	rm -rf $(XDG_CONFIG_HOME)/kustomize/plugin/viaduct.ai/v1/ || true
	rm -rf $(HOME)/sigs.k8s.io/kustomize/plugin/viaduct.ai/v1/ || true
	rm -f $(shell command -v $(PLUGIN_NAME))

.PHONY: kustomize
kustomize:
	./scripts/install-kustomize.sh

.PHONY: sops
sops:
	go get -u go.mozilla.org/sops/cmd/sops

.PHONY: download-dependencies
download-dependencies:
	go mod download
	go mod tidy

.PHONY: setup
setup: .git/hooks/pre-push .git/hooks/pre-commit kustomize sops download-dependencies

.PHONY: import-test-keys
import-test-keys:
	gpg --import test/key.asc

.PHONY: test
test: install setup-test-files go-test

.PHONY: setup-test-files
setup-test-files:
	./scripts/setup-test-files.sh

.PHONY: go-install
go-install:
	go install

.PHONY: go-test
go-test:
	go test -v ./...

.PHONY: go-fmt
go-fmt:
	go fmt .

.PHONY: go-vet
go-vet:
	go vet -v ./...


# https://vincent.bernat.ch/en/blog/2019-makefile-build-golang
BIN = $(CURDIR)/bin
$(BIN):
		@mkdir -p $@
$(BIN)/%: | $(BIN)
		@tmp=$$(mktemp -d); \
       env GO111MODULE=off GOPATH=$$tmp GOBIN=$(BIN) go get $(PACKAGE) \
        || ret=$$?; \
       rm -rf $$tmp ; exit $$ret

$(BIN)/golint: PACKAGE=golang.org/x/lint/golint

GOLINT = $(BIN)/golint
lint: | $(GOLINT)
		$(GOLINT) -set_exit_status ./...


################################################################################
# Git Hooks
################################################################################
## Git hooks to validate worktree is clean before commit/push
.git/hooks/pre-push: Makefile
	# Create Git pre-push hook
	echo 'make pre-push' > .git/hooks/pre-push
	chmod +x .git/hooks/pre-push

.git/hooks/pre-commit: Makefile
	# Create Git pre-commit hook
	echo 'make pre-commit' > .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

.PHONY: pre-commit
pre-commit: download-dependencies lint go-fmt go-vet

.PHONY: pre-push
pre-push: test
