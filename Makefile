PLUGIN_NAME := ksops

GOPATH ?= $(shell go env GOPATH)
STATICCHECK := $(GOPATH)/bin/staticcheck
KUSTOMIZE := $(GOPATH)/bin/kustomize
SOPS := $(GOPATH)/bin/sops
GO_MOD_OUTDATED := $(GOPATH)/bin/go-mod-outdated
PREREQS := $(STATICCHECK) $(KUSTOMIZE) $(GO_MOD_OUTDATED)

XDG_CONFIG_HOME ?= $(HOME)/.config
XDG_KSOPS_PATH ?= $(XDG_CONFIG_HOME)/kustomize/plugin/viaduct.ai/v1/$(PLUGIN_NAME)
XDG_KSOPS_LEGACY_PATH ?= $(XDG_CONFIG_HOME)/kustomize/plugin/viaduct.ai/v1/$(PLUGIN_NAME)-exec

GIT_HOOKS := pre-push pre-commit

GO_LD_FLAGS := "-w -s"
GO_BUILD_FLAGS := -trimpath -ldflags $(GO_LD_FLAGS)

IMAGE ?= viaductai/ksops
RELEASE ?=
BUILDX_ARGS ?= $(if $(RELEASE),--platform $(PLATFORMS) --push,--load)
GO_VERSION := $(shell cat go.mod | grep -m1 'go' | awk '{print $$2}')
GIT_VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n	make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "	\033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: FORCE
FORCE: ;

##@ Prerequisites:

.PHONY: deps
deps: ## Download go modules
	@echo "Downloading go modules…"
	@go mod download

$(STATICCHECK): FORCE
	@echo "Installing staticcheck…"
	@go install honnef.co/go/tools/cmd/staticcheck@latest

$(GO_MOD_OUTDATED): FORCE
	@echo "Installing go-mod-outdated…"
	@go install github.com/psampaz/go-mod-outdated@latest

$(SOPS): FORCE
	@echo "Installing sops…"
	@go install go.mozilla.org/sops/cmd/sops@latest

.PHONY: prereqs
prereqs: $(PREREQS) ## Install prerequisites

$(KUSTOMIZE): FORCE
	./scripts/install-kustomize.sh

##@ Development:

.PHONY: setup
setup: git-hooks $(KUSTOMIZE) $(SOPS) deps ## Setup the development environment

.PHONY: git-hooks
git-hooks: $(addprefix .git/hooks/,$(GIT_HOOKS)) ## Install Git hooks

.PHONY: tidy
tidy: ## Run go mod tidy
	@echo "Running go mod tidy…"
	@go mod tidy

$(PLUGIN_NAME): FORCE
	@echo "Building $(PLUGIN_NAME)…"
	@go build $(GO_BUILD_FLAGS) -o $@

.PHONY: build
build: $(PLUGIN_NAME) ## Compile the plugin

.PHONY: test
test: import-test-keys ## Run go tests
	@echo "Running go test…"
	@go test -v -race -count=1 ./...

.PHONY: import-test-keys
import-test-keys: ## Import GPG test keys
	@echo "Importing GPG test keys…"
	gpg --import test/key.asc

.PHONY: fmt
fmt: ## Run go fmt
	@echo "Running go fmt…"
	@go fmt .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet…"
	@go vet -v ./...

.PHONY: lint
lint: $(STATICCHECK) ## Run staticcheck linter
	@echo "Running staticcheck…"
	@$<

.PHONY: outdated
outdated: $(GO_MOD_OUTDATED) ## List outdated dependencies.
	@go list -u -m -json all | $< -update

.PHONY: docker-build
docker-build: ## Build the Docker image.
	@echo "Building Docker image $(IMAGE):$(GIT_VERSION)…"
	@docker buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		$(BUILDX_ARGS) \
		-t $(IMAGE):$(GIT_VERSION) .

##@ Install:

.PHONY: install
install: clean build install-plugin

$(XDG_KSOPS_PATH):
	@mkdir -p $@

$(XDG_KSOPS_LEGACY_PATH):
	@mkdir -p $@

$(XDG_KSOPS_PATH)/$(PLUGIN_NAME): $(PLUGIN_NAME) | $(XDG_KSOPS_PATH)
	@echo "Installing $< to $@…"
	@install -m 755 $< $@

$(XDG_KSOPS_LEGACY_PATH)/$(PLUGIN_NAME)-exec: $(PLUGIN_NAME) | $(XDG_KSOPS_LEGACY_PATH)
	@echo "Installing $< to $@…"
	@install -m 755 $< $@

$(GOPATH)/bin/$(PLUGIN_NAME): $(PLUGIN_NAME)
	@echo "Installing $< to $@…"
	@install -m 755 $< $@

.PHONY: install-plugin
install-plugin: $(GOPATH)/bin/$(PLUGIN_NAME) $(XDG_KSOPS_PATH)/$(PLUGIN_NAME) $(XDG_KSOPS_LEGACY_PATH)/$(PLUGIN_NAME)-exec ## Install the ksops plugin

.PHONY: clean
clean:
	@echo "Cleaning up…"
	@rm -f $(PLUGIN_NAME)
	@rm -f $(GOPATH)/bin/$(PLUGIN_NAME)
	@rm -rf $(XDG_KSOPS_PATH) || true
	@rm -rf $(XDG_KSOPS_LEGACY) || true
	@rm -f $(shell command -v $(PLUGIN_NAME))

################################################################################
# Git Hooks
################################################################################

.git/hooks/pre-push:
	@echo "Installing pre-push hook…"
	@echo 'make pre-push' > .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push

.git/hooks/pre-commit:
	@echo "Installing pre-commit hook…"
	@echo 'make pre-commit' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit

.PHONY: pre-commit
pre-commit: deps lint fmt vet

.PHONY: pre-push
pre-push: test
